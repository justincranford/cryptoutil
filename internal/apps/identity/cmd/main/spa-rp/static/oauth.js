// OAuth 2.1 + OIDC Client with PKCE Support (Enhanced with Diagnostics)
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

// ==================== Loading State Management ====================

/**
 * Show loading indicator
 * @param {string} message - Loading message
 */
function showLoading(message) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = '⏳ ' + message;
    statusEl.className = 'status loading';
    document.querySelectorAll('button').forEach(btn => btn.disabled = true);
    logDiagnostic('debug', 'showLoading', { message });
}

/**
 * Hide loading indicator
 */
function hideLoading() {
    updateUI();
    logDiagnostic('debug', 'hideLoading');
}

// ==================== PKCE Utilities ====================

/**
 * Generate a cryptographically random string for PKCE code verifier
 * @returns {string} Base64URL encoded random string
 */
function generateCodeVerifier() {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    const verifier = base64URLEncode(array);
    logDiagnostic('debug', 'generateCodeVerifier', { length: verifier.length });
    return verifier;
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
    const challenge = base64URLEncode(new Uint8Array(hash));
    logDiagnostic('debug', 'generateCodeChallenge', { verifierLength: verifier.length, challengeLength: challenge.length });
    return challenge;
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
    const state = base64URLEncode(array);
    logDiagnostic('debug', 'generateState', { length: state.length });
    return state;
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
    logDiagnostic('info', 'storeTokens', {
        hasAccessToken: !!tokens.access_token,
        hasIDToken: !!tokens.id_token,
        hasRefreshToken: !!tokens.refresh_token,
        tokenType: tokens.token_type,
        expiresIn: tokens.expires_in
    });
    updateUI();
}

/**
 * Retrieve stored tokens
 * @returns {Object|null} Tokens or null if not found
 */
function getStoredTokens() {
    const tokens = sessionStorage.getItem(TOKEN_STORAGE_KEY);
    const parsed = tokens ? JSON.parse(tokens) : null;
    logDiagnostic('debug', 'getStoredTokens', { found: !!parsed });
    return parsed;
}

/**
 * Clear all stored OAuth data
 */
function clearStorage() {
    sessionStorage.removeItem(TOKEN_STORAGE_KEY);
    sessionStorage.removeItem(PKCE_STORAGE_KEY);
    sessionStorage.removeItem(STATE_STORAGE_KEY);
    logDiagnostic('info', 'clearStorage', { message: 'All OAuth data cleared from sessionStorage' });
}

// ==================== OAuth 2.1 Flow ====================

/**
 * Start OAuth 2.1 authorization code flow with PKCE
 */
async function startLogin() {
    logDiagnostic('info', 'startLogin', { message: 'Initiating OAuth 2.1 authorization code flow with PKCE' });

    try {
        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;
        const redirectUri = document.getElementById('redirectUri').value;
        const scope = document.getElementById('scope').value;

        logDiagnostic('debug', 'startLogin:config', { authzUrl, clientId, redirectUri, scope });

        // Show loading state.
        showLoading('Initiating login flow...');

        // Generate PKCE parameters.
        const codeVerifier = generateCodeVerifier();
        const codeChallenge = await generateCodeChallenge(codeVerifier);
        const state = generateState();

        // Store PKCE data and state for callback.
        sessionStorage.setItem(PKCE_STORAGE_KEY, JSON.stringify({
            verifier: codeVerifier,
            challenge: codeChallenge
        }));
        sessionStorage.setItem(STATE_STORAGE_KEY, state);

        logDiagnostic('debug', 'startLogin:storage', { message: 'PKCE data and state stored in sessionStorage' });

        // Build authorization URL.
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

        logDiagnostic('info', 'startLogin:redirect', {
            authUrl: authUrl.split('?')[0],
            params: Object.fromEntries(params)
        });

        // Redirect to authorization server.
        window.location.href = authUrl;
    } catch (error) {
        logDiagnostic('error', 'startLogin:error', { error: error.message, stack: error.stack });
        showError('Failed to start login: ' + error.message);
        hideLoading();
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

    logDiagnostic('info', 'handleCallback', {
        hasCode: !!code,
        hasState: !!state,
        hasError: !!error,
        error: error,
        errorDescription: params.get('error_description')
    });

    // Check for errors.
    if (error) {
        const errorDesc = params.get('error_description') || error;
        logDiagnostic('error', 'handleCallback:authError', { error, errorDescription: errorDesc });
        showError('Authorization failed: ' + errorDesc);
        clearStorage();
        return;
    }

    // Validate state parameter (CSRF protection).
    const storedState = sessionStorage.getItem(STATE_STORAGE_KEY);
    if (!state || state !== storedState) {
        logDiagnostic('error', 'handleCallback:stateValidation', {
            providedState: state,
            storedState: storedState,
            match: state === storedState
        });
        showError('Invalid state parameter - possible CSRF attack');
        clearStorage();
        return;
    }

    logDiagnostic('info', 'handleCallback:stateValid', { message: 'State parameter validated successfully' });

    // Exchange authorization code for tokens.
    if (code) {
        await exchangeCodeForTokens(code);
    }

    // Clean up URL.
    window.history.replaceState({}, document.title, window.location.pathname);
    logDiagnostic('debug', 'handleCallback:cleanup', { message: 'URL cleaned up' });
}

/**
 * Exchange authorization code for tokens
 * @param {string} code - Authorization code
 */
async function exchangeCodeForTokens(code) {
    logDiagnostic('info', 'exchangeCodeForTokens', { codeLength: code.length });

    try {
        showLoading('Exchanging authorization code for tokens...');

        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;
        const redirectUri = document.getElementById('redirectUri').value;

        // Retrieve PKCE verifier.
        const pkceData = JSON.parse(sessionStorage.getItem(PKCE_STORAGE_KEY) || '{}');
        if (!pkceData.verifier) {
            throw new Error('PKCE verifier not found - possible session expiry');
        }

        logDiagnostic('debug', 'exchangeCodeForTokens:pkce', { verifierLength: pkceData.verifier.length });

        // Build token request.
        const body = new URLSearchParams({
            grant_type: 'authorization_code',
            code: code,
            redirect_uri: redirectUri,
            client_id: clientId,
            code_verifier: pkceData.verifier
        });

        const tokenEndpoint = `${authzUrl}/oauth2/v1/token`;
        logDiagnostic('debug', 'exchangeCodeForTokens:request', {
            endpoint: tokenEndpoint,
            grantType: 'authorization_code'
        });

        // Make token request.
        const response = await fetch(tokenEndpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: body.toString()
        });

        logDiagnostic('debug', 'exchangeCodeForTokens:response', {
            status: response.status,
            statusText: response.statusText,
            ok: response.ok
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            logDiagnostic('error', 'exchangeCodeForTokens:httpError', {
                status: response.status,
                error: errorData.error,
                errorDescription: errorData.error_description
            });
            throw new Error(errorData.error_description || 'Token exchange failed');
        }

        const tokens = await response.json();

        logDiagnostic('info', 'exchangeCodeForTokens:success', {
            tokenType: tokens.token_type,
            expiresIn: tokens.expires_in,
            hasRefreshToken: !!tokens.refresh_token
        });

        // Store tokens and update UI.
        storeTokens(tokens);
        showSuccess('Successfully authenticated!');

        // Clean up PKCE data.
        sessionStorage.removeItem(PKCE_STORAGE_KEY);
        sessionStorage.removeItem(STATE_STORAGE_KEY);

        hideLoading();
    } catch (error) {
        logDiagnostic('error', 'exchangeCodeForTokens:error', { error: error.message, stack: error.stack });
        showError('Token exchange failed: ' + error.message);
        clearStorage();
        hideLoading();
    }
}

/**
 * Refresh access token using refresh token
 */
async function refreshToken() {
    logDiagnostic('info', 'refreshToken', { message: 'Initiating token refresh' });

    try {
        showLoading('Refreshing access token...');

        const tokens = getStoredTokens();
        if (!tokens || !tokens.refresh_token) {
            throw new Error('No refresh token available - please log in again');
        }

        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;

        const body = new URLSearchParams({
            grant_type: 'refresh_token',
            refresh_token: tokens.refresh_token,
            client_id: clientId
        });

        const tokenEndpoint = `${authzUrl}/oauth2/v1/token`;
        logDiagnostic('debug', 'refreshToken:request', { endpoint: tokenEndpoint });

        const response = await fetch(tokenEndpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: body.toString()
        });

        logDiagnostic('debug', 'refreshToken:response', {
            status: response.status,
            statusText: response.statusText,
            ok: response.ok
        });

        if (!response.ok) {
            throw new Error('Token refresh failed - please log in again');
        }

        const newTokens = await response.json();
        logDiagnostic('info', 'refreshToken:success', { expiresIn: newTokens.expires_in });

        storeTokens(newTokens);
        showSuccess('Token refreshed successfully!');
        hideLoading();
    } catch (error) {
        logDiagnostic('error', 'refreshToken:error', { error: error.message, stack: error.stack });
        showError('Token refresh failed: ' + error.message);
        clearStorage();
        updateUI();
        hideLoading();
    }
}

/**
 * Fetch user info from OIDC UserInfo endpoint
 */
async function getUserInfo() {
    logDiagnostic('info', 'getUserInfo', { message: 'Fetching OIDC UserInfo' });

    try {
        showLoading('Fetching user information...');

        const tokens = getStoredTokens();
        if (!tokens || !tokens.access_token) {
            throw new Error('No access token available - please log in first');
        }

        const idpUrl = document.getElementById('idpUrl').value;
        const userInfoEndpoint = `${idpUrl}/oidc/v1/userinfo`;

        logDiagnostic('debug', 'getUserInfo:request', { endpoint: userInfoEndpoint });

        const response = await fetch(userInfoEndpoint, {
            headers: {
                'Authorization': `Bearer ${tokens.access_token}`
            }
        });

        logDiagnostic('debug', 'getUserInfo:response', {
            status: response.status,
            statusText: response.statusText,
            ok: response.ok
        });

        if (!response.ok) {
            throw new Error('UserInfo request failed - token may be expired');
        }

        const userInfo = await response.json();
        logDiagnostic('info', 'getUserInfo:success', { claims: Object.keys(userInfo) });

        // Display user info.
        document.getElementById('userInfoData').textContent = JSON.stringify(userInfo, null, 2);
        document.getElementById('userInfoSection').style.display = 'block';
        showSuccess('User info retrieved successfully!');
        hideLoading();
    } catch (error) {
        logDiagnostic('error', 'getUserInfo:error', { error: error.message, stack: error.stack });
        showError('Failed to get user info: ' + error.message);
        hideLoading();
    }
}

/**
 * Introspect access token
 */
async function introspectToken() {
    logDiagnostic('info', 'introspectToken', { message: 'Introspecting access token' });

    try {
        showLoading('Introspecting token...');

        const tokens = getStoredTokens();
        if (!tokens || !tokens.access_token) {
            throw new Error('No access token available - please log in first');
        }

        const authzUrl = document.getElementById('authzUrl').value;
        const clientId = document.getElementById('clientId').value;

        const body = new URLSearchParams({
            token: tokens.access_token,
            client_id: clientId
        });

        const introspectEndpoint = `${authzUrl}/oauth2/v1/introspect`;
        logDiagnostic('debug', 'introspectToken:request', { endpoint: introspectEndpoint });

        const response = await fetch(introspectEndpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: body.toString()
        });

        logDiagnostic('debug', 'introspectToken:response', {
            status: response.status,
            statusText: response.statusText,
            ok: response.ok
        });

        if (!response.ok) {
            throw new Error('Token introspection failed');
        }

        const introspection = await response.json();
        logDiagnostic('info', 'introspectToken:success', {
            active: introspection.active,
            clientId: introspection.client_id,
            scope: introspection.scope
        });

        alert('Token Introspection:\n\n' + JSON.stringify(introspection, null, 2));
        hideLoading();
    } catch (error) {
        logDiagnostic('error', 'introspectToken:error', { error: error.message, stack: error.stack });
        showError('Failed to introspect token: ' + error.message);
        hideLoading();
    }
}

/**
 * Logout and clear all stored data
 */
function logout() {
    logDiagnostic('info', 'logout', { message: 'Logging out and clearing session' });
    clearStorage();
    updateUI();
    showSuccess('Logged out successfully');

    // Hide user info and token sections.
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

    logDiagnostic('debug', 'updateUI', { isAuthenticated, hasRefreshToken: !!(tokens && tokens.refresh_token) });

    // Update buttons.
    document.getElementById('loginBtn').disabled = isAuthenticated;
    document.getElementById('logoutBtn').disabled = !isAuthenticated;
    document.getElementById('refreshBtn').disabled = !isAuthenticated || !tokens.refresh_token;
    document.getElementById('userInfoBtn').disabled = !isAuthenticated;
    document.getElementById('introspectBtn').disabled = !isAuthenticated;

    // Update status.
    const statusEl = document.getElementById('status');
    if (isAuthenticated) {
        statusEl.textContent = '✓ Authenticated';
        statusEl.className = 'status logged-in';

        // Show tokens.
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
    logDiagnostic('info', 'showSuccess', { message });
}

/**
 * Show error message
 * @param {string} message - Error message
 */
function showError(message) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = '✗ ' + message;
    statusEl.className = 'status error';
    logDiagnostic('error', 'showError', { message });
}

// ==================== Initialization ====================

// Handle OAuth callback on page load.
window.addEventListener('DOMContentLoaded', () => {
    logDiagnostic('info', 'init', { message: 'OAuth SPA initialized', url: window.location.href });
    handleCallback();
    updateUI();
});
