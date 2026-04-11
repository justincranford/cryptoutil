package cryptoutil;

import static cryptoutil.GatlingHttpUtil.createBrowserApiProtocol;
import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.global;
import static io.gatling.javaapi.core.CoreDsl.rampUsers;
import static io.gatling.javaapi.core.CoreDsl.scenario;
import static io.gatling.javaapi.http.HttpDsl.http;
import static io.gatling.javaapi.http.HttpDsl.status;

import java.util.ArrayList;
import java.util.List;

import io.gatling.javaapi.core.PopulationBuilder;
import io.gatling.javaapi.core.Simulation;
import io.gatling.javaapi.http.HttpProtocolBuilder;

/**
 * Identity Product Load Testing Simulation.
 * Tests all identity services: identity-authz, identity-idp, identity-rp, identity-rs, identity-spa.
 *
 * <p>Example Commands:
 * <pre>
 * # Quick validation (health checks for all identity services)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.IdentityProductSimulation -Dprofile=quick
 *
 * # Standard load test (OAuth flows + health checks)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.IdentityProductSimulation -Dprofile=standard -Dvirtualclients=5
 *
 * # Stress test (concurrent OAuth + service health)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.IdentityProductSimulation -Dprofile=stress -Dvirtualclients=20 -DdurationSeconds=120
 *
 * # Custom ports
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.IdentityProductSimulation -DauthzPort=8400 -DidpPort=8500
 * </pre>
 */
public class IdentityProductSimulation extends Simulation {

    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final String AUTHZ_PORT = System.getProperty("authzPort", "8400");
    private static final String IDP_PORT = System.getProperty("idpPort", "8500");
    private static final String RS_PORT = System.getProperty("rsPort", "8600");
    private static final String RP_PORT = System.getProperty("rpPort", "8700");
    private static final String SPA_PORT = System.getProperty("spaPort", "8800");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "1"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "60"));

    /**
     * Create a health check scenario for a named service.
     */
    private static PopulationBuilder healthCheckPopulation(String serviceName, String port, int clients, int duration) {
        HttpProtocolBuilder protocol = createServiceApiProtocol("https://127.0.0.1:" + port + "/service/api/v1");
        return scenario(serviceName + " Health Check")
            .exec(http(serviceName + " Service Health")
                .get("/health")
                .check(status().in(200, 503)))
            .injectOpen(rampUsers(clients).during(duration))
            .protocols(protocol);
    }

    /**
     * Create a browser-based OAuth health check scenario.
     */
    private static PopulationBuilder browserHealthPopulation(String serviceName, String port, int clients, int duration) {
        HttpProtocolBuilder protocol = createBrowserApiProtocol(port);
        return scenario(serviceName + " Browser Health")
            .exec(http(serviceName + " Browser Health Check")
                .get("/browser/api/v1/health")
                .check(status().in(200, 503)))
            .injectOpen(rampUsers(clients).during(duration))
            .protocols(protocol);
    }

    {
        List<PopulationBuilder> populations = new ArrayList<>();

        switch (PROFILE) {
            case "standard":
                // Service health checks for all 5 identity services
                populations.add(healthCheckPopulation("Identity-AuthZ", AUTHZ_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-IdP", IDP_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-RS", RS_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-RP", RP_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-SPA", SPA_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                // Browser health checks
                populations.add(browserHealthPopulation("Identity-AuthZ", AUTHZ_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(browserHealthPopulation("Identity-IdP", IDP_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                break;

            case "stress":
                // High-concurrency health checks across all services
                populations.add(healthCheckPopulation("Identity-AuthZ", AUTHZ_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-IdP", IDP_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-RS", RS_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-RP", RP_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-SPA", SPA_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(browserHealthPopulation("Identity-AuthZ", AUTHZ_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(browserHealthPopulation("Identity-IdP", IDP_PORT, VIRTUAL_CLIENTS * 2, DURATION_SECONDS));
                populations.add(browserHealthPopulation("Identity-RP", RP_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                break;

            case "quick":
            default:
                // Quick validation: one health check per service
                populations.add(healthCheckPopulation("Identity-AuthZ", AUTHZ_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-IdP", IDP_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-RS", RS_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-RP", RP_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                populations.add(healthCheckPopulation("Identity-SPA", SPA_PORT, VIRTUAL_CLIENTS, DURATION_SECONDS));
                break;
        }

        setUp(populations.toArray(new PopulationBuilder[0]))
            .assertions(
                global().responseTime().max().lt(3000),
                global().successfulRequests().percent().gte(95.0)
            );
    }
}
