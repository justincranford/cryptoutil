// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTenantJoinRequest_TableName(t *testing.T) {
	t.Parallel()

	request := TenantJoinRequest{}
	require.Equal(t, "tenant_join_requests", request.TableName())
}

func TestTenantJoinRequest_StructCreation(t *testing.T) {
	t.Parallel()

	id := googleUuid.New()
	userID := googleUuid.New()
	tenantID := googleUuid.New()
	now := time.Now()
	processedBy := googleUuid.New()

	request := TenantJoinRequest{
		ID:          id,
		UserID:      &userID,
		ClientID:    nil,
		TenantID:    tenantID,
		Status:      JoinRequestStatusPending,
		RequestedAt: now,
		ProcessedAt: &now,
		ProcessedBy: &processedBy,
	}

	require.Equal(t, id, request.ID)
	require.NotNil(t, request.UserID)
	require.Equal(t, userID, *request.UserID)
	require.Nil(t, request.ClientID)
	require.Equal(t, tenantID, request.TenantID)
	require.Equal(t, JoinRequestStatusPending, request.Status)
	require.Equal(t, now, request.RequestedAt)
	require.NotNil(t, request.ProcessedAt)
	require.Equal(t, now, *request.ProcessedAt)
	require.NotNil(t, request.ProcessedBy)
	require.Equal(t, processedBy, *request.ProcessedBy)
}

func TestTenantJoinRequest_StatusConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, "pending", JoinRequestStatusPending)
	require.Equal(t, "approved", JoinRequestStatusApproved)
	require.Equal(t, "rejected", JoinRequestStatusRejected)
}

func TestTenantJoinRequest_ClientIDMutuallyExclusive(t *testing.T) {
	t.Parallel()

	clientID := googleUuid.New()
	tenantID := googleUuid.New()
	now := time.Now()

	request := TenantJoinRequest{
		ID:          googleUuid.New(),
		UserID:      nil,
		ClientID:    &clientID,
		TenantID:    tenantID,
		Status:      JoinRequestStatusApproved,
		RequestedAt: now,
	}

	require.Nil(t, request.UserID)
	require.NotNil(t, request.ClientID)
	require.Equal(t, clientID, *request.ClientID)
}
