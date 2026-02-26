// Copyright (c) 2025 Justin Cranford

package handler

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"

	googleUuid "github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestOamOasMapper_ToOasGetElastickeysResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	elasticKeys := []cryptoutilKmsServer.ElasticKey{}

	resp, err := mapper.toOasGetElastickeysResponse(nil, elasticKeys)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeysResponse_Errors tests elastic keys error responses.
func TestOamOasMapper_ToOasGetElastickeysResponse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr("invalid"), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr("not found"), nil),
			wantErr: true,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(cryptoutilSharedMagic.StringError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasGetElastickeysResponse(tc.err, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOamGetMaterialKeysQueryParams tests material keys query param mapping.
func TestOamOasMapper_ToOamGetMaterialKeysQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	params := &cryptoutilKmsServer.GetMaterialkeysParams{}

	result := mapper.toOamGetMaterialKeysQueryParams(params)
	require.NotNil(t, result)
}

// TestOamOasMapper_ToOasGetMaterialKeysResponse_Success tests successful material keys response.
func TestOamOasMapper_ToOasGetMaterialKeysResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	keys := []cryptoutilKmsServer.MaterialKey{}

	resp, err := mapper.toOasGetMaterialKeysResponse(nil, keys)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetMaterialKeysResponse_Errors tests material keys error responses.
func TestOamOasMapper_ToOasGetMaterialKeysResponse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr("invalid"), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr("not found"), nil),
			wantErr: true,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(cryptoutilSharedMagic.StringError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasGetMaterialKeysResponse(tc.err, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasPutElastickeyElasticKeyIDResponse tests PUT ElasticKey responses.
func TestOamOasMapper_ToOasPutElastickeyElasticKeyIDResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr(testInvalidRequest), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr(testEKNotFound), nil),
			wantErr: false,
		},
		{
			name:    "conflict",
			err:     cryptoutilSharedApperr.NewHTTP409Conflict(strPtr("name already exists"), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(testInternalError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			googleUUID, uuidErr := googleUuid.NewV7()
			require.NoError(t, uuidErr)

			uuid := openapiTypes.UUID(googleUUID)
			elasticKey := &cryptoutilKmsServer.ElasticKey{
				ElasticKeyID: &uuid,
			}

			resp, err := mapper.toOasPutElastickeyElasticKeyIDResponse(tc.err, elasticKey)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasDeleteElastickeyElasticKeyIDResponse tests DELETE ElasticKey responses.
func TestOamOasMapper_ToOasDeleteElastickeyElasticKeyIDResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr(testInvalidRequest), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr(testEKNotFound), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(testInternalError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasDeleteElastickeyElasticKeyIDResponse(tc.err)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDImportResponse tests POST Import MaterialKey responses.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDImportResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr(testInvalidRequest), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr(testEKNotFound), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(testInternalError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			googleUUID, uuidErr := googleUuid.NewV7()
			require.NoError(t, uuidErr)

			uuid := openapiTypes.UUID(googleUUID)
			materialKey := &cryptoutilKmsServer.MaterialKey{
				MaterialKeyID: &uuid,
			}

			resp, err := mapper.toOasPostElastickeyElasticKeyIDImportResponse(tc.err, materialKey)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse tests POST Revoke MaterialKey responses.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr(testInvalidRequest), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr(testEKNotFound), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(testInternalError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(tc.err)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasDeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse tests DELETE MaterialKey responses.
func TestOamOasMapper_ToOasDeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr(testInvalidRequest), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr(testEKNotFound), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr(testInternalError), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasDeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(tc.err)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}
