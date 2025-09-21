package application

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilConfig "cryptoutil/internal/common/config"

	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("application_test")
	testServerPublicUrl  = testSettings.BindPublicProtocol + "://" + testSettings.BindPublicAddress + ":" + strconv.Itoa(int(testSettings.BindPublicPort))
	testServerPrivateUrl = testSettings.BindPrivateProtocol + "://" + testSettings.BindPrivateAddress + ":" + strconv.Itoa(int(testSettings.BindPrivatePort))
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	if exitCode != 0 {
		fmt.Printf("Tests failed with exit code %d\n", exitCode)
	}
}

func TestHttpGetHttp200(t *testing.T) {
	startServerListenerApplication, err := StartServerListenerApplication(testSettings)
	if err != nil {
		t.Fatalf("failed to start server application: %v", err)
	}
	go startServerListenerApplication.StartFunction()
	defer startServerListenerApplication.ShutdownFunction()
	cryptoutilClient.WaitUntilReady(&testServerPrivateUrl, 3*time.Second, 100*time.Millisecond, startServerListenerApplication.PrivateTLSServer.RootCAsPool)

	testCases := []struct {
		name       string
		url        string
		tlsRootCAs *x509.CertPool
	}{
		{name: "Swagger UI root", url: testServerPublicUrl + "/ui/swagger", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool},
		{name: "Swagger UI index.html", url: testServerPublicUrl + "/ui/swagger/index.html", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool},
		{name: "OpenAPI Spec", url: testServerPublicUrl + "/ui/swagger/doc.json", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool},
		{name: "GET Elastic Keys", url: testServerPublicUrl + testSettings.PublicServiceAPIContextPath + "/elastickeys", tlsRootCAs: startServerListenerApplication.PublicTLSServer.RootCAsPool},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			contentBytes, err := httpGetResponseBytes(t, http.StatusOK, testCase.url, testCase.tlsRootCAs)
			var contentString string
			if contentBytes != nil {
				contentString = strings.Replace(string(contentBytes), "\n", " ", -1)
			}
			if err == nil {
				t.Logf("PASS: %s, Contents: %s", testCase.url, contentString)
			} else {
				t.Errorf("FAILED: %s, Contents: %s, Error: %v", testCase.url, contentString, err)
			}
		})
	}
}

func httpGetResponseBytes(t *testing.T, expectedStatusCode int, url string, rootCAsPool *x509.CertPool) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err, "failed to create GET request")
	req.Header.Set("Accept", "*/*")

	// Create HTTP client with appropriate TLS configuration
	client := &http.Client{}
	if strings.HasPrefix(url, "https://") {
		transport := &http.Transport{}
		if rootCAsPool != nil {
			// Use provided root CA pool for server certificate validation
			transport.TLSClientConfig = &tls.Config{
				RootCAs:    rootCAsPool,
				MinVersion: tls.VersionTLS12,
			}
		} else {
			// Use system root CA pool for certificate validation
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
		client.Transport = transport
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "failed to make GET request")
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "HTTP Status code: "+strconv.Itoa(resp.StatusCode)+", failed to read error response body")
	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("HTTP Status code: %d, error response body: %v", resp.StatusCode, string(body))
	}
	t.Logf("HTTP Status code: %d, response body: %d bytes", resp.StatusCode, len(body))
	return body, nil
}
