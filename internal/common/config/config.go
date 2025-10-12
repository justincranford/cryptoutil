package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	httpProtocol  = "http"
	httpsProtocol = "https"

	localhost    = "localhost"
	ipv4Loopback = "127.0.0.1"
	ipv6Loopback = "[::1]"

	localhostCIDRv4     = "127.0.0.0/8"
	linkLocalCIDRv4     = "169.254.0.0/16"
	privateClassACIDRv4 = "10.0.0.0/8"
	privateClassBCIDRv4 = "172.16.0.0/12"
	privateClassCCIDRv4 = "192.168.0.0/16"

	localhostCIDRv6 = "::1/128"
	linkLocalCIDRv6 = "fe80::/10"
	privateLANv6    = "fc00::/7"
)

var allRegisteredSettings []*Setting

type Settings struct {
	SubCommand                  string
	Help                        bool
	ConfigFile                  string
	LogLevel                    string
	VerboseMode                 bool
	DevMode                     bool
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
	IPRateLimit                 uint16
	AllowedIPs                  []string
	AllowedCIDRs                []string
	DatabaseContainer           string
	DatabaseURL                 string
	DatabaseInitTotalTimeout    time.Duration
	DatabaseInitRetryWait       time.Duration
	OTLP                        bool
	OTLPConsole                 bool
	OTLPScope                   string
	UnsealMode                  string
	UnsealFiles                 []string
}

// Setting Input values for pflag.*P(name, shortname, value, usage).
type Setting struct {
	name        string // unique long name for the flag
	shorthand   string // unique short name for the flag
	value       any    // default value for the flag
	usage       string // description of the flag for help text
	description string // human-readable description for logging/display
	redacted    bool   // whether to redact the value in logs (except in dev+verbose mode)
}

// Helper functions for safe type assertions in configuration.
func asBool(s *Setting) bool {
	if v, ok := s.value.(bool); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not bool", s.name))
}

func asString(s *Setting) string {
	if v, ok := s.value.(string); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not string", s.name))
}

func asUint16(s *Setting) uint16 {
	if v, ok := s.value.(uint16); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not uint16", s.name))
}

func asStringSlice(s *Setting) []string {
	if v, ok := s.value.([]string); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not []string", s.name))
}

func asStringArray(s *Setting) []string {
	if v, ok := s.value.([]string); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not []string for array", s.name))
}

func asDuration(s *Setting) time.Duration {
	if v, ok := s.value.(time.Duration); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not time.Duration", s.name))
}

var (
	help = *registerSetting(&Setting{
		name:      "help",
		shorthand: "h",
		value:     false,
		usage: "print help; you can run the server with parameters like this:\n" +
			"cmd -l=INFO -v -M -u=postgres://USR:PWD@localhost:5432/DB?sslmode=disable\n",
		description: "Help",
	})
	configFile = *registerSetting(&Setting{
		name:        "config",
		shorthand:   "y",
		value:       "config.yaml",
		usage:       "path to config file",
		description: "Config file",
	})
	logLevel = *registerSetting(&Setting{
		name:        "log-level",
		shorthand:   "l",
		value:       "INFO",
		usage:       "log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF",
		description: "Log Level",
	})
	verboseMode = *registerSetting(&Setting{
		name:        "verbose",
		shorthand:   "v",
		value:       false,
		usage:       "verbose modifier for log level",
		description: "Verbose mode",
	})
	devMode = *registerSetting(&Setting{
		name:        "dev",
		shorthand:   "d",
		value:       false,
		usage:       "run in development mode; enables in-memory SQLite",
		description: "Dev mode",
	})
	bindPublicProtocol = *registerSetting(&Setting{
		name:        "bind-public-protocol",
		shorthand:   "t",
		value:       httpsProtocol,
		usage:       "bind public protocol (http or https)",
		description: "Bind Public Protocol",
	})
	bindPublicAddress = *registerSetting(&Setting{
		name:        "bind-public-address",
		shorthand:   "a",
		value:       "localhost",
		usage:       "bind public address",
		description: "Bind Public Address",
	})
	bindPublicPort = *registerSetting(&Setting{
		name:        "bind-public-port",
		shorthand:   "p",
		value:       uint16(8080),
		usage:       "bind public port",
		description: "Bind Public Port",
	})
	bindPrivateProtocol = *registerSetting(&Setting{
		name:        "bind-private-protocol",
		shorthand:   "T",
		value:       httpProtocol, // TODO https
		usage:       "bind private protocol (http or https)",
		description: "Bind Private Protocol",
	})
	bindPrivateAddress = *registerSetting(&Setting{
		name:        "bind-private-address",
		shorthand:   "A",
		value:       "localhost",
		usage:       "bind private address",
		description: "Bind Private Address",
	})
	bindPrivatePort = *registerSetting(&Setting{
		name:        "bind-private-port",
		shorthand:   "P",
		value:       uint16(9090),
		usage:       "bind private port",
		description: "Bind Private Port",
	})
	tlsPublicDNSNames = *registerSetting(&Setting{
		name:        "tls-public-dns-names",
		shorthand:   "n",
		value:       []string{"localhost"},
		usage:       "TLS public DNS names",
		description: "TLS Public DNS Names",
	})
	tlsPrivateDNSNames = *registerSetting(&Setting{
		name:        "tls-private-dns-names",
		shorthand:   "j",
		value:       []string{"localhost"},
		usage:       "TLS private DNS names",
		description: "TLS Private DNS Names",
	})
	tlsPublicIPAddresses = *registerSetting(&Setting{
		name:        "tls-public-ip-addresses",
		shorthand:   "i",
		value:       []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"},
		usage:       "TLS public IP addresses",
		description: "TLS Public IP Addresses",
	})
	tlsPrivateIPAddresses = *registerSetting(&Setting{
		name:        "tls-private-ip-addresses",
		shorthand:   "k",
		value:       []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"},
		usage:       "TLS private IP addresses",
		description: "TLS Private IP Addresses",
	})
	publicBrowserAPIContextPath = *registerSetting(&Setting{
		name:        "browser-api-context-path",
		shorthand:   "c",
		value:       "/browser/api/v1",
		usage:       "context path for Public Browser API",
		description: "Public Browser API Context Path",
	})
	publicServiceAPIContextPath = *registerSetting(&Setting{
		name:        "service-api-context-path",
		shorthand:   "b",
		value:       "/service/api/v1",
		usage:       "context path for Public Server API",
		description: "Public Service API Context Path",
	})
	corsAllowedOrigins = *registerSetting(&Setting{
		name:        "cors-origins",
		shorthand:   "o",
		value:       defaultCORSAllowedOrigins,
		usage:       "CORS allowed origins",
		description: "CORS Allowed Origins",
	})
	corsAllowedMethods = *registerSetting(&Setting{
		name:        "cors-methods",
		shorthand:   "m",
		value:       defaultCORSAllowedMethods,
		usage:       "CORS allowed methods",
		description: "CORS Allowed Methods",
	})
	corsAllowedHeaders = *registerSetting(&Setting{
		name:        "cors-headers",
		shorthand:   "H",
		value:       defaultCORSAllowedHeaders,
		usage:       "CORS allowed headers",
		description: "CORS Allowed Headers",
	})
	corsMaxAge = *registerSetting(&Setting{
		name:        "cors-max-age",
		shorthand:   "x",
		value:       defaultCORSMaxAge,
		usage:       "CORS max age in seconds",
		description: "CORS Max Age",
	})
	csrfTokenName = *registerSetting(&Setting{
		name:        "csrf-token-name",
		shorthand:   "N",
		value:       defaultCSRFTokenName,
		usage:       "CSRF token name",
		description: "CSRF Token Name",
	})
	csrfTokenSameSite = *registerSetting(&Setting{
		name:        "csrf-token-same-site",
		shorthand:   "S",
		value:       defaultCSRFTokenSameSite,
		usage:       "CSRF token SameSite attribute",
		description: "CSRF Token SameSite",
	})
	csrfTokenMaxAge = *registerSetting(&Setting{
		name:        "csrf-token-max-age",
		shorthand:   "M",
		value:       defaultCSRFTokenMaxAge,
		usage:       "CSRF token max age (expiration)",
		description: "CSRF Token Max Age",
	})
	csrfTokenCookieSecure = *registerSetting(&Setting{
		name:        "csrf-token-cookie-secure",
		shorthand:   "R",
		value:       true,
		usage:       "CSRF token cookie Secure attribute",
		description: "CSRF Token Cookie Secure",
	})
	csrfTokenCookieHTTPOnly = *registerSetting(&Setting{
		name:        "csrf-token-cookie-http-only",
		shorthand:   "J",
		value:       false, // False needed for Swagger UI submit CSRF workaround
		usage:       "CSRF token cookie HttpOnly attribute",
		description: "CSRF Token Cookie HTTPOnly",
	})
	csrfTokenCookieSessionOnly = *registerSetting(&Setting{
		name:        "csrf-token-cookie-session-only",
		shorthand:   "E",
		value:       true,
		usage:       "CSRF token cookie SessionOnly attribute",
		description: "CSRF Token Cookie SessionOnly",
	})
	csrfTokenSingleUseToken = *registerSetting(&Setting{
		name:        "csrf-token-single-use-token",
		shorthand:   "G",
		value:       false,
		usage:       "CSRF token SingleUse attribute",
		description: "CSRF Token SingleUseToken",
	})
	ipRateLimit = *registerSetting(&Setting{
		name:        "rate-limit",
		shorthand:   "r",
		value:       uint16(50),
		usage:       "rate limit requests per second",
		description: "IP Rate Limit",
	})
	allowedIps = *registerSetting(&Setting{
		name:        "allowed-ips",
		shorthand:   "I",
		value:       defaultAllowedIps,
		usage:       "comma-separated list of allowed IPs",
		description: "Allowed IPs",
	})
	allowedCidrs = *registerSetting(&Setting{
		name:        "allowed-cidrs",
		shorthand:   "C",
		value:       defaultAllowedCIDRs,
		usage:       "comma-separated list of allowed CIDRs",
		description: "Allowed CIDRs",
	})
	databaseContainer = *registerSetting(&Setting{
		name:        "database-container",
		shorthand:   "D",
		value:       "disabled",
		usage:       "database container mode; true to use container, false to use local database",
		description: "Database Container",
	})
	databaseURL = *registerSetting(&Setting{
		name:        "database-url",
		shorthand:   "u",
		value:       "postgres://USR:PWD@localhost:5432/DB?sslmode=disable",
		usage:       "database URL; start a container with:\ndocker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=USR -e POSTGRES_PASSWORD=PWD -e POSTGRES_DB=DB postgres:latest\n",
		description: "Database URL",
		redacted:    true,
	})
	databaseInitTotalTimeout = *registerSetting(&Setting{
		name:        "database-init-total-timeout",
		shorthand:   "Z",
		value:       5 * time.Minute,
		usage:       "database init total timeout",
		description: "Database Init Total Timeout",
	})
	databaseInitRetryWait = *registerSetting(&Setting{
		name:        "database-init-retry-wait",
		shorthand:   "W",
		value:       1 * time.Second,
		usage:       "database init retry wait",
		description: "Database Init Retry Wait",
	})
	otlp = *registerSetting(&Setting{
		name:        "otlp",
		shorthand:   "z",
		value:       false,
		usage:       "enable OTLP export",
		description: "OTLP Export",
	})
	otlpConsole = *registerSetting(&Setting{
		name:        "otlp-console",
		shorthand:   "q",
		value:       false,
		usage:       "enable OTLP logging to console (STDOUT)",
		description: "OTLP Console",
	})
	otlpScope = *registerSetting(&Setting{
		name:        "otlp-scope",
		shorthand:   "s",
		value:       "cryptoutil",
		usage:       "OTLP scope",
		description: "OTLP Scope",
	})
	unsealMode = *registerSetting(&Setting{
		name:        "unseal-mode",
		shorthand:   "U",
		value:       "sysinfo",
		usage:       "unseal mode: N, M-of-N, sysinfo; N keys, or M-of-N derived keys from shared secrets, or X-of-Y custom sysinfo as shared secrets",
		description: "Unseal Mode",
	})
	unsealFiles = *registerSetting(&Setting{
		name:      "unseal-files",
		shorthand: "F",
		value:     []string{},
		usage: "unseal files; repeat for multiple files; e.g. " +
			"\"--unseal-files=/docker/secrets/unseal_1of3 --unseal-files=/docker/secrets/unseal_2of3\"; " +
			"used for N unseal keys or M-of-N unseal shared secrets",
		description: "Unseal Files",
	})
)

var defaultBindPostString = strconv.Itoa(int(asUint16(&bindPublicPort)))

var defaultCORSAllowedOrigins = []string{
	httpProtocol + "://" + localhost + ":" + defaultBindPostString,
	httpProtocol + "://" + ipv4Loopback + ":" + defaultBindPostString,
	httpProtocol + "://" + ipv6Loopback + ":" + defaultBindPostString,
	httpsProtocol + "://" + localhost + ":" + defaultBindPostString,
	httpsProtocol + "://" + ipv4Loopback + ":" + defaultBindPostString,
	httpsProtocol + "://" + ipv6Loopback + ":" + defaultBindPostString,
}

var defaultCORSAllowedMethods = []string{
	"POST",
	"GET",
	"PUT",
	"DELETE",
	"OPTIONS",
}

var defaultCORSAllowedHeaders = []string{
	"Content-Type",
	"Authorization",
	"Accept",
	"Origin",
	"X-Requested-With",
	"Cache-Control",
	"Pragma",
	"Expires",
	"_csrf",
}

var defaultCORSMaxAge = uint16(3600)

var defaultCSRFTokenName = "_csrf"

var defaultCSRFTokenSameSite = "Strict"

var defaultCSRFTokenMaxAge = 1 * time.Hour

var defaultAllowedIps = []string{
	"127.0.0.1",        // localhost (IPv4)
	"::1",              // localhost (IPv6)
	"::ffff:127.0.0.1", // localhost (IPv4-mapped IPv6)
}

var defaultAllowedCIDRs = []string{
	localhostCIDRv4,     // localhost (IPv4)
	linkLocalCIDRv4,     // link-local (IPv4)
	privateClassACIDRv4, // private LAN class A (IPv4)
	privateClassBCIDRv4, // private LAN class B (IPv4)
	privateClassCCIDRv4, // private LAN class C (IPv4)
	localhostCIDRv6,     // localhost (IPv6)
	linkLocalCIDRv6,     // link-local (IPv6)
	privateLANv6,        // private LAN (IPv6)
}

// set of valid subcommands.
var subcommands = map[string]struct{}{
	"start": {},
	"stop":  {},
	"init":  {},
}

func registerSetting(setting *Setting) *Setting {
	allRegisteredSettings = append(allRegisteredSettings, setting)
	return setting
}

type analysisResult struct {
	SettingsByNames      map[string][]*Setting
	SettingsByShorthands map[string][]*Setting
	DuplicateNames       []string
	DuplicateShorthands  []string
}

func analyzeSettings(settings []*Setting) analysisResult {
	result := analysisResult{
		SettingsByNames:      make(map[string][]*Setting),
		SettingsByShorthands: make(map[string][]*Setting),
	}
	for _, setting := range settings {
		result.SettingsByNames[setting.name] = append(result.SettingsByNames[setting.name], setting)
		result.SettingsByShorthands[setting.shorthand] = append(result.SettingsByShorthands[setting.shorthand], setting)
	}
	for _, setting := range settings {
		if len(result.SettingsByNames[setting.name]) > 1 {
			result.DuplicateNames = append(result.DuplicateNames, setting.name)
		}
		if len(result.SettingsByShorthands[setting.shorthand]) > 1 {
			result.DuplicateShorthands = append(result.DuplicateShorthands, setting.shorthand)
		}
	}
	return result
}

// Parse parses command line parameters and returns application settings.
func Parse(commandParameters []string, exitIfHelp bool) (*Settings, error) {
	if len(commandParameters) == 0 {
		return nil, fmt.Errorf("missing subcommand: use \"start\", \"stop\", or \"init\"")
	}
	subCommand := commandParameters[0]
	if _, ok := subcommands[subCommand]; !ok {
		return nil, fmt.Errorf("invalid subcommand: use \"start\", \"stop\", or \"init\"")
	}
	subCommandParameters := commandParameters[1:]

	// pflag will parse subCommandParameters, and viper will union them with config file contents (if specified)
	pflag.BoolP(help.name, help.shorthand, asBool(&help), help.usage)
	pflag.StringP(configFile.name, configFile.shorthand, asString(&configFile), configFile.usage)
	pflag.StringP(logLevel.name, logLevel.shorthand, asString(&logLevel), logLevel.usage)
	pflag.BoolP(verboseMode.name, verboseMode.shorthand, asBool(&verboseMode), verboseMode.usage)
	pflag.BoolP(devMode.name, devMode.shorthand, asBool(&devMode), devMode.usage)
	pflag.StringP(bindPublicProtocol.name, bindPublicProtocol.shorthand, asString(&bindPublicProtocol), bindPublicProtocol.usage)
	pflag.StringP(bindPublicAddress.name, bindPublicAddress.shorthand, asString(&bindPublicAddress), bindPublicAddress.usage)
	pflag.Uint16P(bindPublicPort.name, bindPublicPort.shorthand, asUint16(&bindPublicPort), bindPublicPort.usage)
	pflag.StringSliceP(tlsPublicDNSNames.name, tlsPublicDNSNames.shorthand, asStringSlice(&tlsPublicDNSNames), tlsPublicDNSNames.usage)
	pflag.StringSliceP(tlsPublicIPAddresses.name, tlsPublicIPAddresses.shorthand, asStringSlice(&tlsPublicIPAddresses), tlsPublicIPAddresses.usage)
	pflag.StringSliceP(tlsPrivateDNSNames.name, tlsPrivateDNSNames.shorthand, asStringSlice(&tlsPrivateDNSNames), tlsPrivateDNSNames.usage)
	pflag.StringSliceP(tlsPrivateIPAddresses.name, tlsPrivateIPAddresses.shorthand, asStringSlice(&tlsPrivateIPAddresses), tlsPrivateIPAddresses.usage)
	pflag.StringP(bindPrivateProtocol.name, bindPrivateProtocol.shorthand, asString(&bindPrivateProtocol), bindPrivateProtocol.usage)
	pflag.StringP(bindPrivateAddress.name, bindPrivateAddress.shorthand, asString(&bindPrivateAddress), bindPrivateAddress.usage)
	pflag.Uint16P(bindPrivatePort.name, bindPrivatePort.shorthand, asUint16(&bindPrivatePort), bindPrivatePort.usage)
	pflag.StringP(publicBrowserAPIContextPath.name, publicBrowserAPIContextPath.shorthand, asString(&publicBrowserAPIContextPath), publicBrowserAPIContextPath.usage)
	pflag.StringP(publicServiceAPIContextPath.name, publicServiceAPIContextPath.shorthand, asString(&publicServiceAPIContextPath), publicServiceAPIContextPath.usage)
	pflag.StringSliceP(corsAllowedOrigins.name, corsAllowedOrigins.shorthand, asStringSlice(&corsAllowedOrigins), corsAllowedOrigins.usage)
	pflag.StringSliceP(corsAllowedMethods.name, corsAllowedMethods.shorthand, asStringSlice(&corsAllowedMethods), corsAllowedMethods.usage)
	pflag.StringSliceP(corsAllowedHeaders.name, corsAllowedHeaders.shorthand, asStringSlice(&corsAllowedHeaders), corsAllowedHeaders.usage)
	pflag.Uint16P(corsMaxAge.name, corsMaxAge.shorthand, asUint16(&corsMaxAge), corsMaxAge.usage)
	pflag.StringP(csrfTokenName.name, csrfTokenName.shorthand, asString(&csrfTokenName), csrfTokenName.usage)
	pflag.StringP(csrfTokenSameSite.name, csrfTokenSameSite.shorthand, asString(&csrfTokenSameSite), csrfTokenSameSite.usage)
	pflag.DurationP(csrfTokenMaxAge.name, csrfTokenMaxAge.shorthand, asDuration(&csrfTokenMaxAge), csrfTokenMaxAge.usage)
	pflag.BoolP(csrfTokenCookieSecure.name, csrfTokenCookieSecure.shorthand, asBool(&csrfTokenCookieSecure), csrfTokenCookieSecure.usage)
	pflag.BoolP(csrfTokenCookieHTTPOnly.name, csrfTokenCookieHTTPOnly.shorthand, asBool(&csrfTokenCookieHTTPOnly), csrfTokenCookieHTTPOnly.usage)
	pflag.BoolP(csrfTokenCookieSessionOnly.name, csrfTokenCookieSessionOnly.shorthand, asBool(&csrfTokenCookieSessionOnly), csrfTokenCookieSessionOnly.usage)
	pflag.BoolP(csrfTokenSingleUseToken.name, csrfTokenSingleUseToken.shorthand, asBool(&csrfTokenSingleUseToken), csrfTokenSingleUseToken.usage)
	pflag.Uint16P(ipRateLimit.name, ipRateLimit.shorthand, asUint16(&ipRateLimit), ipRateLimit.usage)
	pflag.StringSliceP(allowedIps.name, allowedIps.shorthand, asStringSlice(&allowedIps), allowedIps.usage)
	pflag.StringSliceP(allowedCidrs.name, allowedCidrs.shorthand, asStringSlice(&allowedCidrs), allowedCidrs.usage)
	pflag.StringP(databaseContainer.name, databaseContainer.shorthand, asString(&databaseContainer), databaseContainer.usage)
	pflag.StringP(databaseURL.name, databaseURL.shorthand, asString(&databaseURL), databaseURL.usage)
	pflag.DurationP(databaseInitTotalTimeout.name, databaseInitTotalTimeout.shorthand, asDuration(&databaseInitTotalTimeout), databaseInitTotalTimeout.usage)
	pflag.DurationP(databaseInitRetryWait.name, databaseInitRetryWait.shorthand, asDuration(&databaseInitRetryWait), databaseInitRetryWait.usage)
	pflag.BoolP(otlp.name, otlp.shorthand, asBool(&otlp), otlp.usage)
	pflag.BoolP(otlpConsole.name, otlpConsole.shorthand, asBool(&otlpConsole), otlpConsole.usage)
	pflag.StringP(otlpScope.name, otlpScope.shorthand, asString(&otlpScope), otlpScope.usage)
	pflag.StringP(unsealMode.name, unsealMode.shorthand, asString(&unsealMode), unsealMode.usage)
	pflag.StringArrayP(unsealFiles.name, unsealFiles.shorthand, asStringArray(&unsealFiles), unsealFiles.usage)
	err := pflag.CommandLine.Parse(subCommandParameters)
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	if configFile := viper.GetString(configFile.name); configFile != "" {
		if info, err := os.Stat(configFile); err == nil && !info.IsDir() {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	s := &Settings{
		SubCommand:                  subCommand,
		Help:                        viper.GetBool(help.name),
		ConfigFile:                  viper.GetString(configFile.name),
		LogLevel:                    viper.GetString(logLevel.name),
		VerboseMode:                 viper.GetBool(verboseMode.name),
		DevMode:                     viper.GetBool(devMode.name),
		BindPublicProtocol:          viper.GetString(bindPublicProtocol.name),
		BindPublicAddress:           viper.GetString(bindPublicAddress.name),
		BindPublicPort:              viper.GetUint16(bindPublicPort.name),
		TLSPublicDNSNames:           viper.GetStringSlice(tlsPublicDNSNames.name),
		TLSPublicIPAddresses:        viper.GetStringSlice(tlsPublicIPAddresses.name),
		TLSPrivateDNSNames:          viper.GetStringSlice(tlsPrivateDNSNames.name),
		TLSPrivateIPAddresses:       viper.GetStringSlice(tlsPrivateIPAddresses.name),
		BindPrivateProtocol:         viper.GetString(bindPrivateProtocol.name),
		BindPrivateAddress:          viper.GetString(bindPrivateAddress.name),
		BindPrivatePort:             viper.GetUint16(bindPrivatePort.name),
		PublicBrowserAPIContextPath: viper.GetString(publicBrowserAPIContextPath.name),
		PublicServiceAPIContextPath: viper.GetString(publicServiceAPIContextPath.name),
		CORSAllowedOrigins:          viper.GetStringSlice(corsAllowedOrigins.name),
		CORSAllowedMethods:          viper.GetStringSlice(corsAllowedMethods.name),
		CORSAllowedHeaders:          viper.GetStringSlice(corsAllowedHeaders.name),
		CORSMaxAge:                  viper.GetUint16(corsMaxAge.name),
		CSRFTokenName:               viper.GetString(csrfTokenName.name),
		CSRFTokenSameSite:           viper.GetString(csrfTokenSameSite.name),
		CSRFTokenMaxAge:             viper.GetDuration(csrfTokenMaxAge.name),
		CSRFTokenCookieSecure:       viper.GetBool(csrfTokenCookieSecure.name),
		CSRFTokenCookieHTTPOnly:     viper.GetBool(csrfTokenCookieHTTPOnly.name),
		CSRFTokenCookieSessionOnly:  viper.GetBool(csrfTokenCookieSessionOnly.name),
		CSRFTokenSingleUseToken:     viper.GetBool(csrfTokenSingleUseToken.name),
		IPRateLimit:                 viper.GetUint16(ipRateLimit.name),
		AllowedIPs:                  viper.GetStringSlice(allowedIps.name),
		AllowedCIDRs:                viper.GetStringSlice(allowedCidrs.name),
		DatabaseContainer:           viper.GetString(databaseContainer.name),
		DatabaseURL:                 viper.GetString(databaseURL.name),
		DatabaseInitTotalTimeout:    viper.GetDuration(databaseInitTotalTimeout.name),
		DatabaseInitRetryWait:       viper.GetDuration(databaseInitRetryWait.name),
		OTLP:                        viper.GetBool(otlp.name),
		OTLPConsole:                 viper.GetBool(otlpConsole.name),
		OTLPScope:                   viper.GetString(otlpScope.name),
		UnsealMode:                  viper.GetString(unsealMode.name),
		UnsealFiles:                 viper.GetStringSlice(unsealFiles.name),
	}

	logSettings(s)

	if s.Help {
		pflag.CommandLine.SetOutput(os.Stdout)
		pflag.CommandLine.PrintDefaults()
		if exitIfHelp {
			os.Exit(0)
		}
	}

	return s, nil
}

func logSettings(s *Settings) {
	if s.VerboseMode {
		log.Info("Sub Command: ", s.SubCommand)

		// Create a map to get values by setting name
		valueMap := map[string]any{
			help.name:                        s.Help,
			configFile.name:                  s.ConfigFile,
			logLevel.name:                    s.LogLevel,
			verboseMode.name:                 s.VerboseMode,
			devMode.name:                     s.DevMode,
			bindPublicProtocol.name:          s.BindPublicProtocol,
			bindPublicAddress.name:           s.BindPublicAddress,
			bindPublicPort.name:              s.BindPublicPort,
			tlsPublicDNSNames.name:           s.TLSPublicDNSNames,
			tlsPublicIPAddresses.name:        s.TLSPublicIPAddresses,
			tlsPrivateDNSNames.name:          s.TLSPrivateDNSNames,
			tlsPrivateIPAddresses.name:       s.TLSPrivateIPAddresses,
			bindPrivateProtocol.name:         s.BindPrivateProtocol,
			bindPrivateAddress.name:          s.BindPrivateAddress,
			bindPrivatePort.name:             s.BindPrivatePort,
			publicBrowserAPIContextPath.name: s.PublicBrowserAPIContextPath,
			publicServiceAPIContextPath.name: s.PublicServiceAPIContextPath,
			corsAllowedOrigins.name:          s.CORSAllowedOrigins,
			corsAllowedMethods.name:          s.CORSAllowedMethods,
			corsAllowedHeaders.name:          s.CORSAllowedHeaders,
			corsMaxAge.name:                  s.CORSMaxAge,
			csrfTokenName.name:               s.CSRFTokenName,
			csrfTokenSameSite.name:           s.CSRFTokenSameSite,
			csrfTokenMaxAge.name:             s.CSRFTokenMaxAge,
			csrfTokenCookieSecure.name:       s.CSRFTokenCookieSecure,
			csrfTokenCookieHTTPOnly.name:     s.CSRFTokenCookieHTTPOnly,
			csrfTokenCookieSessionOnly.name:  s.CSRFTokenCookieSessionOnly,
			csrfTokenSingleUseToken.name:     s.CSRFTokenSingleUseToken,
			ipRateLimit.name:                 s.IPRateLimit,
			allowedIps.name:                  s.AllowedIPs,
			allowedCidrs.name:                s.AllowedCIDRs,
			databaseContainer.name:           s.DatabaseContainer,
			databaseURL.name:                 s.DatabaseURL,
			databaseInitTotalTimeout.name:    s.DatabaseInitTotalTimeout,
			databaseInitRetryWait.name:       s.DatabaseInitRetryWait,
			otlp.name:                        s.OTLP,
			otlpConsole.name:                 s.OTLPConsole,
			otlpScope.name:                   s.OTLPScope,
			unsealMode.name:                  s.UnsealMode,
			unsealFiles.name:                 s.UnsealFiles,
		}

		// Iterate through all registered settings and log them
		for _, setting := range allRegisteredSettings {
			value := valueMap[setting.name]
			if setting.redacted && !(s.DevMode && s.VerboseMode) {
				value = "REDACTED"
			}
			log.Info(setting.description+" (-"+setting.shorthand+"): ", value)
		}

		analysis := analyzeSettings(allRegisteredSettings)
		var usedShorthands []string
		var unusedShorthands []string

		// Check all letters (lowercase and uppercase) and digits
		allPossibleShorthands := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		for _, r := range allPossibleShorthands {
			possibleShorthand := string(r)
			if _, ok := analysis.SettingsByShorthands[possibleShorthand]; ok {
				usedShorthands = append(usedShorthands, possibleShorthand)
			} else {
				unusedShorthands = append(unusedShorthands, possibleShorthand)
			}
		}

		sort.Strings(usedShorthands)
		sort.Strings(unusedShorthands)
		log.Info("Shorthands, Used:   ", len(usedShorthands), ", Values: ", usedShorthands)
		log.Info("Shorthands, Unused: ", len(unusedShorthands), ", Values: ", unusedShorthands)
	}
}

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	viper.Reset()
}
