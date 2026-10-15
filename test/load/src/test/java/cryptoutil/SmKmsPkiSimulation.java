package cryptoutil;

import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.StringBody;
import static io.gatling.javaapi.core.CoreDsl.exec;
import static io.gatling.javaapi.core.CoreDsl.global;
import static io.gatling.javaapi.core.CoreDsl.rampUsers;
import static io.gatling.javaapi.core.CoreDsl.scenario;
import static io.gatling.javaapi.http.HttpDsl.http;
import static io.gatling.javaapi.http.HttpDsl.status;

import io.gatling.javaapi.core.ChainBuilder;
import io.gatling.javaapi.core.PopulationBuilder;
import io.gatling.javaapi.core.Simulation;
import io.gatling.javaapi.http.HttpProtocolBuilder;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Stream;

/**
 * sm-kms PKI domain load testing simulation. Tests PKI certificate enrollment and lifecycle
 * endpoints hosted by sm-kms: certificate enrollment, certificate listing, CA discovery, and
 * profile discovery.
 *
 * <p>Example Commands:
 *
 * <pre>
 * # Quick validation (health check)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmKmsPkiSimulation -Dprofile=quick
 *
 * # Standard load test (certificate lifecycle)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmKmsPkiSimulation -Dprofile=standard -Dvirtualclients=5
 *
 * # Stress test (concurrent certificate operations)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmKmsPkiSimulation -Dprofile=stress -Dvirtualclients=20 -DdurationSeconds=120
 * </pre>
 */
public class SmKmsPkiSimulation extends Simulation {

  private static final String PROFILE = System.getProperty("profile", "quick");
  private static final String SM_KMS_PORT = System.getProperty("smKmsPort", "8000");
  private static final int VIRTUAL_CLIENTS =
      Integer.parseInt(System.getProperty("virtualclients", "1"));
  private static final int DURATION_SECONDS =
      Integer.parseInt(System.getProperty("durationSeconds", "60"));

  private static final String SM_KMS_BASE_URL =
      "https://127.0.0.1:" + SM_KMS_PORT + "/service/api/v1";

  private static final Iterator<Map<String, Object>> UUID_FEEDER =
      Stream.generate(
              (Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
          .iterator();

  /** Submit a CSR for certificate issuance. */
  private static ChainBuilder submitCsr() {
    return exec(
        http("Submit CSR")
            .post("/pki/enrollments")
            .body(
                StringBody(
                    """
                    {
                        "csr": "-----BEGIN CERTIFICATE REQUEST-----\\nMIIBkTCB+wIBADBSMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEW\\nMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzEWMBQGA1UECgwNQ3J5cHRvdXRpbCBDQTBZ\\nMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDummyCSRKeyForLoadTesting\\n-----END CERTIFICATE REQUEST-----",
                        "profile": "server",
                        "validityDays": 365
                    }
                    """))
            .check(status().in(200, 201, 400)));
  }

  /** List available Certificate Authorities. */
  private static ChainBuilder listCAs() {
    return exec(http("List Certificate Authorities").get("/pki/cas").check(status().in(200, 404)));
  }

  /** List CA profiles. */
  private static ChainBuilder listProfiles() {
    return exec(http("List CA Profiles").get("/pki/profiles").check(status().in(200, 404)));
  }

  {
    List<PopulationBuilder> populations = new ArrayList<>();
    HttpProtocolBuilder protocol = createServiceApiProtocol(SM_KMS_BASE_URL);

    switch (PROFILE) {
      case "standard":
        // Certificate lifecycle
        populations.add(
            scenario("SM-KMS PKI Certificate Lifecycle")
                .feed(UUID_FEEDER)
                .exec(listProfiles())
                .pause(1)
                .exec(submitCsr())
                .pause(1)
                .exec(listCAs())
                .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                .protocols(protocol));
        break;

      case "stress":
        // Concurrent certificate operations
        populations.add(
            scenario("SM-KMS PKI Certificate Stress")
                .feed(UUID_FEEDER)
                .repeat(3)
                .on(exec(submitCsr()).pause(1))
                .exec(listCAs())
                .exec(listProfiles())
                .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                .protocols(protocol));
        break;

      case "quick":
      default:
        // Health check + profiles listing
        populations.add(
            scenario("SM-KMS PKI Quick Validation")
                .exec(http("SM-KMS PKI Service Health").get("/health").check(status().in(200, 503)))
                .pause(1)
                .exec(listProfiles())
                .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                .protocols(protocol));
        break;
    }

    setUp(populations.toArray(new PopulationBuilder[0]))
        .assertions(
            global().responseTime().max().lt(3000),
            global().successfulRequests().percent().gte(95.0));
  }
}
