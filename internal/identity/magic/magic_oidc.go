// Copyright (c) 2025 Justin Cranford
//
//

package magic

// OIDC claim names.
const (
	ClaimSub               = "sub"                   // Subject claim.
	ClaimName              = "name"                  // Full name claim.
	ClaimGivenName         = "given_name"            // Given name claim.
	ClaimFamilyName        = "family_name"           // Family name claim.
	ClaimMiddleName        = "middle_name"           // Middle name claim.
	ClaimNickname          = "nickname"              // Nickname claim.
	ClaimPreferredUsername = "preferred_username"    // Preferred username claim.
	ClaimProfile           = "profile"               // Profile URL claim.
	ClaimPicture           = "picture"               // Picture URL claim.
	ClaimWebsite           = "website"               // Website URL claim.
	ClaimEmail             = "email"                 // Email claim.
	ClaimEmailVerified     = "email_verified"        // Email verified claim.
	ClaimGender            = "gender"                // Gender claim.
	ClaimBirthdate         = "birthdate"             // Birthdate claim.
	ClaimZoneinfo          = "zoneinfo"              // Timezone claim.
	ClaimLocale            = "locale"                // Locale claim.
	ClaimPhoneNumber       = "phone_number"          // Phone number claim.
	ClaimPhoneVerified     = "phone_number_verified" // Phone verified claim.
	ClaimAddress           = "address"               // Address claim.
	ClaimUpdatedAt         = "updated_at"            // Updated at claim.
)

// OIDC address claim fields.
const (
	AddressFormatted     = "formatted"      // Formatted address.
	AddressStreetAddress = "street_address" // Street address.
	AddressLocality      = "locality"       // Locality (city).
	AddressRegion        = "region"         // Region (state/province).
	AddressPostalCode    = "postal_code"    // Postal code.
	AddressCountry       = "country"        // Country.
)

// OIDC token claims.
const (
	ClaimIss      = "iss"       // Issuer claim.
	ClaimAud      = "aud"       // Audience claim.
	ClaimExp      = "exp"       // Expiration time claim.
	ClaimIat      = "iat"       // Issued at claim.
	ClaimNbf      = "nbf"       // Not before claim.
	ClaimJti      = "jti"       // JWT ID claim.
	ClaimNonce    = "nonce"     // Nonce claim (ID tokens).
	ClaimAcr      = "acr"       // Authentication Context Class Reference.
	ClaimAmr      = "amr"       // Authentication Methods References.
	ClaimAzp      = "azp"       // Authorized party claim.
	ClaimAuthTime = "auth_time" // Authentication time claim.
	ClaimClientID = "client_id" // Client ID claim (OAuth 2.0).
	ClaimScope    = "scope"     // Scope claim (OAuth 2.0).
)

// OIDC endpoint paths.
const (
	PathAuthorize    = "/authorize"                        // Authorization endpoint.
	PathToken        = "/token"                            // Token endpoint.
	PathUserInfo     = "/userinfo"                         // UserInfo endpoint.
	PathJWKS         = "/.well-known/jwks.json"            // JWKS endpoint.
	PathDiscovery    = "/.well-known/openid-configuration" // Discovery endpoint.
	PathRevoke       = "/revoke"                           // Token revocation endpoint.
	PathIntrospect   = "/introspect"                       // Token introspection endpoint.
	PathLogout       = "/logout"                           // Logout endpoint.
	PathEndSession   = "/endsession"                       // End session endpoint.
	PathRegistration = "/register"                         // Dynamic client registration.
)

// OIDC subject types.
const (
	SubjectTypePublic   = "public"   // Public subject type.
	SubjectTypePairwise = "pairwise" // Pairwise subject type.
)

// OIDC display values.
const (
	DisplayPage  = "page"  // Full page display.
	DisplayPopup = "popup" // Popup display.
	DisplayTouch = "touch" // Touch display.
	DisplayWAP   = "wap"   // WAP display.
)

// OIDC prompt values.
const (
	PromptNone          = "none"           // No prompts.
	PromptLogin         = "login"          // Login prompt.
	PromptConsent       = "consent"        // Consent prompt.
	PromptSelectAccount = "select_account" // Account selection prompt.
)

// OIDC ACR values (Authentication Context Class Reference).
const (
	ACRLevel0 = "0" // No authentication.
	ACRLevel1 = "1" // Password authentication.
	ACRLevel2 = "2" // Multi-factor authentication.
	ACRLevel3 = "3" // Hardware-based authentication.
)

// OIDC AMR values (Authentication Methods References).
const (
	AMRPassword    = "pwd"     // Password authentication.
	AMRMultiFactor = "mfa"     // Multi-factor authentication.
	AMRTOTP        = "otp"     // One-time password.
	AMRSMS         = "sms"     // SMS-based authentication.
	AMRBiometric   = "bio"     // Biometric authentication.
	AMRHardware    = "hwk"     // Hardware key authentication.
	AMRPasskey     = "passkey" // Passkey authentication.
	AMRMTLS        = "mTLS"    // mTLS authentication.
)
