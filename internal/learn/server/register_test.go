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

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
)

func TestHandleRegisterUser_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	require.NotEmpty(t, respBody["user_id"])
	require.NotEmpty(t, respBody["public_key"])

	publicKeyBytes, err := hex.DecodeString(respBody["public_key"])
	require.NoError(t, err)

	const expectedPublicKeyLength = 65 // X9.62 uncompressed format.
	require.Len(t, publicKeyBytes, expectedPublicKeyLength)
}

func TestHandleRegisterUser_UsernameTooShort(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	reqBody := map[string]string{
		"username": "ab",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "username must be 3-50 characters")
}

func TestHandleRegisterUser_PasswordTooShort(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "pass",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "password must be at least 8 characters")
}

func TestHandleRegisterUser_DuplicateUsername(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	passwordHash, err := cryptoutilCrypto.HashPassword("password123")
	require.NoError(t, err)

	privateKey, publicKeyBytes, err := cryptoutilCrypto.GenerateECDHKeyPair()
	require.NoError(t, err)

	_ = privateKey
	_ = publicKeyBytes

	existingUser := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     "existinguser",
		PasswordHash: hex.EncodeToString(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(context.Background(), existingUser)
	require.NoError(t, err)

	reqBody := map[string]string{
		"username": "existinguser",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusConflict, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "username already exists")
}

// TestHandleRegisterUser_MalformedJSON tests registration with malformed JSON.
func TestHandleRegisterUser_MalformedJSON(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte("{invalid json")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleRegisterUser_RepositoryError tests registration when repository fails.
func TestHandleRegisterUser_RepositoryError(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	_ = sqlDB.Close()

	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	registerReq := `{"username": "testuser", "password": "password123"}`
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte(registerReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestHandleRegisterUser_UsernameValidation tests username length validation.
func TestHandleRegisterUser_UsernameValidation(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	tests := []struct {
		name     string
		username string
		wantCode int
	}{
		{"TooShort", "ab", http.StatusBadRequest},
		{"MinLength", "abc", http.StatusCreated},
		{"MaxLength", "12345678901234567890123456789012345678901234567890", http.StatusCreated},
		{"TooLong", "123456789012345678901234567890123456789012345678901", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registerReq := fmt.Sprintf(`{"username": "%s", "password": "password123"}`, tt.username)
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte(registerReq)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.wantCode, resp.StatusCode)
		})
	}
}

// TestHandleRegisterUser_PasswordValidation tests password length validation.
func TestHandleRegisterUser_PasswordValidation(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	tests := []struct {
		name     string
		password string
		wantCode int
	}{
		{"TooShort", "short", http.StatusBadRequest},
		{"MinLength", "12345678", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			username := fmt.Sprintf("user_%s", googleUuid.New().String()[:8])
			registerReq := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, tt.password)
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte(registerReq)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.wantCode, resp.StatusCode)
		})
	}
}

// TestNew_Success tests successful server creation with valid config.
func TestNew_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	cfg := &server.Config{
		PublicPort: 0,
		AdminPort:  0,
		DB:         db,
	}

	srv, err := server.New(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	_ = srv.PublicPort()

	_, _ = srv.AdminPort()
}

// TestHandleRegisterUser_InvalidBody tests registration with malformed JSON.
func TestHandleRegisterUser_InvalidBody(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte("{invalid-json")))
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

// TestNew_NilContext tests server creation with nil context.
func TestNew_NilContext(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	cfg := &server.Config{
		PublicPort: 0,
		AdminPort:  0,
		DB:         db,
	}

	_, err := server.New(context.TODO(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNew_NilConfig tests server creation with nil config.
func TestNew_NilConfig(t *testing.T) {
	t.Parallel()

	_, err := server.New(context.Background(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestNew_NilDatabase tests server creation with nil database.
func TestNew_NilDatabase(t *testing.T) {
	t.Parallel()

	cfg := &server.Config{
		PublicPort: 0,
		AdminPort:  0,
		DB:         nil,
	}

	_, err := server.New(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database cannot be nil")
}
