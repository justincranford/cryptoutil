package cryptoutil;

import static cryptoutil.CryptoutiApi.CLEARTEXT;
import static cryptoutil.CryptoutiApi.DEFAULT_PROVIDER;
import static cryptoutil.CryptoutiApi.createElasticKey;
import static cryptoutil.CryptoutiApi.decrypt;
import static cryptoutil.CryptoutiApi.encrypt;
import static cryptoutil.CryptoutiApi.generateMaterialKey;
import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.global;
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

import io.gatling.javaapi.core.PopulationBuilder;
import io.gatling.javaapi.core.Simulation;
import io.gatling.javaapi.http.HttpProtocolBuilder;

/**
 * Cryptoutil Suite Load Testing Simulation.
 * Exercises all 5 products (10 services) simultaneously to validate system-wide performance.
 *
 * <p>Port assignments follow the service catalog:
 * <ul>
 *   <li>sm-kms: 8000</li>
 *   <li>sm-im: 8100</li>
 *   <li>jose-ja: 8200</li>
 *   <li>pki-ca: 8300</li>
 *   <li>identity-authz: 8400</li>
 *   <li>identity-idp: 8500</li>
 *   <li>identity-rs: 8600</li>
 *   <li>identity-rp: 8700</li>
 *   <li>identity-spa: 8800</li>
 *   <li>skeleton-template: 8900</li>
 * </ul>
 *
 * <p>Example Commands:
 * <pre>
 * # Quick validation (health checks for all 10 services)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.CryptoutilSuiteSimulation -Dprofile=quick
 *
 * # Standard load test (mixed workload across all products)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.CryptoutilSuiteSimulation -Dprofile=standard -Dvirtualclients=3
 *
 * # Stress test (high concurrency across all services)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.CryptoutilSuiteSimulation -Dprofile=stress -Dvirtualclients=10 -DdurationSeconds=180
 * </pre>
 */
public class CryptoutilSuiteSimulation extends Simulation {

    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "1"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "60"));

    // Service ports from the canonical service catalog.
    private static final String SM_KMS_PORT = System.getProperty("smKmsPort", "8000");
    private static final String SM_IM_PORT = System.getProperty("smImPort", "8100");
    private static final String JOSE_JA_PORT = System.getProperty("joseJaPort", "8200");
    private static final String PKI_CA_PORT = System.getProperty("pkiCaPort", "8300");
    private static final String AUTHZ_PORT = System.getProperty("authzPort", "8400");
    private static final String IDP_PORT = System.getProperty("idpPort", "8500");
    private static final String RS_PORT = System.getProperty("rsPort", "8600");
    private static final String RP_PORT = System.getProperty("rpPort", "8700");
    private static final String SPA_PORT = System.getProperty("spaPort", "8800");
    private static final String SKELETON_PORT = System.getProperty("skeletonPort", "8900");

    private static final Iterator<Map<String, Object>> UUID_FEEDER =
        Stream.generate((Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
            .iterator();

    /**
     * Service-level health check population.
     */
    private static PopulationBuilder serviceHealthCheck(String name, String port, int clients, int duration) {
        HttpProtocolBuilder protocol = createServiceApiProtocol("https://127.0.0.1:" + port + "/service/api/v1");
        return scenario(name + " Health")
            .exec(http(name + " Health Check")
                .get("/health")
                .check(status().in(200, 503)))
            .injectOpen(rampUsers(clients).during(duration))
            .protocols(protocol);
    }

    {
        List<PopulationBuilder> populations = new ArrayList<>();

        // All 10 services.
        String[][] services = {
            {"SM-KMS", SM_KMS_PORT},
            {"SM-IM", SM_IM_PORT},
            {"JOSE-JA", JOSE_JA_PORT},
            {"PKI-CA", PKI_CA_PORT},
            {"Identity-AuthZ", AUTHZ_PORT},
            {"Identity-IdP", IDP_PORT},
            {"Identity-RS", RS_PORT},
            {"Identity-RP", RP_PORT},
            {"Identity-SPA", SPA_PORT},
            {"Skeleton-Template", SKELETON_PORT},
        };

        switch (PROFILE) {
            case "standard":
                // Health checks for all services
                for (String[] svc : services) {
                    populations.add(serviceHealthCheck(svc[0], svc[1], VIRTUAL_CLIENTS, DURATION_SECONDS));
                }
                // SM-KMS crypto operations
                HttpProtocolBuilder kmsProtocol = createServiceApiProtocol(
                    "https://127.0.0.1:" + SM_KMS_PORT + "/service/api/v1");
                populations.add(
                    scenario("Suite SM-KMS Crypto")
                        .feed(UUID_FEEDER)
                        .exec(createElasticKey(
                            "Suite-#{uuid}", "Suite load test",
                            "A256GCM/A256KW", DEFAULT_PROVIDER, false, true, "elasticKeyId"
                        ))
                        .exec(generateMaterialKey("elasticKeyId", "materialKeyId", null))
                        .exec(encrypt("elasticKeyId", CLEARTEXT, null, "ciphertext"))
                        .exec(decrypt("elasticKeyId", "ciphertext", "decrypted"))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(kmsProtocol)
                );
                break;

            case "stress":
                // High-concurrency health checks for all services
                for (String[] svc : services) {
                    populations.add(serviceHealthCheck(svc[0], svc[1], VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                }
                // SM-KMS crypto under stress
                HttpProtocolBuilder kmsStressProtocol = createServiceApiProtocol(
                    "https://127.0.0.1:" + SM_KMS_PORT + "/service/api/v1");
                populations.add(
                    scenario("Suite SM-KMS Crypto Stress")
                        .feed(UUID_FEEDER)
                        .exec(createElasticKey(
                            "SuiteStress-#{uuid}", "Suite stress test",
                            "A256GCM/A256KW", DEFAULT_PROVIDER, false, true, "elasticKeyId"
                        ))
                        .repeat(3).on(
                            io.gatling.javaapi.core.CoreDsl.exec(
                                generateMaterialKey("elasticKeyId", "materialKeyId", null))
                                .exec(encrypt("elasticKeyId", CLEARTEXT, null, "ciphertext"))
                                .exec(decrypt("elasticKeyId", "ciphertext", "decrypted"))
                        )
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS * 2).during(DURATION_SECONDS))
                        .protocols(kmsStressProtocol)
                );
                break;

            case "quick":
            default:
                // Quick validation: one health check per service
                for (String[] svc : services) {
                    populations.add(serviceHealthCheck(svc[0], svc[1], VIRTUAL_CLIENTS, DURATION_SECONDS));
                }
                break;
        }

        setUp(populations.toArray(new PopulationBuilder[0]))
            .assertions(
                global().responseTime().max().lt(5000),
                global().successfulRequests().percent().gte(90.0)
            );
    }
}
