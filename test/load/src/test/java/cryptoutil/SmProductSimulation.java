package cryptoutil;

import static cryptoutil.CryptoutiApi.CLEARTEXT;
import static cryptoutil.CryptoutiApi.DATA_KEY_ALGORITHMS;
import static cryptoutil.CryptoutiApi.DEFAULT_PROVIDER;
import static cryptoutil.CryptoutiApi.ELASTIC_KEY_ENCRYPTION_ALGORITHMS;
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
 * SM Product Load Testing Simulation.
 * Combines sm-kms (Key Management Service) and sm-im (Instant Messenger) load tests.
 *
 * <p>Example Commands:
 * <pre>
 * # Quick validation (health checks only)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmProductSimulation -Dprofile=quick
 *
 * # Standard load test (sm-kms crypto + sm-im health)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmProductSimulation -Dprofile=standard -Dvirtualclients=5
 *
 * # Stress test (all algorithms)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmProductSimulation -Dprofile=stress -Dvirtualclients=10 -DdurationSeconds=120
 *
 * # Custom ports
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SmProductSimulation -DsmKmsPort=8000 -DsmImPort=8100
 * </pre>
 */
public class SmProductSimulation extends Simulation {

    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final String SM_KMS_PORT = System.getProperty("smKmsPort", "8000");
    private static final String SM_IM_PORT = System.getProperty("smImPort", "8100");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "1"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "60"));

    private static final String SM_KMS_BASE_URL = "https://127.0.0.1:" + SM_KMS_PORT + "/service/api/v1";
    private static final String SM_IM_BASE_URL = "https://127.0.0.1:" + SM_IM_PORT + "/service/api/v1";

    private static final Iterator<Map<String, Object>> UUID_FEEDER =
        Stream.generate((Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
            .iterator();

    {
        List<PopulationBuilder> populations = new ArrayList<>();
        HttpProtocolBuilder kmsProtocol = createServiceApiProtocol(SM_KMS_BASE_URL);
        HttpProtocolBuilder imProtocol = createServiceApiProtocol(SM_IM_BASE_URL);

        switch (PROFILE) {
            case "standard":
                // sm-kms: cipher and signature scenarios
                for (String algo : ELASTIC_KEY_ENCRYPTION_ALGORITHMS) {
                    populations.add(
                        buildEncryptDecryptScenario(algo, CLEARTEXT)
                            .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                            .protocols(kmsProtocol)
                    );
                }
                // sm-im: health check scenario
                populations.add(
                    scenario("SM-IM Health Check")
                        .exec(http("SM-IM Service Health")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(imProtocol)
                );
                break;

            case "stress":
                // sm-kms: complete workflow with all algorithms
                for (String algo : ELASTIC_KEY_ENCRYPTION_ALGORITHMS) {
                    populations.add(
                        buildCompleteWorkflowScenario(algo, CLEARTEXT, DATA_KEY_ALGORITHMS)
                            .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                            .protocols(kmsProtocol)
                    );
                }
                for (String algo : ELASTIC_KEY_SIGNATURE_ALGORITHMS) {
                    populations.add(
                        buildSignVerifyScenario(algo, CLEARTEXT)
                            .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                            .protocols(kmsProtocol)
                    );
                }
                // sm-im: health check under load
                populations.add(
                    scenario("SM-IM Health Check Stress")
                        .exec(http("SM-IM Service Health")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS * 2).during(DURATION_SECONDS))
                        .protocols(imProtocol)
                );
                break;

            case "quick":
            default:
                // Quick validation: health checks for both services
                populations.add(
                    scenario("SM-KMS Quick Validation")
                        .feed(UUID_FEEDER)
                        .exec(createElasticKey(
                            "SmProduct-#{uuid}", "SM product quick test",
                            "A256GCM/A256KW", DEFAULT_PROVIDER, false, true, "elasticKeyId"
                        ))
                        .exec(generateMaterialKey("elasticKeyId", "materialKeyId", null))
                        .exec(encrypt("elasticKeyId", CLEARTEXT, null, "ciphertext"))
                        .exec(decrypt("elasticKeyId", "ciphertext", "decrypted"))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(kmsProtocol)
                );
                populations.add(
                    scenario("SM-IM Health Check")
                        .exec(http("SM-IM Service Health")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(imProtocol)
                );
                break;
        }

        setUp(populations.toArray(new PopulationBuilder[0]))
            .assertions(
                global().responseTime().max().lt(5000),
                global().successfulRequests().percent().gte(95.0)
            );
    }
}
