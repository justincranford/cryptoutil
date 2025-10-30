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
}
