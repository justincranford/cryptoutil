package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpGetHttp200(t *testing.T) {
	start, stop := NewServer("localhost:8080", true)
	go start()
	defer stop()

	testCases := []struct {
		name string
		url  string
	}{
		{name: "Swagger UI root", url: "http://localhost:8080/swagger"},
		{name: "Swagger UI index.html", url: "http://localhost:8080/swagger/index.html"},
		{name: "OpenAPI Spec", url: "http://localhost:8080/swagger/doc.json"},
		{name: "GET Key Pools", url: "http://localhost:8080/keypool"},
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
				t.Errorf("FAILED: %s, Contents: %s", testCase.url, contentString)
			}
		})
	}
}

func httpGetResponseBytes(t *testing.T, expectedStatusCode int, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if !assert.NoError(t, err) {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if !assert.NoError(t, err) {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return nil, fmt.Errorf("HTTP Status code: %d, failed to read error response body: %w", resp.StatusCode, err)
	} else if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("HTTP Status code: %d, error response body: %v", resp.StatusCode, body)
	}
	t.Logf("HTTP Status code: %d, response body: %d bytes", resp.StatusCode, len(body))
	return body, nil
}
