// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilLearnCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilLearnDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
)

func TestHandleLoginUser_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	passwordHash, err := cryptoutilLearnCrypto.HashPassword("password123")
	require.NoError(t, err)

	user := &cryptoutilLearnDomain.User{
		ID:           googleUuid.New(),
		Username:     "loginuser",
		PasswordHash: hex.EncodeToString(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	reqBody := map[string]string{
		"username": "loginuser",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	require.NotEmpty(t, respBody["token"], "JWT token should be returned")
	require.NotEmpty(t, respBody["expires_at"])
}

func TestHandleLoginUser_WrongPassword(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	passwordHash, err := cryptoutilLearnCrypto.HashPassword("password123")
	require.NoError(t, err)

	user := &cryptoutilLearnDomain.User{
		ID:           googleUuid.New(),
		Username:     "wrongpassuser",
		PasswordHash: hex.EncodeToString(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	reqBody := map[string]string{
		"username": "wrongpassuser",
		"password": "wrongpassword",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid credentials")
}

func TestHandleLoginUser_UserNotFound(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	reqBody := map[string]string{
		"username": "nonexistent",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid credentials")
}

// TestHandleLoginUser_MalformedJSON tests login with malformed JSON.
func TestHandleLoginUser_MalformedJSON(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader([]byte("{invalid json")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleLoginUser_CorruptPasswordHash tests login with corrupted password hash in database.
func TestHandleLoginUser_CorruptPasswordHash(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	username := "testuser"
	password := "password123"

	registerReq := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte(registerReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	_ = resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.FindByUsername(context.Background(), username)
	require.NoError(t, err)

	user.PasswordHash = "not-valid-hex"

	err = userRepo.Update(context.Background(), user)
	require.NoError(t, err)

	loginReq := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password)
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader([]byte(loginReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestHandleLoginUser_HexDecodeError tests handling of corrupted password hash.
func TestHandleLoginUser_HexDecodeError(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	ctx := context.Background()

	userID := googleUuid.New()
	user := &cryptoutilLearnDomain.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: "zzzzzz",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	loginReq := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	loginJSON, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(loginJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "failed to decode password hash", result["error"])
}

// TestHandleLoginUser_EmptyCredentials tests login with empty username or password.
func TestHandleLoginUser_EmptyCredentials(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"EmptyUsername", "", "password123"},
		{"EmptyPassword", "username", ""},
		{"BothEmpty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			reqJSON, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)

			var result map[string]any

			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			require.Equal(t, "username and password are required", result["error"])
		})
	}
}

// TestHandleLoginUser_InvalidBody tests login with malformed JSON.
func TestHandleLoginUser_InvalidBody(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader([]byte("{not:valid")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "invalid request body", result["error"])
}
