package cryptoutil;

import static cryptoutil.GatlingHttpUtil.createBrowserApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;

/**
 * Browser API Load Testing Simulation.
 * Tests OAuth 2.1 Authorization Code + PKCE flow under load.
 *
 * Example Commands:
 * <pre>
 * # Quick validation test (5 users, 30s)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation -Dprofile=quick -Dvirtualclients=5 -DdurationSeconds=30
 *
 * # Standard load test (50 users, 300s/5min)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation -Dprofile=standard -Dvirtualclients=50 -DdurationSeconds=300
 *
 * # Stress test (100 users, 600s/10min)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation -Dprofile=stress -Dvirtualclients=100 -DdurationSeconds=600
 * </pre>
 *
 * Scenarios:
 * - UI Health Check: GET /ui/swagger/doc.json
 * - OAuth Authorization Code + PKCE: Full browser-based flow
 * - Session Management: Cookie-based authentication
 * - CSRF Protection: X-CSRF-Token header validation
 */
public class BrowserApiSimulation extends Simulation {
    // Configuration from system properties
    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "5"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "30"));
    private static final String PORT = System.getProperty("port", "8180"); // Identity AuthZ default

    private static final SecureRandom SECURE_RANDOM = new SecureRandom();
    private static final Base64.Encoder BASE64_URL_ENCODER = Base64.getUrlEncoder().withoutPadding();

    // OAuth 2.1 endpoints
    private static final String AUTHORIZE_ENDPOINT = "/browser/v1/oauth/authorize";
    private static final String TOKEN_ENDPOINT = "/browser/v1/oauth/token";
    private static final String USERINFO_ENDPOINT = "/browser/v1/oauth/userinfo";
    private static final String LOGOUT_ENDPOINT = "/browser/v1/oauth/logout";

    // Test client credentials (must exist in database)
    private static final String CLIENT_ID = "test-browser-client";
    private static final String REDIRECT_URI = "https://localhost:8183/callback";
    private static final String SCOPE = "openid profile email";

    // Test user credentials (must exist in database)
    private static final String USERNAME = "test@cryptoutil.local";
    private static final String PASSWORD = "Test123!@#";

    /**
     * Generate PKCE code verifier (43-128 character random string).
     */
    private static String generateCodeVerifier() {
        byte[] bytes = new byte[32]; // 32 bytes = 43 base64url characters
        SECURE_RANDOM.nextBytes(bytes);
        return BASE64_URL_ENCODER.encodeToString(bytes);
    }

    /**
     * Generate PKCE code challenge from verifier (SHA-256 hash, base64url encoded).
     */
    private static String generateCodeChallenge(String codeVerifier) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(codeVerifier.getBytes(StandardCharsets.US_ASCII));
            return BASE64_URL_ENCODER.encodeToString(hash);
        } catch (Exception e) {
            throw new RuntimeException("Failed to generate code challenge", e);
        }
    }

    /**
     * Scenario: UI Health Check - Test /ui/swagger/doc.json endpoint.
     */
    private static ChainBuilder uiHealthCheckScenario() {
        return exec(
            http("UI Health Check")
                .get("/ui/swagger/doc.json")
                .header("Accept", "application/json")
                .check(status().is(200))
                .check(jsonPath("$.openapi").exists())
                .check(jsonPath("$.info.title").exists())
        );
    }

    /**
     * Scenario: OAuth 2.1 Authorization Code + PKCE Flow.
     *
     * Steps:
     * 1. Generate PKCE code_verifier and code_challenge
     * 2. GET /oauth/authorize (receive redirect to login)
     * 3. POST /oauth/login (authenticate user)
     * 4. GET /oauth/authorize (receive authorization code)
     * 5. POST /oauth/token (exchange code for access token)
     * 6. GET /oauth/userinfo (validate access token)
     * 7. POST /oauth/logout (end session)
     */
    private static ChainBuilder oauthAuthorizationCodePkceFlow() {
        return exec(session -> {
            // Generate PKCE parameters
            String codeVerifier = generateCodeVerifier();
            String codeChallenge = generateCodeChallenge(codeVerifier);
            String state = UUID.randomUUID().toString();
            String nonce = UUID.randomUUID().toString();

            return session
                .set("codeVerifier", codeVerifier)
                .set("codeChallenge", codeChallenge)
                .set("state", state)
                .set("nonce", nonce);
        })
        // Step 1: Initial authorization request (should redirect to login)
        .exec(
            http("OAuth Authorize - Initial")
                .get(AUTHORIZE_ENDPOINT)
                .queryParam("client_id", CLIENT_ID)
                .queryParam("redirect_uri", REDIRECT_URI)
                .queryParam("response_type", "code")
                .queryParam("scope", SCOPE)
                .queryParam("state", "#{state}")
                .queryParam("nonce", "#{nonce}")
                .queryParam("code_challenge", "#{codeChallenge}")
                .queryParam("code_challenge_method", "S256")
                .header("Accept", "text/html")
                .check(status().in(200, 302))
                .check(headerRegex("Set-Cookie", "session_id=([^;]+)").optional().saveAs("sessionCookie"))
                .check(headerRegex("X-CSRF-Token", "(.+)").optional().saveAs("csrfToken"))
        )
        // Step 2: User login (POST credentials)
        .exec(
            http("OAuth Login")
                .post("/browser/v1/oauth/login")
                .formParam("username", USERNAME)
                .formParam("password", PASSWORD)
                .header("Content-Type", "application/x-www-form-urlencoded")
                .header("X-CSRF-Token", "#{csrfToken}")
                .header("Cookie", "session_id=#{sessionCookie}")
                .check(status().in(200, 302))
                .check(headerRegex("Location", "code=([^&]+)").optional().saveAs("authorizationCode"))
        )
        // Step 3: Consent screen (if required) - typically auto-approved for trusted clients
        .doIf(session -> session.contains("authorizationCode") && session.getString("authorizationCode") == null).then(
            exec(
                http("OAuth Consent")
                    .post("/browser/v1/oauth/consent")
                    .formParam("approve", "true")
                    .header("X-CSRF-Token", "#{csrfToken}")
                    .header("Cookie", "session_id=#{sessionCookie}")
                    .check(status().in(200, 302))
                    .check(headerRegex("Location", "code=([^&]+)").saveAs("authorizationCode"))
            )
        )
        // Step 4: Exchange authorization code for access token
        .exec(
            http("OAuth Token Exchange")
                .post(TOKEN_ENDPOINT)
                .formParam("grant_type", "authorization_code")
                .formParam("client_id", CLIENT_ID)
                .formParam("code", "#{authorizationCode}")
                .formParam("redirect_uri", REDIRECT_URI)
                .formParam("code_verifier", "#{codeVerifier}")
                .header("Content-Type", "application/x-www-form-urlencoded")
                .header("Cookie", "session_id=#{sessionCookie}")
                .check(status().is(200))
                .check(jsonPath("$.access_token").saveAs("accessToken"))
                .check(jsonPath("$.id_token").optional().saveAs("idToken"))
                .check(jsonPath("$.refresh_token").optional().saveAs("refreshToken"))
        )
        // Step 5: Call userinfo endpoint to validate access token
        .exec(
            http("OAuth UserInfo")
                .get(USERINFO_ENDPOINT)
                .header("Authorization", "Bearer #{accessToken}")
                .header("Cookie", "session_id=#{sessionCookie}")
                .check(status().is(200))
                .check(jsonPath("$.sub").exists())
                .check(jsonPath("$.email").exists())
        )
        // Step 6: Logout
        .exec(
            http("OAuth Logout")
                .post(LOGOUT_ENDPOINT)
                .header("X-CSRF-Token", "#{csrfToken}")
                .header("Cookie", "session_id=#{sessionCookie}")
                .check(status().in(200, 302))
        );
    }

    /**
     * Scenario: Certificate Request Workflow (CA integration).
     * Tests browser-based CSR submission and certificate retrieval.
     */
    private static ChainBuilder certificateRequestFlow() {
        return exec(session -> session.set("csrCommonName", "test-" + UUID.randomUUID()))
        // Step 1: Generate CSR (browser would do this via JavaScript)
        .exec(
            http("CA Generate CSR")
                .post("/browser/v1/ca/csr/generate")
                .body(StringBody("""
                    {
                        "subject": {
                            "commonName": "#{csrCommonName}",
                            "organization": "Cryptoutil Test",
                            "country": "US"
                        },
                        "keyAlgorithm": "ecdsa",
                        "keySize": 256
                    }
                    """))
                .header("Content-Type", "application/json")
                .header("Cookie", "session_id=#{sessionCookie}")
                .header("X-CSRF-Token", "#{csrfToken}")
                .check(status().is(200))
                .check(jsonPath("$.csrPem").saveAs("csrPem"))
        )
        // Step 2: Submit CSR for signing
        .exec(
            http("CA Submit CSR")
                .post("/browser/v1/ca/certificates")
                .body(StringBody("""
                    {
                        "csr": "#{csrPem}",
                        "profile": "server",
                        "validityDays": 365
                    }
                    """))
                .header("Content-Type", "application/json")
                .header("Cookie", "session_id=#{sessionCookie}")
                .header("X-CSRF-Token", "#{csrfToken}")
                .check(status().is(201))
                .check(jsonPath("$.certificateId").saveAs("certificateId"))
        )
        // Step 3: Retrieve signed certificate
        .exec(
            http("CA Get Certificate")
                .get("/browser/v1/ca/certificates/#{certificateId}")
                .header("Cookie", "session_id=#{sessionCookie}")
                .check(status().is(200))
                .check(jsonPath("$.certificatePem").exists())
        );
    }

    /**
     * Build population based on profile.
     */
    private PopulationBuilder buildPopulation(String profile, int virtualClients, int durationSeconds) {
        ScenarioBuilder scenario;

        switch (profile) {
            case "quick":
                // Quick validation: UI health check only
                scenario = scenario("Quick Health Check")
                    .exec(uiHealthCheckScenario());
                break;

            case "standard":
                // Standard load: OAuth flow
                scenario = scenario("Standard OAuth Flow")
                    .exec(oauthAuthorizationCodePkceFlow());
                break;

            case "certificate":
                // Certificate workflow
                scenario = scenario("Certificate Request Flow")
                    .exec(oauthAuthorizationCodePkceFlow())
                    .exec(certificateRequestFlow());
                break;

            case "stress":
                // Stress test: Mixed scenarios
                scenario = scenario("Stress Test - Mixed")
                    .randomSwitch().on(
                        percent(50.0).then(exec(oauthAuthorizationCodePkceFlow())),
                        percent(30.0).then(exec(certificateRequestFlow())),
                        percent(20.0).then(exec(uiHealthCheckScenario()))
                    );
                break;

            default:
                throw new IllegalArgumentException("Unknown profile: " + profile);
        }

        // Ramp users over duration
        return scenario.injectOpen(rampUsers(virtualClients).during(durationSeconds));
    }

    {
        HttpProtocolBuilder httpProtocol = createBrowserApiProtocol(PORT);

        setUp(
            buildPopulation(PROFILE, VIRTUAL_CLIENTS, DURATION_SECONDS)
                .protocols(httpProtocol)
        ).assertions(
            // Performance targets
            global().responseTime().percentile3().lt(500),  // p95 <500ms
            global().responseTime().percentile4().lt(1000), // p99 <1000ms
            global().successfulRequests().percent().gt(95.0) // >95% success rate
        );
    }
}
