// Copyright (c) 2025 Justin Cranford
//

package assertions_test

import (
	"io"
	http "net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilTestingAssertions "cryptoutil/internal/apps/framework/service/testing/assertions"
)

// newTestResponse builds a minimal *http.Response for assertion tests.
func newTestResponse(statusCode int, contentType, body string) *http.Response {
	header := http.Header{}

	if contentType != "" {
		header.Set("Content-Type", contentType)
	}

	return &http.Response{
		StatusCode: statusCode,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestAssertHealthy_Success(t *testing.T) {
	t.Parallel()

	resp := newTestResponse(http.StatusOK, "application/json", `{"status":"healthy"}`)

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertHealthy(t, resp)
}

func TestAssertHealthy_WrongStatus(t *testing.T) {
	t.Parallel()

	mockT := &mockTB{}
	resp := newTestResponse(http.StatusServiceUnavailable, "application/json", `{"status":"shutting down"}`)

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertHealthy(mockT, resp)

	require.True(t, mockT.hasFailure(), "should fail for non-200 status")
}

func TestAssertErrorResponse_MatchingCode(t *testing.T) {
	t.Parallel()

	resp := newTestResponse(http.StatusBadRequest, "application/json", `{"code":"INVALID","message":"bad request"}`)

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertErrorResponse(t, resp, http.StatusBadRequest)
}

func TestAssertErrorResponse_WrongCode(t *testing.T) {
	t.Parallel()

	mockT := &mockTB{}
	resp := newTestResponse(http.StatusNotFound, "application/json", `{"code":"NOT_FOUND","message":"not found"}`)

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertErrorResponse(mockT, resp, http.StatusBadRequest)

	require.True(t, mockT.hasFailure(), "should fail when status code does not match")
}

func TestAssertTraceID_Present(t *testing.T) {
	t.Parallel()

	resp := newTestResponse(http.StatusOK, "application/json", "{}")

	t.Cleanup(func() { _ = resp.Body.Close() })

	resp.Header.Set("X-Request-Id", "abc-123")
	cryptoutilTestingAssertions.AssertTraceID(t, resp)
}

func TestAssertTraceID_Missing(t *testing.T) {
	t.Parallel()

	mockT := &mockTB{}
	resp := newTestResponse(http.StatusOK, "application/json", "{}")

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertTraceID(mockT, resp)

	require.True(t, mockT.hasFailure(), "should fail when X-Request-Id header is absent")
}

func TestAssertJSONContentType_Present(t *testing.T) {
	t.Parallel()

	resp := newTestResponse(http.StatusOK, "application/json; charset=utf-8", "{}")

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertJSONContentType(t, resp)
}

func TestAssertJSONContentType_Missing(t *testing.T) {
	t.Parallel()

	mockT := &mockTB{}
	resp := newTestResponse(http.StatusOK, "text/plain", "{}")

	t.Cleanup(func() { _ = resp.Body.Close() })

	cryptoutilTestingAssertions.AssertJSONContentType(mockT, resp)

	require.True(t, mockT.hasFailure(), "should fail when Content-Type is not application/json")
}

// mockTB captures assertion failures without aborting the test.
type mockTB struct {
	testing.TB
	failed bool
}

func (m *mockTB) Helper() {}

func (m *mockTB) Name() string { return "mockTB" }

func (m *mockTB) Log(_ ...any) {}

func (m *mockTB) Cleanup(f func()) { f() }

func (m *mockTB) Errorf(_ string, _ ...any) {
	m.failed = true
}

func (m *mockTB) Fatalf(_ string, _ ...any) {
	m.failed = true
}

func (m *mockTB) FailNow() {
	m.failed = true
}

func (m *mockTB) hasFailure() bool {
	return m.failed
}
