// Copyright (c) 2025 Justin Cranford

package client

import (
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.Must(googleUuid.NewV7())

	tests := []struct {
		name    string
		handler http.HandlerFunc
		sendFn  func(*http.Client, string, string, string, ...googleUuid.UUID) (string, error)
		wantID  string
		wantErr string
	}{
		{
			name: "service happy path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"message_id": messageID.String()})
			},
			sendFn: SendMessage,
			wantID: messageID.String(),
		},
		{
			name: "service unauthorized",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid token"})
			},
			sendFn:  SendMessage,
			wantErr: "Invalid token",
		},
		{
			name: "service missing message_id",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{})
			},
			sendFn:  SendMessage,
			wantErr: "response missing message_id field",
		},
		{
			name: "browser happy path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"message_id": messageID.String()})
			},
			sendFn: SendMessageBrowser,
			wantID: messageID.String(),
		},
		{
			name: "browser unauthorized",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid browser session"})
			},
			sendFn:  SendMessageBrowser,
			wantErr: "Invalid browser session",
		},
		{
			name: "browser missing message_id",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{})
			},
			sendFn:  SendMessageBrowser,
			wantErr: "response missing message_id field",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tc.handler)
			defer server.Close()

			receiver := googleUuid.New()

			resultID, err := tc.sendFn(http.DefaultClient, server.URL, "test message", "test-token", receiver)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantID, resultID)
		})
	}
}

func TestReceiveMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		handler   http.HandlerFunc
		receiveFn func(*http.Client, string, string) ([]map[string]any, error)
		wantLen   int
		wantErr   string
	}{
		{
			name: "service happy path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"messages": []map[string]any{
						{"id": googleUuid.New().String(), "message": "Hello"},
						{"id": googleUuid.New().String(), "message": "World"},
					},
				})
			},
			receiveFn: ReceiveMessagesService,
			wantLen:   2,
		},
		{
			name: "service unauthorized",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid token"})
			},
			receiveFn: ReceiveMessagesService,
			wantErr:   "Invalid token",
		},
		{
			name: "service missing messages field",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{})
			},
			receiveFn: ReceiveMessagesService,
			wantErr:   "response missing messages field",
		},
		{
			name: "service invalid json error fallback",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte("Internal Server Error"))
			},
			receiveFn: ReceiveMessagesService,
			wantErr:   "request failed with status",
		},
		{
			name: "browser happy path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"messages": []map[string]any{
						{"id": googleUuid.New().String(), "message": "Browser message 1"},
						{"id": googleUuid.New().String(), "message": "Browser message 2"},
						{"id": googleUuid.New().String(), "message": "Browser message 3"},
					},
				})
			},
			receiveFn: ReceiveMessagesBrowser,
			wantLen:   3,
		},
		{
			name: "browser unauthorized",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid browser token"})
			},
			receiveFn: ReceiveMessagesBrowser,
			wantErr:   "Invalid browser token",
		},
		{
			name: "browser missing messages field",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{})
			},
			receiveFn: ReceiveMessagesBrowser,
			wantErr:   "response missing messages field",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tc.handler)
			defer server.Close()

			messages, err := tc.receiveFn(http.DefaultClient, server.URL, "test-token")
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Len(t, messages, tc.wantLen)
		})
	}
}

func TestDeleteMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		handler  http.HandlerFunc
		deleteFn func(*http.Client, string, string, string) error
		wantErr  string
	}{
		{
			name: "service happy path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			deleteFn: DeleteMessageService,
		},
		{
			name: "service not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Error{Code: "NOT_FOUND", Message: "Message not found"})
			},
			deleteFn: DeleteMessageService,
			wantErr:  "Message not found",
		},
		{
			name: "browser happy path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			deleteFn: DeleteMessageBrowser,
		},
		{
			name: "browser not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Error{Code: "NOT_FOUND", Message: "Browser message not found"})
			},
			deleteFn: DeleteMessageBrowser,
			wantErr:  "Browser message not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tc.handler)
			defer server.Close()

			messageID := googleUuid.New()

			err := tc.deleteFn(http.DefaultClient, server.URL, messageID.String(), "test-token")
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	err := &Error{
		Code:    "TEST_ERROR",
		Message: "This is a test error",
		Details: map[string]string{"field": "value"},
	}

	require.Equal(t, "This is a test error", err.Error())
}
