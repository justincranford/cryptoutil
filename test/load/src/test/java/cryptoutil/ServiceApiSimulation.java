package cryptoutil;

import static cryptoutil.CryptoutiApi.ELASTIC_KEY_ENCRYPTION_ALGORITHMS;
import static cryptoutil.CryptoutiApi.CLEARTEXT;
import static cryptoutil.CryptoutiApi.DATA_KEY_ALGORITHMS;
import static cryptoutil.CryptoutiApi.DEFAULT_PROVIDER;
import static cryptoutil.CryptoutiApi.ELASTIC_KEY_SIGNATURE_ALGORITHMS;
import static cryptoutil.CryptoutiApi.buildCompleteWorkflowScenario;
import static cryptoutil.CryptoutiApi.buildEncryptDecryptScenario;
import static cryptoutil.CryptoutiApi.buildSignVerifyScenario;
import static cryptoutil.CryptoutiApi.createElasticKey;
import static cryptoutil.CryptoutiApi.decrypt;
import static cryptoutil.CryptoutiApi.encrypt;
import static cryptoutil.CryptoutiApi.generateMaterialKey;
import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.global;
import static io.gatling.javaapi.core.CoreDsl.rampUsers;
import static io.gatling.javaapi.core.CoreDsl.scenario;

import java.util.Iterator;
import java.util.Map;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Stream;

import io.gatling.javaapi.core.Assertion;
import io.gatling.javaapi.core.FeederBuilder;
import io.gatling.javaapi.core.PopulationBuilder;
import io.gatling.javaapi.core.Simulation;

/**
 * Service API Load Testing Simulation.
 * Matches TestAllElasticKeyCipherAlgorithms from internal/client/client_test.go.
 *
 * Example Commands:
 * <pre>
 * # Quick validation test (fastest, single algorithm)
 * .\mvnw.cmd gatling:test -Dprofile=quick -Dvirtualclients=1 -DdurationSeconds=30
 *
 * # Test all cipher algorithms
 * .\mvnw.cmd gatling:test -Dprofile=cipher -Dvirtualclients=2 -DdurationSeconds=45
 *
 * # Test all signature algorithms
 * .\mvnw.cmd gatling:test -Dprofile=signature -Dvirtualclients=2 -DdurationSeconds=45
 *
 * # Complete workflow with data key generation
 * .\mvnw.cmd gatling:test -Dprofile=complete -Dvirtualclients=1 -DdurationSeconds=60
 *
 * # Custom port (for cryptoutil-postgres instances)
 * .\mvnw.cmd gatling:test -Dprofile=quick -Dport=8081 -Dvirtualclients=3
 * </pre>
 *
 * Test flow per algorithm:
 * 1. Create Elastic Key with specific algorithm
 * 2. Generate Material Key (stores key metadata and optional public key)
 * 3. Encrypt cleartext → returns JWE ciphertext
 * 4. Generate another Material Key (key rotation)
 * 5. Decrypt ciphertext → returns plaintext
 * 6. Validate decrypted text matches original
 * 7. For each data key algorithm (RSA/2048, EC/P256, oct/256, etc.):
 *    - Generate encrypted data key (JWE)
 *    - Decrypt data key to get JWK
 *    - Validate JWK structure
 */
public class ServiceApiSimulation extends Simulation {

  // Load configuration from system properties
  private static final String testProfile = System.getProperty("profile", "quick");
  private static final String port = System.getProperty("port", "8080");
  private static final String baseUrl = "https://127.0.0.1:" + port + "/service/api/v1";
  private static final int virtualclients = Integer.getInteger("virtualclients", 1);
  private static final int durationSeconds = Integer.getInteger("durationSeconds", 60);

  // Performance assertions - using separate assertions (Gatling 3.14.6 API)
  private static final Assertion responseTimeAssertion = global().responseTime().max().lt(5000);  // max response time < 5s (crypto operations are slower)
  private static final Assertion successRateAssertion = global().successfulRequests().percent().gte(95.0);  // 95% success rate

  // Feeder for generating unique UUIDs
  private static final Iterator<Map<String, Object>> uuidFeeder =
      Stream.generate((Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
          .iterator();

  // Injection profile and test setup
  {
    PopulationBuilder[] scenarios;

    switch (testProfile) {
      case "cipher":
        // Test representative cipher algorithms
        scenarios = new PopulationBuilder[ELASTIC_KEY_ENCRYPTION_ALGORITHMS.length];
        for (int i = 0; i < ELASTIC_KEY_ENCRYPTION_ALGORITHMS.length; i++) {
          scenarios[i] = buildEncryptDecryptScenario(ELASTIC_KEY_ENCRYPTION_ALGORITHMS[i], CLEARTEXT)
              .injectOpen(rampUsers(virtualclients).during(durationSeconds));
        }
        break;

      case "signature":
        // Test all signature algorithms
        scenarios = new PopulationBuilder[ELASTIC_KEY_SIGNATURE_ALGORITHMS.length];
        for (int i = 0; i < ELASTIC_KEY_SIGNATURE_ALGORITHMS.length; i++) {
          scenarios[i] = buildSignVerifyScenario(ELASTIC_KEY_SIGNATURE_ALGORITHMS[i], CLEARTEXT)
              .injectOpen(rampUsers(virtualclients).during(durationSeconds));
        }
        break;

      case "complete":
        // Test complete workflow with data key generation (most comprehensive)
        scenarios = new PopulationBuilder[ELASTIC_KEY_ENCRYPTION_ALGORITHMS.length];
        for (int i = 0; i < ELASTIC_KEY_ENCRYPTION_ALGORITHMS.length; i++) {
          scenarios[i] = buildCompleteWorkflowScenario(ELASTIC_KEY_ENCRYPTION_ALGORITHMS[i], CLEARTEXT, DATA_KEY_ALGORITHMS)
              .injectOpen(rampUsers(virtualclients).during(durationSeconds));
        }
        break;

      case "quick":
      default:
        // Quick validation with single algorithm
        scenarios = new PopulationBuilder[]{
            scenario("Quick Validation")
      .feed(uuidFeeder)
      .exec(createElasticKey(
          "QuickTest-#{uuid}",
          "Quick validation test",
          "A256GCM/A256KW",
          DEFAULT_PROVIDER,
          false,
          true,
          "elasticKeyId"
      ))
      .pause(1)
      .exec(generateMaterialKey("elasticKeyId", "materialKeyId", null))
      .pause(1)
      .exec(encrypt("elasticKeyId", CLEARTEXT, null, "ciphertext"))
      .pause(1)
      .exec(decrypt("elasticKeyId", "ciphertext", "decrypted"))
      .exec(session -> {
        String decrypted = session.getString("decrypted");
        if (!CLEARTEXT.equals(decrypted)) {
          throw new RuntimeException("Quick validation failed");
        }
        return session;
      }).injectOpen(rampUsers(virtualclients).during(durationSeconds))
        };
        break;
    }

    setUp(scenarios)
        .assertions(responseTimeAssertion, successRateAssertion)
        .protocols(createServiceApiProtocol(baseUrl));
  }
}
