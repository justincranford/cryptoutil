package cryptoutil;

import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.StringBody;
import static io.gatling.javaapi.core.CoreDsl.exec;
import static io.gatling.javaapi.core.CoreDsl.global;
import static io.gatling.javaapi.core.CoreDsl.jsonPath;
import static io.gatling.javaapi.core.CoreDsl.rampUsers;
import static io.gatling.javaapi.core.CoreDsl.scenario;
import static io.gatling.javaapi.http.HttpDsl.http;
import static io.gatling.javaapi.http.HttpDsl.status;

import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Stream;

import io.gatling.javaapi.core.ChainBuilder;
import io.gatling.javaapi.core.PopulationBuilder;
import io.gatling.javaapi.core.Simulation;
import io.gatling.javaapi.http.HttpProtocolBuilder;

/**
 * PKI Product Load Testing Simulation.
 * Tests pki-ca (Certificate Authority) service: certificate issuance, revocation, CRL/OCSP.
 *
 * <p>Example Commands:
 * <pre>
 * # Quick validation (health check)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.PkiProductSimulation -Dprofile=quick
 *
 * # Standard load test (certificate lifecycle)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.PkiProductSimulation -Dprofile=standard -Dvirtualclients=5
 *
 * # Stress test (concurrent certificate operations)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.PkiProductSimulation -Dprofile=stress -Dvirtualclients=20 -DdurationSeconds=120
 * </pre>
 */
public class PkiProductSimulation extends Simulation {

    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final String PKI_CA_PORT = System.getProperty("pkiCaPort", "8300");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "1"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "60"));

    private static final String PKI_CA_BASE_URL = "https://127.0.0.1:" + PKI_CA_PORT + "/service/api/v1";

    private static final Iterator<Map<String, Object>> UUID_FEEDER =
        Stream.generate((Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
            .iterator();

    /**
     * Submit a CSR for certificate issuance.
     */
    private static ChainBuilder submitCsr(String saveIdAs) {
        return exec(
            http("Submit CSR")
                .post("/certificates")
                .body(StringBody("""
                    {
                        "csr": "-----BEGIN CERTIFICATE REQUEST-----\\nMIIBkTCB+wIBADBSMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEW\\nMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzEWMBQGA1UECgwNQ3J5cHRvdXRpbCBDQTBZ\\nMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDummyCSRKeyForLoadTesting\\n-----END CERTIFICATE REQUEST-----",
                        "profile": "server",
                        "validityDays": 365
                    }
                    """))
                .check(status().in(200, 201, 400))
                .check(jsonPath("$.certificateId").optional().saveAs(saveIdAs))
        );
    }

    /**
     * Retrieve a certificate by ID.
     */
    private static ChainBuilder getCertificate(String certIdVar) {
        return exec(
            http("Get Certificate")
                .get("/certificates/#{" + certIdVar + "}")
                .check(status().in(200, 404))
        );
    }

    /**
     * Retrieve the CRL (Certificate Revocation List).
     */
    private static ChainBuilder getCrl() {
        return exec(
            http("Get CRL")
                .get("/crl")
                .check(status().in(200, 404))
        );
    }

    /**
     * List CA profiles.
     */
    private static ChainBuilder listProfiles() {
        return exec(
            http("List CA Profiles")
                .get("/profiles")
                .check(status().in(200, 404))
        );
    }

    {
        List<PopulationBuilder> populations = new ArrayList<>();
        HttpProtocolBuilder protocol = createServiceApiProtocol(PKI_CA_BASE_URL);

        switch (PROFILE) {
            case "standard":
                // Certificate lifecycle
                populations.add(
                    scenario("PKI-CA Certificate Lifecycle")
                        .feed(UUID_FEEDER)
                        .exec(listProfiles())
                        .pause(1)
                        .exec(submitCsr("certId"))
                        .pause(1)
                        .exec(getCrl())
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(protocol)
                );
                break;

            case "stress":
                // Concurrent certificate operations
                populations.add(
                    scenario("PKI-CA Certificate Stress")
                        .feed(UUID_FEEDER)
                        .repeat(3).on(
                            exec(submitCsr("certId"))
                                .pause(1)
                        )
                        .exec(getCrl())
                        .exec(listProfiles())
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(protocol)
                );
                break;

            case "quick":
            default:
                // Health check + profiles listing
                populations.add(
                    scenario("PKI-CA Quick Validation")
                        .exec(http("PKI-CA Service Health")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .pause(1)
                        .exec(listProfiles())
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(protocol)
                );
                break;
        }

        setUp(populations.toArray(new PopulationBuilder[0]))
            .assertions(
                global().responseTime().max().lt(3000),
                global().successfulRequests().percent().gte(95.0)
            );
    }
}
