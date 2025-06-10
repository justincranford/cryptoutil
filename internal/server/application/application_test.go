package application

import (
	"cryptoutil/internal/client"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testServerHostname = "localhost"
	testServerPort     = 8081
	testServerBaseUrl  = "http://" + testServerHostname + ":" + strconv.Itoa(testServerPort) + "/"
)

func TestHttpGetHttp200(t *testing.T) {
	start, stop, err := StartServerApplication("localhost", testServerPort, true)
	if err != nil {
		t.Fatalf("failed to start server application: %v", err)
	}
	go start()
	defer stop()
	client.WaitUntilReady(testServerBaseUrl, 5*time.Second, 100*time.Millisecond)

	testCases := []struct {
		name string
		url  string
	}{
		{name: "Swagger UI root", url: testServerBaseUrl + "swagger"},
		{name: "Swagger UI index.html", url: testServerBaseUrl + "swagger/index.html"},
		{name: "OpenAPI Spec", url: testServerBaseUrl + "swagger/doc.json"},
		{name: "GET Key Pools", url: testServerBaseUrl + "keypools"},
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
