package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	httpProtocol  = "http"
	httpsProtocol = "https"

	localhost                 = "localhost"
	ipv4Loopback              = "127.0.0.1"
	ipv6Loopback              = "::1"
	ipv4MappedIPv6Loopback    = "::ffff:127.0.0.1"
	ipv6LoopbackURL           = "[::1]"
	ipv4MappedIPv6LoopbackURL = "[::ffff:127.0.0.1]"

	localhostCIDRv4        = "127.0.0.0/8"
	linkLocalCIDRv4        = "169.254.0.0/16"
	privateLANClassACIDRv4 = "10.0.0.0/8"
	privateLANClassBCIDRv4 = "172.16.0.0/12"
	privateLANClassCCIDRv4 = "192.168.0.0/16"

	localhostCIDRv6  = "::1/128"
	linkLocalCIDRv6  = "fe80::/10"
	privateLANCIDRv6 = "fc00::/7"

	defaultConfigFile                  = "config.yaml"
	defaultLogLevel                    = "INFO"
	defaultBindPublicProtocol          = httpsProtocol
	defaultBindPublicAddress           = localhost
	defaultBindPublicPort              = uint16(8080)
	defaultBindPrivateProtocol         = httpProtocol
	defaultBindPrivateAddress          = localhost
	defaultBindPrivatePort             = uint16(9090)
	defaultPublicBrowserAPIContextPath = "/browser/api/v1"
	defaultPublicServiceAPIContextPath = "/service/api/v1"
	defaultCORSMaxAge                  = uint16(3600)
	defaultCSRFTokenName               = "_csrf"
	defaultCSRFTokenSameSite           = "Strict"
	defaultCSRFTokenMaxAge             = 1 * time.Hour
	defaultCSRFTokenCookieSecure       = true
	defaultCSRFTokenCookieHTTPOnly     = false
	defaultCSRFTokenCookieSessionOnly  = true
	defaultCSRFTokenSingleUseToken     = false
	defaultIPRateLimit                 = uint16(50)
	defaultDatabaseContainer           = "disabled"
	defaultDatabaseURL                 = "postgres://USR:PWD@localhost:5432/DB?sslmode=disable"
	defaultDatabaseInitTotalTimeout    = 5 * time.Minute
	defaultDatabaseInitRetryWait       = 1 * time.Second
	defaultHelp                        = false
	defaultVerboseMode                 = false
	defaultDevMode                     = false
	defaultDryRun                      = false
	defaultOTLP                        = false
	defaultOTLPConsole                 = false
	defaultOTLPScope                   = "cryptoutil"
	defaultUnsealMode                  = "sysinfo"
)

var defaultBindPostString = strconv.Itoa(int(registerAsUint16Setting(&bindPublicPort)))

var defaultCORSAllowedOrigins = []string{
	httpProtocol + "://" + localhost + ":" + defaultBindPostString,
	httpProtocol + "://" + ipv4Loopback + ":" + defaultBindPostString,
	httpProtocol + "://" + ipv6LoopbackURL + ":" + defaultBindPostString,
	httpProtocol + "://" + ipv4MappedIPv6LoopbackURL + ":" + defaultBindPostString,
	httpsProtocol + "://" + localhost + ":" + defaultBindPostString,
	httpsProtocol + "://" + ipv4Loopback + ":" + defaultBindPostString,
	httpsProtocol + "://" + ipv6LoopbackURL + ":" + defaultBindPostString,
	httpsProtocol + "://" + ipv4MappedIPv6LoopbackURL + ":" + defaultBindPostString,
}

var defaultAllowedIps = []string{localhost, ipv4Loopback, ipv6Loopback, ipv4MappedIPv6Loopback}

var defaultTLSPublicDNSNames = []string{localhost}

var defaultTLSPublicIPAddresses = []string{ipv4Loopback, ipv6Loopback, ipv4MappedIPv6Loopback}

var defaultTLSPrivateDNSNames = []string{localhost}

var defaultTLSPrivateIPAddresses = []string{ipv4Loopback, ipv6Loopback, ipv4MappedIPv6Loopback}

var defaultAllowedCIDRs = []string{
	localhostCIDRv4,
	linkLocalCIDRv4,
	privateLANClassACIDRv4,
	privateLANClassBCIDRv4,
	privateLANClassCCIDRv4,
	localhostCIDRv6,
	linkLocalCIDRv6,
	privateLANCIDRv6,
}

var defaultCORSAllowedMethods = []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"}

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

var defaultUnsealFiles = []string{}

var defaultConfigFiles = []string{}

// set of valid subcommands.
var subcommands = map[string]struct{}{
	"start": {},
	"stop":  {},
	"init":  {},
}

var allRegisteredSettings []*Setting

type Settings struct {
	SubCommand                  string
	Help                        bool
	ConfigFile                  []string
	LogLevel                    string
	VerboseMode                 bool
	DevMode                     bool
	DryRun                      bool
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

type analysisResult struct {
	SettingsByNames      map[string][]*Setting
	SettingsByShorthands map[string][]*Setting
	DuplicateNames       []string
	DuplicateShorthands  []string
}

var (
	help = *registerSetting(&Setting{
		name:      "help",
		shorthand: "h",
		value:     defaultHelp,
		usage: "print help; you can run the server with parameters like this:\n" +
			"cmd -l=INFO -v -M -u=postgres://USR:PWD@localhost:5432/DB?sslmode=disable\n",
		description: "Help",
	})
	configFile = *registerSetting(&Setting{
		name:        "config",
		shorthand:   "y",
		value:       defaultConfigFiles,
		usage:       "path to config file (can be specified multiple times)",
		description: "Config files",
	})
	logLevel = *registerSetting(&Setting{
		name:        "log-level",
		shorthand:   "l",
		value:       defaultLogLevel,
		usage:       "log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF",
		description: "Log Level",
	})
	verboseMode = *registerSetting(&Setting{
		name:        "verbose",
		shorthand:   "v",
		value:       defaultVerboseMode,
		usage:       "verbose modifier for log level",
		description: "Verbose mode",
	})
	devMode = *registerSetting(&Setting{
		name:        "dev",
		shorthand:   "d",
		value:       defaultDevMode,
		usage:       "run in development mode; enables in-memory SQLite",
		description: "Dev mode",
	})
	dryRun = *registerSetting(&Setting{
		name:        "dry-run",
		shorthand:   "Y",
		value:       defaultDryRun,
		usage:       "validate configuration and exit without starting server",
		description: "Dry run",
	})
	bindPublicProtocol = *registerSetting(&Setting{
		name:        "bind-public-protocol",
		shorthand:   "t",
		value:       defaultBindPublicProtocol,
		usage:       "bind public protocol (http or https)",
		description: "Bind Public Protocol",
	})
	bindPublicAddress = *registerSetting(&Setting{
		name:        "bind-public-address",
		shorthand:   "a",
		value:       defaultBindPublicAddress,
		usage:       "bind public address",
		description: "Bind Public Address",
	})
	bindPublicPort = *registerSetting(&Setting{
		name:        "bind-public-port",
		shorthand:   "p",
		value:       defaultBindPublicPort,
		usage:       "bind public port",
		description: "Bind Public Port",
	})
	bindPrivateProtocol = *registerSetting(&Setting{
		name:        "bind-private-protocol",
		shorthand:   "T",
		value:       defaultBindPrivateProtocol, // TODO https
		usage:       "bind private protocol (http or https)",
		description: "Bind Private Protocol",
	})
	bindPrivateAddress = *registerSetting(&Setting{
		name:        "bind-private-address",
		shorthand:   "A",
		value:       defaultBindPrivateAddress,
		usage:       "bind private address",
		description: "Bind Private Address",
	})
	bindPrivatePort = *registerSetting(&Setting{
		name:        "bind-private-port",
		shorthand:   "P",
		value:       defaultBindPrivatePort,
		usage:       "bind private port",
		description: "Bind Private Port",
	})
	tlsPublicDNSNames = *registerSetting(&Setting{
		name:        "tls-public-dns-names",
		shorthand:   "n",
		value:       defaultTLSPublicDNSNames,
		usage:       "TLS public DNS names",
		description: "TLS Public DNS Names",
	})
	tlsPrivateDNSNames = *registerSetting(&Setting{
		name:        "tls-private-dns-names",
		shorthand:   "j",
		value:       defaultTLSPrivateDNSNames,
		usage:       "TLS private DNS names",
		description: "TLS Private DNS Names",
	})
	tlsPublicIPAddresses = *registerSetting(&Setting{
		name:        "tls-public-ip-addresses",
		shorthand:   "i",
		value:       defaultTLSPublicIPAddresses,
		usage:       "TLS public IP addresses",
		description: "TLS Public IP Addresses",
	})
	tlsPrivateIPAddresses = *registerSetting(&Setting{
		name:        "tls-private-ip-addresses",
		shorthand:   "k",
		value:       defaultTLSPrivateIPAddresses,
		usage:       "TLS private IP addresses",
		description: "TLS Private IP Addresses",
	})
	publicBrowserAPIContextPath = *registerSetting(&Setting{
		name:        "browser-api-context-path",
		shorthand:   "c",
		value:       defaultPublicBrowserAPIContextPath,
		usage:       "context path for Public Browser API",
		description: "Public Browser API Context Path",
	})
	publicServiceAPIContextPath = *registerSetting(&Setting{
		name:        "service-api-context-path",
		shorthand:   "b",
		value:       defaultPublicServiceAPIContextPath,
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
		value:       defaultCSRFTokenCookieSecure,
		usage:       "CSRF token cookie Secure attribute",
		description: "CSRF Token Cookie Secure",
	})
	csrfTokenCookieHTTPOnly = *registerSetting(&Setting{
		name:        "csrf-token-cookie-http-only",
		shorthand:   "J",
		value:       defaultCSRFTokenCookieHTTPOnly, // False needed for Swagger UI submit CSRF workaround
		usage:       "CSRF token cookie HttpOnly attribute",
		description: "CSRF Token Cookie HTTPOnly",
	})
	csrfTokenCookieSessionOnly = *registerSetting(&Setting{
		name:        "csrf-token-cookie-session-only",
		shorthand:   "E",
		value:       defaultCSRFTokenCookieSessionOnly,
		usage:       "CSRF token cookie SessionOnly attribute",
		description: "CSRF Token Cookie SessionOnly",
	})
	csrfTokenSingleUseToken = *registerSetting(&Setting{
		name:        "csrf-token-single-use-token",
		shorthand:   "G",
		value:       defaultCSRFTokenSingleUseToken,
		usage:       "CSRF token SingleUse attribute",
		description: "CSRF Token SingleUseToken",
	})
	ipRateLimit = *registerSetting(&Setting{
		name:        "rate-limit",
		shorthand:   "r",
		value:       defaultIPRateLimit,
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
		value:       defaultDatabaseContainer,
		usage:       "database container mode; true to use container, false to use local database",
		description: "Database Container",
	})
	databaseURL = *registerSetting(&Setting{
		name:        "database-url",
		shorthand:   "u",
		value:       defaultDatabaseURL,
		usage:       "database URL; start a container with:\ndocker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=USR -e POSTGRES_PASSWORD=PWD -e POSTGRES_DB=DB postgres:latest\n",
		description: "Database URL",
		redacted:    true,
	})
	databaseInitTotalTimeout = *registerSetting(&Setting{
		name:        "database-init-total-timeout",
		shorthand:   "Z",
		value:       defaultDatabaseInitTotalTimeout,
		usage:       "database init total timeout",
		description: "Database Init Total Timeout",
	})
	databaseInitRetryWait = *registerSetting(&Setting{
		name:        "database-init-retry-wait",
		shorthand:   "W",
		value:       defaultDatabaseInitRetryWait,
		usage:       "database init retry wait",
		description: "Database Init Retry Wait",
	})
	otlp = *registerSetting(&Setting{
		name:        "otlp",
		shorthand:   "z",
		value:       defaultOTLP,
		usage:       "enable OTLP export",
		description: "OTLP Export",
	})
	otlpConsole = *registerSetting(&Setting{
		name:        "otlp-console",
		shorthand:   "q",
		value:       defaultOTLPConsole,
		usage:       "enable OTLP logging to console (STDOUT)",
		description: "OTLP Console",
	})
	otlpScope = *registerSetting(&Setting{
		name:        "otlp-scope",
		shorthand:   "s",
		value:       defaultOTLPScope,
		usage:       "OTLP scope",
		description: "OTLP Scope",
	})
	unsealMode = *registerSetting(&Setting{
		name:        "unseal-mode",
		shorthand:   "U",
		value:       defaultUnsealMode,
		usage:       "unseal mode: N, M-of-N, sysinfo; N keys, or M-of-N derived keys from shared secrets, or X-of-Y custom sysinfo as shared secrets",
		description: "Unseal Mode",
	})
	unsealFiles = *registerSetting(&Setting{
		name:      "unseal-files",
		shorthand: "F",
		value:     defaultUnsealFiles,
		usage: "unseal files; repeat for multiple files; e.g. " +
			"\"--unseal-files=/docker/secrets/unseal_1of3 --unseal-files=/docker/secrets/unseal_2of3\"; " +
			"used for N unseal keys or M-of-N unseal shared secrets",
		description: "Unseal Files",
	})
)

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

	// Enable environment variable support with CRYPTOUTIL_ prefix BEFORE parsing flags
	viper.SetEnvPrefix("CRYPTOUTIL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Explicitly bind boolean environment variables (viper.AutomaticEnv may not handle booleans correctly)
	viper.BindEnv("verbose", "CRYPTOUTIL_VERBOSE")
	viper.BindEnv("dev", "CRYPTOUTIL_DEV_MODE")
	viper.BindEnv("dry-run", "CRYPTOUTIL_DRY_RUN")
	viper.BindEnv("otlp", "CRYPTOUTIL_OTLP")
	viper.BindEnv("otlp-console", "CRYPTOUTIL_OTLP_CONSOLE")
	viper.BindEnv("csrf-token-cookie-secure", "CRYPTOUTIL_CSRF_TOKEN_COOKIE_SECURE")
	viper.BindEnv("csrf-token-cookie-http-only", "CRYPTOUTIL_CSRF_TOKEN_COOKIE_HTTP_ONLY")
	viper.BindEnv("csrf-token-cookie-session-only", "CRYPTOUTIL_CSRF_TOKEN_COOKIE_SESSION_ONLY")
	viper.BindEnv("csrf-token-single-use-token", "CRYPTOUTIL_CSRF_TOKEN_SINGLE_USE_TOKEN")

	// pflag will parse subCommandParameters, and viper will union them with config file contents (if specified)
	pflag.BoolP(help.name, help.shorthand, registerAsBoolSetting(&help), help.usage)
	pflag.StringSliceP(configFile.name, configFile.shorthand, registerAsStringSliceSetting(&configFile), configFile.usage)
	pflag.StringP(logLevel.name, logLevel.shorthand, registerAsStringSetting(&logLevel), logLevel.usage)
	pflag.BoolP(verboseMode.name, verboseMode.shorthand, registerAsBoolSetting(&verboseMode), verboseMode.usage)
	pflag.BoolP(devMode.name, devMode.shorthand, registerAsBoolSetting(&devMode), devMode.usage)
	pflag.BoolP(dryRun.name, dryRun.shorthand, registerAsBoolSetting(&dryRun), dryRun.usage)
	pflag.StringP(bindPublicProtocol.name, bindPublicProtocol.shorthand, registerAsStringSetting(&bindPublicProtocol), bindPublicProtocol.usage)
	pflag.StringP(bindPublicAddress.name, bindPublicAddress.shorthand, registerAsStringSetting(&bindPublicAddress), bindPublicAddress.usage)
	pflag.Uint16P(bindPublicPort.name, bindPublicPort.shorthand, registerAsUint16Setting(&bindPublicPort), bindPublicPort.usage)
	pflag.StringSliceP(tlsPublicDNSNames.name, tlsPublicDNSNames.shorthand, registerAsStringSliceSetting(&tlsPublicDNSNames), tlsPublicDNSNames.usage)
	pflag.StringSliceP(tlsPublicIPAddresses.name, tlsPublicIPAddresses.shorthand, registerAsStringSliceSetting(&tlsPublicIPAddresses), tlsPublicIPAddresses.usage)
	pflag.StringSliceP(tlsPrivateDNSNames.name, tlsPrivateDNSNames.shorthand, registerAsStringSliceSetting(&tlsPrivateDNSNames), tlsPrivateDNSNames.usage)
	pflag.StringSliceP(tlsPrivateIPAddresses.name, tlsPrivateIPAddresses.shorthand, registerAsStringSliceSetting(&tlsPrivateIPAddresses), tlsPrivateIPAddresses.usage)
	pflag.StringP(bindPrivateProtocol.name, bindPrivateProtocol.shorthand, registerAsStringSetting(&bindPrivateProtocol), bindPrivateProtocol.usage)
	pflag.StringP(bindPrivateAddress.name, bindPrivateAddress.shorthand, registerAsStringSetting(&bindPrivateAddress), bindPrivateAddress.usage)
	pflag.Uint16P(bindPrivatePort.name, bindPrivatePort.shorthand, registerAsUint16Setting(&bindPrivatePort), bindPrivatePort.usage)
	pflag.StringP(publicBrowserAPIContextPath.name, publicBrowserAPIContextPath.shorthand, registerAsStringSetting(&publicBrowserAPIContextPath), publicBrowserAPIContextPath.usage)
	pflag.StringP(publicServiceAPIContextPath.name, publicServiceAPIContextPath.shorthand, registerAsStringSetting(&publicServiceAPIContextPath), publicServiceAPIContextPath.usage)
	pflag.StringSliceP(corsAllowedOrigins.name, corsAllowedOrigins.shorthand, registerAsStringSliceSetting(&corsAllowedOrigins), corsAllowedOrigins.usage)
	pflag.StringSliceP(corsAllowedMethods.name, corsAllowedMethods.shorthand, registerAsStringSliceSetting(&corsAllowedMethods), corsAllowedMethods.usage)
	pflag.StringSliceP(corsAllowedHeaders.name, corsAllowedHeaders.shorthand, registerAsStringSliceSetting(&corsAllowedHeaders), corsAllowedHeaders.usage)
	pflag.Uint16P(corsMaxAge.name, corsMaxAge.shorthand, registerAsUint16Setting(&corsMaxAge), corsMaxAge.usage)
	pflag.StringP(csrfTokenName.name, csrfTokenName.shorthand, registerAsStringSetting(&csrfTokenName), csrfTokenName.usage)
	pflag.StringP(csrfTokenSameSite.name, csrfTokenSameSite.shorthand, registerAsStringSetting(&csrfTokenSameSite), csrfTokenSameSite.usage)
	pflag.DurationP(csrfTokenMaxAge.name, csrfTokenMaxAge.shorthand, registerAsDurationSetting(&csrfTokenMaxAge), csrfTokenMaxAge.usage)
	pflag.BoolP(csrfTokenCookieSecure.name, csrfTokenCookieSecure.shorthand, registerAsBoolSetting(&csrfTokenCookieSecure), csrfTokenCookieSecure.usage)
	pflag.BoolP(csrfTokenCookieHTTPOnly.name, csrfTokenCookieHTTPOnly.shorthand, registerAsBoolSetting(&csrfTokenCookieHTTPOnly), csrfTokenCookieHTTPOnly.usage)
	pflag.BoolP(csrfTokenCookieSessionOnly.name, csrfTokenCookieSessionOnly.shorthand, registerAsBoolSetting(&csrfTokenCookieSessionOnly), csrfTokenCookieSessionOnly.usage)
	pflag.BoolP(csrfTokenSingleUseToken.name, csrfTokenSingleUseToken.shorthand, registerAsBoolSetting(&csrfTokenSingleUseToken), csrfTokenSingleUseToken.usage)
	pflag.Uint16P(ipRateLimit.name, ipRateLimit.shorthand, registerAsUint16Setting(&ipRateLimit), ipRateLimit.usage)
	pflag.StringSliceP(allowedIps.name, allowedIps.shorthand, registerAsStringSliceSetting(&allowedIps), allowedIps.usage)
	pflag.StringSliceP(allowedCidrs.name, allowedCidrs.shorthand, registerAsStringSliceSetting(&allowedCidrs), allowedCidrs.usage)
	pflag.StringP(databaseContainer.name, databaseContainer.shorthand, registerAsStringSetting(&databaseContainer), databaseContainer.usage)
	pflag.StringP(databaseURL.name, databaseURL.shorthand, registerAsStringSetting(&databaseURL), databaseURL.usage)
	pflag.DurationP(databaseInitTotalTimeout.name, databaseInitTotalTimeout.shorthand, registerAsDurationSetting(&databaseInitTotalTimeout), databaseInitTotalTimeout.usage)
	pflag.DurationP(databaseInitRetryWait.name, databaseInitRetryWait.shorthand, registerAsDurationSetting(&databaseInitRetryWait), databaseInitRetryWait.usage)
	pflag.BoolP(otlp.name, otlp.shorthand, registerAsBoolSetting(&otlp), otlp.usage)
	pflag.BoolP(otlpConsole.name, otlpConsole.shorthand, registerAsBoolSetting(&otlpConsole), otlpConsole.usage)
	pflag.StringP(otlpScope.name, otlpScope.shorthand, registerAsStringSetting(&otlpScope), otlpScope.usage)
	pflag.StringP(unsealMode.name, unsealMode.shorthand, registerAsStringSetting(&unsealMode), unsealMode.usage)
	pflag.StringArrayP(unsealFiles.name, unsealFiles.shorthand, registerAsStringArraySetting(&unsealFiles), unsealFiles.usage)
	err := pflag.CommandLine.Parse(subCommandParameters)
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	configFiles := viper.GetStringSlice(configFile.name)
	if len(configFiles) > 0 {
		// Set the first config file
		if info, err := os.Stat(configFiles[0]); err == nil && !info.IsDir() {
			viper.SetConfigFile(configFiles[0])
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file %s: %w", configFiles[0], err)
			}
		}
		// Merge additional config files
		for i := 1; i < len(configFiles); i++ {
			if info, err := os.Stat(configFiles[i]); err == nil && !info.IsDir() {
				viper.SetConfigFile(configFiles[i])
				if err := viper.MergeInConfig(); err != nil {
					return nil, fmt.Errorf("error merging config file %s: %w", configFiles[i], err)
				}
			}
		}
	}

	s := &Settings{
		SubCommand:                  subCommand,
		Help:                        viper.GetBool(help.name),
		ConfigFile:                  viper.GetStringSlice(configFile.name),
		LogLevel:                    viper.GetString(logLevel.name),
		VerboseMode:                 viper.GetBool(verboseMode.name),
		DevMode:                     viper.GetBool(devMode.name),
		DryRun:                      viper.GetBool(dryRun.name),
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
		fmt.Println("cryptoutil - Cryptographic utility server")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  cryptoutil [subcommand] [flags]")
		fmt.Println()
		fmt.Println("SUBCOMMANDS:")
		fmt.Println("  start    Start the server")
		fmt.Println("  stop     Stop the server")
		fmt.Println("  init     Initialize the server")
		fmt.Println()
		fmt.Println("SERVER SETTINGS:")
		fmt.Println("  -a, --bind-public-address string    bind public address (default \"localhost\")")
		fmt.Println("  -p, --bind-public-port uint16       bind public port (default 8080)")
		fmt.Println("  -t, --bind-public-protocol string   bind public protocol (http or https) (default \"https\")")
		fmt.Println("  -A, --bind-private-address string   bind private address (default \"localhost\")")
		fmt.Println("  -P, --bind-private-port uint16      bind private port (default 9090)")
		fmt.Println("  -T, --bind-private-protocol string  bind private protocol (http or https) (default \"http\")")
		fmt.Println("  -c, --browser-api-context-path string  context path for Public Browser API (default \"/browser/api/v1\")")
		fmt.Println("  -b, --service-api-context-path string  context path for Public Service API (default \"/service/api/v1\")")
		fmt.Println()
		fmt.Println("SECURITY SETTINGS:")
		fmt.Println()
		fmt.Println("NETWORK SECURITY SETTINGS:")
		fmt.Println("  -I, --allowed-ips strings           comma-separated list of allowed IPs (default [localhost,127.0.0.1,::1,::ffff:127.0.0.1])")
		fmt.Println("  -C, --allowed-cidrs strings         comma-separated list of allowed CIDRs (default [127.0.0.0/8,169.254.0.0/16,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,::1/128,fe80::/10,fc00::/7])")
		fmt.Println("  -r, --rate-limit uint16             rate limit requests per second (default 50)")
		fmt.Println()
		fmt.Println("BROWSER CORS SECURITY SETTINGS:")
		fmt.Println("  -o, --cors-origins strings          CORS allowed origins")
		fmt.Println("  -m, --cors-methods strings          CORS allowed methods (default [POST,GET,PUT,DELETE,OPTIONS])")
		fmt.Println("  -H, --cors-headers strings          CORS allowed headers (default [Content-Type,Authorization,Accept,Origin,X-Requested-With,Cache-Control,Pragma,Expires,_csrf])")
		fmt.Println("  -x, --cors-max-age uint16           CORS max age in seconds (default 3600)")
		fmt.Println()
		fmt.Println("BROWSER CSRF SECURITY SETTINGS:")
		fmt.Println("  -N, --csrf-token-name string        CSRF token name (default \"_csrf\")")
		fmt.Println("  -S, --csrf-token-same-site string   CSRF token SameSite attribute (default \"Strict\")")
		fmt.Println("  -M, --csrf-token-max-age duration   CSRF token max age (expiration) (default 1h0m0s)")
		fmt.Println("  -R, --csrf-token-cookie-secure      CSRF token cookie Secure attribute (default true)")
		fmt.Println("  -J, --csrf-token-cookie-http-only   CSRF token cookie HttpOnly attribute (default false)")
		fmt.Println("  -E, --csrf-token-cookie-session-only CSRF token cookie SessionOnly attribute (default true)")
		fmt.Println("  -G, --csrf-token-single-use-token   CSRF token SingleUse attribute (default false)")
		fmt.Println()
		fmt.Println("TLS SETTINGS:")
		fmt.Println("  -n, --tls-public-dns-names strings  TLS public DNS names (default [localhost])")
		fmt.Println("  -i, --tls-public-ip-addresses strings TLS public IP addresses (default [127.0.0.1,::1,::ffff:127.0.0.1])")
		fmt.Println("  -j, --tls-private-dns-names strings TLS private DNS names (default [localhost])")
		fmt.Println("  -k, --tls-private-ip-addresses strings TLS private IP addresses (default [127.0.0.1,::1,::ffff:127.0.0.1])")
		fmt.Println()
		fmt.Println("DATABASE SETTINGS:")
		fmt.Println("  -u, --database-url string           database URL (default \"postgres://USR:PWD@localhost:5432/DB?sslmode=disable\")")
		fmt.Println("  -D, --database-container string     database container mode (default \"disabled\")")
		fmt.Println("  -Z, --database-init-total-timeout duration database init total timeout (default 5m0s)")
		fmt.Println("  -W, --database-init-retry-wait duration database init retry wait (default 1s)")
		fmt.Println()
		fmt.Println("BARRIER SETTINGS:")
		fmt.Println("  -U, --unseal-mode string            unseal mode: N, M-of-N, sysinfo (default \"sysinfo\")")
		fmt.Println("  -F, --unseal-files strings          unseal files")
		fmt.Println()
		fmt.Println("OBSERVABILITY SETTINGS:")
		fmt.Println("  -l, --log-level string              log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF (default \"INFO\")")
		fmt.Println("  -v, --verbose                       verbose modifier for log level")
		fmt.Println("  -z, --otlp                          enable OTLP export")
		fmt.Println("  -q, --otlp-console                  enable OTLP logging to console (STDOUT)")
		fmt.Println("  -s, --otlp-scope string             OTLP scope (default \"cryptoutil\")")
		fmt.Println()
		fmt.Println("CONFIGURATION SETTINGS:")
		fmt.Println("  -y, --config strings                path to config file (can be specified multiple times)")
		fmt.Println("  -d, --dev                           run in development mode; enables in-memory SQLite")
		fmt.Println("  -Y, --dry-run                       validate configuration and exit without starting server")
		fmt.Println("  -h, --help                          print help")
		fmt.Println()
		fmt.Println("ENVIRONMENT VARIABLES:")
		fmt.Println("  All flags can be set via environment variables using the CRYPTOUTIL_ prefix.")
		fmt.Println("  Examples: CRYPTOUTIL_LOG_LEVEL=DEBUG, CRYPTOUTIL_DATABASE_URL=...")
		fmt.Println()
		fmt.Println("Quickstart Examples:")
		fmt.Println("  cryptoutil start --dev                             # Start server with in-memory SQLite")
		fmt.Println("  cryptoutil start --database-container required     # Start server with PostgreSQL container")
		fmt.Println("  cryptoutil start --config global.yml --config preprod.yml  # Start server with settings in YAML config files")
		fmt.Println("  cryptoutil start --dry-run --config config.yml     # Validate configuration without starting")
		fmt.Println("  cryptoutil stop                                     # Stop server")
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
			dryRun.name:                      s.DryRun,
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

func registerSetting(setting *Setting) *Setting {
	allRegisteredSettings = append(allRegisteredSettings, setting)
	return setting
}

// Helper functions for safe type assertions in configuration.
func registerAsBoolSetting(s *Setting) bool {
	if v, ok := s.value.(bool); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not bool", s.name))
}

func registerAsStringSetting(s *Setting) string {
	if v, ok := s.value.(string); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not string", s.name))
}

func registerAsUint16Setting(s *Setting) uint16 {
	if v, ok := s.value.(uint16); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not uint16", s.name))
}

func registerAsStringSliceSetting(s *Setting) []string {
	if v, ok := s.value.([]string); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not []string", s.name))
}

func registerAsStringArraySetting(s *Setting) []string {
	if v, ok := s.value.([]string); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not []string for array", s.name))
}

func registerAsDurationSetting(s *Setting) time.Duration {
	if v, ok := s.value.(time.Duration); ok {
		return v
	}
	panic(fmt.Sprintf("setting %s value is not time.Duration", s.name))
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
