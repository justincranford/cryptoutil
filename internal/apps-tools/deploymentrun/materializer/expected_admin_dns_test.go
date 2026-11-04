// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestExpectedAdminServerDNSUsesExactPodIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		pool      string
		want      string
		wantError string
	}{
		{
			name: "PostgreSQL pod",
			pool: cryptoutilSharedMagic.DeploymentMaterializerPoolPG,
			want: "identity-authz-pg-0.identity-authz-pg-pods.test-ns.svc.cluster.local",
		},
		{
			name: "disk pod",
			pool: cryptoutilSharedMagic.DeploymentMaterializerPoolDisk,
			want: "identity-authz-disk-2.identity-authz-disk-pods.test-ns.svc.cluster.local",
		},
		{
			name:      "invalid cluster DNS suffix",
			pool:      cryptoutilSharedMagic.DeploymentMaterializerPoolPG,
			wantError: "derived administrative pod DNS identity is invalid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			podName := "identity-authz-pg-0"
			clusterDomain := "cluster.local"

			if test.pool == cryptoutilSharedMagic.DeploymentMaterializerPoolDisk {
				podName = "identity-authz-disk-2"
			}

			if test.wantError != "" {
				clusterDomain = strings.Repeat("a", 254)
			}

			got, err := expectedAdminServerDNS(
				InstanceConfig{PSID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, Pool: test.pool, ClusterDomain: clusterDomain},
				InstanceService{PodName: podName},
				"test-ns",
			)

			if test.wantError != "" {
				require.ErrorContains(t, err, test.wantError)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
