// Copyright (c) 2025 Justin Cranford

package network_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"
)

func TestParseIPAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     []string
		wantLen   int
		wantFirst string // Expected first IP in result (if success)
		wantErr   bool
	}{
		{
			name:      "Valid_IPv4_addresses",
			input:     []string{"127.0.0.1", "192.168.1.1", "10.0.0.1"},
			wantLen:   3,
			wantFirst: "127.0.0.1",
			wantErr:   false,
		},
		{
			name:      "Valid_IPv6_addresses",
			input:     []string{"::1", "2001:db8::1", "fe80::1"},
			wantLen:   3,
			wantFirst: "::1",
			wantErr:   false,
		},
		{
			name:      "Mixed_IPv4_and_IPv6",
			input:     []string{"127.0.0.1", "::1", "192.168.1.1"},
			wantLen:   3,
			wantFirst: "127.0.0.1",
			wantErr:   false,
		},
		{
			name:    "Invalid_IP_address",
			input:   []string{"127.0.0.1", "invalid-ip", "192.168.1.1"},
			wantErr: true,
		},
		{
			name:      "Empty_slice",
			input:     []string{},
			wantLen:   0,
			wantFirst: "",
			wantErr:   false,
		},
		{
			name:    "Invalid_format",
			input:   []string{"999.999.999.999"},
			wantErr: true,
		},
		{
			name:    "Partial_IP",
			input:   []string{"192.168.1"},
			wantErr: true,
		},
		{
			name:      "IPv4_mapped_IPv6",
			input:     []string{"::ffff:192.168.1.1"},
			wantLen:   1,
			wantFirst: "192.168.1.1", // net.ParseIP normalizes IPv4-mapped IPv6 to IPv4
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(tc.input)

			if tc.wantErr {
				require.Error(t, err, "ParseIPAddresses should return error")
				require.Contains(t, err.Error(), "failed to parse IP address", "Error should mention parse failure")
			} else {
				require.NoError(t, err, "ParseIPAddresses should not return error")
				require.Len(t, result, tc.wantLen, "Parsed IP count should match")

				if tc.wantLen > 0 && tc.wantFirst != "" {
					require.Equal(t, tc.wantFirst, result[0].String(), "First parsed IP should match")
				}
			}
		})
	}
}

func TestNormalizeIPv4Addresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []string // Input as strings for easier test definition
		expected []string // Expected normalized IPs as strings
	}{
		{
			name:     "IPv4_addresses_unchanged",
			input:    []string{"127.0.0.1", "192.168.1.1", "10.0.0.1"},
			expected: []string{"127.0.0.1", "192.168.1.1", "10.0.0.1"},
		},
		{
			name:     "IPv6_addresses_unchanged",
			input:    []string{"::1", "2001:db8::1", "fe80::1"},
			expected: []string{"::1", "2001:db8::1", "fe80::1"},
		},
		{
			name:     "IPv4_mapped_IPv6_to_IPv4",
			input:    []string{"::ffff:127.0.0.1", "::ffff:192.168.1.1"},
			expected: []string{"127.0.0.1", "192.168.1.1"}, // Converted to IPv4
		},
		{
			name:     "Mixed_addresses",
			input:    []string{"127.0.0.1", "::1", "::ffff:192.168.1.1"},
			expected: []string{"127.0.0.1", "::1", "192.168.1.1"}, // Last one converted
		},
		{
			name:     "Empty_slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single_IPv4",
			input:    []string{"127.0.0.1"},
			expected: []string{"127.0.0.1"},
		},
		{
			name:     "Single_IPv6",
			input:    []string{"::1"},
			expected: []string{"::1"},
		},
		{
			name:     "Single_IPv4_mapped_IPv6",
			input:    []string{"::ffff:10.0.0.1"},
			expected: []string{"10.0.0.1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Convert string IPs to net.IP
			inputIPs := make([]net.IP, len(tc.input))
			for i, ipStr := range tc.input {
				inputIPs[i] = net.ParseIP(ipStr)
				require.NotNil(t, inputIPs[i], "Test setup: input IP should parse successfully: %s", ipStr)
			}

			// Call the function
			result := cryptoutilSharedUtilNetwork.NormalizeIPv4Addresses(inputIPs)

			// Verify results
			require.Len(t, result, len(tc.expected), "Normalized IP count should match")

			for i, expectedStr := range tc.expected {
				require.Equal(t, expectedStr, result[i].String(), "Normalized IP at index %d should match", i)
			}
		})
	}
}

func TestNormalizeIPv4Addresses_NilInput(t *testing.T) {
	t.Parallel()

	// Test with nil slice
	result := cryptoutilSharedUtilNetwork.NormalizeIPv4Addresses(nil)
	require.Empty(t, result, "Nil input should return empty slice")
}
