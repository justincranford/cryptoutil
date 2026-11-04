// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testVerifyInputFile     = "pki-verification.json"
	testVerifyInputArgument = "--input=/run/cryptoutil/secrets/owned/" + testVerifyInputFile
	testVerifySecretName    = "owned"
	testVerifyNamespace     = "owned-ns"
	testVerifySupported     = "supported"
	testVerifyMissingSource = "missing source"
	testVerifyProtectedFile = "protected/verification"
)

func TestPKIVerificationJobProtectedContract(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		args   []string
		names  []string
		source bool
		want   bool
	}{
		{testVerifySupported, []string{lifecycleVerifyCommand, testVerifyInputArgument}, []string{testVerifySecretName}, true, false},
		{"missing argument", []string{lifecycleVerifyCommand}, nil, true, true},
		{"wrong flag", []string{lifecycleVerifyCommand, "--output-dir=/tmp"}, []string{testVerifySecretName}, true, true},
		{"unsafe path", []string{lifecycleVerifyCommand, "--input=/etc/private"}, []string{testVerifySecretName}, true, true},
		{"unmounted", []string{lifecycleVerifyCommand, testVerifyInputArgument}, nil, true, true},
		{testVerifyMissingSource, []string{lifecycleVerifyCommand, testVerifyInputArgument}, []string{testVerifySecretName}, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			sources := []LifecycleSecretSource{}
			if test.source {
				sources = append(sources, LifecycleSecretSource{Namespace: testVerifyNamespace, Name: testVerifySecretName, Files: map[string]string{testVerifyInputFile: "protected/plan"}})
			}

			err := validatePKIVerificationCommand(LifecycleJobInput{Namespace: testVerifyNamespace, Args: test.args, SecretNames: test.names}, sources)
			if test.want {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBuildLifecycleVerifyOnlyPKIContract(t *testing.T) {
	t.Parallel()

	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "no token or writes", true: "reject absent manifest"}[failure], func(t *testing.T) {
			t.Parallel()

			model, _, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
			for jobIndex := range input.Jobs {
				job := &input.Jobs[jobIndex]
				if job.Name != lifecycleInfraPKIName {
					continue
				}

				for sourceIndex := range input.Secrets {
					source := &input.Secrets[sourceIndex]
					if source.Namespace != job.Namespace {
						continue
					}

					source.Files[testVerifyInputFile] = testVerifyProtectedFile
					job.Args = []string{lifecycleVerifyCommand, "--input=/run/cryptoutil/secrets/" + source.Name + "/" + testVerifyInputFile}
					job.SecretNames = []string{source.Name}
					job.WriteSecretNames = []string{}

					if failure {
						delete(source.Files, testVerifyInputFile)
					}

					break
				}

				break
			}

			revision, err := LifecycleRevision(input)
			require.NoError(t, err)

			model.Revision = revision
			inventory, err := Build(model, testServiceSMKMSPGSelector)
			require.NoError(t, err)

			values, err := BuildLifecycleChartValues(model, inventory, input)
			if failure {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			seen := false

			for _, job := range values.Jobs {
				if len(job.Args) > 0 && job.Args[0] == lifecycleVerifyCommand {
					require.False(t, job.APIToken)
					require.Empty(t, job.WriteSecretNames)

					seen = true
				}
			}

			require.True(t, seen)
		})
	}
}
