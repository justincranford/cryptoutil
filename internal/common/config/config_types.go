// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"time"
)

// Settings struct holds all application configuration settings.
type Settings struct {
	SubCommand                  string
	Help                        bool
	ConfigFile                  []string
	LogLevel                    string
	VerboseMode                 bool
	DevMode                     bool
	DryRun                      bool
	Profile                     string // Configuration profile: dev, stg, prod, test.
	BindPublicProtocol          string
	BindPublicAddress           string
	BindPublicPort              uint16
	BindPrivateProtocol         string
	BindPrivateAddress          string
	BindPrivatePort             uint16
	TLSPublicDNSNames           []string
	TLSPublicIPAddresses        []string
	TLSPrivateDNSNames          []string
	TLSPrivateIPAddresses       []string
	PublicBrowserAPIContextPath string
	PublicServiceAPIContextPath string
	CORSAllowedOrigins          []string
	CORSAllowedMethods          []string
	CORSAllowedHeaders          []string
	CORSMaxAge                  uint16
	CSRFTokenName               string
	CSRFTokenSameSite           string
	CSRFTokenMaxAge             time.Duration
	CSRFTokenCookieSecure       bool
	CSRFTokenCookieHTTPOnly     bool
	CSRFTokenCookieSessionOnly  bool
	CSRFTokenSingleUseToken     bool
	RequestBodyLimit            int
	BrowserIPRateLimit          uint16
	ServiceIPRateLimit          uint16
	AllowedIPs                  []string
	AllowedCIDRs                []string
	DatabaseContainer           string
	DatabaseURL                 string
	DatabaseInitTotalTimeout    time.Duration
	DatabaseInitRetryWait       time.Duration
	ServerShutdownTimeout       time.Duration
	OTLPEnabled                 bool
	OTLPConsole                 bool
	OTLPService                 string
	OTLPInstance                string
	OTLPVersion                 string
	OTLPEnvironment             string
	OTLPHostname                string
	OTLPEndpoint                string
	UnsealMode                  string
	UnsealFiles                 []string
}

// PrivateBaseURL returns the private base URL constructed from protocol, address, and port.
func (s *Settings) PrivateBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.BindPrivateProtocol, s.BindPrivateAddress, s.BindPrivatePort)
}

// PublicBaseURL returns the public base URL constructed from protocol, address, and port.
func (s *Settings) PublicBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.BindPublicProtocol, s.BindPublicAddress, s.BindPublicPort)
}

// Setting Input values for pflag.*P(name, shortname, value, usage).
type Setting struct {
	name        string // unique long name for the flag.
	shorthand   string // unique short name for the flag.
	value       any    // default value for the flag.
	usage       string // description of the flag for help text.
	description string // human-readable description for logging/display.
	redacted    bool   // whether to redact the value in logs (except in dev+verbose mode).
}

type analysisResult struct {
	SettingsByNames      map[string][]*Setting
	SettingsByShorthands map[string][]*Setting
	DuplicateNames       []string
	DuplicateShorthands  []string
}

var allRegisteredSettings []*Setting
