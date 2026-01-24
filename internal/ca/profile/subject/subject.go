// Copyright (c) 2025 Justin Cranford

// Package subject provides subject template resolution for certificate subjects.
// It implements YAML-driven subject profile rendering with validation.
package subject

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile represents a subject profile configuration.
type Profile struct {
	// Name is the unique identifier for this profile.
	Name string `yaml:"name"`

	// Description provides human-readable context.
	Description string `yaml:"description"`

	// Subject defines the X.509 distinguished name components.
	Subject DN `yaml:"subject"`

	// SubjectAltNames defines allowed SAN types and patterns.
	SubjectAltNames SANConfig `yaml:"subject_alt_names"`

	// Constraints defines validation constraints.
	Constraints Constraints `yaml:"constraints"`
}

// DN represents X.509 Distinguished Name fields.
type DN struct {
	// CommonName (CN) - typically the entity name.
	CommonName string `yaml:"common_name"`

	// Organization (O) - company or organization name.
	Organization []string `yaml:"organization"`

	// OrganizationalUnit (OU) - department or unit.
	OrganizationalUnit []string `yaml:"organizational_unit"`

	// Country (C) - two-letter ISO 3166-1 country code.
	Country []string `yaml:"country"`

	// State (ST) - state or province.
	State []string `yaml:"state"`

	// Locality (L) - city or town.
	Locality []string `yaml:"locality"`

	// StreetAddress - street address.
	StreetAddress []string `yaml:"street_address"`

	// PostalCode - postal or ZIP code.
	PostalCode []string `yaml:"postal_code"`

	// SerialNumber - entity serial number (not certificate serial).
	SerialNumber string `yaml:"serial_number"`
}

// SANConfig defines Subject Alternative Name configuration.
type SANConfig struct {
	// DNSNames defines allowed DNS name patterns.
	DNSNames SANPatterns `yaml:"dns_names"`

	// IPAddresses defines allowed IP address ranges.
	IPAddresses SANPatterns `yaml:"ip_addresses"`

	// EmailAddresses defines allowed email patterns.
	EmailAddresses SANPatterns `yaml:"email_addresses"`

	// URIs defines allowed URI patterns.
	URIs SANPatterns `yaml:"uris"`
}

// SANPatterns defines patterns and constraints for SAN values.
type SANPatterns struct {
	// Allowed indicates if this SAN type is permitted.
	Allowed bool `yaml:"allowed"`

	// Required indicates if at least one value must be provided.
	Required bool `yaml:"required"`

	// Patterns lists allowed regex patterns.
	Patterns []string `yaml:"patterns"`

	// MaxCount limits the number of values.
	MaxCount int `yaml:"max_count"`
}

// Constraints defines validation constraints for subject profiles.
type Constraints struct {
	// RequireCommonName enforces CN must be present.
	RequireCommonName bool `yaml:"require_common_name"`

	// RequireOrganization enforces O must be present.
	RequireOrganization bool `yaml:"require_organization"`

	// RequireCountry enforces C must be present.
	RequireCountry bool `yaml:"require_country"`

	// AllowWildcard permits wildcard patterns in CN/DNS.
	AllowWildcard bool `yaml:"allow_wildcard"`

	// ValidCountries limits allowed country codes.
	ValidCountries []string `yaml:"valid_countries"`
}

// Request represents a certificate subject request.
type Request struct {
	// CommonName is the requested CN value.
	CommonName string `yaml:"common_name"`

	// Organization is the requested O value.
	Organization []string `yaml:"organization"`

	// OrganizationalUnit is the requested OU value.
	OrganizationalUnit []string `yaml:"organizational_unit"`

	// Country is the requested C value.
	Country []string `yaml:"country"`

	// State is the requested ST value.
	State []string `yaml:"state"`

	// Locality is the requested L value.
	Locality []string `yaml:"locality"`

	// DNSNames are the requested DNS SANs.
	DNSNames []string `yaml:"dns_names"`

	// IPAddresses are the requested IP SANs.
	IPAddresses []string `yaml:"ip_addresses"`

	// EmailAddresses are the requested email SANs.
	EmailAddresses []string `yaml:"email_addresses"`

	// URIs are the requested URI SANs.
	URIs []string `yaml:"uris"`
}

// ResolvedSubject contains the validated and resolved subject.
type ResolvedSubject struct {
	// DN is the resolved distinguished name.
	DN pkix.Name

	// DNSNames are the validated DNS SANs.
	DNSNames []string

	// IPAddresses are the validated IP SANs.
	IPAddresses []net.IP

	// EmailAddresses are the validated email SANs.
	EmailAddresses []string

	// URIs are the validated URI SANs.
	URIs []*url.URL
}

// LoadProfile loads a subject profile from a YAML file.
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read subject profile: %w", err)
	}

	return ParseProfile(data)
}

// ParseProfile parses a subject profile from YAML data.
func ParseProfile(data []byte) (*Profile, error) {
	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse subject profile YAML: %w", err)
	}

	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid subject profile: %w", err)
	}

	return &profile, nil
}

// Validate validates the profile configuration.
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	// Validate SAN patterns are valid regex.
	if err := validatePatterns(p.SubjectAltNames.DNSNames.Patterns); err != nil {
		return fmt.Errorf("invalid DNS name pattern: %w", err)
	}

	if err := validatePatterns(p.SubjectAltNames.IPAddresses.Patterns); err != nil {
		return fmt.Errorf("invalid IP address pattern: %w", err)
	}

	if err := validatePatterns(p.SubjectAltNames.EmailAddresses.Patterns); err != nil {
		return fmt.Errorf("invalid email address pattern: %w", err)
	}

	if err := validatePatterns(p.SubjectAltNames.URIs.Patterns); err != nil {
		return fmt.Errorf("invalid URI pattern: %w", err)
	}

	// Validate country codes.
	for _, country := range p.Constraints.ValidCountries {
		if len(country) != countryCodeLength {
			return fmt.Errorf("invalid country code: %s (must be 2 characters)", country)
		}
	}

	return nil
}

// countryCodeLength defines the ISO 3166-1 alpha-2 country code length.
const countryCodeLength = 2

// Resolve resolves a certificate request against this profile.
func (p *Profile) Resolve(req *Request) (*ResolvedSubject, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resolved := &ResolvedSubject{}

	// Resolve DN fields.
	if err := p.resolveDN(req, resolved); err != nil {
		return nil, err
	}

	// Resolve SANs.
	if err := p.resolveSANs(req, resolved); err != nil {
		return nil, err
	}

	return resolved, nil
}

func (p *Profile) resolveDN(req *Request, resolved *ResolvedSubject) error {
	// Apply common name with fallback to profile default.
	cn := req.CommonName
	if cn == "" {
		cn = p.Subject.CommonName
	}

	if cn == "" && p.Constraints.RequireCommonName {
		return fmt.Errorf("common name is required")
	}

	// Validate wildcard usage.
	if strings.HasPrefix(cn, "*.") && !p.Constraints.AllowWildcard {
		return fmt.Errorf("wildcard common names are not allowed")
	}

	// Apply organization with fallback.
	org := req.Organization
	if len(org) == 0 {
		org = p.Subject.Organization
	}

	if len(org) == 0 && p.Constraints.RequireOrganization {
		return fmt.Errorf("organization is required")
	}

	// Apply country with fallback and validation.
	country := req.Country
	if len(country) == 0 {
		country = p.Subject.Country
	}

	if len(country) == 0 && p.Constraints.RequireCountry {
		return fmt.Errorf("country is required")
	}

	// Validate country codes.
	if len(p.Constraints.ValidCountries) > 0 {
		for _, c := range country {
			if !containsString(p.Constraints.ValidCountries, c) {
				return fmt.Errorf("country %s is not in allowed list", c)
			}
		}
	}

	// Apply remaining fields with fallback.
	ou := req.OrganizationalUnit
	if len(ou) == 0 {
		ou = p.Subject.OrganizationalUnit
	}

	state := req.State
	if len(state) == 0 {
		state = p.Subject.State
	}

	locality := req.Locality
	if len(locality) == 0 {
		locality = p.Subject.Locality
	}

	resolved.DN = pkix.Name{
		CommonName:         cn,
		Organization:       org,
		OrganizationalUnit: ou,
		Country:            country,
		Province:           state,
		Locality:           locality,
		StreetAddress:      p.Subject.StreetAddress,
		PostalCode:         p.Subject.PostalCode,
		SerialNumber:       p.Subject.SerialNumber,
	}

	return nil
}

func (p *Profile) resolveSANs(req *Request, resolved *ResolvedSubject) error {
	// Resolve DNS names.
	if len(req.DNSNames) > 0 {
		if !p.SubjectAltNames.DNSNames.Allowed {
			return fmt.Errorf("DNS names are not allowed in this profile")
		}

		if p.SubjectAltNames.DNSNames.MaxCount > 0 && len(req.DNSNames) > p.SubjectAltNames.DNSNames.MaxCount {
			return fmt.Errorf("too many DNS names: %d (max %d)", len(req.DNSNames), p.SubjectAltNames.DNSNames.MaxCount)
		}

		for _, dns := range req.DNSNames {
			if err := validateAgainstPatterns(dns, p.SubjectAltNames.DNSNames.Patterns); err != nil {
				return fmt.Errorf("DNS name %s does not match allowed patterns: %w", dns, err)
			}

			if strings.HasPrefix(dns, "*.") && !p.Constraints.AllowWildcard {
				return fmt.Errorf("wildcard DNS names are not allowed")
			}
		}

		resolved.DNSNames = req.DNSNames
	} else if p.SubjectAltNames.DNSNames.Required {
		return fmt.Errorf("at least one DNS name is required")
	}

	// Resolve IP addresses.
	if len(req.IPAddresses) > 0 {
		if !p.SubjectAltNames.IPAddresses.Allowed {
			return fmt.Errorf("IP addresses are not allowed in this profile")
		}

		if p.SubjectAltNames.IPAddresses.MaxCount > 0 && len(req.IPAddresses) > p.SubjectAltNames.IPAddresses.MaxCount {
			return fmt.Errorf("too many IP addresses: %d (max %d)", len(req.IPAddresses), p.SubjectAltNames.IPAddresses.MaxCount)
		}

		resolved.IPAddresses = make([]net.IP, 0, len(req.IPAddresses))

		for _, ipStr := range req.IPAddresses {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return fmt.Errorf("invalid IP address: %s", ipStr)
			}

			resolved.IPAddresses = append(resolved.IPAddresses, ip)
		}
	} else if p.SubjectAltNames.IPAddresses.Required {
		return fmt.Errorf("at least one IP address is required")
	}

	// Resolve email addresses.
	if len(req.EmailAddresses) > 0 {
		if !p.SubjectAltNames.EmailAddresses.Allowed {
			return fmt.Errorf("email addresses are not allowed in this profile")
		}

		if p.SubjectAltNames.EmailAddresses.MaxCount > 0 && len(req.EmailAddresses) > p.SubjectAltNames.EmailAddresses.MaxCount {
			return fmt.Errorf("too many email addresses: %d (max %d)", len(req.EmailAddresses), p.SubjectAltNames.EmailAddresses.MaxCount)
		}

		for _, email := range req.EmailAddresses {
			if err := validateAgainstPatterns(email, p.SubjectAltNames.EmailAddresses.Patterns); err != nil {
				return fmt.Errorf("email %s does not match allowed patterns: %w", email, err)
			}
		}

		resolved.EmailAddresses = req.EmailAddresses
	} else if p.SubjectAltNames.EmailAddresses.Required {
		return fmt.Errorf("at least one email address is required")
	}

	// Resolve URIs.
	if len(req.URIs) > 0 {
		if !p.SubjectAltNames.URIs.Allowed {
			return fmt.Errorf("URIs are not allowed in this profile")
		}

		if p.SubjectAltNames.URIs.MaxCount > 0 && len(req.URIs) > p.SubjectAltNames.URIs.MaxCount {
			return fmt.Errorf("too many URIs: %d (max %d)", len(req.URIs), p.SubjectAltNames.URIs.MaxCount)
		}

		resolved.URIs = make([]*url.URL, 0, len(req.URIs))

		for _, uriStr := range req.URIs {
			parsedURI, err := url.Parse(uriStr)
			if err != nil {
				return fmt.Errorf("invalid URI %s: %w", uriStr, err)
			}

			if err := validateAgainstPatterns(uriStr, p.SubjectAltNames.URIs.Patterns); err != nil {
				return fmt.Errorf("URI %s does not match allowed patterns: %w", uriStr, err)
			}

			resolved.URIs = append(resolved.URIs, parsedURI)
		}
	} else if p.SubjectAltNames.URIs.Required {
		return fmt.Errorf("at least one URI is required")
	}

	return nil
}

func validatePatterns(patterns []string) error {
	for _, pattern := range patterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern %s: %w", pattern, err)
		}
	}

	return nil
}

func validateAgainstPatterns(value string, patterns []string) error {
	if len(patterns) == 0 {
		return nil // No patterns means any value is allowed.
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, value)
		if err != nil {
			return fmt.Errorf("pattern match error: %w", err)
		}

		if matched {
			return nil
		}
	}

	return fmt.Errorf("value does not match any allowed pattern")
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
