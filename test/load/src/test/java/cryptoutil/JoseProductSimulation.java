package cryptoutil;

import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
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
 * JOSE Product Load Testing Simulation.
 * Tests jose-ja (JWK Authority) service: elastic JWK management, key rotation, JWKS endpoints.
 *
 * <p>Example Commands:
 * <pre>
 * # Quick validation (health + JWKS endpoint)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.JoseProductSimulation -Dprofile=quick
 *
 * # Standard load test (JWK lifecycle operations)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.JoseProductSimulation -Dprofile=standard -Dvirtualclients=5
 *
 * # Stress test (concurrent key operations)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.JoseProductSimulation -Dprofile=stress -Dvirtualclients=20 -DdurationSeconds=120
 * </pre>
 */
public class JoseProductSimulation extends Simulation {

    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final String JOSE_JA_PORT = System.getProperty("joseJaPort", "8200");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "1"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "60"));

    private static final String JOSE_JA_BASE_URL = "https://127.0.0.1:" + JOSE_JA_PORT + "/service/api/v1";

    private static final Iterator<Map<String, Object>> UUID_FEEDER =
        Stream.generate((Supplier<Map<String, Object>>) () -> Map.of("uuid", UUID.randomUUID().toString()))
            .iterator();

    /**
     * Create an elastic JWK set.
     */
    private static ChainBuilder createElasticJwks(String saveIdAs) {
        return io.gatling.javaapi.core.CoreDsl.exec(
            http("Create Elastic JWKS")
                .post("/elastic-jwks")
                .body(io.gatling.javaapi.core.CoreDsl.StringBody("""
                    {
                        "name": "load-test-jwks",
                        "description": "Gatling load test JWKS",
                        "algorithm": "RS256",
                        "use": "sig"
                    }
                    """))
                .check(status().is(200))
                .check(jsonPath("$.id").saveAs(saveIdAs))
        );
    }

    /**
     * Retrieve the public JWKS endpoint.
     */
    private static ChainBuilder getPublicJwks(String jwksIdVar) {
        return io.gatling.javaapi.core.CoreDsl.exec(
            http("Get Public JWKS")
                .get("/elastic-jwks/#{" + jwksIdVar + "}/jwks.json")
                .check(status().is(200))
                .check(jsonPath("$.keys").exists())
        );
    }

    /**
     * Rotate a key in the elastic JWKS.
     */
    private static ChainBuilder rotateKey(String jwksIdVar) {
        return io.gatling.javaapi.core.CoreDsl.exec(
            http("Rotate JWKS Key")
                .post("/elastic-jwks/#{" + jwksIdVar + "}/rotate")
                .check(status().in(200, 201))
        );
    }

    {
        List<PopulationBuilder> populations = new ArrayList<>();
        HttpProtocolBuilder protocol = createServiceApiProtocol(JOSE_JA_BASE_URL);

        switch (PROFILE) {
            case "standard":
                // JWK lifecycle: create, get, rotate
                populations.add(
                    scenario("JOSE-JA JWK Lifecycle")
                        .feed(UUID_FEEDER)
                        .exec(createElasticJwks("jwksId"))
                        .pause(1)
                        .exec(getPublicJwks("jwksId"))
                        .pause(1)
                        .exec(rotateKey("jwksId"))
                        .pause(1)
                        .exec(getPublicJwks("jwksId"))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(protocol)
                );
                break;

            case "stress":
                // Concurrent JWKS operations
                populations.add(
                    scenario("JOSE-JA JWKS Stress")
                        .feed(UUID_FEEDER)
                        .exec(createElasticJwks("jwksId"))
                        .repeat(5).on(
                            io.gatling.javaapi.core.CoreDsl.exec(rotateKey("jwksId"))
                                .pause(1)
                                .exec(getPublicJwks("jwksId"))
                        )
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(protocol)
                );
                break;

            case "quick":
            default:
                // Health check + basic JWKS read
                populations.add(
                    scenario("JOSE-JA Quick Validation")
                        .exec(http("JOSE-JA Service Health")
                            .get("/health")
                            .check(status().in(200, 503)))
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
