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

type Settings struct {
	VerboseMode              bool
	LogLevel                 string
	DevMode                  bool
	OTLP                     bool
	OTLPConsole              bool
	OTLPScope                string
	ConfigFile               string
	BindAddress              string
	BindPort                 uint16
	ContextPath              string
	CorsOrigins              string
	CorsMethods              string
	CorsHeaders              string
	CorsMaxAge               uint16
	RateLimit                uint16
	AllowedIPs               string
	AllowedCIDRs             string
	DatabaseURL              string
	DatabaseInitTotalTimeout time.Duration
	DatabaseInitRetryWait    time.Duration
	Migrations               bool
}

// Setting Input values for pflag.*P(name, shortname, value, usage)
type Setting struct {
	name      string // unique long name for the flag
	shorthand string // unique short name for the flag
	value     any    // default value for the flag
	usage     string // description of the flag
}

var (
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
		usage:     "log level: TRACE, DEBUG, INFO, WARN, ERROR, FATAL",
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
		usage:     "run in development mode; enables in-memory SQLite and migrations",
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
		value:     false,
		usage:     "OTLP scope",
	}
	bindAddress = Setting{
		name:      "bind-address",
		shorthand: "a",
		value:     "localhost",
		usage:     "default bind address",
	}
	bindPort = Setting{
		name:      "bind-port",
		shorthand: "p",
		value:     uint16(8080),
		usage:     "default bind port",
	}
	contextPath = Setting{
		name:      "context-path",
		shorthand: "c",
		value:     "/api/v1",
		usage:     "context path for API",
	}
	corsOrigins = Setting{
		name:      "cors-origins",
		shorthand: "o",
		value:     defaultAllowedCORSOrigins,
		usage:     "CORS allowed origins",
	}
	corsMethods = Setting{
		name:      "cors-methods",
		shorthand: "m",
		value:     defaultAllowedCORSMethods,
		usage:     "CORS allowed methods",
	}
	corsHeaders = Setting{
		name:      "cors-headers",
		shorthand: "h",
		value:     defaultAllowedCORSHeaders,
		usage:     "CORS allowed headers",
	}
	corsMaxAge = Setting{
		name:      "cors-max-age",
		shorthand: "x",
		value:     uint16(3600),
		usage:     "CORS max age in seconds",
	}
	rateLimit = Setting{
		name:      "rate-limit",
		shorthand: "r",
		value:     uint16(50),
		usage:     "rate limit requests per second",
	}
	allowedIps = Setting{
		name:      "allowed-ips",
		shorthand: "I",
		value:     "",
		usage:     "comma-separated list of allowed IPs",
	}
	allowedCidrs = Setting{
		name:      "allowed-cidrs",
		shorthand: "C",
		value:     defaultAllowedCIDRs,
		usage:     "comma-separated list of allowed CIDRs",
	}
	databaseURL = Setting{
		name:      "database-url",
		shorthand: "u",
		value:     "postgres://postgres:PASSWORD@localhost:5432/readcommend?sslmode=disable", // default is for usage, omit PASSWORD
		usage:     "database URL",
	}
	databaseInitTotalTimeout = Setting{
		name:      "database-init-total-timeout",
		shorthand: "T",
		value:     5 * time.Minute,
		usage:     "database init total timeout",
	}
	databaseInitRetryWait = Setting{
		name:      "database-init-retry-wait",
		shorthand: "W",
		value:     1 * time.Second,
		usage:     "database init retry wait",
	}
	migrations = Setting{
		name:      "migrations",
		shorthand: "M",
		value:     false,
		usage:     "run DB migrations",
	}
)

var defaultAllowedCORSOrigins = func() string {
	defaultBindPostString := strconv.Itoa(int(bindPort.value.(uint16)))
	return strings.Join([]string{
		"http://localhost:" + defaultBindPostString,
		"http://127.0.0.1:" + defaultBindPostString,
		"http://[::1]:" + defaultBindPostString,
		"https://localhost:" + defaultBindPostString,
		"https://127.0.0.1:" + defaultBindPostString,
		"https://[::1]:" + defaultBindPostString,
	}, ",")
}()

var defaultAllowedCORSMethods = func() string {
	return strings.Join([]string{
		"POST",
		"GET",
		"PUT",
		"DELETE",
		"OPTIONS",
	}, ",")
}()

var defaultAllowedCORSHeaders = func() string {
	defaultHeaders := []string{
		"Content-Type",
		"Authorization",
		"Accept",
		"Origin",
		"X-Requested-With",
		"Cache-Control",
		"Pragma",
		"Expires",
	}
	return strings.Join(defaultHeaders, ",")
}()

var defaultAllowedCIDRs = func() string {
	return strings.Join([]string{
		"127.0.0.0/8",    // localhost (IPv4)
		"169.254.0.0/16", // link-local (IPv4)
		"10.0.0.0/8",     // private LAN class A (IPv4)
		"172.16.0.0/12",  // private LAN class B (IPv4)
		"192.168.0.0/16", // private LAN class C (IPv4)
		"::1/128",        // localhost (IPv6)
		"fe80::/10",      // link-local (IPv6)
		"fc00::/7",       // private LAN (IPv6)
	}, ",")
}()

func Parse() (*Settings, error) {
	pflag.StringP(configFile.name, configFile.shorthand, configFile.value.(string), configFile.usage)
	pflag.StringP(logLevel.name, logLevel.shorthand, logLevel.value.(string), logLevel.usage)
	pflag.BoolP(verboseMode.name, verboseMode.shorthand, verboseMode.value.(bool), verboseMode.usage)
	pflag.BoolP(devMode.name, devMode.shorthand, devMode.value.(bool), devMode.usage)
	pflag.BoolP(otlp.name, otlp.shorthand, otlp.value.(bool), otlp.usage)
	pflag.BoolP(otlpConsole.name, otlpConsole.shorthand, otlpConsole.value.(bool), otlpConsole.usage)
	pflag.BoolP(otlpScope.name, otlpScope.shorthand, otlpScope.value.(bool), otlpScope.usage)
	pflag.StringP(bindAddress.name, bindAddress.shorthand, bindAddress.value.(string), bindAddress.usage)
	pflag.Uint16P(bindPort.name, bindPort.shorthand, bindPort.value.(uint16), bindPort.usage)
	pflag.StringP(contextPath.name, contextPath.shorthand, contextPath.value.(string), contextPath.usage)
	pflag.StringP(corsOrigins.name, corsOrigins.shorthand, corsOrigins.value.(string), corsOrigins.usage)
	pflag.StringP(corsMethods.name, corsMethods.shorthand, corsMethods.value.(string), corsMethods.usage)
	pflag.StringP(corsHeaders.name, corsHeaders.shorthand, corsHeaders.value.(string), corsHeaders.usage)
	pflag.Uint16P(corsMaxAge.name, corsMaxAge.shorthand, corsMaxAge.value.(uint16), corsMaxAge.usage)
	pflag.Uint16P(rateLimit.name, rateLimit.shorthand, rateLimit.value.(uint16), rateLimit.usage)
	pflag.StringP(allowedIps.name, allowedIps.shorthand, allowedIps.value.(string), allowedIps.usage)
	pflag.StringP(allowedCidrs.name, allowedCidrs.shorthand, allowedCidrs.value.(string), allowedCidrs.usage)
	pflag.StringP(databaseURL.name, databaseURL.shorthand, databaseURL.value.(string), databaseURL.usage)
	pflag.DurationP(databaseInitTotalTimeout.name, databaseInitTotalTimeout.shorthand, databaseInitTotalTimeout.value.(time.Duration), databaseInitTotalTimeout.usage)
	pflag.DurationP(databaseInitRetryWait.name, databaseInitRetryWait.shorthand, databaseInitRetryWait.value.(time.Duration), databaseInitRetryWait.usage)
	pflag.BoolP(migrations.name, migrations.shorthand, migrations.value.(bool), migrations.usage)
	err := pflag.CommandLine.Parse(os.Args[1:])
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
		ConfigFile:               viper.GetString(configFile.name),
		LogLevel:                 viper.GetString(logLevel.name),
		VerboseMode:              viper.GetBool(verboseMode.name),
		DevMode:                  viper.GetBool(devMode.name),
		OTLP:                     viper.GetBool(otlp.name),
		OTLPConsole:              viper.GetBool(otlpConsole.name),
		OTLPScope:                viper.GetString(otlpScope.name),
		BindAddress:              viper.GetString(bindAddress.name),
		BindPort:                 viper.GetUint16(bindPort.name),
		ContextPath:              viper.GetString(contextPath.name),
		CorsOrigins:              viper.GetString(corsOrigins.name),
		CorsMethods:              viper.GetString(corsMethods.name),
		CorsHeaders:              viper.GetString(corsHeaders.name),
		CorsMaxAge:               viper.GetUint16(corsMaxAge.name),
		RateLimit:                viper.GetUint16(rateLimit.name),
		AllowedIPs:               viper.GetString(allowedIps.name),
		AllowedCIDRs:             viper.GetString(allowedCidrs.name),
		DatabaseURL:              viper.GetString(databaseURL.name),
		Migrations:               viper.GetBool(migrations.name),
		DatabaseInitTotalTimeout: viper.GetDuration(databaseInitTotalTimeout.name),
		DatabaseInitRetryWait:    viper.GetDuration(databaseInitRetryWait.name),
	}
	logSettings(s)
	return s, nil
}

func logSettings(s *Settings) {
	if s.VerboseMode {
		log.Info("Config file: ", s.ConfigFile)
		log.Info("Log Level: ", s.LogLevel)
		log.Info("Verbose mode: ", s.VerboseMode)
		log.Info("Dev mode: ", s.DevMode)
		log.Info("OTLP Export: ", s.OTLP)
		log.Info("OTLP Console: ", s.OTLPConsole)
		log.Info("OTLP Scope: ", s.OTLPScope)
		log.Info("Bind Address: ", s.BindAddress)
		log.Info("Bind Port: ", s.BindPort)
		log.Info("Context Path: ", s.ContextPath)
		log.Info("CORS Origins: ", s.CorsOrigins)
		log.Info("CORS Methods: ", s.CorsMethods)
		log.Info("CORS Headers: ", s.CorsHeaders)
		log.Info("CORS Max Age: ", s.CorsMaxAge)
		log.Info("Rate Limit: ", s.RateLimit)
		log.Info("Allowed IPs: ", s.AllowedIPs)
		log.Info("Allowed CIDRs: ", s.AllowedCIDRs)
		// only give option to log in dev mode (i.e. don't give option to log in production mode)
		if s.DevMode {
			log.Info("Database URL: ", s.DatabaseURL) // sensitive value (i.e. PostgreSQL URLs may contain password)
		}
		log.Info("Database Init Total Timeout: ", s.DatabaseInitTotalTimeout)
		log.Info("Database Init Retry Wait: ", s.DatabaseInitRetryWait)
		log.Info("Migrations: ", s.Migrations)
	}
}

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	viper.Reset()
}
