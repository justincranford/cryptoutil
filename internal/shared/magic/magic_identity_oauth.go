// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// OAuth 2.1 grant types.
const (
	GrantTypeAuthorizationCode = "authorization_code"                           // Authorization code grant type.
	GrantTypeClientCredentials = "client_credentials"                           // Client credentials grant type.
	GrantTypeRefreshToken      = "refresh_token"                                // Refresh token grant type.
	GrantTypeDeviceCode        = "urn:ietf:params:oauth:grant-type:device_code" // Device code grant type (RFC 8628).
)

// OAuth 2.1 response types.
const (
	ResponseTypeCode = "code" // Authorization code response type.
)

// OAuth 2.1 response modes.
const (
	ResponseModeQuery    = "query"     // Query string response mode.
	ResponseModeFragment = "fragment"  // Fragment response mode.
	ResponseModeFormPost = "form_post" // Form POST response mode.
)

// OAuth 2.1 scopes.
const (
	ScopeOpenID        = "openid"         // OpenID scope.
	ScopeProfile       = "profile"        // Profile scope.
	ScopeEmail         = "email"          // Email scope.
	ScopeAddress       = "address"        // Address scope.
	ScopePhone         = "phone"          // Phone scope.
	ScopeOfflineAccess = "offline_access" // Offline access scope (refresh tokens).
)

// OAuth 2.1 token types.
const (
	TokenTypeBearer       = "Bearer"        // Bearer token type.
	TokenTypeAccessToken  = "access_token"  // Access token type hint.
	TokenTypeRefreshToken = "refresh_token" // Refresh token type hint.
)

// OAuth 2.1 PKCE methods.
const (
	PKCEMethodPlain = "plain" // Plain PKCE method.
	PKCEMethodS256  = "S256"  // SHA-256 PKCE method.
)

// Token formats.
const (
	TokenFormatJWS  = "jws"  // JWS token format (signed JWT).
	TokenFormatJWE  = "jwe"  // JWE token format (encrypted JWT).
	TokenFormatUUID = "uuid" // UUID token format (opaque token).
)

// OAuth 2.1 parameter names.
const (
	ParamClientID            = "client_id"             // Client ID parameter.
	ParamClientSecret        = "client_secret"         // Client secret parameter.
	ParamGrantType           = "grant_type"            // Grant type parameter.
	ParamCode                = "code"                  // Authorization code parameter.
	ParamRedirectURI         = "redirect_uri"          // Redirect URI parameter.
	ParamState               = "state"                 // State parameter.
	ParamScope               = "scope"                 // Scope parameter.
	ParamResponseType        = "response_type"         // Response type parameter.
	ParamResponseMode        = "response_mode"         // Response mode parameter.
	ParamCodeChallenge       = "code_challenge"        // PKCE code challenge parameter.
	ParamCodeChallengeMethod = "code_challenge_method" // PKCE code challenge method parameter.
	ParamCodeVerifier        = "code_verifier"         // PKCE code verifier parameter.
	ParamAccessToken         = "access_token"          // Access token parameter.
	ParamRefreshToken        = "refresh_token"         // Refresh token parameter.
	ParamTokenType           = "token_type"            // Token type parameter.
	ParamExpiresIn           = "expires_in"            // Token expiration parameter.
	ParamIDToken             = "id_token"              // ID token parameter.
	ParamToken               = "token"                 // Token parameter (introspection/revocation).
	ParamTokenTypeHint       = "token_type_hint"       // Token type hint parameter.
	ParamDeviceCode          = "device_code"           // Device code parameter (RFC 8628).
	ParamUserCode            = "user_code"             // User code parameter (RFC 8628).
	ParamRequestURI          = "request_uri"           // Request URI parameter (RFC 9126).
)

// OAuth 2.1 error codes.
const (
	ErrorInvalidRequest          = "invalid_request"           // Invalid request error.
	ErrorInvalidClient           = "invalid_client"            // Invalid client error.
	ErrorInvalidGrant            = "invalid_grant"             // Invalid grant error.
	ErrorUnauthorizedClient      = "unauthorized_client"       // Unauthorized client error.
	ErrorUnsupportedGrantType    = "unsupported_grant_type"    // Unsupported grant type error.
	ErrorInvalidScope            = "invalid_scope"             // Invalid scope error.
	ErrorAccessDenied            = "access_denied"             // Access denied error.
	ErrorUnsupportedResponseType = "unsupported_response_type" // Unsupported response type error.
	ErrorServerError             = "server_error"              // Server error.
	ErrorTemporarilyUnavailable  = "temporarily_unavailable"   // Temporarily unavailable error.
	ErrorInvalidToken            = "invalid_token"             // Invalid token error (RFC 6750).
	ErrorInsufficientScope       = "insufficient_scope"        // Insufficient scope error (RFC 6750).
	ErrorAuthorizationPending    = "authorization_pending"     // User has not yet authorized (RFC 8628).
	ErrorSlowDown                = "slow_down"                 // Polling too fast (RFC 8628).
	ErrorExpiredToken            = "expired_token"             // Device code expired (RFC 8628).
	ErrorInvalidRequestURI       = "invalid_request_uri"       // Invalid request_uri (RFC 9126).
	ErrorInvalidRequestObject    = "invalid_request_object"    // Invalid request object (RFC 9126).
)

// OAuth 2.1 client authentication methods.
const (
	ClientAuthMethodSecretBasic       = "client_secret_basic"         // HTTP Basic authentication.
	ClientAuthMethodSecretPost        = "client_secret_post"          // POST body authentication.
	ClientAuthMethodSecretJWT         = "client_secret_jwt"           // JWT signed with client secret.
	ClientAuthMethodPrivateKeyJWT     = "private_key_jwt"             // JWT signed with private key.
	ClientAuthMethodTLSClientAuth     = "tls_client_auth"             // mTLS with CA-issued certificate.
	ClientAuthMethodSelfSignedTLSAuth = "self_signed_tls_client_auth" // mTLS with self-signed certificate.
	ClientAuthMethodBearerToken       = "bearer_token"                // Bearer token authentication.
	ClientAuthMethodNone              = "none"                        // No authentication (public clients).
)

// JWT assertion validation constants.
const (
	JWTAssertionMaxLifetime = 10 * time.Minute // Maximum allowed lifetime for client authentication JWT assertions (RFC 7523 Section 3).
)

// Rate limiting constants.
const (
	RateLimitRequestsPerWindow = 100 // Maximum requests per time window.
	RateLimitWindowSeconds     = 60  // Time window in seconds.
)

// Fiber HTTP server timeout constants (in seconds).
const (
	FiberReadTimeoutSeconds  = 30  // Fiber read timeout in seconds.
	FiberWriteTimeoutSeconds = 30  // Fiber write timeout in seconds.
	FiberIdleTimeoutSeconds  = 120 // Fiber idle timeout in seconds.
	ShutdownTimeoutSeconds   = 30  // Graceful shutdown timeout in seconds.
)

// Default server ports.
// Port ranges per service catalog (architecture.md):
// - identity-authz: 8200-8299
// - identity-idp: 8300-8399
// - identity-rs: 8400-8499
// - identity-rp: 8500-8599
// - identity-spa: 8600-8699.
const (
	DefaultAuthZPort = 8200 // Default OAuth 2.1 authorization server port.
	DefaultIDPPort   = 8300 // Default OIDC identity provider server port.
	DefaultRSPort    = 8400 // Default resource server port.
	DefaultSPARPPort = 8500 // Default SPA relying party demo server port.
)
