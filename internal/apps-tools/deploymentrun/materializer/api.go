// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	http "net/http"
	"net/url"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var labelNamePattern = regexp.MustCompile(`^[A-Za-z0-9](?:[-A-Za-z0-9_.]*[A-Za-z0-9])?$`)

type apiSecret struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   apiSecretMetadata `json:"metadata"`
	Type       string            `json:"type"`
	Data       map[string][]byte `json:"data"`
	StringData map[string]string `json:"stringData"`
}

type apiSecretMetadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type redactedCause struct {
	message string
	cause   error
}

func (cause *redactedCause) Error() string {
	return cause.message
}

func (cause *redactedCause) Unwrap() error {
	return cause.cause
}

func redactError(err error, secret string) error {
	if err == nil || secret == "" {
		return err
	}

	message := strings.ReplaceAll(err.Error(), secret, "[REDACTED]")
	if message == err.Error() {
		return err
	}

	return &redactedCause{message: message, cause: err}
}

func newAPIHTTPClient(caPEM []byte, serverName string) (*http.Client, error) {
	if !validAPIServerName(serverName) {
		return nil, fmt.Errorf("invalid Kubernetes API TLS server name")
	}

	certificates, err := parseCertificatePEM(caPEM)
	if err != nil {
		return nil, fmt.Errorf("parse projected Kubernetes API CA bundle: %w", err)
	}

	roots := x509.NewCertPool()

	for _, certificate := range certificates {
		if !certificate.IsCA || certificate.KeyUsage&x509.KeyUsageCertSign == 0 {
			return nil, fmt.Errorf("projected Kubernetes API CA bundle contains a non-CA certificate")
		}

		if err := validateApprovedCertificate(certificate); err != nil {
			return nil, fmt.Errorf("projected Kubernetes API CA bundle is not FIPS-approved: %w", err)
		}

		roots.AddCert(certificate)
	}

	transport := &http.Transport{
		Proxy: nil,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
			RootCAs:    roots,
			ServerName: serverName,
		},
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

func fetchSecret(
	ctx context.Context,
	doer HTTPDoer,
	token string,
	apiEndpoint string,
	namespace string,
	name string,
	config InstanceConfig,
	service InstanceService,
	role string,
) (map[string][]byte, error) {
	if err := contextError(ctx); err != nil {
		return nil, fmt.Errorf("secret GET was canceled: %w", err)
	}

	requestContext, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.DeploymentMaterializerAPIRequestTimeout)
	defer cancel()

	requestURL, err := makeSecretURL(apiEndpoint, namespace, name)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("construct Kubernetes Secret GET: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+token)

	response, err := doer.Do(request)
	if err != nil {
		requestErr := fmt.Errorf("secret GET transport failed: %w", redactError(err, token))
		if contextErr := requestContext.Err(); contextErr != nil {
			requestErr = errors.Join(
				fmt.Errorf("secret GET was canceled: %w", contextErr),
				requestErr,
			)
		}

		if response != nil && response.Body != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				return nil, errors.Join(requestErr, fmt.Errorf("close Kubernetes API response: %w", closeErr))
			}
		}

		return nil, requestErr
	}

	if response == nil || response.Body == nil {
		return nil, fmt.Errorf("secret GET returned an empty response")
	}

	if response.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("the Kubernetes API returned HTTP %d for Secret GET", response.StatusCode)
		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, errors.Join(statusErr, fmt.Errorf("close Kubernetes API response: %w", closeErr))
		}

		return nil, statusErr
	}

	if response.ContentLength > cryptoutilSharedMagic.DeploymentMaterializerMaximumSecretBodyBytes {
		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("the Kubernetes Secret response exceeds the size limit"),
				fmt.Errorf("close Kubernetes API response: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("the Kubernetes Secret response exceeds the size limit")
	}

	body, err := readBoundedBody(response.Body, cryptoutilSharedMagic.DeploymentMaterializerMaximumSecretBodyBytes)
	if err != nil {
		if contextErr := requestContext.Err(); contextErr != nil {
			return nil, errors.Join(
				fmt.Errorf("the Kubernetes Secret GET canceled: %w", contextErr),
				err,
			)
		}

		return nil, err
	}

	if err := contextError(requestContext); err != nil {
		return nil, fmt.Errorf("the Kubernetes Secret GET canceled: %w", err)
	}

	var secret apiSecret
	if err := json.Unmarshal(body, &secret); err != nil {
		return nil, fmt.Errorf("decode Kubernetes Secret response: %w", err)
	}

	if err := validateSecretIdentity(secret, namespace, name, config, service, role); err != nil {
		return nil, err
	}

	return secret.Data, nil
}

func readBoundedBody(body io.ReadCloser, maximum int64) ([]byte, error) {
	data, readErr := io.ReadAll(io.LimitReader(body, maximum+1))
	closeErr := body.Close()

	if readErr != nil {
		readError := fmt.Errorf("read Kubernetes Secret response: %w", readErr)
		if closeErr != nil {
			return nil, errors.Join(readError, fmt.Errorf("close Kubernetes API response: %w", closeErr))
		}

		return nil, readError
	}

	if closeErr != nil {
		return nil, fmt.Errorf("close Kubernetes API response: %w", closeErr)
	}

	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("the Kubernetes Secret response exceeds the size limit")
	}

	return data, nil
}

func validateSecretIdentity(
	secret apiSecret,
	namespace string,
	name string,
	config InstanceConfig,
	service InstanceService,
	role string,
) error {
	if secret.APIVersion != cryptoutilSharedMagic.DeploymentMaterializerKubernetesAPIVersion || secret.Kind != cryptoutilSharedMagic.DeploymentMaterializerKubernetesSecretKind || secret.Type != cryptoutilSharedMagic.DeploymentMaterializerSecretTypeOpaque {
		return fmt.Errorf("the Kubernetes API object is not an Opaque v1 Secret")
	}

	if secret.Metadata.Name != name || secret.Metadata.Namespace != namespace {
		return fmt.Errorf("the Kubernetes Secret metadata identity does not match the request")
	}

	if secret.StringData != nil {
		return fmt.Errorf("the Kubernetes Secret response unexpectedly contains stringData")
	}

	for key, value := range secret.Metadata.Labels {
		if !validLabelKey(key) || !validLabelValue(value) {
			return fmt.Errorf("the Kubernetes Secret contains a malformed label")
		}
	}

	expectedLabels := map[string]string{
		labelWorkload: config.PSID + "-" + config.Pool,
		labelPSID:     config.PSID,
		labelPool:     config.Pool,
		labelRun:      config.Run,
		labelRelease:  config.Release,
		labelRole:     role,
	}
	for key, expected := range expectedLabels {
		if secret.Metadata.Labels[key] != expected {
			return fmt.Errorf("the Kubernetes Secret identity label %q does not match", key)
		}
	}

	return nil
}

func validLabelKey(key string) bool {
	if key == "" {
		return false
	}

	parts := strings.Split(key, cryptoutilSharedMagic.DeploymentMaterializerPathRoot)
	if len(parts) > 2 {
		return false
	}

	if !labelNamePattern.MatchString(parts[len(parts)-1]) || len(parts[len(parts)-1]) > 63 {
		return false
	}

	if len(parts) == 2 && !validDNSSubdomain(parts[0]) {
		return false
	}

	return true
}

func validLabelValue(value string) bool {
	return len(value) <= 63 && (value == "" || labelValuePattern.MatchString(value))
}

func makeSecretURL(apiEndpoint string, namespace string, name string) (string, error) {
	if !validDNSLabel(namespace) || !validDNSSubdomain(name) {
		return "", fmt.Errorf("the Kubernetes Secret request identity is invalid")
	}

	endpointURL, err := validateAPIEndpoint(apiEndpoint, "")
	if err != nil {
		return "", err
	}

	endpointURL.Path = cryptoutilSharedMagic.DeploymentMaterializerSecretAPIPrefix +
		namespace + cryptoutilSharedMagic.DeploymentMaterializerSecretAPIMiddle + name

	return endpointURL.String(), nil
}

func validateAPIEndpoint(apiEndpoint string, serverName string) (*url.URL, error) {
	endpointURL, err := url.Parse(apiEndpoint)
	if err != nil || endpointURL.Scheme != cryptoutilSharedMagic.ProtocolHTTPS || endpointURL.Opaque != "" ||
		endpointURL.User != nil || endpointURL.Host == "" ||
		(endpointURL.Path != "" && endpointURL.Path != cryptoutilSharedMagic.DeploymentMaterializerPathRoot) ||
		endpointURL.RawQuery != "" || endpointURL.ForceQuery ||
		endpointURL.Fragment != "" || strings.Contains(apiEndpoint, "#") ||
		endpointURL.Port() != "443" || net.ParseIP(endpointURL.Hostname()) == nil ||
		!validAPIServerName(serverName) {
		return nil, fmt.Errorf("the Kubernetes API endpoint or TLS server name is invalid")
	}

	return endpointURL, nil
}

func validAPIServerName(serverName string) bool {
	return serverName == "" || net.ParseIP(serverName) != nil || validDNSSubdomain(serverName)
}

func validateToken(tokenBytes []byte) (string, error) {
	if len(tokenBytes) > cryptoutilSharedMagic.DeploymentMaterializerMaximumTokenBytes {
		return "", fmt.Errorf("projected service-account token exceeds the size limit")
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" || strings.ContainsAny(token, " \t\r\n") {
		return "", fmt.Errorf("projected service-account token is empty or malformed")
	}

	return token, nil
}
