package application

import (
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
	start, stop, err := StartServerListenerApplication(testSettings)
	if err != nil {
		t.Fatalf("failed to start server application: %v", err)
	}
	go start()
	defer stop()
	cryptoutilClient.WaitUntilReady(&testServerPrivateUrl, 3*time.Second, 100*time.Millisecond)

	testCases := []struct {
		name string
		url  string
	}{
		{name: "Swagger UI root", url: testServerPublicUrl + "/ui/swagger"},
		{name: "Swagger UI index.html", url: testServerPublicUrl + "/ui/swagger/index.html"},
		{name: "OpenAPI Spec", url: testServerPublicUrl + "/ui/swagger/doc.json"},
		{name: "GET Elastic Keys", url: testServerPublicUrl + testSettings.PublicUIAPIContextPath + "/elastickeys"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			contentBytes, err := httpGetResponseBytes(t, http.StatusOK, testCase.url)
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

func httpGetResponseBytes(t *testing.T, expectedStatusCode int, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err, "failed to create GET request")
	req.Header.Set("Accept", "*/*")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to make GET request")
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "HTTP Status code: "+strconv.Itoa(resp.StatusCode)+", failed to read error response body")
	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("HTTP Status code: %d, error response body: %v", resp.StatusCode, string(body))
	}
	t.Logf("HTTP Status code: %d, response body: %d bytes", resp.StatusCode, len(body))
	return body, nil
}
