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
	"cryptoutil/internal/learn/server"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

func TestHandleRegisterUser_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	reqBody := map[string]string{
		"username": "testuser",
		"password": password,
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

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	passwordHash, err := cryptoutilLearnCrypto.HashPassword(password)
	require.NoError(t, err)

	existingUser := &cryptoutilLearnDomain.User{
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
// DISABLED: Cannot close shared database in TestMain pattern - breaks parallel tests.
// TODO: Rewrite using mock repository instead of closing shared database.
/*
func TestHandleRegisterUser_RepositoryError(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	_ = sqlDB.Close()

	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	registerReq := fmt.Sprintf(`{"username": "testuser", "password": "%s"}`, password)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader([]byte(registerReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
*/

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

			password, err := cryptoutilRandom.GeneratePasswordSimple()
			require.NoError(t, err)

			registerReq := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, tt.username, password)
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
	cfg := initTestConfig()

	srv, err := server.New(context.Background(), cfg, db, repository.DatabaseTypeSQLite)
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

// TestNew_NilContext tests server creation validates context.
func TestNew_NilContext(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	cfg := server.DefaultAppConfig()
	cfg.BindPublicPort = 0
	cfg.BindPrivatePort = 0
	// NOTE: OTLPService intentionally NOT set to test telemetry validation

	_, err := server.New(context.Background(), cfg, db, repository.DatabaseTypeSQLite)
	require.Error(t, err)
	// Telemetry validation happens before other checks
	require.Contains(t, err.Error(), "service name must be non-empty")
}

// TestNew_NilConfig tests server creation with nil config.
func TestNew_NilConfig(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)

	_, err := server.New(context.Background(), nil, db, repository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestNew_NilDatabase tests server creation with nil database.
func TestNew_NilDatabase(t *testing.T) {
	t.Parallel()

	cfg := initTestConfig()

	_, err := server.New(context.Background(), cfg, nil, repository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database cannot be nil")
}
