// Copyright (c) 2025 Justin Cranford

package subject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProfile_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		wantName string
	}{
		{
			name: "minimal",
			yaml: `
name: minimal
description: Minimal subject profile
`,
			wantName: "minimal",
		},
		{
			name: "full-tls-server",
			yaml: `
name: tls-server
description: TLS Server certificate subject profile
subject:
  common_name: ""
  organization:
    - Example Corp
  country:
    - US
subject_alt_names:
  dns_names:
    allowed: true
    required: true
    max_count: 10
  ip_addresses:
    allowed: true
    required: false
    max_count: 5
  email_addresses:
    allowed: false
  uris:
    allowed: false
constraints:
  require_common_name: false
  require_organization: true
  require_country: true
  allow_wildcard: true
  valid_countries:
    - US
    - CA
    - GB
`,
			wantName: "tls-server",
		},
		{
			name: "with-patterns",
			yaml: `
name: internal-only
description: Internal hosts only
subject_alt_names:
  dns_names:
    allowed: true
    patterns:
      - "^.*\\.internal\\.example\\.com$"
      - "^.*\\.corp\\.example\\.com$"
`,
			wantName: "internal-only",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile, err := ParseProfile([]byte(tc.yaml))
			require.NoError(t, err)
			require.NotNil(t, profile)
			require.Equal(t, tc.wantName, profile.Name)
		})
	}
}

func TestParseProfile_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "empty-name",
			yaml:    `description: No name`,
			wantErr: "profile name is required",
		},
		{
			name: "invalid-dns-pattern",
			yaml: `
name: bad-pattern
subject_alt_names:
  dns_names:
    allowed: true
    patterns:
      - "[invalid"
`,
			wantErr: "invalid DNS name pattern",
		},
		{
			name: "invalid-country-code",
			yaml: `
name: bad-country
constraints:
  valid_countries:
    - USA
`,
			wantErr: "invalid country code",
		},
		{
			name:    "invalid-yaml",
			yaml:    `{{{not yaml`,
			wantErr: "failed to parse subject profile YAML",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile, err := ParseProfile([]byte(tc.yaml))
			require.Error(t, err)
			require.Nil(t, profile)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestProfile_Resolve_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		profile    *Profile
		request    *Request
		wantCN     string
		wantOrg    []string
		wantDNS    int
		wantIPs    int
		wantEmails int
		wantURIs   int
	}{
		{
			name: "simple-request",
			profile: &Profile{
				Name: "test",
				Subject: DN{
					Organization: []string{"Default Org"},
				},
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				DNSNames:   []string{"test.example.com"},
			},
			wantCN:  "test.example.com",
			wantOrg: []string{"Default Org"},
			wantDNS: 1,
		},
		{
			name: "override-defaults",
			profile: &Profile{
				Name: "test",
				Subject: DN{
					Organization: []string{"Default Org"},
					Country:      []string{"US"},
				},
				SubjectAltNames: SANConfig{
					DNSNames:    SANPatterns{Allowed: true},
					IPAddresses: SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName:   "custom.example.com",
				Organization: []string{"Custom Org"},
				DNSNames:     []string{"custom.example.com", "api.example.com"},
				IPAddresses:  []string{"192.168.1.1", "10.0.0.1"},
			},
			wantCN:  "custom.example.com",
			wantOrg: []string{"Custom Org"},
			wantDNS: 2,
			wantIPs: 2,
		},
		{
			name: "wildcard-allowed",
			profile: &Profile{
				Name: "wildcard",
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true},
				},
				Constraints: Constraints{
					AllowWildcard: true,
				},
			},
			request: &Request{
				CommonName: "*.example.com",
				DNSNames:   []string{"*.example.com"},
			},
			wantCN:  "*.example.com",
			wantDNS: 1,
		},
		{
			name: "with-uris-and-emails",
			profile: &Profile{
				Name: "full",
				SubjectAltNames: SANConfig{
					DNSNames:       SANPatterns{Allowed: true},
					EmailAddresses: SANPatterns{Allowed: true},
					URIs:           SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName:     "service.example.com",
				DNSNames:       []string{"service.example.com"},
				EmailAddresses: []string{"admin@example.com"},
				URIs:           []string{"https://example.com/service"},
			},
			wantCN:     "service.example.com",
			wantDNS:    1,
			wantEmails: 1,
			wantURIs:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resolved, err := tc.profile.Resolve(tc.request)
			require.NoError(t, err)
			require.NotNil(t, resolved)
			require.Equal(t, tc.wantCN, resolved.DN.CommonName)

			if len(tc.wantOrg) > 0 {
				require.Equal(t, tc.wantOrg, resolved.DN.Organization)
			}

			require.Len(t, resolved.DNSNames, tc.wantDNS)
			require.Len(t, resolved.IPAddresses, tc.wantIPs)
			require.Len(t, resolved.EmailAddresses, tc.wantEmails)
			require.Len(t, resolved.URIs, tc.wantURIs)
		})
	}
}

func TestProfile_Resolve_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		profile *Profile
		request *Request
		wantErr string
	}{
		{
			name:    "nil-request",
			profile: &Profile{Name: "test"},
			request: nil,
			wantErr: "request cannot be nil",
		},
		{
			name: "cn-required",
			profile: &Profile{
				Name: "test",
				Constraints: Constraints{
					RequireCommonName: true,
				},
			},
			request: &Request{},
			wantErr: "common name is required",
		},
		{
			name: "org-required",
			profile: &Profile{
				Name: "test",
				Constraints: Constraints{
					RequireOrganization: true,
				},
			},
			request: &Request{
				CommonName: "test.example.com",
			},
			wantErr: "organization is required",
		},
		{
			name: "country-required",
			profile: &Profile{
				Name: "test",
				Constraints: Constraints{
					RequireCountry: true,
				},
			},
			request: &Request{
				CommonName: "test.example.com",
			},
			wantErr: "country is required",
		},
		{
			name: "wildcard-cn-not-allowed",
			profile: &Profile{
				Name: "test",
				Constraints: Constraints{
					AllowWildcard: false,
				},
			},
			request: &Request{
				CommonName: "*.example.com",
			},
			wantErr: "wildcard common names are not allowed",
		},
		{
			name: "wildcard-dns-not-allowed",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true},
				},
				Constraints: Constraints{
					AllowWildcard: false,
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				DNSNames:   []string{"*.example.com"},
			},
			wantErr: "wildcard DNS names are not allowed",
		},
		{
			name: "invalid-country",
			profile: &Profile{
				Name: "test",
				Constraints: Constraints{
					ValidCountries: []string{"US", "CA"},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				Country:    []string{"XX"},
			},
			wantErr: "country XX is not in allowed list",
		},
		{
			name: "dns-not-allowed",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				DNSNames:   []string{"test.example.com"},
			},
			wantErr: "DNS names are not allowed",
		},
		{
			name: "dns-required",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true, Required: true},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
			},
			wantErr: "at least one DNS name is required",
		},
		{
			name: "too-many-dns",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true, MaxCount: 2},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				DNSNames:   []string{"a.example.com", "b.example.com", "c.example.com"},
			},
			wantErr: "too many DNS names",
		},
		{
			name: "dns-pattern-mismatch",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{
						Allowed:  true,
						Patterns: []string{`^.*\.internal\.example\.com$`},
					},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				DNSNames:   []string{"external.example.org"},
			},
			wantErr: "does not match allowed patterns",
		},
		{
			name: "ip-not-allowed",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					IPAddresses: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName:  "test.example.com",
				IPAddresses: []string{"192.168.1.1"},
			},
			wantErr: "IP addresses are not allowed",
		},
		{
			name: "invalid-ip",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					IPAddresses: SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName:  "test.example.com",
				IPAddresses: []string{"not-an-ip"},
			},
			wantErr: "invalid IP address",
		},
		{
			name: "email-not-allowed",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					EmailAddresses: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName:     "test.example.com",
				EmailAddresses: []string{"test@example.com"},
			},
			wantErr: "email addresses are not allowed",
		},
		{
			name: "uri-not-allowed",
			profile: &Profile{
				Name: "test",
				SubjectAltNames: SANConfig{
					URIs: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName: "test.example.com",
				URIs:       []string{"https://example.com"},
			},
			wantErr: "URIs are not allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resolved, err := tc.profile.Resolve(tc.request)
			require.Error(t, err)
			require.Nil(t, resolved)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
