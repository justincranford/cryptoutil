// Copyright (c) 2025-2026 Justin Cranford.
package lint_deployments

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestComposeIncludeUnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr string
	}{
		{name: "scalar include rejected", input: testSharedYML, wantErr: "include must be a mapping"},
		{name: "missing path rejected", input: "other: " + testSharedYML, wantErr: "missing a path"},
		{name: testSinglePath, input: "path: " + testSharedYML, want: []string{testSharedYML}},
		{name: "multiple paths", input: "path: [" + testSharedYML + ", extra.yml]", want: []string{testSharedYML, "extra.yml"}},
		{name: "unsupported path node rejected", input: "path: {}", wantErr: testStringOrSequence},
		{name: "numeric scalar path", input: "path: 42", want: []string{"42"}},
		{name: "numeric sequence path", input: "path: [shared.yml, 42]", want: []string{"shared.yml", "42"}},
		{name: "mapping sequence path rejected", input: "path: [{nested: value}]", wantErr: "decode include path"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got composeInclude

			err := yaml.Unmarshal([]byte(tc.input), &got)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got.Path)
		})
	}
}

func TestComposeVolumeUnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		wantType     string
		wantSource   string
		wantTarget   string
		wantReadOnly bool
		wantErr      string
	}{
		{
			name:         "short bind mount",
			input:        testCertVolumeMountSource + ":" + certVolumeMountTarget + ":ro,z",
			wantType:     cryptoutilSharedMagic.DockerComposeVolumeMountTypeBind,
			wantSource:   testCertVolumeMountSource,
			wantTarget:   certVolumeMountTarget,
			wantReadOnly: true,
		},
		{
			name:       "short mount with conflicting read-write option is writable",
			input:      testCertVolumeMountSource + ":" + certVolumeMountTarget + ":ro,rw",
			wantType:   cryptoutilSharedMagic.DockerComposeVolumeMountTypeBind,
			wantSource: testCertVolumeMountSource,
			wantTarget: certVolumeMountTarget,
		},
		{
			name:       "long bind mount infers type",
			input:      testComposeSourceField + ": " + testCertVolumeMountSource + "\n" + testComposeTargetField + ": " + certVolumeMountTarget,
			wantType:   cryptoutilSharedMagic.DockerComposeVolumeMountTypeBind,
			wantSource: testCertVolumeMountSource,
			wantTarget: certVolumeMountTarget,
		},
		{name: "short mount without absolute target", input: "data:relative", wantType: cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume},
		{
			name:         "long mount",
			input:        "type: volume\n" + testComposeSourceField + ": cert-data\n" + testComposeTargetField + ": " + certVolumeMountTarget + "\nread_only: true",
			wantType:     cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume,
			wantSource:   "cert-data",
			wantTarget:   certVolumeMountTarget,
			wantReadOnly: true,
		},
		{name: "unsupported node rejected", input: "[source, target]", wantErr: testStringOrMapping},
		{name: "invalid binary scalar rejected", input: "!!binary '***'", wantErr: "decode volume mount"},
		{name: "numeric scalar mount", input: "42", wantType: cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume},
		{name: "sequence mapping field rejected", input: "type: volume\nsource: [cert, data]\ntarget: " + certVolumeMountTarget, wantErr: "decode long volume mount"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got composeVolume

			err := yaml.Unmarshal([]byte(tc.input), &got)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantType, got.Type)
			require.Equal(t, tc.wantSource, got.Source)
			require.Equal(t, tc.wantTarget, got.Target)
			require.Equal(t, tc.wantReadOnly, got.ReadOnly)
		})
	}
}
