// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentBootstrapSystemWorkloads(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		kind, name string
		want       bool
	}{
		{kind: developmentPodKind, name: "coredns-owned-id", want: true},
		{kind: developmentPodKind, name: "local-path-provisioner-owned-id", want: true},
		{kind: developmentPodKind, name: "metrics-server-owned-id", want: true},
		{kind: developmentPodKind, name: "traefik-owned-id", want: true},
		{kind: developmentPodKind, name: "svclb-traefik-owned-id", want: true},
		{kind: developmentPodKind, name: "helm-install-traefik-crd-owned-id", want: true},
		{kind: developmentJobKind, name: "helm-install-traefik", want: true},
		{kind: developmentJobKind, name: "helm-install-traefik-crd", want: true},
		{kind: developmentJobKind, name: "arbitrary-app-job"},
		{kind: developmentPodKind, name: "arbitrary-app-pod"},
		{kind: developmentPodKind, name: "coredns-"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, developmentBootstrapSystemWorkload(tc.kind, tc.name))
		})
	}
}
