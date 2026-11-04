// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testPublicationMissingOwned = "missing owned"

func TestBuildPublication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*testPKIFixture, *RuntimeInputs, *[]PublicationSecret)
		want string
	}{
		{name: "verified publication"},
		{name: "oversized publication", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			(*v)[0].Data[SecretDataPublicCertificate] = bytes.Repeat([]byte("x"), cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes+1)
		}, want: "payload bound"},
		{name: "invalid config", edit: func(f *testPKIFixture, _ *RuntimeInputs, _ *[]PublicationSecret) {
			f.config.PSID = ""
		}, want: "instance contract"},
		{name: "publication invalid namespace", edit: func(_ *testPKIFixture, v *RuntimeInputs, _ *[]PublicationSecret) {
			v.Namespace = ""
		}, want: "explicit namespace"},
		{name: "zero time", edit: func(f *testPKIFixture, _ *RuntimeInputs, _ *[]PublicationSecret) {
			f.now = time.Time{}
		}, want: "validation time"},
		{name: "unknown pod", edit: func(_ *testPKIFixture, v *RuntimeInputs, _ *[]PublicationSecret) {
			v.PodName = "unknown-0"
		}, want: "exactly one service"},
		{name: "wrong admin", edit: func(_ *testPKIFixture, v *RuntimeInputs, _ *[]PublicationSecret) {
			v.ExpectedAdminDNS = testWrongDNSName
		}, want: "administrative"},
		{name: "bad trust", edit: func(f *testPKIFixture, _ *RuntimeInputs, _ *[]PublicationSecret) {
			f.trust.PublicServerPEM = nil
		}, want: "CA anchors"},
		{name: "missing import", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) { *v = (*v)[1:] }, want: testPublicationMissingOwned},
		{name: "duplicate publication", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			*v = append(*v, (*v)[0])
		}, want: "duplicate"},
		{name: "wrong owner", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			(*v)[0].Metadata.Labels["cu.io/owner"] = "forged"
		}, want: testPublicationMissingOwned},
		{name: "forged role", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			(*v)[0].Metadata.Labels[labelRole] = roleHealth
		}, want: "forged identity"},
		{name: "wrong type", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			(*v)[0].Type = "invalid-type"
		}, want: testOpaqueV1SecretError},
		{name: "publication missing key", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			delete((*v)[0].Data, SecretDataPublicPrivateKey)
		}, want: "missing or unexpected"},
		{name: "mismatched key", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			(*v)[0].Data[SecretDataPublicPrivateKey] = (*v)[0].Data[SecretDataAdminPrivateKey]
		}, want: "private key do not match"},
		{name: "forged cert", edit: func(_ *testPKIFixture, _ *RuntimeInputs, v *[]PublicationSecret) {
			(*v)[0].Data[SecretDataPublicCertificate] = []byte("forged")
		}, want: "parse certificate"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := newTestPKIFixture(t)
			admin, err := expectedAdminServerDNS(f.config, f.service, testNamespace)
			require.NoError(t, err)

			inputs := RuntimeInputs{
				Namespace: testNamespace, PodName: f.service.PodName,
				ExpectedAdminDNS: admin, ExpectedPublicDNSNames: testExpectedPublicDNSNames(f.config, f.service, testNamespace),
			}

			imports := make([]PublicationSecret, 0, 3)
			for index, name := range []string{f.service.TLSLeafSecretName, f.service.HealthClientSecretName, f.service.OTLPClientSecretName} {
				imports = append(imports, PublicationSecret{
					APIVersion: cryptoutilSharedMagic.DeploymentMaterializerKubernetesAPIVersion, Kind: cryptoutilSharedMagic.DeploymentMaterializerKubernetesSecretKind, Type: cryptoutilSharedMagic.DeploymentMaterializerSecretTypeOpaque,
					Metadata: PublicationMetadata{
						Name: name, Namespace: testNamespace,
						Labels: map[string]string{labelRun: f.config.Run, "cu.io/owner": "external-secret-import"},
					},
					Data: []map[string][]byte{f.tlsBundle, f.health, f.otlp}[index],
				})
			}

			if test.edit != nil {
				test.edit(f, &inputs, &imports)
			}

			got, err := BuildPublication(f.config, inputs, f.trust, imports, f.now)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, got)

				return
			}

			require.NoError(t, err)
			require.Len(t, got, 3)
			require.Equal(t, roleTLSBundle, got[0].Metadata.Labels[labelRole])
			require.NotContains(t, imports[0].Metadata.Labels, labelRole)

			got[0].Data[SecretDataPublicPrivateKey][0] ^= 1
			require.NotEqual(t, got[0].Data[SecretDataPublicPrivateKey], imports[0].Data[SecretDataPublicPrivateKey])
		})
	}
}
