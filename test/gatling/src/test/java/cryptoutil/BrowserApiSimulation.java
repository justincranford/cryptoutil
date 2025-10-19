package cryptoutil;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;

public class BrowserApiSimulation extends Simulation {

  // Load configuration from system properties
  private static final String baseUrl = System.getProperty("browserApiBaseUrl", "https://localhost:8080/browser/api/v1");
  private static final int virtualusers = Integer.getInteger("virtualusers", 1);

  // Define HTTP configuration for browser API
  private static final HttpProtocolBuilder httpProtocol = http
      .baseUrl(baseUrl)
      .acceptHeader("application/json")
      .contentTypeHeader("application/json")
      .userAgentHeader("Gatling-Cryptoutil-Browser-API-Test/1.0")
      .inferHtmlResources() // Handle potential HTML resources
      .disableFollowRedirect(); // Disable redirects for API testing
      // SSL certificate validation disabled via JVM system properties

  // CSRF token retrieval chain
  private static final ChainBuilder getCsrfTokenChain = exec(http("Get CSRF Token")
      .get("/csrf-token")
      .check(status().is(200))
      .check(jsonPath("$.token").saveAs("csrfToken")));

  // CSRF token retrieval scenario
  private static final ScenarioBuilder getCsrfToken = scenario("Get CSRF Token")
      .exec(getCsrfTokenChain);

  // Key generation chain with CSRF
  private static final ChainBuilder keyGenChain = exec(getCsrfTokenChain)
      .exec(http("Generate RSA Key")
          .post("/elastickey")
          .header("X-CSRF-Token", "#{csrfToken}")
          .body(StringBody("{\"name\":\"test-key\",\"algorithm\":\"RSA\",\"keySize\":2048,\"provider\":\"CRYPTOUTIL\"}"))
          .check(status().is(201))
          .check(jsonPath("$.id").exists())
          .check(jsonPath("$.algorithm").is("RSA")));

  // Key generation scenario with CSRF
  private static final ScenarioBuilder keyGenScenario = scenario("Browser API - Key Generation")
      .exec(keyGenChain);

  // Encryption/Decryption chain with CSRF
  private static final ChainBuilder cryptoChain = exec(getCsrfTokenChain)
      .exec(http("Generate Key for Crypto")
          .post("/elastickey")
          .header("X-CSRF-Token", "#{csrfToken}")
          .body(StringBody("{\"name\":\"crypto-key\",\"algorithm\":\"RSA\",\"keySize\":2048,\"provider\":\"CRYPTOUTIL\"}"))
          .check(status().is(201))
          .check(jsonPath("$.id").saveAs("keyId")))

      .exec(getCsrfTokenChain)
      .exec(http("Encrypt Data")
          .post("/crypto/encrypt")
          .header("X-CSRF-Token", "#{csrfToken}")
          .body(StringBody("{\"elasticKeyId\":\"#{keyId}\",\"plaintext\":\"SGVsbG8gV29ybGQ=\",\"algorithm\":\"RSA-OAEP\"}"))
          .check(status().is(200))
          .check(jsonPath("$.ciphertext").exists())
          .check(jsonPath("$.ciphertext").saveAs("ciphertext")))

      .exec(getCsrfTokenChain)
      .exec(http("Decrypt Data")
          .post("/crypto/decrypt")
          .header("X-CSRF-Token", "#{csrfToken}")
          .body(StringBody("{\"elasticKeyId\":\"#{keyId}\",\"ciphertext\":\"#{ciphertext}\",\"algorithm\":\"RSA-OAEP\"}"))
          .check(status().is(200))
          .check(jsonPath("$.plaintext").is("SGVsbG8gV29ybGQ="))); // Base64 encoded "Hello World"

  // Encryption/Decryption scenario with CSRF
  private static final ScenarioBuilder cryptoScenario = scenario("Browser API - Encryption/Decryption")
      .exec(cryptoChain);

  // Key retrieval chain with CSRF
  private static final ChainBuilder keyRetrievalChain = exec(getCsrfTokenChain)
      .exec(http("Create Key for Retrieval")
          .post("/elastickey")
          .header("X-CSRF-Token", "#{csrfToken}")
          .body(StringBody("{\"name\":\"retrieve-key\",\"algorithm\":\"RSA\",\"keySize\":2048,\"provider\":\"CRYPTOUTIL\"}"))
          .check(status().is(201))
          .check(jsonPath("$.id").saveAs("retrieveKeyId")))

      .exec(http("Get Key by ID")
          .get("/elastickey/#{retrieveKeyId}")
          .check(status().is(200))
          .check(jsonPath("$.id").is("#{retrieveKeyId}"))
          .check(jsonPath("$.algorithm").is("RSA")));

  // Key retrieval scenario with CSRF
  private static final ScenarioBuilder keyRetrievalScenario = scenario("Browser API - Key Retrieval")
      .exec(keyRetrievalChain);

  // Combined scenario
  private static final ScenarioBuilder fullScenario = scenario("Browser API - Full Crypto Workflow")
      .exec(keyGenChain)
      .pause(1)
      .exec(cryptoChain)
      .pause(1)
      .exec(keyRetrievalChain);

  // Performance assertions
  private static final Assertion assertions = global()
      .responseTime().max().lt(1500);  // max response time < 1.5s (browser API may be slower due to CSRF)

  // Injection profile and test setup
  {
    setUp(
        keyGenScenario.injectOpen(rampUsers(virtualusers).during(30)),
        cryptoScenario.injectOpen(rampUsers(virtualusers).during(60)),
        keyRetrievalScenario.injectOpen(rampUsers(virtualusers).during(20)),
        fullScenario.injectOpen(rampUsers(virtualusers * 2).during(120))
    )
    .assertions(assertions)
    .protocols(httpProtocol);
  }
}
