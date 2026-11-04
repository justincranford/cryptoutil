// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"crypto/tls"
	"crypto/x509"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"time"
)

func validateDevelopmentKubeconfigIdentity(config, cluster, context map[string]any) error {
	binding, ok := context[developmentKubeContext].(map[string]any)
	if !ok || binding["cluster"] != developmentDefaultContext || binding[developmentKubeUser] != developmentDefaultContext {
		return errors.New("owned K3s context identity rejected")
	}

	if cluster["insecure-skip-tls-verify"] != nil || cluster["proxy-url"] != nil {
		return errors.New("owned K3s explicit verified transport required")
	}

	decode := func(value any) ([]byte, error) {
		text, ok := value.(string)
		if !ok || len(text) == 0 || len(text) > developmentMaximumBody {
			return nil, errors.New("owned K3s protected TLS encoding rejected")
		}

		raw, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, errors.New("owned K3s protected TLS encoding rejected")
		}

		return raw, nil
	}

	trust, err := decode(cluster[developmentKubeCA])
	if err != nil {
		return err
	}

	block, rest := pem.Decode(trust)
	if block == nil || block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate || len(rest) != 0 {
		return errors.New("owned K3s explicit CA encoding rejected")
	}

	ca, err := x509.ParseCertificate(block.Bytes)

	now := time.Now().UTC()
	if err != nil || !ca.IsCA || !ca.BasicConstraintsValid || ca.KeyUsage&x509.KeyUsageCertSign == 0 ||
		now.Before(ca.NotBefore) || !now.Before(ca.NotAfter) {
		return errors.New("owned K3s explicit CA policy rejected")
	}

	users, ok := config[developmentKubeUsers].([]any)
	if !ok || len(users) != 1 {
		return errors.New("owned K3s explicit client required")
	}

	entry, ok := users[0].(map[string]any)
	if !ok || entry[developmentName] != developmentDefaultContext {
		return errors.New("owned K3s client identity rejected")
	}

	user, ok := entry[developmentKubeUser].(map[string]any)
	if !ok || len(user) != 2 {
		return errors.New("owned K3s certificate-only client required")
	}

	cert, err := decode(user[developmentKubeCertificate])
	if err != nil {
		return err
	}

	key, err := decode(user[developmentKubeKey])
	if err != nil {
		return err
	}

	pair, err := tls.X509KeyPair(cert, key)
	if err != nil || len(pair.Certificate) == 0 {
		return errors.New("owned K3s client keypair rejected")
	}

	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil || leaf.IsCA || now.Before(leaf.NotBefore) || !now.Before(leaf.NotAfter) {
		return errors.New("owned K3s client certificate policy rejected")
	}

	for _, usage := range leaf.ExtKeyUsage {
		if usage == x509.ExtKeyUsageClientAuth {
			return nil
		}
	}

	return errors.New("owned K3s client authentication usage required")
}
