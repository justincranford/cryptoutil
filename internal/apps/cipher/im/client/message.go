// Copyright (c) 2025 Justin Cranford

// Package client provides reusable client utilities for cipher services.
// Extracted from E2E testing helpers to support client implementations.
package client

import (
	json "encoding/json"
	"fmt"
	http "net/http"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceClient "cryptoutil/internal/apps/template/service/client"
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

	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodPut, baseURL+"/service/api/v1/messages/tx", token, reqJSON)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", decodeErrorResponse(resp)
	}

	var respBody map[string]string
	if err := cryptoutilAppsTemplateServiceClient.DecodeJSONResponse(resp, &respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	messageID, ok := respBody["message_id"]
	if !ok {
		return "", &Error{Message: "response missing message_id field"}
	}

	return messageID, nil
}

// ReceiveMessagesService retrieves messages for the specified receiver via /service/api/v1/messages/rx.
func ReceiveMessagesService(client *http.Client, baseURL, token string) ([]map[string]any, error) {
	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodGet, baseURL+"/service/api/v1/messages/rx", token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()

		return nil, decodeErrorResponse(resp)
	}

	var respBody map[string][]map[string]any
	if err := cryptoutilAppsTemplateServiceClient.DecodeJSONResponse(resp, &respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	messages, ok := respBody["messages"]
	if !ok {
		return nil, &Error{Message: "response missing messages field"}
	}

	return messages, nil
}

// DeleteMessageService deletes a message via /service/api/v1/messages/:id.
func DeleteMessageService(client *http.Client, baseURL, messageID, token string) error {
	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, token, nil)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
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

	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodPut, baseURL+"/browser/api/v1/messages/tx", token, reqJSON)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", decodeErrorResponse(resp)
	}

	var respBody map[string]string
	if err := cryptoutilAppsTemplateServiceClient.DecodeJSONResponse(resp, &respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	messageID, ok := respBody["message_id"]
	if !ok {
		return "", &Error{Message: "response missing message_id field"}
	}

	return messageID, nil
}

// ReceiveMessagesBrowser retrieves messages for the specified receiver via /browser/api/v1/messages/rx.
func ReceiveMessagesBrowser(client *http.Client, baseURL, token string) ([]map[string]any, error) {
	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodGet, baseURL+"/browser/api/v1/messages/rx", token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()

		return nil, decodeErrorResponse(resp)
	}

	var respBody map[string][]map[string]any
	if err := cryptoutilAppsTemplateServiceClient.DecodeJSONResponse(resp, &respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	messages, ok := respBody["messages"]
	if !ok {
		return nil, &Error{Message: "response missing messages field"}
	}

	return messages, nil
}

// DeleteMessageBrowser deletes a message via /browser/api/v1/messages/:id.
func DeleteMessageBrowser(client *http.Client, baseURL, messageID, token string) error {
	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodDelete, baseURL+"/browser/api/v1/messages/"+messageID, token, nil)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return decodeErrorResponse(resp)
	}

	return nil
}

// Error represents an error returned by the cipher service.
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// decodeErrorResponse attempts to decode an error response from the service.
func decodeErrorResponse(resp *http.Response) error {
	var errorResp Error
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		// If we can't decode the error, return a generic error
		return &Error{Message: "request failed with status " + resp.Status}
	}

	return &errorResp
}
