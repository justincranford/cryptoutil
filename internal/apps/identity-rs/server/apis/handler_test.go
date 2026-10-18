package apis

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterMinimalHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mux  bool
	}{
		{name: "nil mux"},
		{name: "registered handler", mux: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var mux *http.ServeMux
			if tc.mux {
				mux = http.NewServeMux()
			}

			RegisterMinimalHandler(mux)

			if tc.mux {
				response := httptest.NewRecorder()
				mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil))
				require.Equal(t, http.StatusNoContent, response.Code)
			}
		})
	}
}
