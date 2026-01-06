// Copyright (c) 2025 Justin Cranford

// Package client provides reusable client utilities for cipher services.
// Extracted from E2E testing helpers to support client implementations.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	googleUuid "github.com/google/uuid"
)

// SendMessage sends a message to one or more receivers via /service/api/v1/messages/tx.
func SendMessage(client *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) (string, error) {
	receiverIDStrs := make([]string, len(receiverIDs))
	for i, id := range receiverIDs {
		receiverIDStrs[i] = id.String()
	}

	reqBody := map[string]any{
		"message":      message,
		"receiver_ids": receiverIDStrs,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := SendAuthenticatedRequest(client, http.MethodPut, baseURL+"/service/api/v1/messages/tx", token, reqJSON)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", decodeErrorResponse(resp)
	}

	var respBody map[string]string
	if err := DecodeJSONResponse(resp, &respBody); err != nil {
		return "", err
	}

	messageID, ok := respBody["message_id"]
	if !ok {
		return "", &ClientError{Message: "response missing message_id field"}
	}

	return messageID, nil
}

// ReceiveMessagesService retrieves messages for the specified receiver via /service/api/v1/messages/rx.
func ReceiveMessagesService(client *http.Client, baseURL, token string) ([]map[string]any, error) {
	resp, err := SendAuthenticatedRequest(client, http.MethodGet, baseURL+"/service/api/v1/messages/rx", token, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()

		return nil, decodeErrorResponse(resp)
	}

	var respBody map[string][]map[string]any
	if err := DecodeJSONResponse(resp, &respBody); err != nil {
		return nil, err
	}

	messages, ok := respBody["messages"]
	if !ok {
		return nil, &ClientError{Message: "response missing messages field"}
	}

	return messages, nil
}

// DeleteMessageService deletes a message via /service/api/v1/messages/:id.
func DeleteMessageService(client *http.Client, baseURL, messageID, token string) error {
	resp, err := SendAuthenticatedRequest(client, http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, token, nil)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return decodeErrorResponse(resp)
	}

	return nil
}

// SendMessageBrowser sends a message to one or more receivers via /browser/api/v1/messages/tx.
func SendMessageBrowser(client *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) (string, error) {
	receiverIDStrs := make([]string, len(receiverIDs))
	for i, id := range receiverIDs {
		receiverIDStrs[i] = id.String()
	}

	reqBody := map[string]any{
		"message":      message,
		"receiver_ids": receiverIDStrs,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := SendAuthenticatedRequest(client, http.MethodPut, baseURL+"/browser/api/v1/messages/tx", token, reqJSON)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", decodeErrorResponse(resp)
	}

	var respBody map[string]string
	if err := DecodeJSONResponse(resp, &respBody); err != nil {
		return "", err
	}

	messageID, ok := respBody["message_id"]
	if !ok {
		return "", &ClientError{Message: "response missing message_id field"}
	}

	return messageID, nil
}

// ReceiveMessagesBrowser retrieves messages for the specified receiver via /browser/api/v1/messages/rx.
func ReceiveMessagesBrowser(client *http.Client, baseURL, token string) ([]map[string]any, error) {
	resp, err := SendAuthenticatedRequest(client, http.MethodGet, baseURL+"/browser/api/v1/messages/rx", token, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()

		return nil, decodeErrorResponse(resp)
	}

	var respBody map[string][]map[string]any
	if err := DecodeJSONResponse(resp, &respBody); err != nil {
		return nil, err
	}

	messages, ok := respBody["messages"]
	if !ok {
		return nil, &ClientError{Message: "response missing messages field"}
	}

	return messages, nil
}

// DeleteMessageBrowser deletes a message via /browser/api/v1/messages/:id.
func DeleteMessageBrowser(client *http.Client, baseURL, messageID, token string) error {
	resp, err := SendAuthenticatedRequest(client, http.MethodDelete, baseURL+"/browser/api/v1/messages/"+messageID, token, nil)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return decodeErrorResponse(resp)
	}

	return nil
}

// ClientError represents an error returned by the cipher service.
type ClientError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *ClientError) Error() string {
	return e.Message
}

// decodeErrorResponse attempts to decode an error response from the service.
func decodeErrorResponse(resp *http.Response) error {
	var errorResp ClientError
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		// If we can't decode the error, return a generic error
		return &ClientError{Message: "request failed with status " + resp.Status}
	}

	return &errorResp
}

// SendAuthenticatedRequest sends an HTTP request with Bearer token authorization.
// Reusable for all services implementing JWT-based authentication.
func SendAuthenticatedRequest(client *http.Client, method, url, token string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// DecodeJSONResponse decodes HTTP response body into provided target struct.
// Reusable for all services returning JSON responses.
func DecodeJSONResponse(resp *http.Response, target any) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
