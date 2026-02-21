// Copyright (c) 2025 Justin Cranford

package client

import (
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSendMessage_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	closedURL := server.URL
	server.Close()

	receiver := googleUuid.New()

	_, err := SendMessage(http.DefaultClient, closedURL, "test", "token", receiver)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send message")
}

func TestSendMessage_DecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("not-valid-json")) //nolint:errcheck // Test handler.
	}))
	defer server.Close()

	receiver := googleUuid.New()

	_, err := SendMessage(http.DefaultClient, server.URL, "test", "token", receiver)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response")
}

func TestReceiveMessagesService_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	closedURL := server.URL
	server.Close()

	_, err := ReceiveMessagesService(http.DefaultClient, closedURL, "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to receive messages")
}

func TestReceiveMessagesService_DecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-valid-json")) //nolint:errcheck // Test handler.
	}))
	defer server.Close()

	_, err := ReceiveMessagesService(http.DefaultClient, server.URL, "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response")
}

func TestDeleteMessageService_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	closedURL := server.URL
	server.Close()

	err := DeleteMessageService(http.DefaultClient, closedURL, "msg-id", "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete message")
}

func TestSendMessageBrowser_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	closedURL := server.URL
	server.Close()

	receiver := googleUuid.New()

	_, err := SendMessageBrowser(http.DefaultClient, closedURL, "test", "token", receiver)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send message")
}

func TestSendMessageBrowser_DecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("not-valid-json")) //nolint:errcheck // Test handler.
	}))
	defer server.Close()

	receiver := googleUuid.New()

	_, err := SendMessageBrowser(http.DefaultClient, server.URL, "test", "token", receiver)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response")
}

func TestReceiveMessagesBrowser_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	closedURL := server.URL
	server.Close()

	_, err := ReceiveMessagesBrowser(http.DefaultClient, closedURL, "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to receive messages")
}

func TestReceiveMessagesBrowser_DecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-valid-json")) //nolint:errcheck // Test handler.
	}))
	defer server.Close()

	_, err := ReceiveMessagesBrowser(http.DefaultClient, server.URL, "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response")
}

func TestDeleteMessageBrowser_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	closedURL := server.URL
	server.Close()

	err := DeleteMessageBrowser(http.DefaultClient, closedURL, "msg-id", "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete message")
}

// TestSendMessage_MarshalError covers the json.Marshal error path.
func TestSendMessage_MarshalError(t *testing.T) {
	t.Parallel()

	// Save original and restore after test.
	originalFn := jsonMarshalFn

	defer func() { jsonMarshalFn = originalFn }()

	jsonMarshalFn = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("simulated marshal failure")
	}

	receiver := googleUuid.New()

	_, err := SendMessage(http.DefaultClient, "http://localhost", "test", "token", receiver)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to marshal request")
}
