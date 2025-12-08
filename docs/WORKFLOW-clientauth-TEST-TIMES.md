=== RUN   TestBasicAuthenticator_MethodName
=== PAUSE TestBasicAuthenticator_MethodName
=== RUN   TestBasicAuthenticator_Authenticate
=== PAUSE TestBasicAuthenticator_Authenticate
=== RUN   TestBasicAuthenticator_ValidateAuthMethod
=== PAUSE TestBasicAuthenticator_ValidateAuthMethod
=== RUN   TestRegistry_AllAuthMethods
=== PAUSE TestRegistry_AllAuthMethods
=== RUN   TestRegistry_UnknownMethod
=== PAUSE TestRegistry_UnknownMethod
=== RUN   TestRegistry_RegisterCustomAuthenticator
=== PAUSE TestRegistry_RegisterCustomAuthenticator
=== RUN   TestBasicAuthenticator_Method
=== PAUSE TestBasicAuthenticator_Method
=== RUN   TestPostAuthenticator_Method
=== PAUSE TestPostAuthenticator_Method
=== RUN   TestClientAuthentication_MultiSecretValidation
=== PAUSE TestClientAuthentication_MultiSecretValidation
=== RUN   TestClientAuthentication_OldSecretExpired
=== PAUSE TestClientAuthentication_OldSecretExpired
=== RUN   TestClientAuthentication_NewSecretImmediate
=== PAUSE TestClientAuthentication_NewSecretImmediate
=== RUN   TestClientAuthentication_RevokedSecretRejected
=== PAUSE TestClientAuthentication_RevokedSecretRejected
=== RUN   TestClientSecretJWTAuthenticator_Method
=== PAUSE TestClientSecretJWTAuthenticator_Method
=== RUN   TestClientSecretJWTAuthenticator_Authenticate_MissingAssertion
=== PAUSE TestClientSecretJWTAuthenticator_Authenticate_MissingAssertion
=== RUN   TestPrivateKeyJWTAuthenticator_Method
=== PAUSE TestPrivateKeyJWTAuthenticator_Method
=== RUN   TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion
=== PAUSE TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion
=== RUN   TestTLSClientAuthenticator_Method
=== PAUSE TestTLSClientAuthenticator_Method
=== RUN   TestTLSClientAuthenticator_Authenticate_MissingCredential
=== PAUSE TestTLSClientAuthenticator_Authenticate_MissingCredential
=== RUN   TestSelfSignedAuthenticator_Method
=== PAUSE TestSelfSignedAuthenticator_Method
=== RUN   TestSelfSignedAuthenticator_Authenticate_MissingCredential
=== PAUSE TestSelfSignedAuthenticator_Authenticate_MissingCredential
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_Success
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_Success
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_NoJWKSet
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_NoJWKSet
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_InvalidJWKSet
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_InvalidJWKSet
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_InvalidSignature
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_InvalidSignature
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_ExpiredToken
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_ExpiredToken
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_InvalidIssuer
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_InvalidIssuer
=== RUN   TestPrivateKeyJWTValidator_ValidateJWT_InvalidAudience
=== PAUSE TestPrivateKeyJWTValidator_ValidateJWT_InvalidAudience
=== RUN   TestClientSecretJWTValidator_ValidateJWT_Success
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_Success
=== RUN   TestClientSecretJWTValidator_ValidateJWT_NoClientSecret
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_NoClientSecret
=== RUN   TestClientSecretJWTValidator_ValidateJWT_InvalidSignature
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_InvalidSignature
=== RUN   TestClientSecretJWTValidator_ValidateJWT_ExpiredToken
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_ExpiredToken
=== RUN   TestClientSecretJWTValidator_ValidateJWT_MissingExpirationClaim
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_MissingExpirationClaim
=== RUN   TestClientSecretJWTValidator_ValidateJWT_MissingIssuedAtClaim
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_MissingIssuedAtClaim
=== RUN   TestClientSecretJWTValidator_ValidateJWT_MalformedJWT
=== PAUSE TestClientSecretJWTValidator_ValidateJWT_MalformedJWT
=== RUN   TestPrivateKeyJWTValidator_ExtractClaims_AllClaimsPresent
=== PAUSE TestPrivateKeyJWTValidator_ExtractClaims_AllClaimsPresent
=== RUN   TestClientSecretJWTValidator_ExtractClaims_AllClaimsPresent
=== PAUSE TestClientSecretJWTValidator_ExtractClaims_AllClaimsPresent
=== RUN   TestPostAuthenticator_MethodName
=== PAUSE TestPostAuthenticator_MethodName
=== RUN   TestPostAuthenticator_Authenticate
=== PAUSE TestPostAuthenticator_Authenticate
=== RUN   TestPostAuthenticator_ValidateAuthMethod
=== PAUSE TestPostAuthenticator_ValidateAuthMethod
=== RUN   TestCRLCache_GetCRL
=== PAUSE TestCRLCache_GetCRL
=== RUN   TestCRLCache_GetCRL_ServerError
=== PAUSE TestCRLCache_GetCRL_ServerError
=== RUN   TestCRLCache_GetCRL_InvalidCRL
=== PAUSE TestCRLCache_GetCRL_InvalidCRL
=== RUN   TestCRLCache_IsRevoked
=== PAUSE TestCRLCache_IsRevoked
=== RUN   TestCRLRevocationChecker_CheckRevocation
=== PAUSE TestCRLRevocationChecker_CheckRevocation
=== RUN   TestOCSPRevocationChecker_CheckRevocation_NoOCSPServer
=== PAUSE TestOCSPRevocationChecker_CheckRevocation_NoOCSPServer
=== RUN   TestOCSPRevocationChecker_CheckRevocation_Good
=== PAUSE TestOCSPRevocationChecker_CheckRevocation_Good
=== RUN   TestCombinedRevocationChecker_CheckRevocation
=== PAUSE TestCombinedRevocationChecker_CheckRevocation
=== RUN   TestPBKDF2Hasher_HashSecret
=== PAUSE TestPBKDF2Hasher_HashSecret
=== RUN   TestPBKDF2Hasher_CompareSecret
=== PAUSE TestPBKDF2Hasher_CompareSecret
=== RUN   TestMigrateClientSecrets
=== PAUSE TestMigrateClientSecrets
=== RUN   TestSecretBasedAuthenticator_AuthenticatePost
=== PAUSE TestSecretBasedAuthenticator_AuthenticatePost
=== RUN   TestSecretBasedAuthenticator_MigrateSecrets
=== PAUSE TestSecretBasedAuthenticator_MigrateSecrets
=== RUN   TestSecretBasedAuthenticator_AuthenticateBasic
=== PAUSE TestSecretBasedAuthenticator_AuthenticateBasic
=== RUN   TestSelfSignedAuthenticator_Authenticate
=== PAUSE TestSelfSignedAuthenticator_Authenticate
=== RUN   TestSelfSignedAuthenticator_Method_Custom
=== PAUSE TestSelfSignedAuthenticator_Method_Custom
=== RUN   TestTLSClientAuthenticator_Authenticate_Cert
=== PAUSE TestTLSClientAuthenticator_Authenticate_Cert
=== RUN   TestTLSClientAuthenticator_Method_Cert
=== PAUSE TestTLSClientAuthenticator_Method_Cert
=== RUN   TestCACertificateValidator_ValidCertificate
=== PAUSE TestCACertificateValidator_ValidCertificate
=== RUN   TestCACertificateValidator_ExpiredCertificate
=== PAUSE TestCACertificateValidator_ExpiredCertificate
=== RUN   TestCACertificateValidator_UntrustedCA
=== PAUSE TestCACertificateValidator_UntrustedCA
=== RUN   TestCACertificateValidator_NilCertificate
=== PAUSE TestCACertificateValidator_NilCertificate
=== RUN   TestCACertificateValidator_IsRevoked_Deprecated
=== PAUSE TestCACertificateValidator_IsRevoked_Deprecated
=== RUN   TestSelfSignedCertificateValidator_ValidateCertificate
=== PAUSE TestSelfSignedCertificateValidator_ValidateCertificate
=== RUN   TestSelfSignedCertificateValidator_IsRevoked
=== PAUSE TestSelfSignedCertificateValidator_IsRevoked
=== RUN   TestCertificateParser_ParsePEMCertificate
=== PAUSE TestCertificateParser_ParsePEMCertificate
=== RUN   TestNewPBKDF2Hasher
=== PAUSE TestNewPBKDF2Hasher
=== RUN   TestPBKDF2Hasher_HashSecret
=== PAUSE TestPBKDF2Hasher_HashSecret
=== RUN   TestPBKDF2Hasher_HashUniqueness
=== PAUSE TestPBKDF2Hasher_HashUniqueness
=== RUN   TestPBKDF2Hasher_CompareSecret
=== PAUSE TestPBKDF2Hasher_CompareSecret
=== RUN   TestPBKDF2Hasher_CompareSecret_ConstantTime
=== PAUSE TestPBKDF2Hasher_CompareSecret_ConstantTime
=== RUN   TestPBKDF2Hasher_FIPS140_3Compliance
=== PAUSE TestPBKDF2Hasher_FIPS140_3Compliance
=== RUN   TestPBKDF2Hasher_SaltRandomness
=== PAUSE TestPBKDF2Hasher_SaltRandomness
=== RUN   TestPBKDF2Hasher_EdgeCases
=== PAUSE TestPBKDF2Hasher_EdgeCases
=== RUN   TestPBKDF2Hasher_CompareSecret_VectorTests
=== PAUSE TestPBKDF2Hasher_CompareSecret_VectorTests
=== RUN   TestRegistry_Creation
=== PAUSE TestRegistry_Creation
=== RUN   TestRegistry_GetAuthenticator
=== PAUSE TestRegistry_GetAuthenticator
=== RUN   TestRegistry_GetAuthenticator_NotFound
=== PAUSE TestRegistry_GetAuthenticator_NotFound
=== RUN   TestRegistry_GetHasher
=== PAUSE TestRegistry_GetHasher
=== RUN   TestHashSecret
=== PAUSE TestHashSecret
=== RUN   TestHashSecret_Uniqueness
=== PAUSE TestHashSecret_Uniqueness
=== RUN   TestCompareSecret
=== PAUSE TestCompareSecret
=== RUN   TestCompareSecret_InvalidFormat
=== PAUSE TestCompareSecret_InvalidFormat
=== RUN   TestCompareSecret_ConstantTime
=== PAUSE TestCompareSecret_ConstantTime
=== CONT  TestBasicAuthenticator_MethodName
=== CONT  TestCRLCache_IsRevoked
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_NoJWKSet
--- PASS: TestBasicAuthenticator_MethodName (0.00s)
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_Success
=== CONT  TestClientSecretJWTValidator_ValidateJWT_MissingIssuedAtClaim
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_NoJWKSet (0.00s)
=== CONT  TestSelfSignedAuthenticator_Method
=== CONT  TestTLSClientAuthenticator_Authenticate_MissingCredential
=== CONT  TestTLSClientAuthenticator_Method
=== CONT  TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion
=== CONT  TestSelfSignedAuthenticator_Authenticate_MissingCredential
--- PASS: TestClientSecretJWTValidator_ValidateJWT_MissingIssuedAtClaim (0.00s)
=== CONT  TestPrivateKeyJWTAuthenticator_Method
--- PASS: TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion (0.01s)
=== CONT  TestClientSecretJWTAuthenticator_Authenticate_MissingAssertion
--- PASS: TestPrivateKeyJWTAuthenticator_Method (0.03s)
=== CONT  TestClientSecretJWTAuthenticator_Method
--- PASS: TestSelfSignedAuthenticator_Authenticate_MissingCredential (0.04s)
=== CONT  TestClientAuthentication_RevokedSecretRejected
--- PASS: TestTLSClientAuthenticator_Method (0.05s)
=== CONT  TestClientAuthentication_NewSecretImmediate
=== CONT  TestClientAuthentication_OldSecretExpired
--- PASS: TestTLSClientAuthenticator_Authenticate_MissingCredential (0.06s)
--- PASS: TestSelfSignedAuthenticator_Method (0.08s)
=== CONT  TestClientAuthentication_MultiSecretValidation
--- PASS: TestClientSecretJWTAuthenticator_Authenticate_MissingAssertion (0.07s)
=== CONT  TestPostAuthenticator_Method
--- PASS: TestClientSecretJWTAuthenticator_Method (0.08s)
=== CONT  TestBasicAuthenticator_Method
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_Success (0.12s)
=== CONT  TestRegistry_RegisterCustomAuthenticator
=== CONT  TestRegistry_UnknownMethod
--- PASS: TestPostAuthenticator_Method (0.07s)
--- PASS: TestBasicAuthenticator_Method (0.07s)
=== CONT  TestRegistry_AllAuthMethods
--- PASS: TestRegistry_RegisterCustomAuthenticator (0.06s)
=== CONT  TestBasicAuthenticator_ValidateAuthMethod
=== RUN   TestBasicAuthenticator_ValidateAuthMethod/valid_auth_method
=== PAUSE TestBasicAuthenticator_ValidateAuthMethod/valid_auth_method
=== RUN   TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_POST
=== PAUSE TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_POST
=== RUN   TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt
=== PAUSE TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt
=== CONT  TestBasicAuthenticator_Authenticate
=== RUN   TestCRLCache_IsRevoked/certificate_is_revoked
=== RUN   TestCRLCache_IsRevoked/certificate_is_not_revoked
=== RUN   TestCRLCache_IsRevoked/empty_CRL
--- PASS: TestCRLCache_IsRevoked (0.20s)
    --- PASS: TestCRLCache_IsRevoked/certificate_is_revoked (0.00s)
    --- PASS: TestCRLCache_IsRevoked/certificate_is_not_revoked (0.00s)
    --- PASS: TestCRLCache_IsRevoked/empty_CRL (0.00s)
=== CONT  TestPostAuthenticator_Authenticate
--- PASS: TestRegistry_UnknownMethod (0.04s)
=== CONT  TestCRLCache_GetCRL_InvalidCRL
--- PASS: TestRegistry_AllAuthMethods (0.04s)
=== CONT  TestCRLCache_GetCRL_ServerError
--- PASS: TestCRLCache_GetCRL_InvalidCRL (0.02s)
=== CONT  TestCRLCache_GetCRL
--- PASS: TestCRLCache_GetCRL_ServerError (0.03s)
=== CONT  TestPostAuthenticator_ValidateAuthMethod
=== RUN   TestPostAuthenticator_ValidateAuthMethod/valid_auth_method
=== PAUSE TestPostAuthenticator_ValidateAuthMethod/valid_auth_method
=== RUN   TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_BASIC
=== PAUSE TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_BASIC
=== RUN   TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt
=== PAUSE TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt
=== CONT  TestSelfSignedCertificateValidator_IsRevoked
--- PASS: TestSelfSignedCertificateValidator_IsRevoked (0.00s)
=== CONT  TestCompareSecret_ConstantTime
=== RUN   TestCRLCache_GetCRL/fetch_CRL_first_time
=== RUN   TestCRLCache_GetCRL/use_cached_CRL
=== RUN   TestCRLCache_GetCRL/refetch_after_cache_expiration
--- PASS: TestCRLCache_GetCRL (0.37s)
    --- PASS: TestCRLCache_GetCRL/fetch_CRL_first_time (0.00s)
    --- PASS: TestCRLCache_GetCRL/use_cached_CRL (0.00s)
    --- PASS: TestCRLCache_GetCRL/refetch_after_cache_expiration (0.20s)
=== CONT  TestCompareSecret_InvalidFormat
=== RUN   TestCompareSecret_InvalidFormat/missing_separator
=== PAUSE TestCompareSecret_InvalidFormat/missing_separator
=== RUN   TestCompareSecret_InvalidFormat/invalid_base64_salt
=== PAUSE TestCompareSecret_InvalidFormat/invalid_base64_salt
=== RUN   TestCompareSecret_InvalidFormat/invalid_base64_hash
=== PAUSE TestCompareSecret_InvalidFormat/invalid_base64_hash
=== RUN   TestCompareSecret_InvalidFormat/empty_hashed
=== PAUSE TestCompareSecret_InvalidFormat/empty_hashed
=== CONT  TestCompareSecret
=== RUN   TestCompareSecret/matching_secret
=== PAUSE TestCompareSecret/matching_secret
=== RUN   TestCompareSecret/non-matching_secret
=== PAUSE TestCompareSecret/non-matching_secret
=== RUN   TestCompareSecret/empty_secret_matches_empty
=== PAUSE TestCompareSecret/empty_secret_matches_empty
=== RUN   TestCompareSecret/empty_secret_does_not_match_non-empty
=== PAUSE TestCompareSecret/empty_secret_does_not_match_non-empty
=== CONT  TestHashSecret_Uniqueness
=== RUN   TestBasicAuthenticator_Authenticate/valid_basic_auth
=== PAUSE TestBasicAuthenticator_Authenticate/valid_basic_auth
=== RUN   TestBasicAuthenticator_Authenticate/invalid_client_secret
=== PAUSE TestBasicAuthenticator_Authenticate/invalid_client_secret
=== RUN   TestBasicAuthenticator_Authenticate/client_not_found
=== PAUSE TestBasicAuthenticator_Authenticate/client_not_found
=== CONT  TestHashSecret
=== RUN   TestHashSecret/valid_secret
=== PAUSE TestHashSecret/valid_secret
=== RUN   TestHashSecret/empty_secret
=== PAUSE TestHashSecret/empty_secret
=== RUN   TestHashSecret/long_secret
=== PAUSE TestHashSecret/long_secret
=== CONT  TestRegistry_GetHasher
--- PASS: TestRegistry_GetHasher (0.01s)
=== CONT  TestRegistry_GetAuthenticator_NotFound
=== RUN   TestPostAuthenticator_Authenticate/valid_post_auth
=== PAUSE TestPostAuthenticator_Authenticate/valid_post_auth
=== RUN   TestPostAuthenticator_Authenticate/invalid_client_secret
=== PAUSE TestPostAuthenticator_Authenticate/invalid_client_secret
=== RUN   TestPostAuthenticator_Authenticate/client_not_found
=== PAUSE TestPostAuthenticator_Authenticate/client_not_found
=== CONT  TestRegistry_GetAuthenticator
--- PASS: TestRegistry_GetAuthenticator_NotFound (0.02s)
=== CONT  TestRegistry_Creation
--- PASS: TestRegistry_GetAuthenticator (0.01s)
=== CONT  TestPBKDF2Hasher_CompareSecret_VectorTests
--- PASS: TestRegistry_Creation (0.02s)
=== CONT  TestPBKDF2Hasher_EdgeCases
=== RUN   TestPBKDF2Hasher_EdgeCases/very_long_password_(10KB)
=== PAUSE TestPBKDF2Hasher_EdgeCases/very_long_password_(10KB)
=== RUN   TestPBKDF2Hasher_EdgeCases/special_characters
=== PAUSE TestPBKDF2Hasher_EdgeCases/special_characters
=== RUN   TestPBKDF2Hasher_EdgeCases/whitespace_only
=== PAUSE TestPBKDF2Hasher_EdgeCases/whitespace_only
=== RUN   TestPBKDF2Hasher_EdgeCases/null_bytes_(valid_UTF-8)
=== PAUSE TestPBKDF2Hasher_EdgeCases/null_bytes_(valid_UTF-8)
=== CONT  TestPBKDF2Hasher_SaltRandomness
--- PASS: TestPBKDF2Hasher_CompareSecret_VectorTests (0.25s)
=== CONT  TestPBKDF2Hasher_FIPS140_3Compliance
--- PASS: TestPBKDF2Hasher_FIPS140_3Compliance (0.09s)
=== CONT  TestPBKDF2Hasher_CompareSecret_ConstantTime
--- PASS: TestClientAuthentication_MultiSecretValidation (1.10s)
=== CONT  TestPBKDF2Hasher_CompareSecret
--- PASS: TestClientAuthentication_NewSecretImmediate (1.13s)
=== CONT  TestPBKDF2Hasher_HashUniqueness
=== RUN   TestPBKDF2Hasher_CompareSecret/correct_password_matches
=== PAUSE TestPBKDF2Hasher_CompareSecret/correct_password_matches
=== RUN   TestPBKDF2Hasher_CompareSecret/incorrect_password_does_not_match
=== PAUSE TestPBKDF2Hasher_CompareSecret/incorrect_password_does_not_match
=== RUN   TestPBKDF2Hasher_CompareSecret/empty_password_does_not_match
=== PAUSE TestPBKDF2Hasher_CompareSecret/empty_password_does_not_match
=== RUN   TestPBKDF2Hasher_CompareSecret/case-sensitive_comparison
=== PAUSE TestPBKDF2Hasher_CompareSecret/case-sensitive_comparison
=== RUN   TestPBKDF2Hasher_CompareSecret/malformed_hash_(3_parts)
=== PAUSE TestPBKDF2Hasher_CompareSecret/malformed_hash_(3_parts)
=== RUN   TestPBKDF2Hasher_CompareSecret/malformed_hash_(wrong_prefix)
=== PAUSE TestPBKDF2Hasher_CompareSecret/malformed_hash_(wrong_prefix)
=== RUN   TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_iterations)
=== PAUSE TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_iterations)
=== RUN   TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_salt_base64)
=== PAUSE TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_salt_base64)
=== RUN   TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_hash_base64)
=== PAUSE TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_hash_base64)
=== CONT  TestPBKDF2Hasher_HashSecret
=== RUN   TestPBKDF2Hasher_HashSecret/valid_strong_password
=== PAUSE TestPBKDF2Hasher_HashSecret/valid_strong_password
=== RUN   TestPBKDF2Hasher_HashSecret/valid_weak_password
=== PAUSE TestPBKDF2Hasher_HashSecret/valid_weak_password
=== RUN   TestPBKDF2Hasher_HashSecret/valid_empty_password
=== PAUSE TestPBKDF2Hasher_HashSecret/valid_empty_password
=== RUN   TestPBKDF2Hasher_HashSecret/valid_unicode_password
=== PAUSE TestPBKDF2Hasher_HashSecret/valid_unicode_password
=== RUN   TestPBKDF2Hasher_HashSecret/valid_long_password_(256_chars)
=== PAUSE TestPBKDF2Hasher_HashSecret/valid_long_password_(256_chars)
=== CONT  TestNewPBKDF2Hasher
--- PASS: TestNewPBKDF2Hasher (0.00s)
=== CONT  TestCertificateParser_ParsePEMCertificate
--- PASS: TestPBKDF2Hasher_HashUniqueness (0.15s)
=== CONT  TestClientSecretJWTValidator_ValidateJWT_Success
--- PASS: TestClientSecretJWTValidator_ValidateJWT_Success (0.00s)
=== CONT  TestClientSecretJWTValidator_ValidateJWT_MissingExpirationClaim
--- PASS: TestClientSecretJWTValidator_ValidateJWT_MissingExpirationClaim (0.00s)
=== CONT  TestClientSecretJWTValidator_ValidateJWT_ExpiredToken
--- PASS: TestClientSecretJWTValidator_ValidateJWT_ExpiredToken (0.00s)
=== CONT  TestClientSecretJWTValidator_ValidateJWT_InvalidSignature
--- PASS: TestClientSecretJWTValidator_ValidateJWT_InvalidSignature (0.00s)
=== CONT  TestClientSecretJWTValidator_ValidateJWT_NoClientSecret
--- PASS: TestClientSecretJWTValidator_ValidateJWT_NoClientSecret (0.00s)
=== CONT  TestSelfSignedAuthenticator_Authenticate
=== RUN   TestSelfSignedAuthenticator_Authenticate/valid_self-signed_certificate
=== PAUSE TestSelfSignedAuthenticator_Authenticate/valid_self-signed_certificate
=== RUN   TestSelfSignedAuthenticator_Authenticate/missing_certificate
=== PAUSE TestSelfSignedAuthenticator_Authenticate/missing_certificate
=== RUN   TestSelfSignedAuthenticator_Authenticate/subject_mismatch
=== PAUSE TestSelfSignedAuthenticator_Authenticate/subject_mismatch
=== RUN   TestSelfSignedAuthenticator_Authenticate/fingerprint_mismatch
=== PAUSE TestSelfSignedAuthenticator_Authenticate/fingerprint_mismatch
=== CONT  TestSelfSignedCertificateValidator_ValidateCertificate
=== RUN   TestCertificateParser_ParsePEMCertificate/valid_PEM_certificate
=== RUN   TestCertificateParser_ParsePEMCertificate/invalid_PEM_data
=== RUN   TestCertificateParser_ParsePEMCertificate/invalid_certificate_data
--- PASS: TestCertificateParser_ParsePEMCertificate (0.17s)
    --- PASS: TestCertificateParser_ParsePEMCertificate/valid_PEM_certificate (0.00s)
    --- PASS: TestCertificateParser_ParsePEMCertificate/invalid_PEM_data (0.00s)
    --- PASS: TestCertificateParser_ParsePEMCertificate/invalid_certificate_data (0.00s)
=== CONT  TestCACertificateValidator_IsRevoked_Deprecated
--- PASS: TestCACertificateValidator_IsRevoked_Deprecated (0.00s)
=== CONT  TestCACertificateValidator_NilCertificate
--- PASS: TestCACertificateValidator_NilCertificate (0.00s)
=== CONT  TestCACertificateValidator_UntrustedCA
--- PASS: TestCACertificateValidator_UntrustedCA (0.00s)
=== CONT  TestCACertificateValidator_ExpiredCertificate
--- PASS: TestCACertificateValidator_ExpiredCertificate (0.00s)
=== CONT  TestCACertificateValidator_ValidCertificate
--- PASS: TestCACertificateValidator_ValidCertificate (0.00s)
=== CONT  TestTLSClientAuthenticator_Method_Cert
--- PASS: TestTLSClientAuthenticator_Method_Cert (0.00s)
=== CONT  TestTLSClientAuthenticator_Authenticate_Cert
=== RUN   TestTLSClientAuthenticator_Authenticate_Cert/valid_certificate
=== PAUSE TestTLSClientAuthenticator_Authenticate_Cert/valid_certificate
=== RUN   TestTLSClientAuthenticator_Authenticate_Cert/missing_certificate
=== PAUSE TestTLSClientAuthenticator_Authenticate_Cert/missing_certificate
=== RUN   TestTLSClientAuthenticator_Authenticate_Cert/subject_mismatch
=== PAUSE TestTLSClientAuthenticator_Authenticate_Cert/subject_mismatch
=== RUN   TestTLSClientAuthenticator_Authenticate_Cert/fingerprint_mismatch
=== PAUSE TestTLSClientAuthenticator_Authenticate_Cert/fingerprint_mismatch
=== CONT  TestSelfSignedAuthenticator_Method_Custom
--- PASS: TestSelfSignedAuthenticator_Method_Custom (0.00s)
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_ExpiredToken
=== RUN   TestSelfSignedCertificateValidator_ValidateCertificate/valid_pinned_certificate
--- PASS: TestPBKDF2Hasher_CompareSecret_ConstantTime (0.52s)
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_InvalidAudience
=== RUN   TestSelfSignedCertificateValidator_ValidateCertificate/nil_certificate
=== RUN   TestSelfSignedCertificateValidator_ValidateCertificate/expired_certificate
=== RUN   TestSelfSignedCertificateValidator_ValidateCertificate/not_yet_valid_certificate
=== RUN   TestSelfSignedCertificateValidator_ValidateCertificate/certificate_not_pinned
--- PASS: TestSelfSignedCertificateValidator_ValidateCertificate (0.28s)
    --- PASS: TestSelfSignedCertificateValidator_ValidateCertificate/valid_pinned_certificate (0.00s)
    --- PASS: TestSelfSignedCertificateValidator_ValidateCertificate/nil_certificate (0.00s)
    --- PASS: TestSelfSignedCertificateValidator_ValidateCertificate/expired_certificate (0.00s)
    --- PASS: TestSelfSignedCertificateValidator_ValidateCertificate/not_yet_valid_certificate (0.00s)
    --- PASS: TestSelfSignedCertificateValidator_ValidateCertificate/certificate_not_pinned (0.00s)
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_InvalidIssuer
--- PASS: TestHashSecret_Uniqueness (1.04s)
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_InvalidSignature
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_ExpiredToken (0.22s)
=== CONT  TestPBKDF2Hasher_CompareSecret
--- PASS: TestClientAuthentication_RevokedSecretRejected (1.64s)
=== CONT  TestSecretBasedAuthenticator_AuthenticateBasic
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_InvalidIssuer (0.09s)
=== CONT  TestSecretBasedAuthenticator_MigrateSecrets
=== RUN   TestPBKDF2Hasher_CompareSecret/matching_password
=== PAUSE TestPBKDF2Hasher_CompareSecret/matching_password
=== RUN   TestPBKDF2Hasher_CompareSecret/wrong_password
=== PAUSE TestPBKDF2Hasher_CompareSecret/wrong_password
=== RUN   TestPBKDF2Hasher_CompareSecret/invalid_hash_format
=== PAUSE TestPBKDF2Hasher_CompareSecret/invalid_hash_format
=== RUN   TestPBKDF2Hasher_CompareSecret/empty_plaintext
=== PAUSE TestPBKDF2Hasher_CompareSecret/empty_plaintext
=== CONT  TestSecretBasedAuthenticator_AuthenticatePost
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_InvalidAudience (0.13s)
=== CONT  TestMigrateClientSecrets
=== RUN   TestMigrateClientSecrets/migrate_plaintext_secrets
=== PAUSE TestMigrateClientSecrets/migrate_plaintext_secrets
=== RUN   TestMigrateClientSecrets/skip_public_clients
=== PAUSE TestMigrateClientSecrets/skip_public_clients
=== CONT  TestOCSPRevocationChecker_CheckRevocation_Good
=== RUN   TestSecretBasedAuthenticator_AuthenticateBasic/valid_credentials
=== PAUSE TestSecretBasedAuthenticator_AuthenticateBasic/valid_credentials
=== RUN   TestSecretBasedAuthenticator_AuthenticateBasic/wrong_secret
=== PAUSE TestSecretBasedAuthenticator_AuthenticateBasic/wrong_secret
=== RUN   TestSecretBasedAuthenticator_AuthenticateBasic/disabled_client
=== PAUSE TestSecretBasedAuthenticator_AuthenticateBasic/disabled_client
=== CONT  TestPBKDF2Hasher_HashSecret
=== RUN   TestPBKDF2Hasher_HashSecret/valid_password
=== PAUSE TestPBKDF2Hasher_HashSecret/valid_password
=== RUN   TestPBKDF2Hasher_HashSecret/empty_password
=== PAUSE TestPBKDF2Hasher_HashSecret/empty_password
=== RUN   TestPBKDF2Hasher_HashSecret/unicode_password
=== PAUSE TestPBKDF2Hasher_HashSecret/unicode_password
=== CONT  TestCombinedRevocationChecker_CheckRevocation
--- PASS: TestSecretBasedAuthenticator_MigrateSecrets (0.16s)
=== CONT  TestOCSPRevocationChecker_CheckRevocation_NoOCSPServer
--- PASS: TestCombinedRevocationChecker_CheckRevocation (0.15s)
=== CONT  TestClientSecretJWTValidator_ExtractClaims_AllClaimsPresent
--- PASS: TestClientSecretJWTValidator_ExtractClaims_AllClaimsPresent (0.00s)
=== CONT  TestPostAuthenticator_MethodName
--- PASS: TestPostAuthenticator_MethodName (0.00s)
=== CONT  TestPrivateKeyJWTValidator_ExtractClaims_AllClaimsPresent
--- PASS: TestPrivateKeyJWTValidator_ExtractClaims_AllClaimsPresent (0.00s)
=== CONT  TestPrivateKeyJWTValidator_ValidateJWT_InvalidJWKSet
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_InvalidJWKSet (0.00s)
=== CONT  TestCRLRevocationChecker_CheckRevocation
--- PASS: TestOCSPRevocationChecker_CheckRevocation_Good (0.19s)
=== CONT  TestClientSecretJWTValidator_ValidateJWT_MalformedJWT
--- PASS: TestClientSecretJWTValidator_ValidateJWT_MalformedJWT (0.00s)
=== CONT  TestBasicAuthenticator_ValidateAuthMethod/valid_auth_method
=== CONT  TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt
=== CONT  TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_POST
--- PASS: TestBasicAuthenticator_ValidateAuthMethod (0.00s)
    --- PASS: TestBasicAuthenticator_ValidateAuthMethod/valid_auth_method (0.00s)
    --- PASS: TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt (0.00s)
    --- PASS: TestBasicAuthenticator_ValidateAuthMethod/invalid_auth_method_-_POST (0.00s)
=== CONT  TestPostAuthenticator_ValidateAuthMethod/valid_auth_method
=== CONT  TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt
=== CONT  TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_BASIC
--- PASS: TestPostAuthenticator_ValidateAuthMethod (0.00s)
    --- PASS: TestPostAuthenticator_ValidateAuthMethod/valid_auth_method (0.00s)
    --- PASS: TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_private_key_jwt (0.00s)
    --- PASS: TestPostAuthenticator_ValidateAuthMethod/invalid_auth_method_-_BASIC (0.00s)
=== CONT  TestCompareSecret_InvalidFormat/missing_separator
=== CONT  TestCompareSecret/matching_secret
--- PASS: TestPrivateKeyJWTValidator_ValidateJWT_InvalidSignature (0.31s)
=== CONT  TestCompareSecret_InvalidFormat/empty_hashed
=== CONT  TestCompareSecret_InvalidFormat/invalid_base64_hash
=== CONT  TestCompareSecret_InvalidFormat/invalid_base64_salt
--- PASS: TestCompareSecret_InvalidFormat (0.00s)
    --- PASS: TestCompareSecret_InvalidFormat/missing_separator (0.00s)
    --- PASS: TestCompareSecret_InvalidFormat/empty_hashed (0.00s)
    --- PASS: TestCompareSecret_InvalidFormat/invalid_base64_hash (0.00s)
    --- PASS: TestCompareSecret_InvalidFormat/invalid_base64_salt (0.00s)
=== CONT  TestCompareSecret/empty_secret_does_not_match_non-empty
--- PASS: TestSecretBasedAuthenticator_AuthenticatePost (0.27s)
=== CONT  TestCompareSecret/empty_secret_matches_empty
=== RUN   TestCRLRevocationChecker_CheckRevocation/certificate_is_revoked_but_CRL_signature_fails
=== RUN   TestCRLRevocationChecker_CheckRevocation/certificate_has_no_CRL_distribution_points
--- PASS: TestCRLRevocationChecker_CheckRevocation (0.16s)
    --- PASS: TestCRLRevocationChecker_CheckRevocation/certificate_is_revoked_but_CRL_signature_fails (0.00s)
    --- PASS: TestCRLRevocationChecker_CheckRevocation/certificate_has_no_CRL_distribution_points (0.00s)
=== CONT  TestCompareSecret/non-matching_secret
--- PASS: TestOCSPRevocationChecker_CheckRevocation_NoOCSPServer (0.24s)
=== CONT  TestBasicAuthenticator_Authenticate/valid_basic_auth
=== CONT  TestBasicAuthenticator_Authenticate/client_not_found
=== CONT  TestBasicAuthenticator_Authenticate/invalid_client_secret
=== CONT  TestHashSecret/valid_secret
=== CONT  TestHashSecret/long_secret
=== CONT  TestHashSecret/empty_secret
--- PASS: TestCompareSecret (0.00s)
    --- PASS: TestCompareSecret/empty_secret_does_not_match_non-empty (1.08s)
    --- PASS: TestCompareSecret/matching_secret (1.13s)
    --- PASS: TestCompareSecret/empty_secret_matches_empty (1.15s)
    --- PASS: TestCompareSecret/non-matching_secret (1.14s)
=== CONT  TestPostAuthenticator_Authenticate/valid_post_auth
--- PASS: TestBasicAuthenticator_Authenticate (0.54s)
    --- PASS: TestBasicAuthenticator_Authenticate/valid_basic_auth (0.55s)
    --- PASS: TestBasicAuthenticator_Authenticate/client_not_found (0.00s)
    --- PASS: TestBasicAuthenticator_Authenticate/invalid_client_secret (0.61s)
=== CONT  TestPostAuthenticator_Authenticate/client_not_found
=== CONT  TestPostAuthenticator_Authenticate/invalid_client_secret
=== CONT  TestPBKDF2Hasher_EdgeCases/very_long_password_(10KB)
=== CONT  TestPBKDF2Hasher_EdgeCases/whitespace_only
--- PASS: TestHashSecret (0.00s)
    --- PASS: TestHashSecret/valid_secret (0.59s)
    --- PASS: TestHashSecret/long_secret (0.59s)
    --- PASS: TestHashSecret/empty_secret (0.63s)
=== CONT  TestPBKDF2Hasher_EdgeCases/null_bytes_(valid_UTF-8)
=== CONT  TestPBKDF2Hasher_EdgeCases/special_characters
=== CONT  TestPBKDF2Hasher_CompareSecret/correct_password_matches
=== CONT  TestPBKDF2Hasher_HashSecret/valid_strong_password
--- PASS: TestClientAuthentication_OldSecretExpired (3.83s)
=== CONT  TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_hash_base64)
=== CONT  TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_salt_base64)
=== CONT  TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_iterations)
=== CONT  TestPBKDF2Hasher_CompareSecret/malformed_hash_(wrong_prefix)
=== CONT  TestPBKDF2Hasher_CompareSecret/malformed_hash_(3_parts)
=== CONT  TestPBKDF2Hasher_CompareSecret/case-sensitive_comparison
=== CONT  TestPBKDF2Hasher_CompareSecret/empty_password_does_not_match
--- PASS: TestPostAuthenticator_Authenticate (0.56s)
    --- PASS: TestPostAuthenticator_Authenticate/client_not_found (0.00s)
    --- PASS: TestPostAuthenticator_Authenticate/valid_post_auth (0.58s)
    --- PASS: TestPostAuthenticator_Authenticate/invalid_client_secret (0.63s)
=== CONT  TestPBKDF2Hasher_CompareSecret/incorrect_password_does_not_match
=== CONT  TestPBKDF2Hasher_HashSecret/valid_unicode_password
=== CONT  TestPBKDF2Hasher_HashSecret/valid_long_password_(256_chars)
=== CONT  TestPBKDF2Hasher_HashSecret/valid_empty_password
--- PASS: TestPBKDF2Hasher_EdgeCases (0.00s)
    --- PASS: TestPBKDF2Hasher_EdgeCases/very_long_password_(10KB) (0.20s)
    --- PASS: TestPBKDF2Hasher_EdgeCases/whitespace_only (0.20s)
    --- PASS: TestPBKDF2Hasher_EdgeCases/special_characters (0.20s)
    --- PASS: TestPBKDF2Hasher_EdgeCases/null_bytes_(valid_UTF-8) (0.21s)
=== CONT  TestPBKDF2Hasher_HashSecret/valid_weak_password
=== CONT  TestSelfSignedAuthenticator_Authenticate/valid_self-signed_certificate
=== CONT  TestSelfSignedAuthenticator_Authenticate/subject_mismatch
=== CONT  TestSelfSignedAuthenticator_Authenticate/fingerprint_mismatch
=== CONT  TestSelfSignedAuthenticator_Authenticate/missing_certificate
--- PASS: TestSelfSignedAuthenticator_Authenticate (0.00s)
    --- PASS: TestSelfSignedAuthenticator_Authenticate/valid_self-signed_certificate (0.00s)
    --- PASS: TestSelfSignedAuthenticator_Authenticate/subject_mismatch (0.00s)
    --- PASS: TestSelfSignedAuthenticator_Authenticate/fingerprint_mismatch (0.00s)
    --- PASS: TestSelfSignedAuthenticator_Authenticate/missing_certificate (0.00s)
=== CONT  TestTLSClientAuthenticator_Authenticate_Cert/valid_certificate
=== CONT  TestTLSClientAuthenticator_Authenticate_Cert/subject_mismatch
=== CONT  TestTLSClientAuthenticator_Authenticate_Cert/fingerprint_mismatch
=== CONT  TestTLSClientAuthenticator_Authenticate_Cert/missing_certificate
--- PASS: TestTLSClientAuthenticator_Authenticate_Cert (0.00s)
    --- PASS: TestTLSClientAuthenticator_Authenticate_Cert/valid_certificate (0.00s)
    --- PASS: TestTLSClientAuthenticator_Authenticate_Cert/subject_mismatch (0.00s)
    --- PASS: TestTLSClientAuthenticator_Authenticate_Cert/fingerprint_mismatch (0.00s)
    --- PASS: TestTLSClientAuthenticator_Authenticate_Cert/missing_certificate (0.00s)
=== CONT  TestPBKDF2Hasher_CompareSecret/matching_password
--- PASS: TestPBKDF2Hasher_CompareSecret (0.09s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_hash_base64) (0.00s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_salt_base64) (0.00s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/malformed_hash_(invalid_iterations) (0.00s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/malformed_hash_(wrong_prefix) (0.00s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/malformed_hash_(3_parts) (0.00s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/correct_password_matches (0.10s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/case-sensitive_comparison (0.10s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/empty_password_does_not_match (0.10s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/incorrect_password_does_not_match (0.11s)
=== CONT  TestPBKDF2Hasher_CompareSecret/invalid_hash_format
=== CONT  TestPBKDF2Hasher_CompareSecret/empty_plaintext
=== CONT  TestPBKDF2Hasher_CompareSecret/wrong_password
=== CONT  TestMigrateClientSecrets/migrate_plaintext_secrets
=== CONT  TestMigrateClientSecrets/skip_public_clients
=== CONT  TestSecretBasedAuthenticator_AuthenticateBasic/valid_credentials
--- PASS: TestPBKDF2Hasher_HashSecret (0.00s)
    --- PASS: TestPBKDF2Hasher_HashSecret/valid_strong_password (0.10s)
    --- PASS: TestPBKDF2Hasher_HashSecret/valid_unicode_password (0.10s)
    --- PASS: TestPBKDF2Hasher_HashSecret/valid_empty_password (0.10s)
    --- PASS: TestPBKDF2Hasher_HashSecret/valid_long_password_(256_chars) (0.10s)
    --- PASS: TestPBKDF2Hasher_HashSecret/valid_weak_password (0.10s)
=== CONT  TestPBKDF2Hasher_HashSecret/valid_password
=== CONT  TestSecretBasedAuthenticator_AuthenticateBasic/disabled_client
=== CONT  TestSecretBasedAuthenticator_AuthenticateBasic/wrong_secret
=== CONT  TestPBKDF2Hasher_HashSecret/unicode_password
--- PASS: TestPBKDF2Hasher_CompareSecret (0.08s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/invalid_hash_format (0.00s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/matching_password (0.10s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/empty_plaintext (0.10s)
    --- PASS: TestPBKDF2Hasher_CompareSecret/wrong_password (0.10s)
=== CONT  TestPBKDF2Hasher_HashSecret/empty_password
--- PASS: TestMigrateClientSecrets (0.00s)
    --- PASS: TestMigrateClientSecrets/skip_public_clients (0.00s)
    --- PASS: TestMigrateClientSecrets/migrate_plaintext_secrets (0.10s)
--- PASS: TestSecretBasedAuthenticator_AuthenticateBasic (0.09s)
    --- PASS: TestSecretBasedAuthenticator_AuthenticateBasic/disabled_client (0.00s)
    --- PASS: TestSecretBasedAuthenticator_AuthenticateBasic/wrong_secret (0.10s)
    --- PASS: TestSecretBasedAuthenticator_AuthenticateBasic/valid_credentials (0.11s)
--- PASS: TestPBKDF2Hasher_HashSecret (0.00s)
    --- PASS: TestPBKDF2Hasher_HashSecret/valid_password (0.10s)
    --- PASS: TestPBKDF2Hasher_HashSecret/unicode_password (0.10s)
    --- PASS: TestPBKDF2Hasher_HashSecret/empty_password (0.09s)
--- PASS: TestCompareSecret_ConstantTime (5.42s)
--- PASS: TestPBKDF2Hasher_SaltRandomness (6.74s)
PASS
ok   cryptoutil/internal/identity/authz/clientauth 8.334s
