// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestExpandedDNSNameBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		extraByte   int
		domainBytes int
		wantError   string
		wantDomain  bool
	}{
		{name: "full name at DNS limit"},
		{name: "expanded name exceeds limit", extraByte: 1, wantError: testInvalidDNSName},
		{name: "suffix above legacy cutoff remains valid", domainBytes: 201, wantError: testInvalidDNSName, wantDomain: true},
		{name: "maximum suffix length is permitted", domainBytes: 253, wantError: testInvalidDNSName, wantDomain: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			prefix := "sm-kms-pg-0.sm-kms-pg-pods.cu-" + m.Run + "-s-sm-kms-pg.svc."

			domainLength := 253 - len(prefix) + tc.extraByte
			if tc.domainBytes > 0 {
				domainLength = tc.domainBytes
			}

			var labels []string
			for remaining := domainLength; remaining > 0; {
				if len(labels) > 0 {
					remaining--
				}

				labelLength := min(63, remaining)
				labels = append(labels, "1"+strings.Repeat("a", labelLength-1))
				remaining -= labelLength
			}

			m.ClusterDomain = strings.Join(labels, ".")
			if tc.wantDomain {
				require.Greater(t, len(m.ClusterDomain), 200)
				require.NoError(t, validate(m, r))
			}

			inv, err := build(m, r, testServiceSMKMSPGSelector)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.Equal(t, 253, len(inv.Releases[0].Workloads[0].Pods[0].FQDN))
			require.True(t, dnsLabel(strings.Split(m.ClusterDomain, ".")[0]))
		})
	}
}

func TestDNSValidationHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		label       string
		dnsName     string
		wantLabel   bool
		wantDNSName bool
	}{
		{name: "numeric label", label: "123", dnsName: "123.example", wantLabel: true, wantDNSName: true},
		{name: "zero-length name", dnsName: "", wantDNSName: false},
		{name: "missing domain label", label: "", dnsName: "example..local", wantDNSName: false},
		{name: "label too long", label: strings.Repeat("a", cryptoutilSharedMagic.FQDNLabelMaxLength+1), dnsName: "example.local", wantDNSName: true},
		{name: "name too long", label: testExampleValue, dnsName: strings.Repeat("a", cryptoutilSharedMagic.FQDNMaxLength+1), wantLabel: true, wantDNSName: false},
		{name: "leading hyphen", label: "-example", dnsName: "-example.local", wantDNSName: false},
		{name: "trailing hyphen", label: "example-", dnsName: "example-.local", wantDNSName: false},
		{name: "uppercase", label: "Example", dnsName: "Example.local", wantDNSName: false},
		{name: "invalid character", label: "bad_name", dnsName: "bad_name.local", wantDNSName: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantLabel, dnsLabel(tc.label))
			require.Equal(t, tc.wantDNSName, dnsName(tc.dnsName))
		})
	}
}
