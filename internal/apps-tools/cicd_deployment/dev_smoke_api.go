// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	json "encoding/json"
	"errors"
	"io"
	"net"
	http "net/http"
	"os/exec"
	"strconv"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func verifyDevelopmentPublicAPI(ctx context.Context, input gen.BootstrapPreflightInput, secrets []gen.LifecycleSecret, config execution.KubectlConfig) error {
	return verifyDevelopmentPublicAPIWithForward(ctx, input, secrets, config, startDevelopmentForward)
}

func startDevelopmentForward(ctx context.Context, binary string, args []string) (func(), error) {
	forwardCtx, cancel := context.WithCancel(ctx)
	command := exec.CommandContext(forwardCtx, binary, args...)

	command.Stdout, command.Stderr = io.Discard, io.Discard
	if err := command.Start(); err != nil {
		cancel()

		return nil, errors.New("owned API loopback forwarding unavailable")
	}

	return func() { cancel(); _ = command.Wait() }, nil
}

func verifyDevelopmentPublicAPIWithForward(ctx context.Context, input gen.BootstrapPreflightInput, secrets []gen.LifecycleSecret,
	config execution.KubectlConfig, start func(context.Context, string, []string) (func(), error),
) error {
	if len(input.RuntimeValues) != 1 || len(input.RuntimeValues[0].Workloads) != 1 || len(input.RuntimeValues[0].Workloads[0].InstanceServices) == 0 {
		return errors.New("owned API explicit workload identity unavailable")
	}

	release := input.RuntimeValues[0]
	workload := release.Workloads[0]
	instance := workload.InstanceServices[0]

	var trust []byte

	for _, secret := range secrets {
		if secret.Metadata.Namespace == release.Namespace && secret.Metadata.Name == workload.TrustSecretName {
			trust = secret.Data["ca.crt"]
		}
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(trust) {
		return errors.New("owned API explicit trust unavailable")
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return errors.New("owned API loopback port unavailable")
	}

	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		_ = listener.Close()

		return errors.New("owned loopback address rejected")
	}

	port := address.Port

	if err := listener.Close(); err != nil {
		return errors.New("owned API loopback lease closure failed")
	}

	closeForward, err := start(ctx, config.Kubectl, []string{
		developmentKubeconfigFlag, config.Kubeconfig, developmentContextFlag, config.Context,
		"-n", release.Namespace, "port-forward", "--address=127.0.0.1", "pod/" + instance.PodName, strconv.Itoa(port) + ":8080",
	})
	if err != nil {
		return errors.New("owned API loopback forwarding unavailable")
	}
	defer closeForward()

	transport := &http.Transport{DisableKeepAlives: true, TLSClientConfig: &tls.Config{
		MinVersion: tls.VersionTLS13, RootCAs: roots,
		ServerName: instance.Name + "." + release.Namespace + ".svc." + release.ClusterDomain,
	}}
	defer transport.CloseIdleConnections()

	client := &http.Client{Transport: transport, Timeout: developmentRequestTimeout}
	endpoint := cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + strconv.Itoa(port)

	if err := waitDevelopment(ctx, func() error {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+cryptoutilSharedMagic.DefaultServiceHealthEndpoint, nil)
		if err != nil {
			return errors.New("owned API request rejected")
		}

		response, err := client.Do(request)
		if err != nil {
			return errors.New("owned API not yet reachable")
		}

		_ = response.Body.Close()

		return nil
	}); err != nil {
		return errors.New("owned API verified TLS readiness unavailable")
	}

	return verifyDevelopmentAuthentication(ctx, client, endpoint)
}

func verifyDevelopmentAuthentication(ctx context.Context, client *http.Client, endpoint string) error {
	call := func(method, route, user, password, token string, body []byte, expected int, result any) error {
		request, err := http.NewRequestWithContext(ctx, method, endpoint+route, bytes.NewReader(body))
		if err != nil {
			return errors.New("owned authenticated request rejected")
		}

		request.Header.Set("Content-Type", "application/json")

		if user != "" {
			request.SetBasicAuth(user, password)
		}

		if token != "" {
			request.Header.Set("Authorization", cryptoutilSharedMagic.AuthorizationBearerPrefix+token)
		}

		response, err := client.Do(request)
		if err != nil {
			return errors.New("owned authenticated transport unavailable")
		}
		defer func() { _ = response.Body.Close() }()

		if response.StatusCode != expected {
			return errors.New("owned authentication response contract rejected")
		}

		if result != nil {
			if err := json.NewDecoder(io.LimitReader(response.Body, developmentMaximumBody)).Decode(result); err != nil {
				return errors.New("owned authentication response encoding rejected")
			}
		}

		return nil
	}

	var credentials struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		TenantID     string `json:"tenant_id"`
	}

	if err := call(http.MethodGet, cryptoutilSharedMagic.DefaultServiceHealthEndpoint, "", "", "", nil, http.StatusUnauthorized, nil); err != nil {
		return err
	}

	if err := call(http.MethodPost, cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath, "", "", "",
		[]byte(`{"tenant_name":"p2sm1103-owned-smoke"}`), http.StatusCreated, &credentials); err != nil {
		return err
	}

	if credentials.ClientID == "" || credentials.ClientSecret == "" || credentials.TenantID == "" {
		return errors.New("owned client registration incomplete")
	}

	if err := call(http.MethodPost, cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
		credentials.ClientID, "deliberately-invalid", "", nil, http.StatusUnauthorized, nil); err != nil {
		return err
	}

	var issued struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"`
	}
	if err := call(http.MethodPost, cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
		credentials.ClientID, credentials.ClientSecret, "", nil, http.StatusOK, &issued); err != nil {
		return err
	}

	if issued.Token == "" || issued.TokenType != cryptoutilSharedMagic.AuthorizationBearer {
		return errors.New("owned service token incomplete")
	}

	return call(http.MethodGet, cryptoutilSharedMagic.DefaultServiceHealthEndpoint, "", "", issued.Token, nil, http.StatusOK, nil)
}
