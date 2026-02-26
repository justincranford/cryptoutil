// Copyright (c) 2025 Justin Cranford

package subject

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadProfile_Valid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "profile.yml")
	err := os.WriteFile(profilePath, []byte("name: test-profile\ndescription: test\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	p, err := LoadProfile(profilePath)
	require.NoError(t, err)
	require.Equal(t, "test-profile", p.Name)
}

func TestLoadProfile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadProfile("/nonexistent/path/profile.yml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read subject profile")
}

func TestValidate_InvalidPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "invalid IP pattern",
			yaml:    "name: test\nsubject_alt_names:\n  ip_addresses:\n    allowed: true\n    patterns:\n      - \"[invalid\"",
			wantErr: "invalid IP address pattern",
		},
		{
			name:    "invalid email pattern",
			yaml:    "name: test\nsubject_alt_names:\n  email_addresses:\n    allowed: true\n    patterns:\n      - \"[invalid\"",
			wantErr: "invalid email address pattern",
		},
		{
			name:    "invalid URI pattern",
			yaml:    "name: test\nsubject_alt_names:\n  uris:\n    allowed: true\n    patterns:\n      - \"[invalid\"",
			wantErr: "invalid URI pattern",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseProfile([]byte(tc.yaml))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestResolveSANs_IPErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		req     *Request
		wantErr string
	}{
		{
			name:    "too many IPs",
			yaml:    "name: test\nsubject_alt_names:\n  ip_addresses:\n    allowed: true\n    max_count: 1",
			req:     &Request{IPAddresses: []string{"1.2.3.4", "5.6.7.8"}},
			wantErr: "too many IP addresses",
		},
		{
			name:    "IP required",
			yaml:    "name: test\nsubject_alt_names:\n  ip_addresses:\n    allowed: true\n    required: true",
			req:     &Request{},
			wantErr: "at least one IP address is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := ParseProfile([]byte(tc.yaml))
			require.NoError(t, err)
			_, err = p.Resolve(tc.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestResolveSANs_EmailErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		req     *Request
		wantErr string
	}{
		{
			name:    "too many emails",
			yaml:    "name: test\nsubject_alt_names:\n  email_addresses:\n    allowed: true\n    max_count: 1",
			req:     &Request{EmailAddresses: []string{"a@b.com", "c@d.com"}},
			wantErr: "too many email addresses",
		},
		{
			name:    "email pattern mismatch",
			yaml:    "name: test\nsubject_alt_names:\n  email_addresses:\n    allowed: true\n    patterns:\n      - \"^.*@example\\\\.com$\"",
			req:     &Request{EmailAddresses: []string{"a@other.com"}},
			wantErr: cryptoutilSharedMagic.ClaimEmail,
		},
		{
			name:    "email required",
			yaml:    "name: test\nsubject_alt_names:\n  email_addresses:\n    allowed: true\n    required: true",
			req:     &Request{},
			wantErr: "at least one email address is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := ParseProfile([]byte(tc.yaml))
			require.NoError(t, err)
			_, err = p.Resolve(tc.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestResolveSANs_URIErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		req     *Request
		wantErr string
	}{
		{
			name:    "too many URIs",
			yaml:    "name: test\nsubject_alt_names:\n  uris:\n    allowed: true\n    max_count: 1",
			req:     &Request{URIs: []string{"https://a.com", "https://b.com"}},
			wantErr: "too many URIs",
		},
		{
			name:    "URI pattern mismatch",
			yaml:    "name: test\nsubject_alt_names:\n  uris:\n    allowed: true\n    patterns:\n      - \"^https://example\\\\.com/.*$\"",
			req:     &Request{URIs: []string{"https://other.com/path"}},
			wantErr: "URI",
		},
		{
			name:    "URI required",
			yaml:    "name: test\nsubject_alt_names:\n  uris:\n    allowed: true\n    required: true",
			req:     &Request{},
			wantErr: "at least one URI is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := ParseProfile([]byte(tc.yaml))
			require.NoError(t, err)
			_, err = p.Resolve(tc.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestContainsString_Found(t *testing.T) {
	t.Parallel()

	require.True(t, containsString([]string{"a", "b", "c"}, "b"))
}
