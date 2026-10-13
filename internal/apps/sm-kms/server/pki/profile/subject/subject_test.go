// Copyright (c) 2025-2026 Justin Cranford.
package subject

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestParseProfile_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		wantName string
	}{
		{
			name: testLiteralMinimal,
			yaml: `
name: minimal
description: Minimal subject profile
`,
			wantName: testLiteralMinimal,
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
			wantName: cryptoutilSharedMagic.PKIProfileIDTLSServer,
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
				Name: testLiteralTest,
				Subject: DN{
					Organization: []string{testLiteralDefaultOrg},
				},
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
				DNSNames:   []string{testLiteralTestExampleCom},
			},
			wantCN:  testLiteralTestExampleCom,
			wantOrg: []string{testLiteralDefaultOrg},
			wantDNS: 1,
		},
		{
			name: "override-defaults",
			profile: &Profile{
				Name: testLiteralTest,
				Subject: DN{
					Organization: []string{testLiteralDefaultOrg},
					Country:      []string{testLiteralUS},
				},
				SubjectAltNames: SANConfig{
					DNSNames:    SANPatterns{Allowed: true},
					IPAddresses: SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName:   testLiteralCustomExampleCom,
				Organization: []string{testLiteralCustomOrg},
				DNSNames:     []string{testLiteralCustomExampleCom, "api.example.com"},
				IPAddresses:  []string{testLiteral19216811, "10.0.0.1"},
			},
			wantCN:  testLiteralCustomExampleCom,
			wantOrg: []string{testLiteralCustomOrg},
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
				CommonName: testLiteralWildcardExampleCom,
				DNSNames:   []string{testLiteralWildcardExampleCom},
			},
			wantCN:  testLiteralWildcardExampleCom,
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
				CommonName:     testLiteralServiceExampleCom,
				DNSNames:       []string{testLiteralServiceExampleCom},
				EmailAddresses: []string{"admin@example.com"},
				URIs:           []string{"https://example.com/service"},
			},
			wantCN:     testLiteralServiceExampleCom,
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
			profile: &Profile{Name: testLiteralTest},
			request: nil,
			wantErr: "request cannot be nil",
		},
		{
			name: "cn-required",
			profile: &Profile{
				Name: testLiteralTest,
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
				Name: testLiteralTest,
				Constraints: Constraints{
					RequireOrganization: true,
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
			},
			wantErr: "organization is required",
		},
		{
			name: "country-required",
			profile: &Profile{
				Name: testLiteralTest,
				Constraints: Constraints{
					RequireCountry: true,
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
			},
			wantErr: "country is required",
		},
		{
			name: "wildcard-cn-not-allowed",
			profile: &Profile{
				Name: testLiteralTest,
				Constraints: Constraints{
					AllowWildcard: false,
				},
			},
			request: &Request{
				CommonName: testLiteralWildcardExampleCom,
			},
			wantErr: "wildcard common names are not allowed",
		},
		{
			name: "wildcard-dns-not-allowed",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true},
				},
				Constraints: Constraints{
					AllowWildcard: false,
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
				DNSNames:   []string{testLiteralWildcardExampleCom},
			},
			wantErr: "wildcard DNS names are not allowed",
		},
		{
			name: "invalid-country",
			profile: &Profile{
				Name: testLiteralTest,
				Constraints: Constraints{
					ValidCountries: []string{testLiteralUS, "CA"},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
				Country:    []string{"XX"},
			},
			wantErr: "country XX is not in allowed list",
		},
		{
			name: "dns-not-allowed",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
				DNSNames:   []string{testLiteralTestExampleCom},
			},
			wantErr: "DNS names are not allowed",
		},
		{
			name: "dns-required",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true, Required: true},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
			},
			wantErr: "at least one DNS name is required",
		},
		{
			name: "too-many-dns",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{Allowed: true, MaxCount: 2},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
				DNSNames:   []string{"a.example.com", "b.example.com", "c.example.com"},
			},
			wantErr: "too many DNS names",
		},
		{
			name: "dns-pattern-mismatch",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					DNSNames: SANPatterns{
						Allowed:  true,
						Patterns: []string{`^.*\.internal\.example\.com$`},
					},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
				DNSNames:   []string{"external.example.org"},
			},
			wantErr: "does not match allowed patterns",
		},
		{
			name: "ip-not-allowed",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					IPAddresses: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName:  testLiteralTestExampleCom,
				IPAddresses: []string{testLiteral19216811},
			},
			wantErr: "IP addresses are not allowed",
		},
		{
			name: "invalid-ip",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					IPAddresses: SANPatterns{Allowed: true},
				},
			},
			request: &Request{
				CommonName:  testLiteralTestExampleCom,
				IPAddresses: []string{"not-an-ip"},
			},
			wantErr: "invalid IP address",
		},
		{
			name: "email-not-allowed",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					EmailAddresses: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName:     testLiteralTestExampleCom,
				EmailAddresses: []string{"test@example.com"},
			},
			wantErr: "email addresses are not allowed",
		},
		{
			name: "uri-not-allowed",
			profile: &Profile{
				Name: testLiteralTest,
				SubjectAltNames: SANConfig{
					URIs: SANPatterns{Allowed: false},
				},
			},
			request: &Request{
				CommonName: testLiteralTestExampleCom,
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
