package cryptoutil;

import static io.gatling.javaapi.core.CoreDsl.StringBody;
import static io.gatling.javaapi.core.CoreDsl.bodyString;
import static io.gatling.javaapi.core.CoreDsl.doIf;
import static io.gatling.javaapi.core.CoreDsl.exec;
import static io.gatling.javaapi.core.CoreDsl.jsonPath;
import static io.gatling.javaapi.core.CoreDsl.scenario;
import static io.gatling.javaapi.http.HttpDsl.http;
import static io.gatling.javaapi.http.HttpDsl.status;

import java.util.Iterator;
import java.util.Map;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Stream;

import io.gatling.javaapi.core.ChainBuilder;
import io.gatling.javaapi.core.ScenarioBuilder;

public class CryptoutiApi {
    public static final String CLEARTEXT = "Hello World Load Test";

    public static final String DEFAULT_PROVIDER = "Internal";

    // Feeder for generating unique UUIDs
    private static final Iterator<Map<String, Object>> uuidFeeder =
        Stream.generate((Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
            .iterator();

    public static final String[] ELASTIC_KEY_ENCRYPTION_ALGORITHMS = {
        // GCM with A256KW
        "A256GCM/A256KW",
        "A192GCM/A256KW",
        "A128GCM/A256KW",
        // GCM with A256GCMKW
        "A256GCM/A256GCMKW",
        "A192GCM/A256GCMKW",
        "A128GCM/A256GCMKW",
        // GCM with dir
        "A256GCM/dir",
        "A192GCM/dir",
        "A128GCM/dir",
        // GCM with RSA-OAEP-512
        "A256GCM/RSA-OAEP-512",
        "A192GCM/RSA-OAEP-512",
        "A128GCM/RSA-OAEP-512",
        // GCM with ECDH-ES+A256KW
        "A256GCM/ECDH-ES+A256KW",
        "A192GCM/ECDH-ES+A256KW",
        "A128GCM/ECDH-ES+A256KW",
        // GCM with ECDH-ES
        "A256GCM/ECDH-ES",
        "A192GCM/ECDH-ES",
        "A128GCM/ECDH-ES",
        // CBC with A256KW
        "A256CBC-HS512/A256KW",
        "A192CBC-HS384/A256KW",
        "A128CBC-HS256/A256KW",
        // CBC with dir
        "A256CBC-HS512/dir",
        "A192CBC-HS384/dir",
        "A128CBC-HS256/dir",
    };

    // Signature algorithms from happyPathElasticKeyTestCasesSign in client_test.go
    public static final String[] ELASTIC_KEY_SIGNATURE_ALGORITHMS = {
        // "RS256",
        // "RS384",
        // "RS512",
        // "PS256",
        // "PS384",
        // "PS512",
        "ES256",
        // "ES384",
        // "ES512",
        // "HS256",
        // "HS384",
        // "HS512",
        // "EdDSA",
    };

    public static final String[] DATA_KEY_ALGORITHMS = {
        // "RSA/4096",
        // "RSA/3072",
        // "RSA/2048",
        // "EC/P521",
        // "EC/P384",
        // "EC/P256",
        // "OKP/Ed25519",
        // "oct/512", // AES256-HS512
        // "oct/384", // AES192-HS384
        // "oct/256", // AES-256, AES256-HS512
        // "oct/192", // AES-192
        "oct/128", // AES-128
    };

    // ========================================
    // Elastic Key Management APIs
    // ========================================

    /**
     * Create a new Elastic Key.
     * POST /elastickey
     *
     * @param name Elastic Key name
     * @param description Elastic Key description
     * @param algorithm Encryption algorithm (e.g., "A256GCM/A256KW")
     * @param provider Provider (e.g., "Internal")
     * @param importAllowed Whether import is allowed
     * @param versioningAllowed Whether versioning is allowed
     * @param saveElasticKeyIdAs Session variable name to save elastic_key_id
     * @return ChainBuilder for creating an Elastic Key
     */
    public static ChainBuilder createElasticKey(
        String name,
        String description,
        String algorithm,
        String provider,
        boolean importAllowed,
        boolean versioningAllowed,
        String saveElasticKeyIdAs
    ) {
        return exec(
            http("Create Elastic Key: " + algorithm)
                .post("/elastickey")
                .body(StringBody(String.format(
                    "{\"name\":\"%s\",\"description\":\"%s\",\"algorithm\":\"%s\",\"provider\":\"%s\",\"import_allowed\":%b,\"versioning_allowed\":%b}",
                    name, description, algorithm, provider, importAllowed, versioningAllowed
                )))
                .check(status().is(200))
                .check(jsonPath("$.elastic_key_id").saveAs(saveElasticKeyIdAs))
                .check(jsonPath("$.name").exists())
                .check(jsonPath("$.description").is(description))
                .check(jsonPath("$.algorithm").is(algorithm))
                .check(jsonPath("$.provider").is(provider))
                .check(jsonPath("$.import_allowed").is(String.valueOf(importAllowed)))
                .check(jsonPath("$.versioning_allowed").is(String.valueOf(versioningAllowed)))
                .check(jsonPath("$.status").is("active"))
        );
    }

    /**
     * Get an Elastic Key by ID.
     * GET /elastickey/{elasticKeyID}
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @return ChainBuilder for getting an Elastic Key
     */
    public static ChainBuilder getElasticKey(String elasticKeyIdSessionVar) {
        return exec(
            http("Get Elastic Key")
                .get("/elastickey/#{" + elasticKeyIdSessionVar + "}")
                .check(status().is(200))
                .check(jsonPath("$.elastic_key_id").exists())
        );
    }

    /**
     * Find Elastic Keys with optional filtering, sorting, and paging.
     * GET /elastickeys
     *
     * @param page Page number (0-indexed)
     * @param size Page size (2-50)
     * @return ChainBuilder for finding Elastic Keys
     */
    public static ChainBuilder findElasticKeys(int page, int size) {
        return exec(
            http("Find Elastic Keys")
                .get("/elastickeys?page=" + page + "&size=" + size)
                .check(status().is(200))
        );
    }

    // ========================================
    // Material Key Management APIs
    // ========================================

    /**
     * Generate a new Material Key in an Elastic Key.
     * POST /elastickey/{elasticKeyID}/materialkey
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param saveMaterialKeyIdAs Session variable name to save material_key_id
     * @param saveClearPublicAs Optional session variable name to save clear_public (null to skip)
     * @return ChainBuilder for generating a Material Key
     */
    public static ChainBuilder generateMaterialKey(
        String elasticKeyIdSessionVar,
        String saveMaterialKeyIdAs,
        String saveClearPublicAs
    ) {
        ChainBuilder chain = exec(
            http("Generate Material Key")
                .post("/elastickey/#{" + elasticKeyIdSessionVar + "}/materialkey")
                .body(StringBody("{}"))
                .check(status().is(200))
                .check(jsonPath("$.elastic_key_id").exists())
                .check(jsonPath("$.material_key_id").saveAs(saveMaterialKeyIdAs))
                .check(jsonPath("$.generate_date").exists())
        );

        if (saveClearPublicAs != null) {
            chain = chain.exec(
                doIf(session -> session.contains("clear_public")).then(
                    exec(session -> session.set(saveClearPublicAs, session.getString("clear_public")))
                )
            );
        }

        return chain;
    }

    /**
     * Get a Material Key by ID.
     * GET /elastickey/{elasticKeyID}/materialkey/{materialKeyID}
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param materialKeyIdSessionVar Session variable containing material_key_id
     * @return ChainBuilder for getting a Material Key
     */
    public static ChainBuilder getMaterialKey(String elasticKeyIdSessionVar, String materialKeyIdSessionVar) {
        return exec(
            http("Get Material Key")
                .get("/elastickey/#{" + elasticKeyIdSessionVar + "}/materialkey/#{" + materialKeyIdSessionVar + "}")
                .check(status().is(200))
                .check(jsonPath("$.elastic_key_id").exists())
                .check(jsonPath("$.material_key_id").exists())
        );
    }

    /**
     * Find Material Keys in an Elastic Key.
     * GET /elastickey/{elasticKeyID}/materialkeys
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param page Page number (0-indexed)
     * @param size Page size (2-50)
     * @return ChainBuilder for finding Material Keys
     */
    public static ChainBuilder findMaterialKeys(String elasticKeyIdSessionVar, int page, int size) {
        return exec(
            http("Find Material Keys")
                .get("/elastickey/#{" + elasticKeyIdSessionVar + "}/materialkeys?page=" + page + "&size=" + size)
                .check(status().is(200))
        );
    }

    /**
     * Find all Material Keys across Elastic Keys.
     * GET /materialkeys
     *
     * @param page Page number (0-indexed)
     * @param size Page size (2-50)
     * @return ChainBuilder for finding all Material Keys
     */
    public static ChainBuilder findAllMaterialKeys(int page, int size) {
        return exec(
            http("Find All Material Keys")
                .get("/materialkeys?page=" + page + "&size=" + size)
                .check(status().is(200))
        );
    }

    // ========================================
    // Cryptographic Operation APIs
    // ========================================

    /**
     * Generate a random key (Secret Key, Key Pair, etc.) encrypted as JWE.
     * POST /elastickey/{elasticKeyID}/generate
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param algorithm Generate algorithm (e.g., "RSA/2048", "EC/P256", "oct/256")
     * @param context Optional context (null to omit)
     * @param saveJweAs Session variable name to save JWE response
     * @return ChainBuilder for generating a key
     */
    public static ChainBuilder generateDataKey(
        String elasticKeyIdSessionVar,
        String algorithm,
        String context,
        String saveJweAs
    ) {
        String queryParams = "";
        if (algorithm != null) {
            queryParams += "?alg=" + algorithm;
        }
        if (context != null) {
            queryParams += (queryParams.isEmpty() ? "?" : "&") + "context=" + context;
        }

        return exec(
            http("Generate Data Key: " + algorithm)
                .post("/elastickey/#{" + elasticKeyIdSessionVar + "}/generate" + queryParams)
                .check(status().is(200))
                .check(bodyString().saveAs(saveJweAs))
                .check(bodyString().transform(s -> s.split("\\.").length == 5 ? "valid" : "invalid").is("valid"))
        );
    }

    /**
     * Encrypt clear text data using latest Material Key.
     * POST /elastickey/{elasticKeyID}/encrypt
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param cleartext Clear text to encrypt
     * @param context Optional context (null to omit)
     * @param saveCiphertextAs Session variable name to save ciphertext JWE
     * @return ChainBuilder for encryption
     */
    public static ChainBuilder encrypt(
        String elasticKeyIdSessionVar,
        String cleartext,
        String context,
        String saveCiphertextAs
    ) {
        String queryParams = context != null ? "?context=" + context : "";

        return exec(
            http("Encrypt")
                .post("/elastickey/#{" + elasticKeyIdSessionVar + "}/encrypt" + queryParams)
                .body(StringBody(cleartext))
                .header("Content-Type", "text/plain")
                .check(status().is(200))
                .check(bodyString().saveAs(saveCiphertextAs))
                .check(bodyString().transform(s -> s.split("\\.").length == 5 ? "valid" : "invalid").is("valid"))
        );
    }

    /**
     * Decrypt JWE message using Material Key identified by kid header.
     * POST /elastickey/{elasticKeyID}/decrypt
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param ciphertextSessionVar Session variable containing ciphertext JWE
     * @param saveDecryptedAs Session variable name to save decrypted plaintext
     * @return ChainBuilder for decryption
     */
    public static ChainBuilder decrypt(
        String elasticKeyIdSessionVar,
        String ciphertextSessionVar,
        String saveDecryptedAs
    ) {
        return exec(
            http("Decrypt")
                .post("/elastickey/#{" + elasticKeyIdSessionVar + "}/decrypt")
                .body(StringBody("#{" + ciphertextSessionVar + "}"))
                .header("Content-Type", "text/plain")
                .check(status().is(200))
                .check(bodyString().saveAs(saveDecryptedAs))
        );
    }

    /**
     * Sign clear text using latest Material Key.
     * POST /elastickey/{elasticKeyID}/sign
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param cleartext Clear text to sign
     * @param context Optional context (null to omit)
     * @param saveSignatureAs Session variable name to save signature JWS
     * @return ChainBuilder for signing
     */
    public static ChainBuilder sign(
        String elasticKeyIdSessionVar,
        String cleartext,
        String context,
        String saveSignatureAs
    ) {
        String queryParams = context != null ? "?context=" + context : "";

        return exec(
            http("Sign")
                .post("/elastickey/#{" + elasticKeyIdSessionVar + "}/sign" + queryParams)
                .body(StringBody(cleartext))
                .header("Content-Type", "text/plain")
                .check(status().is(200))
                .check(bodyString().saveAs(saveSignatureAs))
                .check(bodyString().transform(s -> s.split("\\.").length == 3 ? "valid" : "invalid").is("valid"))
        );
    }

    /**
     * Verify JWS message using Material Key identified by kid header.
     * POST /elastickey/{elasticKeyID}/verify
     *
     * @param elasticKeyIdSessionVar Session variable containing elastic_key_id
     * @param signatureSessionVar Session variable containing signature JWS
     * @return ChainBuilder for verification
     */
    public static ChainBuilder verify(
        String elasticKeyIdSessionVar,
        String signatureSessionVar
    ) {
        return exec(
            http("Verify")
                .post("/elastickey/#{" + elasticKeyIdSessionVar + "}/verify")
                .body(StringBody("#{" + signatureSessionVar + "}"))
                .header("Content-Type", "text/plain")
                .check(status().is(204))
        );
    }

    // ========================================
    // Scenario Builders - Common Test Patterns
    // ========================================

    /**
     * Full encrypt/decrypt cycle scenario matching TestAllElasticKeyCipherAlgorithms pattern.
     * Creates Elastic Key → Generates Material Key → Encrypts → Generates Another Key → Decrypts → Validates.
     *
     * @param algorithm Elastic Key algorithm (e.g., "A256GCM/A256KW")
     * @param cleartext Clear text to encrypt/decrypt
     * @return ScenarioBuilder for full encryption workflow
     */
    public static ScenarioBuilder buildEncryptDecryptScenario(String algorithm, String cleartext) {
        String scenarioName = "Encrypt/Decrypt: " + algorithm;
        String elasticKeyName = "LoadTest-" + algorithm.replace("/", "_") + "-#{uuid}";
        String elasticKeyDesc = "Load test for " + algorithm;

        return scenario(scenarioName)
            .feed(uuidFeeder)
            .exec(createElasticKey(
                elasticKeyName,
                elasticKeyDesc,
                algorithm,
                "Internal",
                false,
                true,
                "elasticKeyId"
            ))
            .pause(1)
            .exec(generateMaterialKey("elasticKeyId", "materialKeyId1", "clearPublic1"))
            .pause(1)
            .exec(encrypt("elasticKeyId", cleartext, null, "ciphertext"))
            .pause(1)
            .exec(generateMaterialKey("elasticKeyId", "materialKeyId2", "clearPublic2"))
            .pause(1)
            .exec(decrypt("elasticKeyId", "ciphertext", "decrypted"))
            .exec(session -> {
                String original = cleartext;
                String decrypted = session.getString("decrypted");
                if (!original.equals(decrypted)) {
                    throw new RuntimeException("Decryption validation failed: expected '" + original + "', got '" + decrypted + "'");
                }
                return session;
            });
    }

    /**
     * Full sign/verify cycle scenario matching TestAllElasticKeySignatureAlgorithms pattern.
     * Creates Elastic Key → Generates Material Key → Signs → Generates Another Key → Verifies.
     *
     * @param algorithm Elastic Key algorithm (e.g., "RS256", "ES256", "EdDSA")
     * @param cleartext Clear text to sign/verify
     * @return ScenarioBuilder for full signature workflow
     */
    public static ScenarioBuilder buildSignVerifyScenario(String algorithm, String cleartext) {
        String scenarioName = "Sign/Verify: " + algorithm;
        String elasticKeyName = "LoadTest-" + algorithm + "-#{uuid}";
        String elasticKeyDesc = "Load test for " + algorithm;

        return scenario(scenarioName)
            .feed(uuidFeeder)
            .exec(createElasticKey(
                elasticKeyName,
                elasticKeyDesc,
                algorithm,
                "Internal",
                false,
                true,
                "elasticKeyId"
            ))
            .pause(1)
            .exec(generateMaterialKey("elasticKeyId", "materialKeyId1", "clearPublic1"))
            .pause(1)
            .exec(sign("elasticKeyId", cleartext, null, "signature"))
            .pause(1)
            .exec(generateMaterialKey("elasticKeyId", "materialKeyId2", "clearPublic2"))
            .pause(1)
            .exec(verify("elasticKeyId", "signature"));
    }

    /**
     * Data key generation scenario matching TestAllElasticKeyCipherAlgorithms generate data key pattern.
     * Creates Elastic Key → Generates Material Key → Generates Data Keys → Decrypts Data Keys → Validates JWKs.
     *
     * @param elasticKeyAlgorithm Elastic Key algorithm (e.g., "A256GCM/A256KW")
     * @param dataKeyAlgorithms Array of data key algorithms (e.g., ["RSA/2048", "EC/P256", "oct/256"])
     * @return ScenarioBuilder for data key generation workflow
     */
    public static ScenarioBuilder buildDataKeyGenerationScenario(String elasticKeyAlgorithm, String[] dataKeyAlgorithms) {
        String scenarioName = "Data Key Generation: " + elasticKeyAlgorithm;
        String elasticKeyName = "LoadTest-DataKey-" + elasticKeyAlgorithm.replace("/", "_") + "-#{uuid}";
        String elasticKeyDesc = "Load test data key generation for " + elasticKeyAlgorithm;

        ScenarioBuilder scenario = scenario(scenarioName)
            .feed(uuidFeeder)
            .exec(createElasticKey(
                elasticKeyName,
                elasticKeyDesc,
                elasticKeyAlgorithm,
                "Internal",
                false,
                true,
                "elasticKeyId"
            ))
            .pause(1);

        for (int i = 0; i < dataKeyAlgorithms.length; i++) {
            String dataKeyAlg = dataKeyAlgorithms[i];
            String saveVar = "dataKeyJwe_" + i;
            String decryptedVar = "dataKeyJwk_" + i;

            scenario = scenario
                .exec(generateDataKey("elasticKeyId", dataKeyAlg, null, saveVar))
                .pause(1)
                .exec(decrypt("elasticKeyId", saveVar, decryptedVar))
                .exec(session -> {
                    // Validate JWK structure (basic validation)
                    String jwk = session.getString(decryptedVar);
                    if (!jwk.contains("kty") || !jwk.contains("kid")) {
                        throw new RuntimeException("Invalid JWK structure: missing required fields");
                    }
                    return session;
                })
                .pause(1);
        }

        return scenario;
    }

    /**
     * Complete workflow scenario matching TestAllElasticKeyCipherAlgorithms with all operations.
     * Combines Material Key generation, encryption/decryption, and data key generation.
     *
     * @param elasticKeyAlgorithm Elastic Key algorithm
     * @param cleartext Clear text for encryption
     * @param dataKeyAlgorithms Array of data key algorithms to test
     * @return ScenarioBuilder for complete workflow
     */
    public static ScenarioBuilder buildCompleteWorkflowScenario(
        String elasticKeyAlgorithm,
        String cleartext,
        String[] dataKeyAlgorithms
    ) {
        String scenarioName = "Complete Workflow: " + elasticKeyAlgorithm;
        String elasticKeyName = "LoadTest-Complete-" + elasticKeyAlgorithm.replace("/", "_") + "-#{uuid}";
        String elasticKeyDesc = "Complete load test for " + elasticKeyAlgorithm;

        ScenarioBuilder scenario = scenario(scenarioName)
            .feed(uuidFeeder)
            .exec(createElasticKey(
                elasticKeyName,
                elasticKeyDesc,
                elasticKeyAlgorithm,
                "Internal",
                false,
                true,
                "elasticKeyId"
            ))
            .pause(1)
            .exec(generateMaterialKey("elasticKeyId", "materialKeyId1", "clearPublic1"))
            .pause(1)
            .exec(encrypt("elasticKeyId", cleartext, null, "ciphertext"))
            .pause(1)
            .exec(generateMaterialKey("elasticKeyId", "materialKeyId2", "clearPublic2"))
            .pause(1)
            .exec(decrypt("elasticKeyId", "ciphertext", "decrypted"))
            .exec(session -> {
                String original = cleartext;
                String decrypted = session.getString("decrypted");
                if (!original.equals(decrypted)) {
                    throw new RuntimeException("Decryption validation failed");
                }
                return session;
            })
            .pause(1);

        for (int i = 0; i < dataKeyAlgorithms.length; i++) {
            String dataKeyAlg = dataKeyAlgorithms[i];
            String saveVar = "dataKeyJwe_" + i;
            String decryptedVar = "dataKeyJwk_" + i;

            scenario = scenario
                .exec(generateDataKey("elasticKeyId", dataKeyAlg, null, saveVar))
                .pause(1)
                .exec(decrypt("elasticKeyId", saveVar, decryptedVar))
                .pause(1);
        }

        return scenario;
    }
}
