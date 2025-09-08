package config

import (
	"fmt"
	"os"
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
	CORSAllowedOrigins          string
	CORSAllowedMethods          string
	CORSAllowedHeaders          string
	CORSMaxAge                  uint16
	CSRFTokenName               string
	CSRFTokenSameSite           string
	CSRFTokenMaxAge             time.Duration
	CSRFTokenCookieSecure       bool
	CSRFTokenCookieHTTPOnly     bool
	CSRFTokenCookieSessionOnly  bool
	IPRateLimit                 uint16
	AllowedIPs                  string // TODO change to []string
	AllowedCIDRs                string // TODO change to []string
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

// Setting Input values for pflag.*P(name, shortname, value, usage)
type Setting struct {
	name      string // unique long name for the flag
	shorthand string // unique short name for the flag
	value     any    // default value for the flag
	usage     string // description of the flag
}

var (
	help = Setting{
		name:      "help",
		shorthand: "h",
		value:     false,
		usage: "print help; you can run the server with parameters like this:\n" +
			"cmd -l=INFO -v -M -u=postgres://USR:PWD@localhost:5432/DB?sslmode=disable\n",
	}
	configFile = Setting{
		name:      "config",
		shorthand: "y",
		value:     "config.yaml",
		usage:     "path to config file",
	}
	logLevel = Setting{
		name:      "log-level",
		shorthand: "l",
		value:     "INFO",
		usage:     "log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF",
	}
	verboseMode = Setting{
		name:      "verbose",
		shorthand: "v",
		value:     false,
		usage:     "verbose modifier for log level",
	}
	devMode = Setting{
		name:      "dev",
		shorthand: "d",
		value:     false,
		usage:     "run in development mode; enables in-memory SQLite",
	}
	bindPublicProtocol = Setting{
		name:      "bind-public-protocol",
		shorthand: "t",
		value:     httpsProtocol,
		usage:     "bind public protocol (http or https)",
	}
	bindPublicAddress = Setting{
		name:      "bind-public-address",
		shorthand: "a",
		value:     "localhost",
		usage:     "bind public address",
	}
	bindPublicPort = Setting{
		name:      "bind-public-port",
		shorthand: "p",
		value:     uint16(8080),
		usage:     "bind public port",
	}
	bindPrivateProtocol = Setting{
		name:      "bind-private-protocol",
		shorthand: "T",
		value:     httpProtocol, // TODO https
		usage:     "bind private protocol (http or https)",
	}
	bindPrivateAddress = Setting{
		name:      "bind-private-address",
		shorthand: "A",
		value:     "localhost",
		usage:     "bind private address",
	}
	bindPrivatePort = Setting{
		name:      "bind-private-port",
		shorthand: "P",
		value:     uint16(9090),
		usage:     "bind private port",
	}
	tlsPublicDnsNames = Setting{
		name:      "tls-public-dns-names",
		shorthand: "n",
		value:     []string{"localhost"},
		usage:     "TLS public DNS names",
	}
	tlsPrivateDnsNames = Setting{
		name:      "tls-private-dns-names",
		shorthand: "j",
		value:     []string{"localhost"},
		usage:     "TLS private DNS names",
	}
	tlsPublicIPAddresses = Setting{
		name:      "tls-public-ip-addresses",
		shorthand: "i",
		value:     []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"},
		usage:     "TLS public IP addresses",
	}
	tlsPrivateIPAddresses = Setting{
		name:      "tls-private-ip-addresses",
		shorthand: "k",
		value:     []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"},
		usage:     "TLS private IP addresses",
	}
	publicBrowserAPIContextPath = Setting{
		name:      "browser-api-context-path",
		shorthand: "c",
		value:     "/browser/api/v1",
		usage:     "context path for Public Browser API",
	}
	publicServiceAPIContextPath = Setting{
		name:      "service-api-context-path",
		shorthand: "b",
		value:     "/service/api/v1",
		usage:     "context path for Public Server API",
	}
	corsAllowedOrigins = Setting{
		name:      "cors-origins",
		shorthand: "o",
		value:     defaultCORSAllowedOrigins,
		usage:     "CORS allowed origins",
	}
	corsAllowedMethods = Setting{
		name:      "cors-methods",
		shorthand: "m",
		value:     defaultCORSAllowedMethods,
		usage:     "CORS allowed methods",
	}
	corsAllowedHeaders = Setting{
		name:      "cors-headers",
		shorthand: "H",
		value:     defaultCORSAllowedHeaders,
		usage:     "CORS allowed headers",
	}
	corsMaxAge = Setting{
		name:      "cors-max-age",
		shorthand: "x",
		value:     defaultCORSMaxAge,
		usage:     "CORS max age in seconds",
	}
	csrfTokenName = Setting{
		name:      "csrf-token-name",
		shorthand: "N",
		value:     defaultCSRFTokenName,
		usage:     "CSRF token name",
	}
	csrfTokenSameSite = Setting{
		name:      "csrf-token-same-site",
		shorthand: "S",
		value:     defaultCSRFTokenSameSite,
		usage:     "CSRF token SameSite attribute",
	}
	csrfTokenMaxAge = Setting{
		name:      "csrf-token-max-age",
		shorthand: "M",
		value:     defaultCSRFTokenMaxAge,
		usage:     "CSRF token max age (expiration)",
	}
	csrfTokenCookieSecure = Setting{
		name:      "csrf-token-cookie-secure",
		shorthand: "R",
		value:     false, // TODO Change to true after implementing HTTPS
		usage:     "CSRF token cookie Secure attribute",
	}
	csrfTokenCookieHTTPOnly = Setting{
		name:      "csrf-token-cookie-http-only",
		shorthand: "J",
		value:     false, // False needed for Swagger UI submit CSRF workaround
		usage:     "CSRF token cookie HttpOnly attribute",
	}
	csrfTokenCookieSessionOnly = Setting{
		name:      "csrf-token-cookie-session-only",
		shorthand: "E",
		value:     true,
		usage:     "CSRF token cookie SessionOnly attribute",
	}
	ipRateLimit = Setting{
		name:      "rate-limit",
		shorthand: "r",
		value:     uint16(50),
		usage:     "rate limit requests per second",
	}
	allowedIps = Setting{
		name:      "allowed-ips",
		shorthand: "I",
		value:     defaultAllowedIps,
		usage:     "comma-separated list of allowed IPs",
	}
	allowedCidrs = Setting{
		name:      "allowed-cidrs",
		shorthand: "C",
		value:     defaultAllowedCIDRs,
		usage:     "comma-separated list of allowed CIDRs",
	}
	databaseContainer = Setting{
		name:      "database-container",
		shorthand: "D",
		value:     "disabled",
		usage:     "database container mode; true to use container, false to use local database",
	}
	databaseURL = Setting{
		name:      "database-url",
		shorthand: "u",
		value:     "postgres://USR:PWD@localhost:5432/DB?sslmode=disable",
		usage:     "database URL; start a container with:\ndocker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=USR -e POSTGRES_PASSWORD=PWD -e POSTGRES_DB=DB postgres:latest\n",
	}
	databaseInitTotalTimeout = Setting{
		name:      "database-init-total-timeout",
		shorthand: "Z",
		value:     5 * time.Minute,
		usage:     "database init total timeout",
	}
	databaseInitRetryWait = Setting{
		name:      "database-init-retry-wait",
		shorthand: "W",
		value:     1 * time.Second,
		usage:     "database init retry wait",
	}
	otlp = Setting{
		name:      "otlp",
		shorthand: "z",
		value:     false,
		usage:     "enable OTLP export",
	}
	otlpConsole = Setting{
		name:      "otlp-console",
		shorthand: "q",
		value:     false,
		usage:     "enable OTLP logging to console (STDOUT)",
	}
	otlpScope = Setting{
		name:      "otlp-scope",
		shorthand: "s",
		value:     "cryptoutil",
		usage:     "OTLP scope",
	}
	unsealMode = Setting{
		name:      "unseal-mode",
		shorthand: "U",
		value:     "sysinfo",
		usage:     "unseal mode: N, M-of-N, sysinfo; N keys, or M-of-N derived keys from shared secrets, or X-of-Y custom sysinfo as shared secrets",
	}
	unsealFiles = Setting{
		name:      "unseal-files",
		shorthand: "F",
		value:     []string{},
		usage: "unseal files; repeat for multiple files; e.g. " +
			"\"--unseal-files=/docker/secrets/unseal_1of3 --unseal-files=/docker/secrets/unseal_2of3\"; " +
			"used for N unseal keys or M-of-N unseal shared secrets",
	}
)

var defaultCORSAllowedOrigins = func() string {
	defaultBindPostString := strconv.Itoa(int(bindPublicPort.value.(uint16)))
	return strings.Join([]string{
		httpProtocol + "://" + localhost + ":" + defaultBindPostString,
		httpProtocol + "://" + ipv4Loopback + ":" + defaultBindPostString,
		httpProtocol + "://" + ipv6Loopback + ":" + defaultBindPostString,
		httpsProtocol + "://" + localhost + ":" + defaultBindPostString,
		httpsProtocol + "://" + ipv4Loopback + ":" + defaultBindPostString,
		httpsProtocol + "://" + ipv6Loopback + ":" + defaultBindPostString,
	}, ",")
}()

var defaultCORSAllowedMethods = func() string {
	return strings.Join([]string{
		"POST",
		"GET",
		"PUT",
		"DELETE",
		"OPTIONS",
	}, ",")
}()

var defaultCORSAllowedHeaders = func() string {
	defaultHeaders := []string{
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
	return strings.Join(defaultHeaders, ",")
}()

var defaultCORSMaxAge = uint16(3600)

var defaultCSRFTokenName = "_csrf"

var defaultCSRFTokenSameSite = "Strict"

var defaultCSRFTokenMaxAge = 1 * time.Hour

var defaultAllowedIps = "" // TODO []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"}

var defaultAllowedCIDRs = func() string {
	return strings.Join([]string{
		localhostCIDRv4,     // localhost (IPv4)
		linkLocalCIDRv4,     // link-local (IPv4)
		privateClassACIDRv4, // private LAN class A (IPv4)
		privateClassBCIDRv4, // private LAN class B (IPv4)
		privateClassCCIDRv4, // private LAN class C (IPv4)
		localhostCIDRv6,     // localhost (IPv6)
		linkLocalCIDRv6,     // link-local (IPv6)
		privateLANv6,        // private LAN (IPv6)
	}, ",")
}()

// set of valid subcommands
var subcommands = map[string]struct{}{
	"start": {},
	"stop":  {},
	"init":  {},
}

// TODO Server only parameters?
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
	pflag.BoolP(help.name, help.shorthand, help.value.(bool), help.usage)
	pflag.StringP(configFile.name, configFile.shorthand, configFile.value.(string), configFile.usage)
	pflag.StringP(logLevel.name, logLevel.shorthand, logLevel.value.(string), logLevel.usage)
	pflag.BoolP(verboseMode.name, verboseMode.shorthand, verboseMode.value.(bool), verboseMode.usage)
	pflag.BoolP(devMode.name, devMode.shorthand, devMode.value.(bool), devMode.usage)
	pflag.StringP(bindPublicProtocol.name, bindPublicProtocol.shorthand, bindPublicProtocol.value.(string), bindPublicProtocol.usage)
	pflag.StringP(bindPublicAddress.name, bindPublicAddress.shorthand, bindPublicAddress.value.(string), bindPublicAddress.usage)
	pflag.Uint16P(bindPublicPort.name, bindPublicPort.shorthand, bindPublicPort.value.(uint16), bindPublicPort.usage)
	pflag.StringSliceP(tlsPublicDnsNames.name, tlsPublicDnsNames.shorthand, tlsPublicDnsNames.value.([]string), tlsPublicDnsNames.usage)
	pflag.StringSliceP(tlsPublicIPAddresses.name, tlsPublicIPAddresses.shorthand, tlsPublicIPAddresses.value.([]string), tlsPublicIPAddresses.usage)
	pflag.StringSliceP(tlsPrivateDnsNames.name, tlsPrivateDnsNames.shorthand, tlsPrivateDnsNames.value.([]string), tlsPrivateDnsNames.usage)
	pflag.StringSliceP(tlsPrivateIPAddresses.name, tlsPrivateIPAddresses.shorthand, tlsPrivateIPAddresses.value.([]string), tlsPrivateIPAddresses.usage)
	pflag.StringP(bindPrivateProtocol.name, bindPrivateProtocol.shorthand, bindPrivateProtocol.value.(string), bindPrivateProtocol.usage)
	pflag.StringP(bindPrivateAddress.name, bindPrivateAddress.shorthand, bindPrivateAddress.value.(string), bindPrivateAddress.usage)
	pflag.Uint16P(bindPrivatePort.name, bindPrivatePort.shorthand, bindPrivatePort.value.(uint16), bindPrivatePort.usage)
	pflag.StringP(publicBrowserAPIContextPath.name, publicBrowserAPIContextPath.shorthand, publicBrowserAPIContextPath.value.(string), publicBrowserAPIContextPath.usage)
	pflag.StringP(publicServiceAPIContextPath.name, publicServiceAPIContextPath.shorthand, publicServiceAPIContextPath.value.(string), publicServiceAPIContextPath.usage)
	pflag.StringP(corsAllowedOrigins.name, corsAllowedOrigins.shorthand, corsAllowedOrigins.value.(string), corsAllowedOrigins.usage)
	pflag.StringP(corsAllowedMethods.name, corsAllowedMethods.shorthand, corsAllowedMethods.value.(string), corsAllowedMethods.usage)
	pflag.StringP(corsAllowedHeaders.name, corsAllowedHeaders.shorthand, corsAllowedHeaders.value.(string), corsAllowedHeaders.usage)
	pflag.Uint16P(corsMaxAge.name, corsMaxAge.shorthand, corsMaxAge.value.(uint16), corsMaxAge.usage)
	pflag.StringP(csrfTokenName.name, csrfTokenName.shorthand, csrfTokenName.value.(string), csrfTokenName.usage)
	pflag.StringP(csrfTokenSameSite.name, csrfTokenSameSite.shorthand, csrfTokenSameSite.value.(string), csrfTokenSameSite.usage)
	pflag.DurationP(csrfTokenMaxAge.name, csrfTokenMaxAge.shorthand, csrfTokenMaxAge.value.(time.Duration), csrfTokenMaxAge.usage)
	pflag.BoolP(csrfTokenCookieSecure.name, csrfTokenCookieSecure.shorthand, csrfTokenCookieSecure.value.(bool), csrfTokenCookieSecure.usage)
	pflag.BoolP(csrfTokenCookieHTTPOnly.name, csrfTokenCookieHTTPOnly.shorthand, csrfTokenCookieHTTPOnly.value.(bool), csrfTokenCookieHTTPOnly.usage)
	pflag.BoolP(csrfTokenCookieSessionOnly.name, csrfTokenCookieSessionOnly.shorthand, csrfTokenCookieSessionOnly.value.(bool), csrfTokenCookieSessionOnly.usage)
	pflag.Uint16P(ipRateLimit.name, ipRateLimit.shorthand, ipRateLimit.value.(uint16), ipRateLimit.usage)
	pflag.StringP(allowedIps.name, allowedIps.shorthand, allowedIps.value.(string), allowedIps.usage)
	pflag.StringP(allowedCidrs.name, allowedCidrs.shorthand, allowedCidrs.value.(string), allowedCidrs.usage)
	pflag.StringP(databaseContainer.name, databaseContainer.shorthand, databaseContainer.value.(string), databaseContainer.usage)
	pflag.StringP(databaseURL.name, databaseURL.shorthand, databaseURL.value.(string), databaseURL.usage)
	pflag.DurationP(databaseInitTotalTimeout.name, databaseInitTotalTimeout.shorthand, databaseInitTotalTimeout.value.(time.Duration), databaseInitTotalTimeout.usage)
	pflag.DurationP(databaseInitRetryWait.name, databaseInitRetryWait.shorthand, databaseInitRetryWait.value.(time.Duration), databaseInitRetryWait.usage)
	pflag.BoolP(otlp.name, otlp.shorthand, otlp.value.(bool), otlp.usage)
	pflag.BoolP(otlpConsole.name, otlpConsole.shorthand, otlpConsole.value.(bool), otlpConsole.usage)
	pflag.StringP(otlpScope.name, otlpScope.shorthand, otlpScope.value.(string), otlpScope.usage)
	pflag.StringP(unsealMode.name, unsealMode.shorthand, unsealMode.value.(string), unsealMode.usage)
	pflag.StringArrayP(unsealFiles.name, unsealFiles.shorthand, unsealFiles.value.([]string), unsealFiles.usage)
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
		TLSPublicDNSNames:           viper.GetStringSlice(tlsPublicDnsNames.name),
		TLSPublicIPAddresses:        viper.GetStringSlice(tlsPublicIPAddresses.name),
		TLSPrivateDNSNames:          viper.GetStringSlice(tlsPrivateDnsNames.name),
		TLSPrivateIPAddresses:       viper.GetStringSlice(tlsPrivateIPAddresses.name),
		BindPrivateProtocol:         viper.GetString(bindPrivateProtocol.name),
		BindPrivateAddress:          viper.GetString(bindPrivateAddress.name),
		BindPrivatePort:             viper.GetUint16(bindPrivatePort.name),
		PublicBrowserAPIContextPath: viper.GetString(publicBrowserAPIContextPath.name),
		PublicServiceAPIContextPath: viper.GetString(publicServiceAPIContextPath.name),
		CORSAllowedOrigins:          viper.GetString(corsAllowedOrigins.name),
		CORSAllowedMethods:          viper.GetString(corsAllowedMethods.name),
		CORSAllowedHeaders:          viper.GetString(corsAllowedHeaders.name),
		CORSMaxAge:                  viper.GetUint16(corsMaxAge.name),
		CSRFTokenName:               viper.GetString(csrfTokenName.name),
		CSRFTokenSameSite:           viper.GetString(csrfTokenSameSite.name),
		CSRFTokenMaxAge:             viper.GetDuration(csrfTokenMaxAge.name),
		CSRFTokenCookieSecure:       viper.GetBool(csrfTokenCookieSecure.name),
		CSRFTokenCookieHTTPOnly:     viper.GetBool(csrfTokenCookieHTTPOnly.name),
		CSRFTokenCookieSessionOnly:  viper.GetBool(csrfTokenCookieSessionOnly.name),
		IPRateLimit:                 viper.GetUint16(ipRateLimit.name),
		AllowedIPs:                  viper.GetString(allowedIps.name),
		AllowedCIDRs:                viper.GetString(allowedCidrs.name),
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
		log.Info("SubCommand: ", s.SubCommand)
		log.Info("Help: ", s.Help)
		log.Info("Config file: ", s.ConfigFile)
		log.Info("Log Level: ", s.LogLevel)
		log.Info("Verbose mode: ", s.VerboseMode)
		log.Info("Dev mode: ", s.DevMode)
		log.Info("Bind Public Protocol: ", s.BindPublicProtocol)
		log.Info("Bind Public Address: ", s.BindPublicAddress)
		log.Info("Bind Public Port: ", s.BindPublicPort)
		log.Info("TLS Public DNS Names: ", s.TLSPublicDNSNames)
		log.Info("TLS Public IP Addresses: ", s.TLSPublicIPAddresses)
		log.Info("TLS Private DNS Names: ", s.TLSPrivateDNSNames)
		log.Info("TLS Private IP Addresses: ", s.TLSPrivateIPAddresses)
		log.Info("Bind Private Protocol: ", s.BindPrivateProtocol)
		log.Info("Bind Private Address: ", s.BindPrivateAddress)
		log.Info("Bind Private Port: ", s.BindPrivatePort)
		log.Info("TLS Private DNS Names: ", s.TLSPrivateDNSNames)
		log.Info("Public Browser API Context Path: ", s.PublicBrowserAPIContextPath)
		log.Info("Public Service API Context Path: ", s.PublicServiceAPIContextPath)
		log.Info("CORS Allowed Origins: ", s.CORSAllowedOrigins)
		log.Info("CORS Allowed Methods: ", s.CORSAllowedMethods)
		log.Info("CORS Allowed Headers: ", s.CORSAllowedHeaders)
		log.Info("CORS Max Age: ", s.CORSMaxAge)
		log.Info("CSRF Token Name: ", s.CSRFTokenName)
		log.Info("CSRF Token SameSite: ", s.CSRFTokenSameSite)
		log.Info("CSRF Token Max Age: ", s.CSRFTokenMaxAge)
		log.Info("CSRF Token Cookie Secure: ", s.CSRFTokenCookieSecure)
		log.Info("CSRF Token Cookie HTTPOnly: ", s.CSRFTokenCookieHTTPOnly)
		log.Info("CSRF Token Cookie SessionOnly: ", s.CSRFTokenCookieSessionOnly)
		log.Info("IP Rate Limit: ", s.IPRateLimit)
		log.Info("Allowed IPs: ", s.AllowedIPs)
		log.Info("Allowed CIDRs: ", s.AllowedCIDRs)
		log.Info("Database Container: ", s.DatabaseContainer)
		// only give option to log in dev mode (i.e. don't give option to log in production mode)
		if s.DevMode {
			log.Info("Database URL: ", s.DatabaseURL) // sensitive value (i.e. PostgreSQL URLs may contain password)
		}
		log.Info("Database Init Total Timeout: ", s.DatabaseInitTotalTimeout)
		log.Info("Database Init Retry Wait: ", s.DatabaseInitRetryWait)
		log.Info("OTLP Export: ", s.OTLP)
		log.Info("OTLP Console: ", s.OTLPConsole)
		log.Info("OTLP Scope: ", s.OTLPScope)
		log.Info("Unseal Mode: ", s.UnsealMode)
		log.Info("Unseal Files: ", s.UnsealFiles)
	}
}

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	viper.Reset()
}
