// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func preparePKIVerification(input gen.BootstrapPreflightInput, request execution.Request, manifest PKIPublicationInput) (execution.Request, error) {
	prepared, err := buildPKIPublication(input.RuntimeValues, request.Secrets, manifest, time.Now().UTC())
	if err != nil {
		return request, errors.New("published PKI certificate contract rejected")
	}
	// Enrich only expected labels; never substitute bytes or omit imports.
	for _, built := range prepared {
		for index, secret := range request.Secrets {
			if secret.Metadata.Namespace == built.Metadata.Namespace && secret.Metadata.Name == built.Metadata.Name {
				request.Secrets[index] = built
			}
		}
	}

	plans := map[string]pki.VerificationInput{}

	for _, job := range request.Input.Jobs {
		if job.Name != "infra-pki" && job.Name != cryptoutilSharedMagic.PSIDPKIInit {
			continue
		}

		if len(job.Args) != 2 || job.Args[0] != pki.VerifyCommand || len(job.WriteSecretNames) != 0 || !strings.HasPrefix(job.Args[1], "--input=") {
			return request, execution.ErrUnsupportedPKI
		}

		file := strings.TrimPrefix(job.Args[1], "--input=")
		if path.Base(file) != pki.VerificationInputName {
			return request, execution.ErrUnsupportedPKI
		}

		read := verificationImports(job.Namespace, job.SecretNames, request.Secrets)

		raw, err := read(file)
		if err != nil {
			return request, errors.New("PKI verification manifest absent from restricted imports")
		}

		plan, err := pki.DecodeVerificationInput(raw)
		if err != nil {
			return request, errors.New("PKI verification manifest rejected")
		}

		if err := pki.VerifyPublishedFiles(context.Background(), plan, read, time.Now().UTC()); err != nil {
			return request, errors.New("PKI verify-only material/certificate policy rejected")
		}
		// Every mounted material key is independently bound, except the plan
		// itself. Unused/unverified keys cannot be hidden behind a valid leaf.
		for _, secret := range request.Secrets {
			if secret.Metadata.Namespace != job.Namespace || !containsPKIName(job.SecretNames, secret.Metadata.Name) {
				continue
			}

			for key := range secret.Data {
				if "/run/cryptoutil/secrets/"+secret.Metadata.Name+"/"+key == file {
					continue
				}

				if plan.Files["/run/cryptoutil/secrets/"+secret.Metadata.Name+"/"+key] == "" {
					return request, errors.New("PKI verification plan does not cover restricted material")
				}
			}
		}

		plans[job.Namespace+"/"+job.Name] = plan
	}

	request.VerifyPKI = func(ctx context.Context, job gen.LifecycleJobValues, client execution.Client) error {
		reader, ok := client.(interface {
			VerifySecret(context.Context, gen.LifecycleSecret) error
			VerifyJob(context.Context, execution.Identity) error
		})
		if !ok {
			return execution.ErrPrerequisite
		}

		id := execution.Identity{Namespace: job.Namespace, Name: job.Name, Kind: "Job", Run: request.Model.Run, Release: job.Release, Revision: request.Model.Revision}
		if err := reader.VerifyJob(ctx, id); err != nil {
			return fmt.Errorf("postJob identity verification: %w", err)
		}

		for _, secret := range request.Secrets {
			if secret.Metadata.Namespace == job.Namespace && containsPKIName(job.SecretNames, secret.Metadata.Name) {
				if err := reader.VerifySecret(ctx, secret); err != nil {
					return fmt.Errorf("postJob material verification: %w", err)
				}
			}
		}

		base := strings.TrimSuffix(job.Name, "-"+request.Model.Revision)
		if err := pki.VerifyPublishedFiles(ctx, plans[job.Namespace+"/"+base], verificationImports(job.Namespace, job.SecretNames, request.Secrets), time.Now().UTC()); err != nil {
			return errors.New("postJob PKI certificate policy rejected")
		}
		// Runtime bundles use exactly the existing stronger ordinal/PG policy.
		if _, err := buildPKIPublication(input.RuntimeValues, request.Secrets, manifest, time.Now().UTC()); err != nil {
			return errors.New("postJob published runtime certificate policy rejected")
		}

		return nil
	}

	return request, nil
}

func containsPKIName(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}

	return false
}

func verificationImports(namespace string, names []string, secrets []gen.LifecycleSecret) func(string) ([]byte, error) {
	return func(file string) ([]byte, error) {
		for _, secret := range secrets {
			if secret.Metadata.Namespace != namespace || !containsPKIName(names, secret.Metadata.Name) {
				continue
			}

			for key, data := range secret.Data {
				if file == "/run/cryptoutil/secrets/"+secret.Metadata.Name+"/"+key {
					return data, nil
				}
			}
		}

		return nil, errors.New("restricted verification material unavailable")
	}
}
