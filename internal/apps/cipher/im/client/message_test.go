// Copyright (c) 2025 Justin Cranford

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSendMessage_HappyPath(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/service/api/v1/messages/tx", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var reqBody map[string]any

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		require.Equal(t, "Hello, World!", reqBody["message"])
		require.Len(t, reqBody["receiver_ids"], 2)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		respBody := map[string]string{"message_id": messageID.String()}
		_ = json.NewEncoder(w).Encode(respBody)
	}))
	defer server.Close()

	receiver1 := googleUuid.New()
	receiver2 := googleUuid.New()

	resultMessageID, err := SendMessage(
		http.DefaultClient,
		server.URL,
		"Hello, World!",
		"test-token",
		receiver1,
		receiver2,
	)

	require.NoError(t, err)
	require.Equal(t, messageID.String(), resultMessageID)
}

func TestSendMessage_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid token"})
	}))
	defer server.Close()

	receiver := googleUuid.New()

	_, err := SendMessage(
		http.DefaultClient,
		server.URL,
		"Test message",
		"invalid-token",
		receiver,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid token")
}

func TestReceiveMessagesService_HappyPath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/service/api/v1/messages/rx", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		respBody := map[string]any{
			"messages": []map[string]any{
				{"id": googleUuid.New().String(), "message": "Hello"},
				{"id": googleUuid.New().String(), "message": "World"},
			},
		}
		_ = json.NewEncoder(w).Encode(respBody)
	}))
	defer server.Close()

	messages, err := ReceiveMessagesService(
		http.DefaultClient,
		server.URL,
		"test-token",
	)

	require.NoError(t, err)
	require.Len(t, messages, 2)
}

func TestReceiveMessagesService_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid token"})
	}))
	defer server.Close()

	_, err := ReceiveMessagesService(
		http.DefaultClient,
		server.URL,
		"invalid-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid token")
}

func TestDeleteMessageService_HappyPath(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/service/api/v1/messages/"+messageID.String(), r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := DeleteMessageService(
		http.DefaultClient,
		server.URL,
		messageID.String(),
		"test-token",
	)

	require.NoError(t, err)
}

func TestDeleteMessageService_NotFound(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Error{Code: "NOT_FOUND", Message: "Message not found"})
	}))
	defer server.Close()

	err := DeleteMessageService(
		http.DefaultClient,
		server.URL,
		messageID.String(),
		"test-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Message not found")
}

func TestSendMessageBrowser_HappyPath(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/browser/api/v1/messages/tx", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var reqBody map[string]any

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		require.Equal(t, "Hello from browser", reqBody["message"])
		require.Len(t, reqBody["receiver_ids"], 1)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		respBody := map[string]string{"message_id": messageID.String()}
		_ = json.NewEncoder(w).Encode(respBody)
	}))
	defer server.Close()

	receiver := googleUuid.New()

	resultMessageID, err := SendMessageBrowser(
		http.DefaultClient,
		server.URL,
		"Hello from browser",
		"test-token",
		receiver,
	)

	require.NoError(t, err)
	require.Equal(t, messageID.String(), resultMessageID)
}

func TestReceiveMessagesBrowser_HappyPath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/browser/api/v1/messages/rx", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		respBody := map[string]any{
			"messages": []map[string]any{
				{"id": googleUuid.New().String(), "message": "Browser message 1"},
				{"id": googleUuid.New().String(), "message": "Browser message 2"},
				{"id": googleUuid.New().String(), "message": "Browser message 3"},
			},
		}
		_ = json.NewEncoder(w).Encode(respBody)
	}))
	defer server.Close()

	messages, err := ReceiveMessagesBrowser(
		http.DefaultClient,
		server.URL,
		"test-token",
	)

	require.NoError(t, err)
	require.Len(t, messages, 3)
}

func TestDeleteMessageBrowser_HappyPath(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/browser/api/v1/messages/"+messageID.String(), r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := DeleteMessageBrowser(
		http.DefaultClient,
		server.URL,
		messageID.String(),
		"test-token",
	)

	require.NoError(t, err)
}

func TestErrorError_ReturnsMessage(t *testing.T) {
	t.Parallel()

	err := &Error{
		Code:    "TEST_ERROR",
		Message: "This is a test error",
		Details: map[string]string{"field": "value"},
	}

	require.Equal(t, "This is a test error", err.Error())
}

func TestReceiveMessagesBrowser_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid browser token"})
	}))
	defer server.Close()

	_, err := ReceiveMessagesBrowser(
		http.DefaultClient,
		server.URL,
		"invalid-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid browser token")
}

func TestReceiveMessagesBrowser_MissingMessagesField(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// Return empty object without "messages" field.
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	_, err := ReceiveMessagesBrowser(
		http.DefaultClient,
		server.URL,
		"test-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response missing messages field")
}

func TestDeleteMessageBrowser_NotFound(t *testing.T) {
	t.Parallel()

	messageID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Error{Code: "NOT_FOUND", Message: "Browser message not found"})
	}))
	defer server.Close()

	err := DeleteMessageBrowser(
		http.DefaultClient,
		server.URL,
		messageID.String(),
		"test-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Browser message not found")
}

func TestSendMessageBrowser_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Error{Code: "UNAUTHORIZED", Message: "Invalid browser session"})
	}))
	defer server.Close()

	receiver := googleUuid.New()

	_, err := SendMessageBrowser(
		http.DefaultClient,
		server.URL,
		"Test message",
		"invalid-token",
		receiver,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid browser session")
}

func TestSendMessageBrowser_MissingMessageID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		// Return empty object without "message_id" field.
		_ = json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer server.Close()

	receiver := googleUuid.New()

	_, err := SendMessageBrowser(
		http.DefaultClient,
		server.URL,
		"Test message",
		"test-token",
		receiver,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response missing message_id field")
}

func TestSendMessage_MissingMessageID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		// Return empty object without "message_id" field.
		_ = json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer server.Close()

	receiver := googleUuid.New()

	_, err := SendMessage(
		http.DefaultClient,
		server.URL,
		"Test message",
		"test-token",
		receiver,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response missing message_id field")
}

func TestReceiveMessagesService_MissingMessagesField(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// Return empty object without "messages" field.
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	_, err := ReceiveMessagesService(
		http.DefaultClient,
		server.URL,
		"test-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response missing messages field")
}

func TestDecodeErrorResponse_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain")
		// Return invalid JSON to trigger fallback error.
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Use ReceiveMessagesService which will trigger decodeErrorResponse.
	_, err := ReceiveMessagesService(
		http.DefaultClient,
		server.URL,
		"test-token",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "request failed with status")
}
