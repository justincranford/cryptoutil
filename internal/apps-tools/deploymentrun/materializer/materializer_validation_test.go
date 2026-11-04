// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewUsesFixedRuntimePaths(t *testing.T) {
	t.Parallel()

	materializer := New(Dependencies{})
	require.Equal(t, defaultOutputDirectory, materializer.outputDirectory)
	require.Equal(t, cryptoutilSharedMagic.DeploymentMaterializerDefaultOutputDirectory, materializer.outputDirectory)
	require.NotNil(t, materializer.readFile)
	require.NotNil(t, materializer.fileSystem)
	require.NotNil(t, materializer.now)
	require.WithinDuration(t, time.Now().UTC(), materializer.now(), time.Second)
	require.Equal(t, cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath, cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath)
	require.Equal(t, cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath, cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath)
}

func TestValidateTrustInputsRespectsPoolScope(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)

	tests := []struct {
		name    string
		config  func(*testing.T) InstanceConfig
		trust   func(TrustInputs) TrustInputs
		wantErr string
	}{
		{
			name:   "pg requires PostgreSQL anchors",
			config: func(_ *testing.T) InstanceConfig { return fixture.config },
			trust: func(trust TrustInputs) TrustInputs {
				trust.PostgreSQLClientPEM = nil

				return trust
			},
			wantErr: "trusted CA anchors for postgres-client",
		},
		{
			name: "non-pg pool needs only four issuer anchors",
			config: func(t *testing.T) InstanceConfig {
				return testInstanceConfigForPool(t, cryptoutilSharedMagic.DeploymentMaterializerPoolMem)
			},
			trust: func(trust TrustInputs) TrustInputs {
				trust.PostgreSQLClientPEM = nil

				return trust
			},
		},
		{
			name:   "pg pool requires binding for each configured pod",
			config: func(_ *testing.T) InstanceConfig { return fixture.config },
			trust: func(trust TrustInputs) TrustInputs {
				delete(trust.PostgreSQLClientBindings, "identity-authz-pg-1")
				trust.PostgreSQLClientBindings["identity-authz-pg-2"] = PostgreSQLClientBinding{
					CommonName: testAnotherPodClientCommonName, DatabaseRole: "identity_authz_client",
				}

				return trust
			},
			wantErr: testExactlyCoverConfiguredPodsError,
		},
		{
			name:   "pg pool rejects duplicate per-pod common names",
			config: func(_ *testing.T) InstanceConfig { return fixture.config },
			trust: func(trust TrustInputs) TrustInputs {
				binding := trust.PostgreSQLClientBindings["identity-authz-pg-0"]
				binding.CommonName = trust.PostgreSQLClientBindings["identity-authz-pg-1"].CommonName
				trust.PostgreSQLClientBindings["identity-authz-pg-0"] = binding

				return trust
			},
			wantErr: testNotUniquePerPodError,
		},
		{
			name:   "pg pool requires explicit database role",
			config: func(_ *testing.T) InstanceConfig { return fixture.config },
			trust: func(trust TrustInputs) TrustInputs {
				binding := trust.PostgreSQLClientBindings["identity-authz-pg-1"]
				binding.DatabaseRole = ""
				trust.PostgreSQLClientBindings["identity-authz-pg-1"] = binding

				return trust
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
		{
			name:   "malformed public anchor is rejected",
			config: func(_ *testing.T) InstanceConfig { return fixture.config },
			trust: func(trust TrustInputs) TrustInputs {
				trust.PublicServerPEM = []byte(testMalformedError)

				return trust
			},
			wantErr: "trusted CA anchors for public-server",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := testCase.config(t)
			trust := cloneTestPKIFixture(fixture).trust

			err := validateTrustInputs(config, testCase.trust(trust), "identity-authz-pg-1")
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateTrustInputsRejectsUnknownCurrentPodBinding(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	err := validateTrustInputs(fixture.config, fixture.trust, "identity-authz-pg-9")
	require.ErrorContains(t, err, "binding for the current pod is missing")
}

func TestPostgreSQLClientBindingValidation(t *testing.T) {
	t.Parallel()

	services := []InstanceService{{PodName: testDatabasePodZero}, {PodName: testDatabasePodNameOne}}
	validBindings := map[string]PostgreSQLClientBinding{
		testDatabasePodZero:    {CommonName: testDatabaseClientCommonNameZero, DatabaseRole: testApplicationDatabaseRole},
		testDatabasePodNameOne: {CommonName: "client-db-1", DatabaseRole: testApplicationDatabaseRole},
	}

	tests := []struct {
		name     string
		bindings map[string]PostgreSQLClientBinding
		wantErr  string
	}{
		{name: "one explicit unique binding per pod", bindings: validBindings},
		{name: "empty bindings", bindings: map[string]PostgreSQLClientBinding{}, wantErr: "bindings are empty"},
		{
			name: "missing configured pod",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero: validBindings[testDatabasePodZero],
			},
			wantErr: testExactlyCoverConfiguredPodsError,
		},
		{
			name: "unconfigured pod",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero: validBindings[testDatabasePodZero], "db-2": validBindings[testDatabasePodNameOne],
			},
			wantErr: testExactlyCoverConfiguredPodsError,
		},
		{
			name: "invalid pod identity",
			bindings: map[string]PostgreSQLClientBinding{
				"../db-0": validBindings[testDatabasePodZero], testDatabasePodNameOne: validBindings[testDatabasePodNameOne],
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
		{
			name: "duplicate common names",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero:    validBindings[testDatabasePodZero],
				testDatabasePodNameOne: {CommonName: testDatabaseClientCommonNameZero, DatabaseRole: testApplicationDatabaseRole},
			},
			wantErr: testNotUniquePerPodError,
		},
		{
			name: "empty common name",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero: {DatabaseRole: testApplicationDatabaseRole}, testDatabasePodNameOne: validBindings[testDatabasePodNameOne],
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
		{
			name: "empty database role",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero: {CommonName: testDatabaseClientCommonNameZero}, testDatabasePodNameOne: validBindings[testDatabasePodNameOne],
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
		{
			name: "control in database role",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero:    {CommonName: testDatabaseClientCommonNameZero, DatabaseRole: "application\x00role"},
				testDatabasePodNameOne: validBindings[testDatabasePodNameOne],
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
		{
			name: "padded common name",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero:    {CommonName: " client-db-0 ", DatabaseRole: testApplicationDatabaseRole},
				testDatabasePodNameOne: validBindings[testDatabasePodNameOne],
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
		{
			name: "oversized database role",
			bindings: map[string]PostgreSQLClientBinding{
				testDatabasePodZero:    {CommonName: testDatabaseClientCommonNameZero, DatabaseRole: strings.Repeat("r", 64)},
				testDatabasePodNameOne: validBindings[testDatabasePodNameOne],
			},
			wantErr: testInvalidPostgreSQLBindingError,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validatePostgreSQLClientBindings(testCase.bindings, services)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func nilContextForTest() context.Context {
	return nil
}

func TestMaterializeRejectsNilContext(t *testing.T) {
	t.Parallel()

	err := New(Dependencies{}).Materialize(nilContextForTest(), nil, RuntimeInputs{}, TrustInputs{})
	require.ErrorContains(t, err, "context is nil")
}

func TestMaterializeRejectsInvalidAPIEndpointBeforeReads(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)

	tests := []struct {
		name       string
		endpoint   string
		serverName string
		wantErr    string
	}{
		{
			name:       "DNS endpoint is not trusted as API source",
			endpoint:   "https://kubernetes.test.svc:443",
			serverName: testAPIServerName,
			wantErr:    testEndpointOrTLSServerNameError,
		},
		{
			name:       "invalid server name",
			endpoint:   testAPIEndpoint,
			serverName: testInvalidAPIServerName,
			wantErr:    testEndpointOrTLSServerNameError,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			harness := newMaterializerHarness(t, fixture)
			harness.inputs.APIEndpoint = testCase.endpoint
			harness.inputs.APIServerName = testCase.serverName
			err := harness.materializer().Materialize(
				context.Background(), harness.configYAML, harness.inputs, harness.trust,
			)
			require.ErrorContains(t, err, testCase.wantErr)
			require.Empty(t, harness.readPaths)
			require.Empty(t, harness.doer.requests)
			require.NoDirExists(t, harness.output)
		})
	}
}

func TestValidateExpectedPublicDNSNames(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	validNames := testExpectedPublicDNSNames(fixture.config, fixture.service, testNamespace)

	tests := []struct {
		name    string
		config  InstanceConfig
		names   []string
		wantErr string
	}{
		{name: "derived service and pod FQDNs", config: fixture.config, names: validNames},
		{
			name:   "controller adds external SNI",
			config: fixture.config,
			names:  append(append([]string(nil), validNames...), testExternalLoginSNI),
		},
		{name: "missing names", config: fixture.config, wantErr: testMissingDerivedServiceOrPodFQDNError},
		{name: "missing pod FQDN", config: fixture.config, names: validNames[:1], wantErr: testMissingDerivedServiceOrPodFQDNError},
		{name: "invalid external SNI", config: fixture.config, names: append(validNames, "bad/name"), wantErr: "identity is invalid"},
		{name: "duplicate expected DNS name", config: fixture.config, names: append(validNames, validNames[0]), wantErr: "identity is duplicated"},
		{
			name:   "too many expected DNS names",
			config: fixture.config,
			names: func() []string {
				names := append([]string(nil), validNames...)
				for index := len(names); index < cryptoutilSharedMagic.DeploymentMaterializerMaximumExpectedPublicDNSNames+1; index++ {
					names = append(names, fmt.Sprintf("sni-%d.example", index))
				}

				return names
			}(),
			wantErr: testSizeLimitError,
		},
		{
			name: "derived Pod FQDN exceeds DNS limit",
			config: func() InstanceConfig {
				config := fixture.config
				config.ClusterDomain = strings.Repeat("a", 240)

				return config
			}(),
			names:   []string{"service.example", "pod.example"},
			wantErr: "derived public service or pod FQDN is invalid",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			names, err := validateExpectedPublicDNSNames(
				testCase.config, fixture.service, testNamespace, testCase.names,
			)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.names, names)

			if len(names) != 0 {
				names[0] = "mutated.example"
				require.NotEqual(t, names[0], testCase.names[0])
			}
		})
	}
}
