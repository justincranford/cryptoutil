time=2025-12-08T06:55:52.255-05:00 level=DEBUG msg="initialized otel logs provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.256-05:00 level=INFO msg="sidecar health check succeeded" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 attempts=0 errors=<nil>
time=2025-12-08T06:55:52.256-05:00 level=DEBUG msg="initialized metrics provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.256-05:00 level=DEBUG msg="initialized traces provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.277-05:00 level=INFO msg="Starting JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 address=127.0.0.1 port=59603

 ΓöîΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÉ
 Γöé               JOSE Authority Server               Γöé
 Γöé                  Fiber v2.52.10                   Γöé
 Γöé              <http://127.0.0.1:59603>               Γöé
 Γöé                                                   Γöé
 Γöé Handlers ............ 22  Processes ........... 1 Γöé
 Γöé Prefork ....... Disabled  PID ............. 17392 Γöé
 ΓööΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÿ

=== RUN   TestKeyStoreStore
=== PAUSE TestKeyStoreStore
=== RUN   TestKeyStoreDuplicateKey
=== PAUSE TestKeyStoreDuplicateKey
=== RUN   TestKeyStoreGet
=== PAUSE TestKeyStoreGet
=== RUN   TestKeyStoreDelete
=== PAUSE TestKeyStoreDelete
=== RUN   TestKeyStoreList
=== PAUSE TestKeyStoreList
=== RUN   TestKeyStoreCount
=== PAUSE TestKeyStoreCount
=== RUN   TestKeyStoreGetJWKS
=== PAUSE TestKeyStoreGetJWKS
=== RUN   TestHealthEndpoints
=== PAUSE TestHealthEndpoints
=== RUN   TestHealthJSON
=== PAUSE TestHealthJSON
=== RUN   TestJWKGenerateAndRetrieve
=== PAUSE TestJWKGenerateAndRetrieve
=== RUN   TestJWKGenerateInvalidAlgorithm
=== PAUSE TestJWKGenerateInvalidAlgorithm
=== RUN   TestJWKGetNotFound
=== PAUSE TestJWKGetNotFound
=== RUN   TestJWKDeleteNotFound
=== PAUSE TestJWKDeleteNotFound
=== RUN   TestJWKList
=== PAUSE TestJWKList
=== RUN   TestJWSSignAndVerify
=== PAUSE TestJWSSignAndVerify
=== RUN   TestJWEEncryptAndDecrypt
=== PAUSE TestJWEEncryptAndDecrypt
=== RUN   TestJWTCreateAndVerify
=== PAUSE TestJWTCreateAndVerify
=== RUN   TestWellKnownJWKS
=== PAUSE TestWellKnownJWKS
=== RUN   TestJWSSignMissingKID
=== PAUSE TestJWSSignMissingKID
=== RUN   TestJWSSignMissingPayload
=== PAUSE TestJWSSignMissingPayload
=== RUN   TestJWSSignKeyNotFound
=== PAUSE TestJWSSignKeyNotFound
=== RUN   TestJWSVerifyMissingJWS
=== PAUSE TestJWSVerifyMissingJWS
=== RUN   TestJWSVerifyKeyNotFound
=== PAUSE TestJWSVerifyKeyNotFound
=== RUN   TestJWEEncryptMissingKID
=== PAUSE TestJWEEncryptMissingKID
=== RUN   TestJWEEncryptMissingPlaintext
=== PAUSE TestJWEEncryptMissingPlaintext
=== RUN   TestJWEEncryptKeyNotFound
=== PAUSE TestJWEEncryptKeyNotFound
=== RUN   TestJWEDecryptMissingJWE
=== PAUSE TestJWEDecryptMissingJWE
=== RUN   TestJWEDecryptMissingKID
=== PAUSE TestJWEDecryptMissingKID
=== RUN   TestJWEDecryptKeyNotFound
=== PAUSE TestJWEDecryptKeyNotFound
=== RUN   TestJWTCreateMissingKID
=== PAUSE TestJWTCreateMissingKID
=== RUN   TestJWTCreateMissingClaims
=== PAUSE TestJWTCreateMissingClaims
=== RUN   TestJWTCreateKeyNotFound
=== PAUSE TestJWTCreateKeyNotFound
=== RUN   TestJWTVerifyMissingJWT
=== PAUSE TestJWTVerifyMissingJWT
=== RUN   TestJWTVerifyKeyNotFound
=== PAUSE TestJWTVerifyKeyNotFound
=== RUN   TestJWKGetMissingKID
=== PAUSE TestJWKGetMissingKID
=== RUN   TestJWKDeleteSuccess
=== PAUSE TestJWKDeleteSuccess
=== RUN   TestInvalidJSONBody
=== PAUSE TestInvalidJSONBody
=== RUN   TestJWSVerifyErrorPaths
=== PAUSE TestJWSVerifyErrorPaths
=== RUN   TestJWTVerifyErrorPaths
=== PAUSE TestJWTVerifyErrorPaths
=== RUN   TestServerLifecycle
=== PAUSE TestServerLifecycle
=== RUN   TestAPIKeyMiddleware
=== PAUSE TestAPIKeyMiddleware
=== RUN   TestNewServerErrorPaths
=== PAUSE TestNewServerErrorPaths
=== RUN   TestStartBlocking
=== PAUSE TestStartBlocking
=== RUN   TestShutdownCoverage
=== PAUSE TestShutdownCoverage
=== CONT  TestKeyStoreStore
=== RUN   TestKeyStoreStore/valid_key
=== PAUSE TestKeyStoreStore/valid_key
=== RUN   TestKeyStoreStore/nil_key
=== PAUSE TestKeyStoreStore/nil_key
=== CONT  TestKeyStoreStore/valid_key
=== CONT  TestShutdownCoverage
=== RUN   TestShutdownCoverage/NormalShutdown
=== PAUSE TestShutdownCoverage/NormalShutdown
=== RUN   TestShutdownCoverage/ShutdownWithoutStart
=== PAUSE TestShutdownCoverage/ShutdownWithoutStart
=== CONT  TestShutdownCoverage/NormalShutdown
time=2025-12-08T06:55:52.567-05:00 level=DEBUG msg="initialized otel logs provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.567-05:00 level=INFO msg="sidecar health check succeeded" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 attempts=0 errors=<nil>
time=2025-12-08T06:55:52.567-05:00 level=DEBUG msg="initialized metrics provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.568-05:00 level=DEBUG msg="initialized traces provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.569-05:00 level=INFO msg="Starting JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 address=127.0.0.1 port=59605

 ΓöîΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÉ
 Γöé               JOSE Authority Server               Γöé
 Γöé                  Fiber v2.52.10                   Γöé
 Γöé              <http://127.0.0.1:59605>               Γöé
 Γöé                                                   Γöé
 Γöé Handlers ............ 22  Processes ........... 1 Γöé
 Γöé Prefork ....... Disabled  PID ............. 17392 Γöé
 ΓööΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÿ

=== CONT  TestStartBlocking
time=2025-12-08T06:55:52.572-05:00 level=DEBUG msg="initialized otel logs provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.595-05:00 level=INFO msg="sidecar health check succeeded" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 attempts=0 errors=<nil>
time=2025-12-08T06:55:52.595-05:00 level=DEBUG msg="initialized metrics provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.595-05:00 level=DEBUG msg="initialized traces provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.595-05:00 level=INFO msg="Starting JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 address=127.0.0.1 port=59606

 ΓöîΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÉ
 Γöé               JOSE Authority Server               Γöé
 Γöé                  Fiber v2.52.10                   Γöé
 Γöé              <http://127.0.0.1:59606>               Γöé
 Γöé                                                   Γöé
 Γöé Handlers ............ 22  Processes ........... 1 Γöé
 Γöé Prefork ....... Disabled  PID ............. 17392 Γöé
 ΓööΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÿ

=== CONT  TestNewServerErrorPaths
=== RUN   TestNewServerErrorPaths/NilContext
=== PAUSE TestNewServerErrorPaths/NilContext
=== RUN   TestNewServerErrorPaths/NilSettings
=== PAUSE TestNewServerErrorPaths/NilSettings
time=2025-12-08T06:55:52.694-05:00 level=INFO msg="Shutting down JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.719-05:00 level=INFO msg="Context cancelled, shutting down server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
--- PASS: TestStartBlocking (0.15s)
=== CONT  TestJWKGetMissingKID
time=2025-12-08T06:55:52.806-05:00 level=DEBUG msg="stopping JWKGenService" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:52.807-05:00 level=DEBUG msg="telemetry providers force flushed" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=0.239127 flush=0.000766
time=2025-12-08T06:55:52.838-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:55:52.838-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:55:52.838-05:00 level=DEBUG msg="telemetry providers shut down" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=0.2706432 flush=0.0315162
=== CONT  TestJWTVerifyKeyNotFound
time=2025-12-08T06:55:52.838-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:55:52.838-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:55:52.947-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:52.947-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:52.951-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:52.954-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:52.954-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:52.954-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:52.955-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:52.977-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:52.978-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:53.094-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:55:53.094-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
=== CONT  TestAPIKeyMiddleware
time=2025-12-08T06:55:53.119-05:00 level=DEBUG msg="initialized otel logs provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:53.119-05:00 level=INFO msg="sidecar health check succeeded" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 attempts=0 errors=<nil>
time=2025-12-08T06:55:53.119-05:00 level=DEBUG msg="initialized metrics provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:53.119-05:00 level=DEBUG msg="initialized traces provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
--- PASS: TestAPIKeyMiddleware (0.00s)
=== CONT  TestJWTVerifyMissingJWT
=== CONT  TestJWKDeleteSuccess
--- PASS: TestJWKGetMissingKID (0.43s)
=== CONT  TestServerLifecycle
=== CONT  TestJWTCreateKeyNotFound
time=2025-12-08T06:55:53.170-05:00 level=DEBUG msg="initialized otel logs provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:53.170-05:00 level=INFO msg="sidecar health check succeeded" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 attempts=0 errors=<nil>
time=2025-12-08T06:55:53.170-05:00 level=DEBUG msg="initialized metrics provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:53.170-05:00 level=DEBUG msg="initialized traces provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:53.171-05:00 level=INFO msg="Starting JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 address=127.0.0.1 port=59612
time=2025-12-08T06:55:53.171-05:00 level=INFO msg="Shutting down JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:53.171-05:00 level=DEBUG msg="stopping JWKGenService" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
=== CONT  TestJWTVerifyErrorPaths
=== CONT  TestJWSVerifyErrorPaths
=== CONT  TestInvalidJSONBody
time=2025-12-08T06:55:53.178-05:00 level=DEBUG msg="telemetry providers force flushed" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=0.0095733 flush=0.0075103

 ΓöîΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÉ
 Γöé               JOSE Authority Server               Γöé
 Γöé                  Fiber v2.52.10                   Γöé
 Γöé              <http://127.0.0.1:59612>               Γöé
 Γöé                                                   Γöé
 Γöé Handlers ............ 22  Processes ........... 1 Γöé
 Γöé Prefork ....... Disabled  PID ............. 17392 Γöé
 ΓööΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÿ

--- PASS: TestJWTVerifyKeyNotFound (0.47s)
=== CONT  TestJWTCreateMissingClaims
=== RUN   TestJWTVerifyErrorPaths/MissingJWT
=== PAUSE TestJWTVerifyErrorPaths/MissingJWT
=== RUN   TestJWTVerifyErrorPaths/KeyNotFound
=== PAUSE TestJWTVerifyErrorPaths/KeyNotFound
=== RUN   TestJWTVerifyErrorPaths/InvalidJWTFormat
=== PAUSE TestJWTVerifyErrorPaths/InvalidJWTFormat
=== RUN   TestJWTVerifyErrorPaths/VerifyWithoutKID
=== PAUSE TestJWTVerifyErrorPaths/VerifyWithoutKID
=== RUN   TestJWSVerifyErrorPaths/MissingJWS
=== PAUSE TestJWSVerifyErrorPaths/MissingJWS
=== RUN   TestJWSVerifyErrorPaths/KeyNotFound
=== PAUSE TestJWSVerifyErrorPaths/KeyNotFound
=== RUN   TestJWSVerifyErrorPaths/InvalidSignature
=== PAUSE TestJWSVerifyErrorPaths/InvalidSignature
=== RUN   TestJWSVerifyErrorPaths/VerifyWithoutKID
=== PAUSE TestJWSVerifyErrorPaths/VerifyWithoutKID
=== CONT  TestJWTCreateMissingKID
=== RUN   TestInvalidJSONBody/generate
=== PAUSE TestInvalidJSONBody/generate
=== RUN   TestInvalidJSONBody/sign
=== PAUSE TestInvalidJSONBody/sign
=== RUN   TestInvalidJSONBody/verify
=== PAUSE TestInvalidJSONBody/verify
=== RUN   TestInvalidJSONBody/encrypt
=== PAUSE TestInvalidJSONBody/encrypt
=== RUN   TestInvalidJSONBody/decrypt
=== PAUSE TestInvalidJSONBody/decrypt
=== RUN   TestInvalidJSONBody/jwtCreate
=== PAUSE TestInvalidJSONBody/jwtCreate
=== RUN   TestInvalidJSONBody/jwtVerify
=== PAUSE TestInvalidJSONBody/jwtVerify
=== CONT  TestJWEDecryptKeyNotFound
=== CONT  TestJWEDecryptMissingKID
time=2025-12-08T06:55:53.556-05:00 level=DEBUG msg="telemetry providers shut down" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=0.3877431 flush=0.3781698
--- PASS: TestServerLifecycle (0.39s)
=== CONT  TestJWEDecryptMissingJWE
time=2025-12-08T06:55:53.556-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:55:53.556-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:55:53.590-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:55:53.590-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:55:53.652-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:55:53.652-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:55:53.686-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:53.686-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:53.689-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:53.698-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:53.699-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:53.699-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:53.699-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:53.699-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:53.699-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:53.699-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:53.753-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:53.754-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:53.754-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:53.754-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:54.023-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-727d-9629-568e98d8438d algorithm=EC/P256
--- PASS: TestJWTVerifyMissingJWT (1.04s)
=== CONT  TestJWEEncryptKeyNotFound
--- PASS: TestJWTCreateKeyNotFound (1.09s)
=== CONT  TestJWEEncryptMissingPlaintext
time=2025-12-08T06:55:54.256-05:00 level=INFO msg="Deleted JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-727d-9629-568e98d8438d
--- PASS: TestJWTCreateMissingKID (0.98s)
=== CONT  TestJWEEncryptMissingKID
--- PASS: TestJWEDecryptKeyNotFound (1.13s)
=== CONT  TestJWSVerifyKeyNotFound
--- PASS: TestJWTCreateMissingClaims (1.41s)
=== CONT  TestJWSVerifyMissingJWS
--- PASS: TestJWEDecryptMissingKID (1.34s)
=== CONT  TestJWSSignKeyNotFound
--- PASS: TestJWEDecryptMissingJWE (1.46s)
=== CONT  TestJWSSignMissingPayload
--- PASS: TestJWEEncryptMissingKID (0.51s)
=== CONT  TestJWSSignMissingKID
--- PASS: TestJWEEncryptKeyNotFound (1.19s)
=== CONT  TestWellKnownJWKS
time=2025-12-08T06:55:55.352-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
--- PASS: TestJWEEncryptMissingPlaintext (1.15s)
=== CONT  TestJWTCreateAndVerify
--- PASS: TestJWKDeleteSuccess (2.77s)
=== CONT  TestJWEEncryptAndDecrypt
=== RUN   TestJWEEncryptAndDecrypt/Oct512
=== PAUSE TestJWEEncryptAndDecrypt/Oct512
=== RUN   TestJWEEncryptAndDecrypt/Oct384
=== PAUSE TestJWEEncryptAndDecrypt/Oct384
=== RUN   TestJWEEncryptAndDecrypt/Oct256
=== PAUSE TestJWEEncryptAndDecrypt/Oct256
=== RUN   TestJWEEncryptAndDecrypt/Oct192
=== PAUSE TestJWEEncryptAndDecrypt/Oct192
=== RUN   TestJWEEncryptAndDecrypt/Oct128
=== PAUSE TestJWEEncryptAndDecrypt/Oct128
=== RUN   TestJWEEncryptAndDecrypt/RSA4096
=== PAUSE TestJWEEncryptAndDecrypt/RSA4096
=== RUN   TestJWEEncryptAndDecrypt/RSA3072
=== PAUSE TestJWEEncryptAndDecrypt/RSA3072
=== RUN   TestJWEEncryptAndDecrypt/RSA2048
=== PAUSE TestJWEEncryptAndDecrypt/RSA2048
=== RUN   TestJWEEncryptAndDecrypt/ECP521
=== PAUSE TestJWEEncryptAndDecrypt/ECP521
=== RUN   TestJWEEncryptAndDecrypt/ECP384
=== PAUSE TestJWEEncryptAndDecrypt/ECP384
=== RUN   TestJWEEncryptAndDecrypt/ECP256
=== PAUSE TestJWEEncryptAndDecrypt/ECP256
=== CONT  TestJWSSignAndVerify
time=2025-12-08T06:55:55.934-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
--- PASS: TestJWSVerifyKeyNotFound (1.49s)
=== CONT  TestJWKList
--- PASS: TestJWSSignMissingPayload (1.17s)
=== CONT  TestJWKDeleteNotFound
--- PASS: TestJWSVerifyMissingJWS (1.49s)
=== CONT  TestJWKGetNotFound
--- PASS: TestJWSSignKeyNotFound (1.32s)
=== CONT  TestJWKGenerateInvalidAlgorithm
--- PASS: TestWellKnownJWKS (1.14s)
=== CONT  TestJWKGenerateAndRetrieve
=== RUN   TestJWKGenerateAndRetrieve/RSA4096
=== PAUSE TestJWKGenerateAndRetrieve/RSA4096
=== RUN   TestJWKGenerateAndRetrieve/RSA3072
=== PAUSE TestJWKGenerateAndRetrieve/RSA3072
=== RUN   TestJWKGenerateAndRetrieve/RSA2048
=== PAUSE TestJWKGenerateAndRetrieve/RSA2048
=== RUN   TestJWKGenerateAndRetrieve/ECP256
=== PAUSE TestJWKGenerateAndRetrieve/ECP256
=== RUN   TestJWKGenerateAndRetrieve/ECP384
=== PAUSE TestJWKGenerateAndRetrieve/ECP384
=== RUN   TestJWKGenerateAndRetrieve/ECP521
=== PAUSE TestJWKGenerateAndRetrieve/ECP521
=== RUN   TestJWKGenerateAndRetrieve/OKPEd25519
=== PAUSE TestJWKGenerateAndRetrieve/OKPEd25519
=== RUN   TestJWKGenerateAndRetrieve/Oct512
=== PAUSE TestJWKGenerateAndRetrieve/Oct512
=== RUN   TestJWKGenerateAndRetrieve/Oct384
=== PAUSE TestJWKGenerateAndRetrieve/Oct384
=== RUN   TestJWKGenerateAndRetrieve/Oct256
=== PAUSE TestJWKGenerateAndRetrieve/Oct256
=== RUN   TestJWKGenerateAndRetrieve/Oct192
=== PAUSE TestJWKGenerateAndRetrieve/Oct192
=== RUN   TestJWKGenerateAndRetrieve/Oct128
=== PAUSE TestJWKGenerateAndRetrieve/Oct128
=== CONT  TestHealthJSON
time=2025-12-08T06:55:56.687-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-727e-af24-fd690af3207c algorithm=EC/P384
--- PASS: TestJWSSignMissingKID (1.64s)
=== CONT  TestHealthEndpoints
=== RUN   TestHealthEndpoints/livez
=== PAUSE TestHealthEndpoints/livez
=== RUN   TestHealthEndpoints/readyz
=== PAUSE TestHealthEndpoints/readyz
=== CONT  TestKeyStoreGetJWKS
--- PASS: TestKeyStoreGetJWKS (0.00s)
=== CONT  TestKeyStoreCount
--- PASS: TestKeyStoreCount (0.00s)
=== CONT  TestKeyStoreList
--- PASS: TestKeyStoreList (0.00s)
=== CONT  TestKeyStoreDelete
--- PASS: TestKeyStoreDelete (0.00s)
=== CONT  TestKeyStoreGet
=== RUN   TestKeyStoreGet/existing_key
=== PAUSE TestKeyStoreGet/existing_key
=== RUN   TestKeyStoreGet/non-existing_key
=== PAUSE TestKeyStoreGet/non-existing_key
=== RUN   TestKeyStoreGet/invalid_uuid
=== PAUSE TestKeyStoreGet/invalid_uuid
=== CONT  TestKeyStoreDuplicateKey
--- PASS: TestKeyStoreDuplicateKey (0.00s)
=== CONT  TestKeyStoreStore/nil_key
--- PASS: TestKeyStoreStore (0.00s)
    --- PASS: TestKeyStoreStore/valid_key (0.00s)
    --- PASS: TestKeyStoreStore/nil_key (0.00s)
=== CONT  TestShutdownCoverage/ShutdownWithoutStart
time=2025-12-08T06:55:56.689-05:00 level=DEBUG msg="initialized otel logs provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:56.689-05:00 level=INFO msg="sidecar health check succeeded" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 attempts=0 errors=<nil>
time=2025-12-08T06:55:56.689-05:00 level=DEBUG msg="initialized metrics provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:56.689-05:00 level=DEBUG msg="initialized traces provider" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:56.784-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-727f-8e57-aec69ee0b431 algorithm=EC/P256
--- PASS: TestJWKGenerateInvalidAlgorithm (0.40s)
=== CONT  TestNewServerErrorPaths/NilContext
=== CONT  TestNewServerErrorPaths/NilSettings
--- PASS: TestNewServerErrorPaths (0.00s)
    --- PASS: TestNewServerErrorPaths/NilContext (0.00s)
    --- PASS: TestNewServerErrorPaths/NilSettings (0.00s)
=== CONT  TestJWTVerifyErrorPaths/MissingJWT
time=2025-12-08T06:55:56.839-05:00 level=INFO msg="Shutting down JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:56.839-05:00 level=DEBUG msg="stopping JWKGenService" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:55:56.841-05:00 level=DEBUG msg="telemetry providers force flushed" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=0.1528198 flush=0.0010009
time=2025-12-08T06:55:56.978-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:55:56.978-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:55:57.043-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:55:57.043-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:55:57.129-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:57.129-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:57.130-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:55:57.130-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:57.130-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:57.131-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:55:57.131-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:57.131-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:57.131-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:55:57.131-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:57.131-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:57.139-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:55:57.139-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:57.139-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:55:57.140-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:57.141-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:57.180-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:55:57.181-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:57.181-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:57.181-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:55:57.181-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:55:57.181-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:55:57.181-05:00 level=DEBUG msg="telemetry providers shut down" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=0.492503 flush=0.3396832
--- PASS: TestShutdownCoverage (0.00s)
    --- PASS: TestShutdownCoverage/NormalShutdown (0.27s)
    --- PASS: TestShutdownCoverage/ShutdownWithoutStart (0.49s)
=== CONT  TestJWSVerifyErrorPaths/MissingJWS
--- PASS: TestJWKList (1.02s)
=== CONT  TestJWTVerifyErrorPaths/VerifyWithoutKID
--- PASS: TestJWKDeleteNotFound (1.13s)
=== CONT  TestJWTVerifyErrorPaths/InvalidJWTFormat
--- PASS: TestJWTCreateAndVerify (1.94s)
=== CONT  TestJWTVerifyErrorPaths/KeyNotFound
time=2025-12-08T06:55:57.458-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7280-85ce-e4a3a1bdd0fa algorithm=EC/P384
--- PASS: TestJWKGetNotFound (1.28s)
=== CONT  TestInvalidJSONBody/generate
=== CONT  TestJWSVerifyErrorPaths/VerifyWithoutKID
--- PASS: TestHealthJSON (1.18s)
=== CONT  TestJWSVerifyErrorPaths/InvalidSignature
=== CONT  TestJWSVerifyErrorPaths/KeyNotFound
time=2025-12-08T06:55:57.873-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:55:58.058-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7281-a84e-12ad9f5e3c17 algorithm=EC/P256
time=2025-12-08T06:55:58.059-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7282-8e04-c9c7d8669513 algorithm=RSA/2048
=== CONT  TestInvalidJSONBody/jwtVerify
time=2025-12-08T06:55:58.124-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7283-a29f-1d07c82cc43b algorithm=oct/256
--- PASS: TestJWSSignAndVerify (2.63s)
=== CONT  TestInvalidJSONBody/jwtCreate
=== CONT  TestInvalidJSONBody/decrypt
=== CONT  TestInvalidJSONBody/encrypt
=== CONT  TestInvalidJSONBody/verify
=== CONT  TestInvalidJSONBody/sign
--- PASS: TestJWTVerifyErrorPaths (0.00s)
    --- PASS: TestJWTVerifyErrorPaths/KeyNotFound (0.35s)
    --- PASS: TestJWTVerifyErrorPaths/MissingJWT (1.31s)
    --- PASS: TestJWTVerifyErrorPaths/InvalidJWTFormat (1.32s)
    --- PASS: TestJWTVerifyErrorPaths/VerifyWithoutKID (1.66s)
=== CONT  TestJWEEncryptAndDecrypt/Oct512
time=2025-12-08T06:55:58.899-05:00 level=DEBUG msg="closing channels" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
=== CONT  TestJWEEncryptAndDecrypt/ECP256
=== CONT  TestJWEEncryptAndDecrypt/ECP384
=== CONT  TestJWEEncryptAndDecrypt/ECP521
=== CONT  TestJWEEncryptAndDecrypt/RSA2048
=== CONT  TestJWEEncryptAndDecrypt/RSA3072
time=2025-12-08T06:55:59.453-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7284-af6f-fb49105ab2ae algorithm=EC/P521
time=2025-12-08T06:55:59.538-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7285-8281-960d7211a3a4 algorithm=oct/512
--- PASS: TestInvalidJSONBody (0.00s)
    --- PASS: TestInvalidJSONBody/jwtVerify (0.43s)
    --- PASS: TestInvalidJSONBody/generate (1.15s)
    --- PASS: TestInvalidJSONBody/decrypt (0.59s)
    --- PASS: TestInvalidJSONBody/verify (0.48s)
    --- PASS: TestInvalidJSONBody/encrypt (0.50s)
    --- PASS: TestInvalidJSONBody/sign (0.73s)
    --- PASS: TestInvalidJSONBody/jwtCreate (1.02s)
=== CONT  TestJWEEncryptAndDecrypt/RSA4096
time=2025-12-08T06:55:59.635-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7286-bfdd-6b38b387357c algorithm=EC/P256
time=2025-12-08T06:55:59.736-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7287-8a6c-f2ae756780c8 algorithm=EC/P384
--- PASS: TestJWSVerifyErrorPaths (0.00s)
    --- PASS: TestJWSVerifyErrorPaths/MissingJWS (0.41s)
    --- PASS: TestJWSVerifyErrorPaths/InvalidSignature (0.93s)
    --- PASS: TestJWSVerifyErrorPaths/KeyNotFound (1.12s)
    --- PASS: TestJWSVerifyErrorPaths/VerifyWithoutKID (2.39s)
time=2025-12-08T06:55:59.976-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7288-9309-5729681ce992 algorithm=RSA/2048
=== CONT  TestJWEEncryptAndDecrypt/Oct192
time=2025-12-08T06:56:00.168-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7289-8c6c-49218de22c39 algorithm=RSA/3072
=== CONT  TestJWEEncryptAndDecrypt/Oct128
time=2025-12-08T06:56:00.261-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-728a-bee6-56d235e991e0 algorithm=RSA/4096
time=2025-12-08T06:56:00.309-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-728b-9465-b9b8ccc4a74f algorithm=oct/192
time=2025-12-08T06:56:00.461-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-728c-bc8a-2d82f65d7bfe algorithm=oct/128
=== CONT  TestJWEEncryptAndDecrypt/Oct256
=== CONT  TestJWEEncryptAndDecrypt/Oct384
=== CONT  TestJWKGenerateAndRetrieve/RSA4096
=== CONT  TestJWKGenerateAndRetrieve/Oct128
time=2025-12-08T06:56:00.730-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-728d-a9b9-a5bdf4fbd31d algorithm=oct/256
time=2025-12-08T06:56:00.732-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-728e-b69f-c570c3ea34f1 algorithm=oct/384
time=2025-12-08T06:56:00.791-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7290-9940-26eaecdb59b3 algorithm=oct/128
time=2025-12-08T06:56:00.845-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-728f-a03f-a91f8d364324 algorithm=RSA/4096
=== CONT  TestJWKGenerateAndRetrieve/Oct192
=== CONT  TestJWKGenerateAndRetrieve/Oct256
time=2025-12-08T06:56:00.994-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e53-7291-94f6-b0e4f0dfd9c8 algorithm=oct/192
time=2025-12-08T06:56:01.089-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-6e55-7941-b4d2-3278d9eb3ae6 algorithm=oct/256
=== CONT  TestJWKGenerateAndRetrieve/Oct384
=== CONT  TestJWKGenerateAndRetrieve/Oct512
=== CONT  TestJWKGenerateAndRetrieve/OKPEd25519
--- PASS: TestJWEEncryptAndDecrypt (0.00s)
    --- PASS: TestJWEEncryptAndDecrypt/ECP521 (0.97s)
    --- PASS: TestJWEEncryptAndDecrypt/ECP384 (1.39s)
    --- PASS: TestJWEEncryptAndDecrypt/ECP256 (1.72s)
    --- PASS: TestJWEEncryptAndDecrypt/Oct512 (1.78s)
    --- PASS: TestJWEEncryptAndDecrypt/RSA2048 (1.62s)
    --- PASS: TestJWEEncryptAndDecrypt/RSA4096 (1.34s)
    --- PASS: TestJWEEncryptAndDecrypt/RSA3072 (1.56s)
    --- PASS: TestJWEEncryptAndDecrypt/Oct192 (1.01s)
    --- PASS: TestJWEEncryptAndDecrypt/Oct128 (0.96s)
    --- PASS: TestJWEEncryptAndDecrypt/Oct384 (0.58s)
    --- PASS: TestJWEEncryptAndDecrypt/Oct256 (0.73s)
time=2025-12-08T06:56:01.224-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-7f90-7bd7-b565-7b879ebf937a algorithm=oct/512
=== CONT  TestJWKGenerateAndRetrieve/ECP521
=== CONT  TestJWKGenerateAndRetrieve/ECP384
=== CONT  TestJWKGenerateAndRetrieve/ECP256
time=2025-12-08T06:56:01.226-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-75ae-7cd2-94d0-2e5698ddf6c3 algorithm=oct/384
time=2025-12-08T06:56:01.341-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-7ff0-7a51-a80a-6f9f2b0006bd algorithm=OKP/Ed25519
time=2025-12-08T06:56:01.343-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-82ab-7ca4-bd1d-6a2786dc221d algorithm=EC/P384
time=2025-12-08T06:56:01.344-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-850d-7a45-9b5c-e602e5ca05d6 algorithm=EC/P521
=== CONT  TestJWKGenerateAndRetrieve/RSA2048
time=2025-12-08T06:56:01.514-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-850d-7a46-af30-2e863fe34bd7 algorithm=RSA/2048
=== CONT  TestJWKGenerateAndRetrieve/RSA3072
=== CONT  TestHealthEndpoints/livez
=== CONT  TestKeyStoreGet/existing_key
=== CONT  TestHealthEndpoints/readyz
time=2025-12-08T06:56:01.615-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-8565-790f-bf27-61e76cee054a algorithm=EC/P256
=== CONT  TestKeyStoreGet/invalid_uuid
=== CONT  TestKeyStoreGet/non-existing_key
--- PASS: TestKeyStoreGet (0.00s)
    --- PASS: TestKeyStoreGet/existing_key (0.00s)
    --- PASS: TestKeyStoreGet/invalid_uuid (0.00s)
    --- PASS: TestKeyStoreGet/non-existing_key (0.00s)
time=2025-12-08T06:56:01.697-05:00 level=INFO msg="Generated JWK" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 kid=019afdd1-8aa0-75d0-a957-22897d2027f1 algorithm=RSA/3072
--- PASS: TestHealthEndpoints (0.00s)
    --- PASS: TestHealthEndpoints/livez (0.20s)
    --- PASS: TestHealthEndpoints/readyz (0.20s)
--- PASS: TestJWKGenerateAndRetrieve (0.00s)
    --- PASS: TestJWKGenerateAndRetrieve/RSA4096 (0.41s)
    --- PASS: TestJWKGenerateAndRetrieve/Oct128 (0.50s)
    --- PASS: TestJWKGenerateAndRetrieve/Oct192 (0.47s)
    --- PASS: TestJWKGenerateAndRetrieve/Oct256 (0.69s)
    --- PASS: TestJWKGenerateAndRetrieve/Oct384 (0.53s)
    --- PASS: TestJWKGenerateAndRetrieve/Oct512 (0.49s)
    --- PASS: TestJWKGenerateAndRetrieve/OKPEd25519 (0.57s)
    --- PASS: TestJWKGenerateAndRetrieve/ECP521 (0.50s)
    --- PASS: TestJWKGenerateAndRetrieve/RSA2048 (0.46s)
    --- PASS: TestJWKGenerateAndRetrieve/ECP256 (0.62s)
    --- PASS: TestJWKGenerateAndRetrieve/ECP384 (0.71s)
    --- PASS: TestJWKGenerateAndRetrieve/RSA3072 (0.46s)
PASS
time=2025-12-08T06:56:02.072-05:00 level=INFO msg="Shutting down JOSE Authority Server" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:56:02.601-05:00 level=DEBUG msg="stopping JWKGenService" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74
time=2025-12-08T06:56:02.602-05:00 level=DEBUG msg="telemetry providers force flushed" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=10.3477901 flush=0.0006282
time=2025-12-08T06:56:02.650-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:56:02.650-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 4096"
time=2025-12-08T06:56:02.650-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:56:02.650-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 3072"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService RSA 2048"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P521"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDSA-P256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P521"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService ECDH-P256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService Ed25519"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-GCM"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-GCM"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-GCM"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-256-CBC HS-512"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-192-CBC HS-384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService AES-128-CBC HS-256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-512"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-384"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService HMAC-256"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg=canceled deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="waiting for workers" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 pool="JWKGenService UUIDv7"
time=2025-12-08T06:56:02.679-05:00 level=DEBUG msg="telemetry providers shut down" deployment.id=dev host.name=localhost service.name=jose-server service.version=0.0.1 service.instance.id=019afdd1-6e3d-7af9-96ab-80d3b4124e74 uptime=10.425244 flush=0.0774539
ok   cryptoutil/internal/jose/server 10.831s
