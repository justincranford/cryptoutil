// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	"context"
	json "encoding/json"
	"errors"
	http "net/http"
	"testing"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type elasticJWKRepositoryStub struct {
	getByIDFn func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error)
	listFn    func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error)
}

func (r *elasticJWKRepositoryStub) Create(context.Context, *cryptoutilAppsSmKmsModel.ElasticJWK) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *elasticJWKRepositoryStub) Get(context.Context, googleUuid.UUID, string) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
	return nil, errors.New(testCompatUnexpectedRepoCall)
}

func (r *elasticJWKRepositoryStub) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
	if r.getByIDFn == nil {
		return nil, errors.New(testCompatUnexpectedRepoCall)
	}

	return r.getByIDFn(ctx, id)
}

func (r *elasticJWKRepositoryStub) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
	if r.listFn == nil {
		return nil, 0, errors.New(testCompatUnexpectedRepoCall)
	}

	return r.listFn(ctx, tenantID, offset, limit)
}

func (r *elasticJWKRepositoryStub) Update(context.Context, *cryptoutilAppsSmKmsModel.ElasticJWK) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *elasticJWKRepositoryStub) Delete(context.Context, googleUuid.UUID) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *elasticJWKRepositoryStub) IncrementMaterialCount(context.Context, googleUuid.UUID) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *elasticJWKRepositoryStub) DecrementMaterialCount(context.Context, googleUuid.UUID) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

type materialJWKRepositoryStub struct {
	getActiveMaterialFn func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error)
}

func (r *materialJWKRepositoryStub) Create(context.Context, *cryptoutilAppsSmKmsModel.MaterialJWK) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) GetByMaterialKID(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	return nil, errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) GetByID(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	return nil, errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	if r.getActiveMaterialFn == nil {
		return nil, errors.New(testCompatUnexpectedRepoCall)
	}

	return r.getActiveMaterialFn(ctx, elasticJWKID)
}

func (r *materialJWKRepositoryStub) ListByElasticJWK(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
	return nil, 0, errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) RotateMaterial(context.Context, googleUuid.UUID, *cryptoutilAppsSmKmsModel.MaterialJWK) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) RetireMaterial(context.Context, googleUuid.UUID) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) Delete(context.Context, googleUuid.UUID) error {
	return errors.New(testCompatUnexpectedRepoCall)
}

func (r *materialJWKRepositoryStub) CountMaterials(context.Context, googleUuid.UUID) (int64, error) {
	return 0, errors.New(testCompatUnexpectedRepoCall)
}

func TestJWKCompatHandler_GetActiveMaterialJWK(t *testing.T) {
	t.Parallel()

	id := newCompatTestUUID(t)
	tests := []struct {
		name        string
		path        string
		material    *cryptoutilAppsSmKmsModel.MaterialJWK
		materialErr error
		wantStatus  int
	}{
		{name: testCompatInvalidKeyIDCase, path: testCompatInvalidUUID, wantStatus: fiber.StatusBadRequest},
		{name: "active material missing", path: id.String(), materialErr: errors.New(testCompatUnexpectedRepoCall), wantStatus: fiber.StatusNotFound},
		{name: "active material found", path: id.String(), material: &cryptoutilAppsSmKmsModel.MaterialJWK{ID: newCompatTestUUID(t), MaterialKID: testCompatMaterialKID, ElasticJWKID: id, Active: true}, wantStatus: fiber.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			materialRepo := &materialJWKRepositoryStub{getActiveMaterialFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
				return tc.material, tc.materialErr
			}}
			handler := NewJWKCompatHandler(&elasticJWKRepositoryStub{}, materialRepo)
			path := testCompatActiveMaterialPathStart + tc.path + testCompatActiveMaterialPathEnd
			status, body := runCompatFiberRequest(t, func(app *fiber.App) {
				app.Get(testCompatActiveMaterialPathStart+testCompatElasticKeyIDParam+testCompatActiveMaterialPathEnd, handler.HandleGetActiveMaterialJWK())
			}, http.MethodGet, path)

			require.Equal(t, tc.wantStatus, status)

			if tc.wantStatus == fiber.StatusOK {
				require.Contains(t, string(body), testCompatMaterialKID)
				require.Contains(t, string(body), "material_kid")
			}
		})
	}
}

func TestJWKCompatHandler_RotateMaterialJWK(t *testing.T) {
	t.Parallel()

	id := newCompatTestUUID(t)
	tests := []struct {
		name       string
		path       string
		elastic    *cryptoutilAppsSmKmsModel.ElasticJWK
		elasticErr error
		wantStatus int
	}{
		{name: testCompatInvalidKeyIDCase, path: testCompatInvalidUUID, wantStatus: fiber.StatusBadRequest},
		{name: "elastic key missing", path: id.String(), elasticErr: errors.New(testCompatUnexpectedRepoCall), wantStatus: fiber.StatusNotFound},
		{name: "rotation requested", path: id.String(), elastic: &cryptoutilAppsSmKmsModel.ElasticJWK{ID: id}, wantStatus: fiber.StatusAccepted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			elasticRepo := &elasticJWKRepositoryStub{getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
				return tc.elastic, tc.elasticErr
			}}
			handler := NewJWKCompatHandler(elasticRepo, &materialJWKRepositoryStub{})
			path := testCompatActiveMaterialPathStart + tc.path + testCompatRotatePathEnd
			status, body := runCompatFiberRequest(t, func(app *fiber.App) {
				app.Post(testCompatActiveMaterialPathStart+":"+"elasticKeyID"+testCompatRotatePathEnd, handler.HandleRotateMaterialJWK())
			}, http.MethodPost, path)

			require.Equal(t, tc.wantStatus, status)

			if tc.wantStatus == fiber.StatusAccepted {
				require.Contains(t, string(body), "rotation requested")
				require.Contains(t, string(body), id.String())
			}
		})
	}
}

func TestJWKSCompatHandler_GetJWKS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tenant        string
		listFn        func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error)
		activeFn      func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error)
		wantStatus    int
		wantKey       bool
		hasKeyFixture bool
	}{
		{name: "missing tenant", wantStatus: fiber.StatusBadRequest},
		{name: "repository failure", tenant: newCompatTestUUID(t).String(), listFn: func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
			return nil, 0, errors.New(testCompatUnexpectedRepoCall)
		}, wantStatus: fiber.StatusInternalServerError},
		{name: "lists active key and skips inactive material", tenant: newCompatTestUUID(t).String(), wantStatus: fiber.StatusOK, wantKey: true, hasKeyFixture: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			firstID := newCompatTestUUID(t)
			secondID := newCompatTestUUID(t)

			if tc.hasKeyFixture {
				tc.listFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
					return []*cryptoutilAppsSmKmsModel.ElasticJWK{{ID: firstID}, {ID: secondID}}, 2, nil
				}
				tc.activeFn = func(_ context.Context, elasticID googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					if elasticID == firstID {
						return &cryptoutilAppsSmKmsModel.MaterialJWK{MaterialKID: testCompatMaterialKID}, nil
					}

					return nil, errors.New(testCompatUnexpectedRepoCall)
				}
			}

			elasticRepo := &elasticJWKRepositoryStub{listFn: tc.listFn}
			materialRepo := &materialJWKRepositoryStub{getActiveMaterialFn: tc.activeFn}
			handler := NewJWKSCompatHandler(elasticRepo, materialRepo)

			path := testCompatJWKSPath
			if tc.tenant != "" {
				path += "?" + testCompatTenantQuery + "=" + tc.tenant
			}

			status, body := runCompatFiberRequest(t, func(app *fiber.App) {
				app.Get(testCompatJWKSPath, handler.HandleGetJWKS())
			}, http.MethodGet, path)

			require.Equal(t, tc.wantStatus, status)

			if tc.wantKey {
				var response struct {
					Keys []struct {
						KID string `json:"kid"`
						KTY string `json:"kty"`
						Use string `json:"use"`
					} `json:"keys"`
				}
				require.NoError(t, json.Unmarshal(body, &response))
				require.Len(t, response.Keys, 1)
				require.Equal(t, testCompatMaterialKID, response.Keys[0].KID)
				require.Equal(t, cryptoutilSharedMagic.KeyTypeOct, response.Keys[0].KTY)
				require.Equal(t, cryptoutilSharedMagic.JoseKeyUseSig, response.Keys[0].Use)
			}
		})
	}
}
