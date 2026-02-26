// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilKmsServerHandler "cryptoutil/internal/apps/sm/kms/server/handler"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"

	"github.com/getkin/kin-openapi/openapi3"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

const (
	fiberAppIDPublic  fiberAppID = "public"
	fiberAppIDPrivate fiberAppID = "private"
)

type fiberAppID string

var ready atomic.Bool

// ServerApplicationListener provides HTTP listener configuration and lifecycle management for the server.
type ServerApplicationListener struct {
	StartFunction     func()
	ShutdownFunction  func()
	PublicTLSServer   *TLSServerConfig
	PrivateTLSServer  *TLSServerConfig
	ActualPublicPort  uint16
	ActualPrivatePort uint16
}

// TLSServerConfig holds TLS configuration including certificates and certificate pools.
type TLSServerConfig struct {
	Certificate         *tls.Certificate
	RootCAsPool         *x509.CertPool
	IntermediateCAsPool *x509.CertPool
	Config              *tls.Config
}

// SendServerListenerLivenessCheck sends a liveness probe to the server's private admin endpoint.
func SendServerListenerLivenessCheck(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.ClientLivenessRequestTimeout)
	defer cancel()

	_, _, result, err := cryptoutilSharedUtilNetwork.HTTPGetLivez(ctx, settings.PrivateBaseURL(), settings.PrivateAdminAPIContextPath, 0, nil, settings.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get liveness check: %w", err)
	}

	return result, nil
}

// SendServerListenerReadinessCheck sends a readiness probe to the server's private admin endpoint.
func SendServerListenerReadinessCheck(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.ClientReadinessRequestTimeout)
	defer cancel()

	_, _, result, err := cryptoutilSharedUtilNetwork.HTTPGetReadyz(ctx, settings.PrivateBaseURL(), settings.PrivateAdminAPIContextPath, 0, nil, settings.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get readiness check: %w", err)
	}

	return result, nil
}

// SendServerListenerShutdownRequest sends a shutdown request to the server's private admin endpoint.
func SendServerListenerShutdownRequest(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.ClientShutdownRequestTimeout)
	defer cancel()

	_, _, _, err := cryptoutilSharedUtilNetwork.HTTPPostShutdown(ctx, settings.PrivateBaseURL(), settings.PrivateAdminAPIContextPath, 0, nil, settings.DevMode)
	if err != nil {
		return fmt.Errorf("failed to send shutdown request: %w", err)
	}

	return nil
}

// StartServerListenerApplication creates and starts a new server application listener.
func StartServerListenerApplication(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*ServerApplicationListener, error) {
	ctx := context.Background()

	serverApplicationCore, err := StartServerApplicationCore(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server application: %w", err)
	} else if settings.BindPublicProtocol != cryptoutilSharedMagic.ProtocolHTTP && settings.BindPublicProtocol != cryptoutilSharedMagic.ProtocolHTTPS {
		return nil, fmt.Errorf("invalid public protocol: expected 'http' or 'https', got '%s'", settings.BindPublicProtocol)
	} else if settings.BindPrivateProtocol != cryptoutilSharedMagic.ProtocolHTTP && settings.BindPrivateProtocol != cryptoutilSharedMagic.ProtocolHTTPS {
		return nil, fmt.Errorf("invalid private protocol: expected 'http' or 'https', got '%s'", settings.BindPrivateProtocol)
	}

	publicTLSServerSubject, privateTLSServerSubject, err := generateTLSServerSubjects(settings, serverApplicationCore.ServerApplicationBasic)
	if err != nil {
		return nil, fmt.Errorf("failed to run new function: %w", err)
	}

	// Public server: TLS 1.3 only, no client certificate required (browser access).
	publicTLSConfig, err := cryptoutilSharedCryptoTls.NewServerConfig(&cryptoutilSharedCryptoTls.ServerConfigOptions{
		Subject:    publicTLSServerSubject,
		ClientAuth: tls.NoClientCert,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build public TLS server config: %w", err)
	}

	publicTLSServer := &TLSServerConfig{
		Certificate:         publicTLSConfig.Certificate,
		RootCAsPool:         publicTLSConfig.RootCAsPool,
		IntermediateCAsPool: publicTLSConfig.IntermediateCAsPool,
		Config:              publicTLSConfig.TLSConfig,
	}

	// Private server: TLS 1.3 only with optional mTLS for internal service-to-service communication.
	// In production, uses RequireAndVerifyClientCert to enforce mutual authentication on admin/internal APIs.
	// In dev mode or container mode (0.0.0.0 binding), uses NoClientCert for easier testing/healthchecks.
	// Container deployments use 0.0.0.0 for network accessibility and wget healthchecks without client certs.
	privateClientAuth := tls.RequireAndVerifyClientCert

	isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
	if settings.DevMode || isContainerMode {
		privateClientAuth = tls.NoClientCert
	}

	privateTLSConfig, err := cryptoutilSharedCryptoTls.NewServerConfig(&cryptoutilSharedCryptoTls.ServerConfigOptions{
		Subject:    privateTLSServerSubject,
		ClientAuth: privateClientAuth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build private TLS server config: %w", err)
	}

	privateTLSServer := &TLSServerConfig{
		Certificate:         privateTLSConfig.Certificate,
		RootCAsPool:         privateTLSConfig.RootCAsPool,
		IntermediateCAsPool: privateTLSConfig.IntermediateCAsPool,
		Config:              privateTLSConfig.TLSConfig,
	}

	// Common base middlewares shared by both private and public apps
	commonBaseMiddlewares := []fiber.Handler{
		recover.New(recover.Config{
			EnableStackTrace: true,
		}),
		requestid.New(),
		commonIPFilterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonOtelFiberTelemetryMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonOtelFiberRequestLoggerMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService),
		commonHTTPGETCacheControlMiddleware(),
		commonUnsupportedHTTPMethodsMiddleware(settings),
	} // Fiber app for Service-to-Service API calls

	privateMiddlewares := append([]fiber.Handler{commonSetFiberRequestAttribute(fiberAppIDPrivate)}, commonBaseMiddlewares...)
	privateMiddlewares = append(privateMiddlewares, commonIPRateLimiterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, int(settings.ServiceIPRateLimit)))
	privateMiddlewares = append(privateMiddlewares, compress.New()) // Enable response compression after rate limiting to avoid compressing blocked responses
	privateMiddlewares = append(privateMiddlewares, commonHTTPGETCacheControlMiddleware())
	privateMiddlewares = append(privateMiddlewares, commonUnsupportedHTTPMethodsMiddleware(settings))
	privateMiddlewares = append(privateMiddlewares, privateHealthCheckMiddlewareFunction(serverApplicationCore))

	privateFiberApp := fiber.New(fiber.Config{Immutable: true, BodyLimit: settings.RequestBodyLimit})
	for _, middleware := range privateMiddlewares {
		privateFiberApp.Use(middleware)
	}

	// Fiber app for Browser-to-Service API calls and Swagger UI

	publicMiddlewares := append([]fiber.Handler{commonSetFiberRequestAttribute(fiberAppIDPublic)}, commonBaseMiddlewares...)
	publicMiddlewares = append(publicMiddlewares, commonIPRateLimiterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, int(settings.BrowserIPRateLimit)))
	publicMiddlewares = append(publicMiddlewares, compress.New()) // Enable response compression after rate limiting to avoid compressing blocked responses
	publicMiddlewares = append(publicMiddlewares, commonHTTPGETCacheControlMiddleware())
	publicMiddlewares = append(publicMiddlewares, commonUnsupportedHTTPMethodsMiddleware(settings))
	publicMiddlewares = append(publicMiddlewares, publicBrowserCORSMiddlewareFunction(settings))
	publicMiddlewares = append(publicMiddlewares, publicBrowserXSSMiddlewareFunction(settings))
	publicMiddlewares = append(publicMiddlewares, publicBrowserAdditionalSecurityHeadersMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings))
	publicMiddlewares = append(publicMiddlewares, publicBrowserCSRFMiddlewareFunction(settings))

	publicFiberApp := fiber.New(fiber.Config{Immutable: true, BodyLimit: settings.RequestBodyLimit})
	for _, middleware := range publicMiddlewares {
		publicFiberApp.Use(middleware)
	}

	// shutdownServerFunction stops privateFiberApp and publicFiberApp, it is called via /shutdown hosted by privateFiberApp
	var shutdownServerFunction func()

	// Private APIs - add admin context path prefix
	privateFiberApp.Post(settings.PrivateAdminAPIContextPath+cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, func(c *fiber.Ctx) error {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Info("shutdown requested via API endpoint")

		if shutdownServerFunction != nil {
			defer func() {
				time.Sleep(cryptoutilSharedMagic.WaitBeforeShutdownDuration) // allow server small amount of time to finish sending response to client
				shutdownServerFunction()
			}()
		}

		return c.SendString("Server shutdown initiated")
	})

	// Public Swagger UI with basic authentication
	swaggerAPI, err := cryptoutilKmsServer.GetSwagger()
	if err != nil {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to get swagger", cryptoutilSharedMagic.StringError, err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	swaggerAPI.Servers = []*openapi3.Server{
		{URL: settings.PublicBrowserAPIContextPath}, // Browser users will access the APIs via this context path, with browser middlewares (CORS, CSRF, etc)
		{URL: settings.PublicServiceAPIContextPath}, // Service clients will access the APIs via this context path, without browser middlewares
	}

	swaggerSpecBytes, err := swaggerAPI.MarshalJSON() // Serialize OpenAPI 3 spec to JSON with the added public server context path
	if err != nil {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", cryptoutilSharedMagic.StringError, err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	publicFiberApp.Get("/ui/swagger/doc.json", swaggerUIBasicAuthMiddleware(settings.SwaggerUIUsername, settings.SwaggerUIPassword), func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		return c.Send(swaggerSpecBytes)
	})
	publicFiberApp.Get("/ui/swagger/*", swaggerUIBasicAuthMiddleware(settings.SwaggerUIUsername, settings.SwaggerUIPassword), func(c *fiber.Ctx) error {
		swaggerHandler := swagger.New(swagger.Config{
			Title:                  "Cryptoutil API",
			URL:                    "/ui/swagger/doc.json",
			TryItOutEnabled:        true,
			DisplayRequestDuration: true,
			ShowCommonExtensions:   true,
			CustomScript:           swaggerUICustomCSRFScript(settings.CSRFTokenName, settings.PublicBrowserAPIContextPath),
		})

		err := swaggerHandler(c)
		if err != nil {
			return err
		}
		// Ensure Content-Type includes charset for HTML responses to satisfy security scanners
		if c.Get("Content-Type") == "text/html" {
			c.Set("Content-Type", "text/html; charset=utf-8")
		}

		return nil
	})
	publicFiberApp.Get(settings.PublicBrowserAPIContextPath+"/csrf-token", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":         "CSRF token set in cookie",
			"csrf_token_name": settings.CSRFTokenName,
			"cookie_secure":   settings.CSRFTokenCookieSecure,
			"same_site":       settings.CSRFTokenSameSite,
		})
	})

	// Swagger APIs, will be double exposed on publicFiberApp, but with different security middlewares (i.e. browser user vs machine client)
	openapiStrictServer := cryptoutilKmsServerHandler.NewOpenapiStrictServer(serverApplicationCore.BusinessLogicService)
	openapiStrictHandler := cryptoutilKmsServer.NewStrictHandler(openapiStrictServer, nil)
	commonOapiMiddlewareFiberRequestValidators := []cryptoutilKmsServer.MiddlewareFunc{
		fibermiddleware.OapiRequestValidatorWithOptions(swaggerAPI, &fibermiddleware.Options{}),
	}
	publicBrowserFiberServerOptions := cryptoutilKmsServer.FiberServerOptions{
		BaseURL:     settings.PublicBrowserAPIContextPath,
		Middlewares: commonOapiMiddlewareFiberRequestValidators,
	}
	publicServiceFiberServerOptions := cryptoutilKmsServer.FiberServerOptions{
		BaseURL:     settings.PublicServiceAPIContextPath,
		Middlewares: commonOapiMiddlewareFiberRequestValidators,
	}

	cryptoutilKmsServer.RegisterHandlersWithOptions(publicFiberApp, openapiStrictHandler, publicBrowserFiberServerOptions)
	cryptoutilKmsServer.RegisterHandlersWithOptions(publicFiberApp, openapiStrictHandler, publicServiceFiberServerOptions)

	// Create listeners - use port 0 for testing (to get OS-assigned ports), configured ports for production
	publicBinding := fmt.Sprintf("%s:%d", settings.BindPublicAddress, settings.BindPublicPort)
	privateBinding := fmt.Sprintf("%s:%d", settings.BindPrivateAddress, settings.BindPrivatePort)

	// Create net listeners to get actual assigned ports (port 0 for tests, configured ports for production)
	publicListener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", publicBinding)
	if err != nil {
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create public listener: %w", err)
	}

	privateListener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", privateBinding)
	if err != nil {
		if closeErr := publicListener.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close public listener during cleanup: %v\n", closeErr)
		}

		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create private listener: %w", err)
	}

	// Extract actual assigned ports
	publicAddr, ok := publicListener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("failed to get public listener address")
	}

	if publicAddr.Port < 0 || publicAddr.Port > int(cryptoutilSharedMagic.MaxPortNumber) {
		return nil, fmt.Errorf("invalid public port: %d", publicAddr.Port)
	}

	actualPublicPort := uint16(publicAddr.Port)

	privateAddr, ok := privateListener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("failed to get private listener address")
	}

	if privateAddr.Port < 0 || privateAddr.Port > int(cryptoutilSharedMagic.MaxPortNumber) {
		return nil, fmt.Errorf("invalid private port: %d", privateAddr.Port)
	}

	actualPrivatePort := uint16(privateAddr.Port)

	serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Info("assigned ports",
		cryptoutilSharedMagic.SubjectTypePublic, actualPublicPort, "private", actualPrivatePort)

	startServerFunction := startServerFuncWithListeners(
		publicListener, privateListener,
		publicFiberApp, privateFiberApp,
		settings.BindPublicProtocol, settings.BindPrivateProtocol,
		publicTLSServer.Config, privateTLSServer.Config,
		serverApplicationCore.ServerApplicationBasic.TelemetryService)
	shutdownServerFunction = stopServerFuncWithListeners(serverApplicationCore, publicFiberApp, privateFiberApp, publicListener, privateListener, settings)

	go stopServerSignalFunc(serverApplicationCore.ServerApplicationBasic.TelemetryService, shutdownServerFunction)() // listen for OS signals to gracefully shutdown the server

	return &ServerApplicationListener{
		StartFunction:     startServerFunction,
		ShutdownFunction:  shutdownServerFunction,
		PublicTLSServer:   publicTLSServer,
		PrivateTLSServer:  privateTLSServer,
		ActualPublicPort:  actualPublicPort,
		ActualPrivatePort: actualPrivatePort,
	}, nil
}

func startServerFuncWithListeners(publicListener, privateListener net.Listener, publicFiberApp, privateFiberApp *fiber.App, publicProtocol, privateProtocol string, publicTLSConfig, privateTLSConfig *tls.Config, telemetryService *cryptoutilSharedTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listeners with pre-created listeners")

		// Mark server as ready immediately since listeners are already bound and accepting connections.
		ready.Store(true)

		go func() {
			telemetryService.Slogger.Info("starting private fiber listener", "addr", privateListener.Addr().String(), "protocol", privateProtocol)

			var err error

			if privateProtocol == cryptoutilSharedMagic.ProtocolHTTPS && privateTLSConfig != nil {
				// Wrap the listener with TLS
				tlsListener := tls.NewListener(privateListener, privateTLSConfig)
				telemetryService.Slogger.Info("private server listening with TLS", "addr", privateListener.Addr().String())

				err = privateFiberApp.Listener(tlsListener)
			} else {
				telemetryService.Slogger.Info("private server listening without TLS", "addr", privateListener.Addr().String())

				err = privateFiberApp.Listener(privateListener)
			}

			if err != nil {
				telemetryService.Slogger.Error("failed to start private fiber listener", cryptoutilSharedMagic.StringError, err)
			}

			telemetryService.Slogger.Debug("private fiber listener stopped")
		}()

		telemetryService.Slogger.Info("starting public fiber listener", "addr", publicListener.Addr().String(), "protocol", publicProtocol)

		var err error

		if publicProtocol == cryptoutilSharedMagic.ProtocolHTTPS && publicTLSConfig != nil {
			// Wrap the listener with TLS
			tlsListener := tls.NewListener(publicListener, publicTLSConfig)
			telemetryService.Slogger.Info("public server listening with TLS", "addr", publicListener.Addr().String())

			err = publicFiberApp.Listener(tlsListener)
		} else {
			telemetryService.Slogger.Info("public server listening without TLS", "addr", publicListener.Addr().String())

			err = publicFiberApp.Listener(publicListener)
		}

		if err != nil {
			telemetryService.Slogger.Error("failed to start public fiber listener", cryptoutilSharedMagic.StringError, err)
		}

		telemetryService.Slogger.Debug("public fiber listener stopped")
	}
}
