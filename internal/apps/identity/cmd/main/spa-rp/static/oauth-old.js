// OAuth 2.1 + OIDC Client with PKCE Support
// Implements RFC 6749, RFC 7636 (PKCE), and OpenID Connect Core 1.0

// ==================== Diagnostic Logging ====================

const DEBUG = true; // Enable diagnostic logging.

/**
 * Log diagnostic message to browser console
 * @param {string} level - Log level (info, warn, error, debug)
 * @param {string} operation - Operation name
 * @param {Object} data - Additional context data
 */
function logDiagnostic(level, operation, data = {}) {
    if (!DEBUG && level === 'debug') return;

    const timestamp = new Date().toISOString();
    const logEntry = {
        timestamp,
        level,
        operation,
        ...data
    };

    const logMethod = console[level] || console.log;
    logMethod(`[OAuth-${level.toUpperCase()}] ${operation}`, logEntry);
}

// ==================== PKCE Utilities ====================

/**
 * Generate a cryptographically random string for PKCE code verifier
 * @returns {string} Base64URL encoded random string
 */
function generateCodeVerifier() {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return base64URLEncode(array);
}

/**
 * Generate code challenge from verifier using SHA-256
 * @param {string} verifier - The code verifier
 * @returns {Promise<string>} Base64URL encoded SHA-256 hash
 */
async function generateCodeChallenge(verifier) {
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);
    const hash = await crypto.subtle.digest('SHA-256', data);
    return base64URLEncode(new Uint8Array(hash));
}

/**
 * Base64URL encoding without padding
 * @param {Uint8Array} buffer - Buffer to encode
 * @returns {string} Base64URL encoded string
 */
function base64URLEncode(buffer) {
    const base64 = btoa(String.fromCharCode.apply(null, buffer));
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

/**
 * Generate a random state parameter for CSRF protection
 * @returns {string} Random state string
 */
function generateState() {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return base64URLEncode(array);
}

// ==================== Token Storage ====================

const TOKEN_STORAGE_KEY = 'oauth_tokens';
const PKCE_STORAGE_KEY = 'pkce_data';
const STATE_STORAGE_KEY = 'oauth_state';

/**
 * Store tokens securely in sessionStorage
 * @param {Object} tokens - Token response from server
 */
function storeTokens(tokens) {
    sessionStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokens));
    updateUI();
}

/**
 * Retrieve stored tokens
 * @returns {Object|null} Tokens or null if not found
 */
function getStoredTokens() {
    const tokens = sessionStorage.getItem(TOKEN_STORAGE_KEY);
    return tokens ? JSON.parse(tokens) : null;
}

/**
 * Clear all stored OAuth data
 */
function clearStorage() {
    sessionStorage.removeItem(TOKEN_STORAGE_KEY);
    sessionStorage.removeItem(PKCE_STORAGE_KEY);
    sessionStorage.removeItem(STATE_STORAGE_KEY);
}

// ==================== OAuth 2.1 Flow ====================

/**
 * Start OAuth 2.1 authorization code flow with PKCE
 */
async function startLogin() {
    try {
        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;
        const redirectUri = document.getElementById('redirectUri').value;
        const scope = document.getElementById('scope').value;

        // Generate PKCE parameters
        const codeVerifier = generateCodeVerifier();
        const codeChallenge = await generateCodeChallenge(codeVerifier);
        const state = generateState();

        // Store PKCE data and state for callback
        sessionStorage.setItem(PKCE_STORAGE_KEY, JSON.stringify({
            verifier: codeVerifier,
            challenge: codeChallenge
        }));
        sessionStorage.setItem(STATE_STORAGE_KEY, state);

        // Build authorization URL
        const params = new URLSearchParams({
            response_type: 'code',
            client_id: clientId,
            redirect_uri: redirectUri,
            scope: scope,
            state: state,
            code_challenge: codeChallenge,
            code_challenge_method: 'S256'
        });

        const authUrl = `${authzUrl}/oauth2/v1/authorize?${params.toString()}`;

        // Redirect to authorization server
        window.location.href = authUrl;
    } catch (error) {
        showError('Failed to start login: ' + error.message);
    }
}

/**
 * Handle OAuth callback and exchange code for tokens
 */
async function handleCallback() {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');
    const error = params.get('error');

    // Check for errors
    if (error) {
        showError('Authorization failed: ' + (params.get('error_description') || error));
        clearStorage();
        return;
    }

    // Validate state parameter (CSRF protection)
    const storedState = sessionStorage.getItem(STATE_STORAGE_KEY);
    if (!state || state !== storedState) {
        showError('Invalid state parameter - possible CSRF attack');
        clearStorage();
        return;
    }

    // Exchange authorization code for tokens
    if (code) {
        await exchangeCodeForTokens(code);
    }

    // Clean up URL
    window.history.replaceState({}, document.title, window.location.pathname);
}

/**
 * Exchange authorization code for tokens
 * @param {string} code - Authorization code
 */
async function exchangeCodeForTokens(code) {
    try {
        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;
        const redirectUri = document.getElementById('redirectUri').value;

        // Retrieve PKCE verifier
        const pkceData = JSON.parse(sessionStorage.getItem(PKCE_STORAGE_KEY) || '{}');
        if (!pkceData.verifier) {
            throw new Error('PKCE verifier not found');
        }

        // Build token request
        const body = new URLSearchParams({
            grant_type: 'authorization_code',
            code: code,
            redirect_uri: redirectUri,
            client_id: clientId,
            code_verifier: pkceData.verifier
        });

        // Make token request
        const response = await fetch(`${authzUrl}/oauth2/v1/token`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: body.toString()
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error_description || 'Token exchange failed');
        }

        const tokens = await response.json();

        // Store tokens and update UI
        storeTokens(tokens);
        showSuccess('Successfully authenticated!');

        // Clean up PKCE data
        sessionStorage.removeItem(PKCE_STORAGE_KEY);
        sessionStorage.removeItem(STATE_STORAGE_KEY);
    } catch (error) {
        showError('Token exchange failed: ' + error.message);
        clearStorage();
    }
}

/**
 * Refresh access token using refresh token
 */
async function refreshToken() {
    try {
        const tokens = getStoredTokens();
        if (!tokens || !tokens.refresh_token) {
            throw new Error('No refresh token available');
        }

        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;

        const body = new URLSearchParams({
            grant_type: 'refresh_token',
            refresh_token: tokens.refresh_token,
            client_id: clientId
        });

        const response = await fetch(`${authzUrl}/oauth2/v1/token`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: body.toString()
        });

        if (!response.ok) {
            throw new Error('Token refresh failed');
        }

        const newTokens = await response.json();
        storeTokens(newTokens);
        showSuccess('Token refreshed successfully!');
    } catch (error) {
        showError('Token refresh failed: ' + error.message);
        clearStorage();
        updateUI();
    }
}

/**
 * Fetch user info from OIDC UserInfo endpoint
 */
async function getUserInfo() {
    try {
        const tokens = getStoredTokens();
        if (!tokens || !tokens.access_token) {
            throw new Error('No access token available');
        }

        const idpUrl = document.getElementById('idpUrl').value;

        const response = await fetch(`${idpUrl}/oidc/v1/userinfo`, {
            headers: {
                'Authorization': `Bearer ${tokens.access_token}`
            }
        });

        if (!response.ok) {
            throw new Error('UserInfo request failed');
        }

        const userInfo = await response.json();

        // Display user info
        document.getElementById('userInfoData').textContent = JSON.stringify(userInfo, null, 2);
        document.getElementById('userInfoSection').style.display = 'block';
        showSuccess('User info retrieved successfully!');
    } catch (error) {
        showError('Failed to get user info: ' + error.message);
    }
}

/**
 * Introspect access token
 */
async function introspectToken() {
    try {
        const tokens = getStoredTokens();
        if (!tokens || !tokens.access_token) {
            throw new Error('No access token available');
        }

        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;

        const body = new URLSearchParams({
            token: tokens.access_token,
            client_id: clientId
        });

        const response = await fetch(`${authzUrl}/oauth2/v1/introspect`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: body.toString()
        });

        if (!response.ok) {
            throw new Error('Token introspection failed');
        }

        const introspection = await response.json();
        alert('Token Introspection:\n\n' + JSON.stringify(introspection, null, 2));
    } catch (error) {
        showError('Failed to introspect token: ' + error.message);
    }
}

/**
 * Logout and clear all stored data
 */
function logout() {
    clearStorage();
    updateUI();
    showSuccess('Logged out successfully');

    // Hide user info and token sections
    document.getElementById('userInfoSection').style.display = 'none';
    document.getElementById('tokenSection').style.display = 'none';
}

// ==================== UI Updates ====================

/**
 * Update UI based on authentication state
 */
function updateUI() {
    const tokens = getStoredTokens();
    const isAuthenticated = !!(tokens && tokens.access_token);

    // Update buttons
    document.getElementById('loginBtn').disabled = isAuthenticated;
    document.getElementById('logoutBtn').disabled = !isAuthenticated;
    document.getElementById('refreshBtn').disabled = !isAuthenticated || !tokens.refresh_token;
    document.getElementById('userInfoBtn').disabled = !isAuthenticated;
    document.getElementById('introspectBtn').disabled = !isAuthenticated;

    // Update status
    const statusEl = document.getElementById('status');
    if (isAuthenticated) {
        statusEl.textContent = '✓ Authenticated';
        statusEl.className = 'status logged-in';

        // Show tokens
        document.getElementById('accessToken').value = tokens.access_token || '';
        document.getElementById('idToken').value = tokens.id_token || '';
        document.getElementById('refreshToken').value = tokens.refresh_token || '';
        document.getElementById('tokenSection').style.display = 'block';
    } else {
        statusEl.textContent = 'Not authenticated';
        statusEl.className = 'status logged-out';
        document.getElementById('tokenSection').style.display = 'none';
    }
}

/**
 * Show success message
 * @param {string} message - Success message
 */
function showSuccess(message) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = '✓ ' + message;
    statusEl.className = 'status logged-in';
}

/**
 * Show error message
 * @param {string} message - Error message
 */
function showError(message) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = '✗ ' + message;
    statusEl.className = 'status error';
}

// ==================== Initialization ====================

// Handle OAuth callback on page load
window.addEventListener('DOMContentLoaded', () => {
    handleCallback();
    updateUI();
});
