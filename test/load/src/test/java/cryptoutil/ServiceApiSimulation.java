package cryptoutil;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;

public class ServiceApiSimulation extends Simulation {

  // Load configuration from system properties
  private static final String baseUrl = System.getProperty("serviceApiBaseUrl", "https://localhost:8080/service/api/v1");
  private static final int virtualclients = Integer.getInteger("virtualclients", 1);

  // Define HTTP configuration for service API
  private static final HttpProtocolBuilder httpProtocol = http
      .baseUrl(baseUrl)
      .acceptHeader("application/json")
      .contentTypeHeader("application/json")
      .userAgentHeader("Gatling-Cryptoutil-Service-API-Test/1.0")
      .disableFollowRedirect(); // Disable redirects for API testing
      // SSL certificate validation disabled via JVM system properties

  // Key generation chain
  private static final ChainBuilder keyGenChain = exec(http("Generate RSA Key")
      .post("/elastickey")
      .body(StringBody("{\"name\":\"test-key\",\"algorithm\":\"RSA\",\"keySize\":2048,\"provider\":\"CRYPTOUTIL\"}"))
      .check(status().is(201))
      .check(jsonPath("$.id").exists())
      .check(jsonPath("$.algorithm").is("RSA")));

  // Key generation scenario
  private static final ScenarioBuilder keyGenScenario = scenario("Service API - Key Generation")
      .exec(keyGenChain);

  // Encryption/Decryption chain
  private static final ChainBuilder cryptoChain = exec(http("Generate Key for Crypto")
      .post("/elastickey")
      .body(StringBody("{\"name\":\"crypto-key\",\"algorithm\":\"RSA\",\"keySize\":2048,\"provider\":\"CRYPTOUTIL\"}"))
      .check(status().is(201))
      .check(jsonPath("$.id").saveAs("keyId")))

      .exec(http("Encrypt Data")
          .post("/crypto/encrypt")
          .body(StringBody("{\"elasticKeyId\":\"#{keyId}\",\"plaintext\":\"SGVsbG8gV29ybGQ=\",\"algorithm\":\"RSA-OAEP\"}"))
          .check(status().is(200))
          .check(jsonPath("$.ciphertext").exists())
          .check(jsonPath("$.ciphertext").saveAs("ciphertext")))

      .exec(http("Decrypt Data")
          .post("/crypto/decrypt")
          .body(StringBody("{\"elasticKeyId\":\"#{keyId}\",\"ciphertext\":\"#{ciphertext}\",\"algorithm\":\"RSA-OAEP\"}"))
          .check(status().is(200))
          .check(jsonPath("$.plaintext").is("SGVsbG8gV29ybGQ="))); // Base64 encoded "Hello World"

  // Encryption/Decryption scenario
  private static final ScenarioBuilder cryptoScenario = scenario("Service API - Encryption/Decryption")
      .exec(cryptoChain);

  // Key retrieval chain
  private static final ChainBuilder keyRetrievalChain = exec(http("Create Key for Retrieval")
      .post("/elastickey")
      .body(StringBody("{\"name\":\"retrieve-key\",\"algorithm\":\"RSA\",\"keySize\":2048,\"provider\":\"CRYPTOUTIL\"}"))
      .check(status().is(201))
      .check(jsonPath("$.id").saveAs("retrieveKeyId")))

      .exec(http("Get Key by ID")
          .get("/elastickey/#{retrieveKeyId}")
          .check(status().is(200))
          .check(jsonPath("$.id").is("#{retrieveKeyId}"))
          .check(jsonPath("$.algorithm").is("RSA")));

  // Key retrieval scenario
  private static final ScenarioBuilder keyRetrievalScenario = scenario("Service API - Key Retrieval")
      .exec(keyRetrievalChain);

  // Combined scenario
  private static final ScenarioBuilder fullScenario = scenario("Service API - Full Crypto Workflow")
      .exec(keyGenChain)
      .pause(1)
      .exec(cryptoChain)
      .pause(1)
      .exec(keyRetrievalChain);

  // Performance assertions
  private static final Assertion assertions = global()
      .responseTime().max().lt(1000);  // max response time < 1s

  // Injection profile and test setup
  {
    setUp(
        keyGenScenario.injectOpen(rampUsers(virtualclients).during(30)),
        cryptoScenario.injectOpen(rampUsers(virtualclients).during(60)),
        keyRetrievalScenario.injectOpen(rampUsers(virtualclients).during(20)),
        fullScenario.injectOpen(rampUsers(virtualclients * 2).during(120))
    )
    .assertions(assertions)
    .protocols(httpProtocol);
  }
}
