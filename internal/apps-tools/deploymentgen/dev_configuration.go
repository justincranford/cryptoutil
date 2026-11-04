// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"errors"
	"fmt"
	"path"
	"strings"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

// ConfigureDevelopmentSmoke replaces budget placeholders with protected mount
// references and explicit listener/client policy. It never inserts credentials.
func ConfigureDevelopmentSmoke(plan DevelopmentRuntimePlan, domain []byte) (DevelopmentRuntimePlan, error) {
	if plan.Model.Run != DevelopmentSmokeRun || plan.Selector != DevelopmentSmokeSelector || len(domain) == 0 || len(domain) > 65536 {
		return DevelopmentRuntimePlan{}, errors.New("explicit disposable smoke domain configuration required")
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(domain, &parsed); err != nil || len(parsed) == 0 {
		return DevelopmentRuntimePlan{}, errors.New("development domain configuration rejected")
	}

	inputs := map[string]RuntimeWorkloadInput{}

	for key, old := range plan.RuntimeInput.Runtime.Workloads {
		input := old

		input.FrameworkConfig = "bind-public-address: 0.0.0.0\nbind-public-port: 8080\nbind-admin-address: 127.0.0.1\nbind-admin-port: 9090\n" +
			"tls-enabled: true\notlp-enabled: false\nlog-level: INFO\ndatabase-url: sqlite://file::memory:?cache=shared\n" +
			"unseal-mode: 3-of-5\nunseal-files:\n"
		for index := 1; index <= 5; index++ {
			input.FrameworkConfig += fmt.Sprintf("  - /run/cryptoutil/credentials/0/unseal-%dof5.secret\n", index)
		}

		for _, name := range []string{"browser-username", "browser-password", "service-username", "service-password"} {
			input.FrameworkConfig += name + "-file: /run/cryptoutil/credentials/2/" + name + ".secret\n"
		}

		input.FrameworkConfig += "server-admin-tls-client-policy: require-and-verify\nserver-admin-tls-ca-file: /run/cryptoutil/trust/ca.crt\n" +
			"server-public-tls-client-policy: none\n" +
			"server-public-tls-cert-file: " + path.Join(pki.DefaultOutputDirectory, pki.PublicServerCertificateFile) + "\n" +
			"server-public-tls-key-file: " + path.Join(pki.DefaultOutputDirectory, pki.PublicServerPrivateKeyFile) + "\n" +
			"server-admin-tls-cert-file: " + path.Join(pki.DefaultOutputDirectory, pki.AdminServerCertificateFile) + "\n" +
			"server-admin-tls-key-file: " + path.Join(pki.DefaultOutputDirectory, pki.AdminServerPrivateKeyFile) + "\n"
		input.DomainConfig = string(domain)
		inputs[key] = input
	}

	images := map[string]ContainerImage{}
	flags := []string{
		"--cacert=/run/cryptoutil/trust/ca.crt",
		"--cert=" + path.Join(pki.DefaultOutputDirectory, pki.HealthClientCertificateFile),
		"--key=" + path.Join(pki.DefaultOutputDirectory, pki.HealthClientPrivateKeyFile),
	}

	for psid, old := range plan.RuntimeInput.Images {
		image := old
		image.Probes.Startup = append(append([]string(nil), old.Probes.Startup...), flags...)
		image.Probes.Liveness = append(append([]string(nil), old.Probes.Liveness...), flags...)
		image.Probes.Readiness = append(append([]string(nil), old.Probes.Readiness...), flags...)
		image.Drain = append(append([]string(nil), old.Drain...), flags...)
		images[psid] = image
	}

	plan.RuntimeInput.Images, plan.RuntimeInput.Runtime.Workloads = images, inputs

	inventory, err := Build(plan.Model, plan.Selector)
	if err != nil {
		return DevelopmentRuntimePlan{}, err
	}

	plan.inventory = inventory

	plan.runtimeValues = nil
	for _, release := range inventory.Releases {
		values, err := BuildRuntimeChartValues(plan.Model, inventory, release.Key, images,
			plan.RuntimeInput.PKIMaterializer, plan.RuntimeInput.Storage, plan.RuntimeInput.Runtime)
		if err != nil {
			return DevelopmentRuntimePlan{}, err
		}

		for _, workload := range values.Workloads {
			if workload.Pool != cryptoutilSharedMagic.DeploymentPoolMem || !strings.Contains(workload.Image.Probes.Startup[len(workload.Image.Probes.Startup)-1], cryptoutilSharedMagic.DeploymentMaterializerSecretRoleHealth) {
				return DevelopmentRuntimePlan{}, errors.New("development protected probe contract rejected")
			}
		}

		plan.runtimeValues = append(plan.runtimeValues, values)
	}

	return plan, nil
}
