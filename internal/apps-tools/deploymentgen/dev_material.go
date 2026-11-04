// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"crypto/elliptic"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	json "encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"path"
	"strings"
	"time"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DevelopmentMaterialPolicy is explicit disposable trust, not production certification.
type DevelopmentMaterialPolicy struct {
	Curve    elliptic.Curve
	Lifetime time.Duration
}

// DevelopmentMaterial holds protected bytes only in memory. Callers must never
// log, serialize to public artifacts, or substitute these for existing data.
type DevelopmentMaterial struct {
	Plan    DevelopmentRuntimePlan
	Secrets []LifecycleSecret
	Token   []byte
}

// GenerateDevelopmentMaterial supports the initial explicitly owned empty-data
// SQLite smoke. PostgreSQL is rejected until complete protected DB mappings exist.
func GenerateDevelopmentMaterial(ctx context.Context, plan DevelopmentRuntimePlan, policy DevelopmentMaterialPolicy) (DevelopmentMaterial, error) {
	return generateDevelopmentMaterial(ctx, plan, policy, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair)
}

func generateDevelopmentMaterial(ctx context.Context, plan DevelopmentRuntimePlan, policy DevelopmentMaterialPolicy,
	generate func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error),
) (DevelopmentMaterial, error) {
	return generateDevelopmentMaterialWithEntropy(ctx, plan, policy, generate, crand.Read)
}

func generateDevelopmentMaterialWithEntropy(ctx context.Context, plan DevelopmentRuntimePlan, policy DevelopmentMaterialPolicy,
	generate func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error), entropy func([]byte) (int, error),
) (DevelopmentMaterial, error) {
	result := DevelopmentMaterial{}
	if ctx == nil || generate == nil || entropy == nil || plan.Model.Run != DevelopmentSmokeRun || plan.Selector != DevelopmentSmokeSelector ||
		policy.Curve == nil || (policy.Curve != elliptic.P256() && policy.Curve != elliptic.P384() && policy.Curve != elliptic.P521()) ||
		policy.Lifetime < time.Hour || policy.Lifetime > 7*24*time.Hour {
		return result, errors.New("explicit bounded disposable smoke policy required")
	}

	inventory, err := Build(plan.Model, plan.Selector)
	if err != nil {
		return result, err
	}

	if _, err := BuildLifecycleChartValues(plan.Model, inventory, plan.Lifecycle); err != nil {
		return result, err
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("development generation canceled: %w", err)
	}

	caKey, err := generate(policy.Curve)
	if err != nil {
		return result, errors.New("development issuer generation failed")
	}

	issuer, err := cryptoutilSharedCryptoCertificate.CreateCASubject(nil, nil, "disposable-"+plan.Model.Run, caKey, policy.Lifetime, 0)
	if err != nil {
		return result, errors.New("development issuer creation failed")
	}

	trust := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: issuer.KeyMaterial.CertificateChain[0].Raw})
	material := map[string]map[string][]byte{}
	verification := map[string]pki.VerificationInput{}

	for _, source := range plan.Lifecycle.Secrets {
		data := map[string][]byte{}

		for key := range source.Files {
			if strings.HasSuffix(key, cryptoutilSharedMagic.PKIInitPublicCertFileExtension) {
				data[key] = append([]byte(nil), trust...)
			} else if key != pki.VerificationInputName && !strings.HasSuffix(key, cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension) {
				raw := make([]byte, 32)
				if count, err := entropy(raw); err != nil || count != len(raw) {
					return result, errors.New("development entropy unavailable")
				}

				data[key] = []byte(hex.EncodeToString(raw))
			}
		}

		material[source.Namespace+"/"+source.Name] = data
	}

	issue := func(namespace, secret, certificateKey, privateKey, usage, commonName string, dns []string, trustSecret string) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("development generation canceled: %w", err)
		}

		kp, err := generate(policy.Curve)
		if err != nil {
			return errors.New("development leaf generation failed")
		}

		eku := x509.ExtKeyUsageClientAuth
		if usage == developmentServerUsage {
			eku = x509.ExtKeyUsageServerAuth
		}

		ips := []net.IP(nil)
		if certificateKey == pki.SecretDataAdminCertificate {
			ips = []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback), net.ParseIP(cryptoutilSharedMagic.IPv6Loopback)}
		}

		leaf, err := cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(issuer, kp, commonName, policy.Lifetime, dns, ips, nil, nil,
			x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{eku})
		if err != nil {
			return errors.New("development leaf creation failed")
		}

		der, err := x509.MarshalPKCS8PrivateKey(kp.Private)
		if err != nil {
			return errors.New("development private encoding failed")
		}

		data := material[namespace+"/"+secret]
		if data == nil {
			return errors.New("development certificate source absent")
		}

		data[certificateKey] = append(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: leaf.KeyMaterial.CertificateChain[0].Raw}), trust...)
		data[privateKey] = pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: der})
		v := verification[namespace]
		v.Version = 1
		v.Certificates = append(v.Certificates, pki.VerificationCertificate{
			Certificate: path.Join("/run/cryptoutil/secrets", secret, certificateKey),
			Key:         path.Join("/run/cryptoutil/secrets", secret, privateKey),
			Trust:       path.Join("/run/cryptoutil/secrets", trustSecret, developmentCAKey),
			Usage:       usage, DNSNames: dns, CommonName: commonName,
		})
		verification[namespace] = v

		return nil
	}

	for _, value := range plan.runtimeValues {
		for _, workload := range value.Workloads {
			var input pki.CommandInput
			if err := json.Unmarshal([]byte(workload.MaterializerConfig), &input); err != nil {
				return result, errors.New("development materializer contract rejected")
			}

			for _, binding := range input.Pods {
				for _, pair := range []struct {
					secret, certificate, key, usage string
					dns                             []string
				}{
					{binding.PodName + "-tls", pki.SecretDataPublicCertificate, pki.SecretDataPublicPrivateKey, developmentServerUsage, binding.ExpectedPublicDNSNames},
					{binding.PodName + "-tls", pki.SecretDataAdminCertificate, pki.SecretDataAdminPrivateKey, developmentServerUsage, []string{binding.ExpectedAdminDNS}},
					{binding.PodName + "-health-client", pki.SecretDataCertificate, pki.SecretDataPrivateKey, developmentClientUsage, nil},
					{binding.PodName + "-otlp-client", pki.SecretDataCertificate, pki.SecretDataPrivateKey, developmentClientUsage, nil},
				} {
					if err := issue(value.Namespace, pair.secret, pair.certificate, pair.key, pair.usage, binding.PodName, pair.dns, workload.TrustSecretName); err != nil {
						return result, err
					}
				}
			}

			if err := issue(value.Namespace, workload.PSID+"-e2e-tls", pki.SecretDataCertificate, pki.SecretDataPrivateKey,
				developmentClientUsage, workload.PSID+"-smoke", nil, workload.TrustSecretName); err != nil {
				return result, err
			}
		}
	}

	for _, pair := range []struct{ secret, usage, service, trust string }{
		{developmentCollectorTLS, developmentServerUsage, developmentCollectorName, "collector-client-ca"},
		{"lgtm-server-tls", developmentServerUsage, "lgtm", "lgtm-server-ca"},
		{"collector-lgtm-client-tls", developmentClientUsage, developmentCollectorName, "lgtm-client-ca"},
	} {
		dns := []string(nil)
		if pair.usage == developmentServerUsage {
			dns = []string{pair.service + "." + inventory.InfraNamespace + ".svc." + plan.Model.ClusterDomain}
		}

		if err := issue(inventory.InfraNamespace, pair.secret, pki.SecretDataCertificate, pki.SecretDataPrivateKey,
			pair.usage, pair.service, dns, pair.trust); err != nil {
			return result, err
		}
	}

	for _, job := range plan.Lifecycle.Jobs {
		v := verification[job.Namespace]
		v.Files = map[string]string{}

		for _, name := range job.SecretNames {
			for key, raw := range material[job.Namespace+"/"+name] {
				hash := sha256.Sum256(raw)
				v.Files[path.Join("/run/cryptoutil/secrets", name, key)] = hex.EncodeToString(hash[:])
			}
		}

		seen := map[string]bool{}
		for _, binding := range v.Certificates {
			if seen[binding.Certificate] {
				return result, fmt.Errorf("duplicate generated verification certificate %s", binding.Certificate)
			}

			seen[binding.Certificate] = true
			if v.Files[binding.Certificate] == "" || v.Files[binding.Key] == "" || v.Files[binding.Trust] == "" {
				return result, fmt.Errorf("generated verification mount missing for public source %s", binding.Certificate)
			}

			for _, file := range []string{binding.Certificate, binding.Key, binding.Trust} {
				parts := strings.Split(strings.TrimPrefix(file, "/run/cryptoutil/secrets/"), "/")
				if len(material[job.Namespace+"/"+parts[0]][parts[1]]) == 0 {
					return result, fmt.Errorf("empty generated verification source %s", file)
				}
			}
		}

		if err := pki.VerifyPublishedFiles(ctx, v, func(file string) ([]byte, error) {
			parts := strings.Split(strings.TrimPrefix(file, "/run/cryptoutil/secrets/"), "/")

			return material[job.Namespace+"/"+parts[0]][parts[1]], nil
		}, time.Now().UTC()); err != nil {
			return result, fmt.Errorf("generated development policy verification: %w", err)
		}

		raw, err := json.Marshal(v)
		if err != nil {
			return result, errors.New("development verification encoding failed")
		}

		name := strings.Split(strings.TrimPrefix(job.Args[1], "--input=/run/cryptoutil/secrets/"), "/")[0]
		material[job.Namespace+"/"+name][pki.VerificationInputName] = raw
	}

	for _, source := range plan.Lifecycle.Secrets {
		data := material[source.Namespace+"/"+source.Name]
		if len(data) != len(source.Files) {
			return result, errors.New("generated development Secret key coverage rejected")
		}

		result.Secrets = append(result.Secrets, LifecycleSecret{
			APIVersion: cryptoutilSharedMagic.DeploymentMaterializerKubernetesAPIVersion, Kind: cryptoutilSharedMagic.KubernetesSecretKind, Type: source.Type, Data: data,
			Metadata: LifecycleMetadata{
				Name: source.Name, Namespace: source.Namespace,
				Labels: map[string]string{pki.SecretLabelRun: plan.Model.Run, developmentOwnerLabel: developmentSecretOwner},
			},
		})
	}

	token := make([]byte, 32)
	if count, err := entropy(token); err != nil || count != len(token) {
		return DevelopmentMaterial{}, errors.New("development token entropy unavailable")
	}

	result.Token, result.Plan = []byte(hex.EncodeToString(token)), plan

	return result, nil
}
