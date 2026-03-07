#!/usr/bin/env python3
"""Write assertions package files without BOM."""
import os

base = os.path.join("internal", "apps", "template", "service", "testing", "assertions")
os.makedirs(base, exist_ok=True)

assertions_go = (
    "// Copyright (c) 2025 Justin Cranford\n"
    "//\n"
    "\n"
    "// Package assertions provides reusable HTTP response assertion helpers for cryptoutil service tests.\n"
    "// Each helper reads and closes the response body; do not call Body.Close() separately.\n"
    "package assertions\n"
    "\n"
    "import (\n"
    '\t"encoding/json"\n'
    '\t"io"\n'
    '\t"net/http"\n'
    '\t"strings"\n'
    '\t"testing"\n'
    "\n"
    '\t"github.com/stretchr/testify/assert"\n'
    '\t"github.com/stretchr/testify/require"\n'
    "\n"
    '\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n'
    ")\n"
    "\n"
    "// AssertHealthy asserts a response is 200 OK with body {\"status\":\"healthy\"}.\n"
    "// The response body is read and closed by this function.\n"
    "func AssertHealthy(t testing.TB, resp *http.Response) {\n"
    "\tt.Helper()\n"
    "\n"
    "\trequire.NotNil(t, resp)\n"
    "\tassert.Equal(t, http.StatusOK, resp.StatusCode)\n"
    "\n"
    "\tbody, err := io.ReadAll(resp.Body)\n"
    "\trequire.NoError(t, err)\n"
    "\trequire.NoError(t, resp.Body.Close())\n"
    "\n"
    "\tvar result map[string]string\n"
    "\n"
    "\trequire.NoError(t, json.Unmarshal(body, &result))\n"
    '\tassert.Equal(t, cryptoutilSharedMagic.DockerServiceHealthHealthy, result[cryptoutilSharedMagic.StringStatus], "health status should be healthy")\n'
    "}\n"
    "\n"
    "// AssertErrorResponse asserts a response has the expected HTTP status code and a valid JSON error body.\n"
    "// The response body is read and closed by this function.\n"
    "func AssertErrorResponse(t testing.TB, resp *http.Response, code int) {\n"
    "\tt.Helper()\n"
    "\n"
    "\trequire.NotNil(t, resp)\n"
    "\tassert.Equal(t, code, resp.StatusCode)\n"
    "\n"
    "\tbody, err := io.ReadAll(resp.Body)\n"
    "\trequire.NoError(t, err)\n"
    "\trequire.NoError(t, resp.Body.Close())\n"
    "\n"
    "\tvar result map[string]any\n"
    "\n"
    '\trequire.NoError(t, json.Unmarshal(body, &result), "error response body must be valid JSON")\n'
    "}\n"
    "\n"
    "// AssertTraceID asserts the response includes a non-empty X-Request-Id header.\n"
    "func AssertTraceID(t testing.TB, resp *http.Response) {\n"
    "\tt.Helper()\n"
    "\n"
    "\trequire.NotNil(t, resp)\n"
    '\tassert.NotEmpty(t, resp.Header.Get("X-Request-Id"), "X-Request-Id header should be present")\n'
    "}\n"
    "\n"
    "// AssertJSONContentType asserts the response Content-Type contains \"application/json\".\n"
    "func AssertJSONContentType(t testing.TB, resp *http.Response) {\n"
    "\tt.Helper()\n"
    "\n"
    "\trequire.NotNil(t, resp)\n"
    '\tcontentType := resp.Header.Get("Content-Type")\n'
    '\tassert.True(t, strings.Contains(contentType, "application/json"), "Content-Type should contain application/json, got: %s", contentType)\n'
    "}\n"
)

assertions_test_go = (
    "// Copyright (c) 2025 Justin Cranford\n"
    "//\n"
    "\n"
    "package assertions_test\n"
    "\n"
    "import (\n"
    '\t"io"\n'
    '\t"net/http"\n'
    '\t"strings"\n'
    '\t"testing"\n'
    "\n"
    '\t"github.com/stretchr/testify/assert"\n'
    "\n"
    '\tcryptoutilTestingAssertions "cryptoutil/internal/apps/template/service/testing/assertions"\n'
    ")\n"
    "\n"
    "// newTestResponse builds a minimal *http.Response for assertion tests.\n"
    "func newTestResponse(statusCode int, contentType, body string) *http.Response {\n"
    "\theader := http.Header{}\n"
    "\n"
    "\tif contentType != \"\" {\n"
    '\t\theader.Set("Content-Type", contentType)\n'
    "\t}\n"
    "\n"
    "\treturn &http.Response{\n"
    "\t\tStatusCode: statusCode,\n"
    "\t\tHeader:     header,\n"
    "\t\tBody:       io.NopCloser(strings.NewReader(body)),\n"
    "\t}\n"
    "}\n"
    "\n"
    "func TestAssertHealthy_Success(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    '\tresp := newTestResponse(http.StatusOK, "application/json", `{"status":"healthy"}`)\n'
    "\tcryptoutilTestingAssertions.AssertHealthy(t, resp)\n"
    "}\n"
    "\n"
    "func TestAssertHealthy_WrongStatus(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tmockT := &mockTB{}\n"
    '\tresp := newTestResponse(http.StatusServiceUnavailable, "application/json", `{"status":"shutting down"}`)\n'
    "\tcryptoutilTestingAssertions.AssertHealthy(mockT, resp)\n"
    "\n"
    '\tassert.True(t, mockT.hasFailure(), "should fail for non-200 status")\n'
    "}\n"
    "\n"
    "func TestAssertErrorResponse_MatchingCode(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    '\tresp := newTestResponse(http.StatusBadRequest, "application/json", `{"code":"INVALID","message":"bad request"}`)\n'
    "\tcryptoutilTestingAssertions.AssertErrorResponse(t, resp, http.StatusBadRequest)\n"
    "}\n"
    "\n"
    "func TestAssertErrorResponse_WrongCode(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tmockT := &mockTB{}\n"
    '\tresp := newTestResponse(http.StatusNotFound, "application/json", `{"code":"NOT_FOUND","message":"not found"}`)\n'
    "\tcryptoutilTestingAssertions.AssertErrorResponse(mockT, resp, http.StatusBadRequest)\n"
    "\n"
    '\tassert.True(t, mockT.hasFailure(), "should fail when status code does not match")\n'
    "}\n"
    "\n"
    "func TestAssertTraceID_Present(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    '\tresp := newTestResponse(http.StatusOK, "application/json", "{}")\n'
    '\tresp.Header.Set("X-Request-Id", "abc-123")\n'
    "\tcryptoutilTestingAssertions.AssertTraceID(t, resp)\n"
    "}\n"
    "\n"
    "func TestAssertTraceID_Missing(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tmockT := &mockTB{}\n"
    '\tresp := newTestResponse(http.StatusOK, "application/json", "{}")\n'
    "\tcryptoutilTestingAssertions.AssertTraceID(mockT, resp)\n"
    "\n"
    '\tassert.True(t, mockT.hasFailure(), "should fail when X-Request-Id header is absent")\n'
    "}\n"
    "\n"
    "func TestAssertJSONContentType_Present(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    '\tresp := newTestResponse(http.StatusOK, "application/json; charset=utf-8", "{}")\n'
    "\tcryptoutilTestingAssertions.AssertJSONContentType(t, resp)\n"
    "}\n"
    "\n"
    "func TestAssertJSONContentType_Missing(t *testing.T) {\n"
    "\tt.Parallel()\n"
    "\n"
    "\tmockT := &mockTB{}\n"
    '\tresp := newTestResponse(http.StatusOK, "text/plain", "{}")\n'
    "\tcryptoutilTestingAssertions.AssertJSONContentType(mockT, resp)\n"
    "\n"
    '\tassert.True(t, mockT.hasFailure(), "should fail when Content-Type is not application/json")\n'
    "}\n"
    "\n"
    "// mockTB captures assertion failures without aborting the test.\n"
    "type mockTB struct {\n"
    "\ttesting.TB\n"
    "\tfailed bool\n"
    "}\n"
    "\n"
    "func (m *mockTB) Helper() {}\n"
    "\n"
    "func (m *mockTB) Errorf(_ string, _ ...any) {\n"
    "\tm.failed = true\n"
    "}\n"
    "\n"
    "func (m *mockTB) Fatalf(_ string, _ ...any) {\n"
    "\tm.failed = true\n"
    "}\n"
    "\n"
    "func (m *mockTB) FailNow() {\n"
    "\tm.failed = true\n"
    "}\n"
    "\n"
    "func (m *mockTB) hasFailure() bool {\n"
    "\treturn m.failed\n"
    "}\n"
)

for fname, content in [("assertions.go", assertions_go), ("assertions_test.go", assertions_test_go)]:
    path = os.path.join(base, fname)
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(content)
    print(f"Written: {path}")
