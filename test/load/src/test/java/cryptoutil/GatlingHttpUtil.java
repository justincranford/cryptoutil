package cryptoutil;

import static io.gatling.javaapi.http.HttpDsl.http;

import io.gatling.javaapi.http.HttpProtocolBuilder;

/**
 * Gatling documentation: https://docs.gatling.io/reference/script/http/tls/
 *
 * "By default, Gatling uses a fake TrustManager that trusts everything. The reason
 * is we want to work out of the box with any server certificate, even when it's
 * issued from a custom Authority. Gatling is a load test tool, not your application
 * managing sensible data to be protected against a man-in-the-middle."
 *
 * To override this behavior (e.g., to use a specific truststore), configure the
 * ssl block in gatling.conf or use perUserKeyManagerFactory() on the protocol.
 */
public class GatlingHttpUtil {
    public static HttpProtocolBuilder createServiceApiProtocol(String baseUrl) {
        return http
            .baseUrl(baseUrl)
            .acceptHeader("application/json")
            .contentTypeHeader("application/json")
            .userAgentHeader("Gatling-Cryptoutil-Service-API-Test/1.0")
            .disableFollowRedirect();
    }

    /**
     * Create HTTP protocol for browser-based API testing.
     * Uses HTTPS, follows redirects (for OAuth flows), accepts HTML/JSON.
     */
    public static HttpProtocolBuilder createBrowserApiProtocol(String port) {
        return http
            .baseUrl("https://localhost:" + port)
            .acceptHeader("text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8")
            .acceptEncodingHeader("gzip, deflate")
            .acceptLanguageHeader("en-US,en;q=0.5")
            .userAgentHeader("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Gatling-Cryptoutil-Browser-API-Test/1.0")
            .inferHtmlResources() // Automatically fetch HTML resources
            .shareConnections(); // Reuse connections (browser-like behavior)
    }
}
