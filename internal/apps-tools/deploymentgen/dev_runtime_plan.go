// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	sha256 "crypto/sha256"
	json "encoding/json"
	"fmt"
	"strings"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	developmentCAKey                   = "ca.crt"
	developmentLGTMAdmin               = "lgtm-admin"
	developmentAPIServerName           = "kubernetes.default.svc"
	lifecycleUsernameKey               = "username"
	lifecyclePostgresFollowerBootstrap = "postgres-follower-bootstrap"
	developmentJobDeadlineSeconds      = 300
)

// DevelopmentRuntimePlan is non-secret preparation, not image/runtime certification.
// It never reads or copies old credentials, databases, certificates or unseal state.
type DevelopmentRuntimePlan struct {
	Model         Model            `json:"model"`
	Selector      string           `json:"selector"`
	RuntimeInput  ChartValuesInput `json:"runtimeInput"`
	Lifecycle     LifecycleInput   `json:"lifecycle"`
	Cluster       LocalK3sPlan     `json:"cluster"`
	inventory     Inventory
	runtimeValues []RuntimeChartValues
}

// BuildDevelopmentRuntimePlan derives disposable identities and exact runnable
// suite argv from the existing graph. Image must be an actual inspected suite
// digest; this function validates its syntax, not its provenance or availability.
// All required protected files are references only, never invented material.
func BuildDevelopmentRuntimePlan(model Model, lock ToolchainLock, run, selector, image string) (DevelopmentRuntimePlan, error) {
	result := DevelopmentRuntimePlan{}
	model.Run = run

	cluster, err := BuildLocalK3sPlan(model, lock)
	if err != nil {
		return result, err
	}

	inventory, err := Build(model, selector)
	if err != nil {
		return result, err
	}

	return buildDevelopmentRuntimePlanForInventory(model, cluster, inventory, selector, image)
}

// Keep graph-to-contract assembly separate so drift/missing-source failures are
// testable at the actual controlling boundary, not by bypassing host admission.
func buildDevelopmentRuntimePlanForInventory(model Model, cluster LocalK3sPlan, inventory Inventory, selector, image string) (DevelopmentRuntimePlan, error) {
	result := DevelopmentRuntimePlan{}

	identities, err := BuildLifecycleInventory(inventory)
	if err != nil {
		return result, err
	}

	resources := ResourceRequirements{Requests: ResourceSet{CPU: "100m", Memory: "128Mi"}, Limits: ResourceSet{CPU: "1", Memory: "512Mi"}}
	jobResources := ResourceRequirements{Requests: ResourceSet{CPU: "50m", Memory: "64Mi"}, Limits: ResourceSet{CPU: "500m", Memory: "256Mi"}}
	input := ChartValuesInput{
		Images: map[string]ContainerImage{},
		PKIMaterializer: PKIMaterializerImage{
			Reference: image, Command: []string{pki.SuiteExecutable},
			Args: []string{pki.Command, "--instance=" + pki.InstancePath, "--input=" + pki.InputPath}, Resources: jobResources,
		},
		Storage: ChartStorage{ClassName: "cu-" + model.Run + "-retain", Size: "1Gi"},
		Runtime: RuntimeChartInputs{Workloads: map[string]RuntimeWorkloadInput{}},
	}
	lifecycle := LifecycleInput{Version: 1, Secrets: []LifecycleSecretSource{}, Jobs: []LifecycleJobInput{}}
	verificationSources := map[string]string{}

	for _, id := range identities.Identities {
		if id.Kind != cryptoutilSharedMagic.KubernetesSecretKind {
			continue
		}

		keys := []string{"value"}

		switch {
		case strings.HasSuffix(id.Name, "-tls"), strings.HasSuffix(id.Name, "-health-client"), strings.HasSuffix(id.Name, "-otlp-client"):
			keys = []string{pki.SecretDataCertificate, pki.SecretDataPrivateKey}
		case strings.HasSuffix(id.Name, "-ca"), strings.HasSuffix(id.Name, "-workload-trust"):
			keys = []string{developmentCAKey}
		case strings.HasSuffix(id.Name, "-credentials"):
			keys = []string{"browser-username.secret", "browser-password.secret", "service-username.secret", "service-password.secret"}
		case strings.HasSuffix(id.Name, "-migration-db"), id.Name == lifecyclePostgresBootstrap, id.Name == lifecyclePostgresFollowerBootstrap:
			keys = []string{lifecycleUsernameKey, lifecyclePasswordKey, developmentCAKey, pki.SecretDataCertificate, pki.SecretDataPrivateKey, db.InputName}
		case id.Name == developmentLGTMAdmin:
			keys = []string{lifecyclePasswordKey}
		case id.Name == "unseal":
			keys = []string{"unseal-1of5.secret", "unseal-2of5.secret", "unseal-3of5.secret", "unseal-4of5.secret", "unseal-5of5.secret"}
		}

		if strings.HasSuffix(id.Name, "-workload-trust") || (id.Namespace == inventory.InfraNamespace && id.Name == developmentCollectorTLS) {
			if verificationSources[id.Namespace] == "" {
				verificationSources[id.Namespace] = id.Name

				keys = append(keys, pki.VerificationInputName)
			}
		}

		files := map[string]string{}
		for _, key := range keys {
			files[key] = id.Namespace + "/" + id.Name + "/" + key
		}

		lifecycle.Secrets = append(lifecycle.Secrets, LifecycleSecretSource{Namespace: id.Namespace, Name: id.Name, Type: cryptoutilSharedMagic.KubernetesSecretTypeOpaque, Files: files})
	}

	for _, release := range inventory.Releases {
		for _, workload := range release.Workloads {
			parts := strings.SplitN(workload.PSID, "-", 2)
			route := parts
			command := []string{pki.SuiteExecutable}
			probe := func(name string) []string {
				return append(append([]string{}, command...), append(append([]string{}, route...), name)...)
			}
			input.Images[workload.PSID] = ContainerImage{
				Reference: image, Command: command, Args: append(append([]string{}, route...), cryptoutilSharedMagic.CMD_SERVER), Drain: probe("shutdown"),
				Probes: ProbeCommands{Startup: probe("startupz"), Liveness: probe("livez"), Readiness: probe("readyz")},
			}

			materializer := pki.CommandInput{
				Namespace: release.Namespace, APIEndpoint: "https://" + model.Bootstrap.APIServiceIP + ":443", APIServerName: developmentAPIServerName,
				TrustFiles: map[string]string{}, Pods: []pki.CommandPod{},
			}
			for _, role := range []string{"public-server", "admin-server", cryptoutilSharedMagic.DeploymentMaterializerSecretRoleHealth, cryptoutilSharedMagic.DeploymentMaterializerSecretRoleOTLP} {
				materializer.TrustFiles[role] = pki.TrustDirectory + "/ca.crt"
			}

			if workload.Pool == cryptoutilSharedMagic.DeploymentPoolPG {
				materializer.TrustFiles["postgres-client"] = pki.TrustDirectory + "/ca.crt"
			}

			for _, pod := range workload.Pods {
				podDNS := pod.Name + "." + workload.Name + "-pods." + release.Namespace + ".svc." + model.ClusterDomain

				binding := pki.CommandPod{
					PodName: pod.Name, ExpectedAdminDNS: podDNS,
					ExpectedPublicDNSNames: []string{
						podDNS, pod.Service + "." + release.Namespace + ".svc." + model.ClusterDomain,
						workload.Name + "." + release.Namespace + ".svc." + model.ClusterDomain,
					},
				}
				if workload.Pool == cryptoutilSharedMagic.DeploymentPoolPG {
					binding.PostgreSQLClientBindings = map[string]pki.PostgreSQLClientBinding{}
					role := ""

					for _, database := range release.Databases {
						if database.PSID == workload.PSID {
							role = database.Role
						}
					}

					for _, other := range workload.Pods {
						identity := sha256.Sum256([]byte(release.Namespace + "/" + other.Name))
						binding.PostgreSQLClientBindings[other.Name] = pki.PostgreSQLClientBinding{CommonName: fmt.Sprintf("cu-%x", identity[:16]), DatabaseRole: role}
					}
				}

				materializer.Pods = append(materializer.Pods, binding)

				for index := range lifecycle.Secrets {
					source := &lifecycle.Secrets[index]
					if source.Name == pod.TLS && source.Namespace == release.Namespace {
						source.Files = map[string]string{}
						for _, key := range []string{pki.SecretDataPublicCertificate, pki.SecretDataPublicPrivateKey, pki.SecretDataAdminCertificate, pki.SecretDataAdminPrivateKey} {
							source.Files[key] = release.Namespace + "/" + pod.TLS + "/" + key
						}

						if workload.Pool == cryptoutilSharedMagic.DeploymentPoolPG {
							for _, key := range []string{pki.SecretDataPostgreSQLCertificate, pki.SecretDataPostgreSQLPrivateKey} {
								source.Files[key] = release.Namespace + "/" + pod.TLS + "/" + key
							}
						}
					}
				}
			}

			raw, err := json.Marshal(materializer)
			if err != nil {
				return result, fmt.Errorf("encode development materializer: %w", err)
			}

			databaseURL := "sqlite://file::memory:?cache=shared"

			switch workload.Pool {
			case cryptoutilSharedMagic.DeploymentPoolDisk:
				databaseURL = "sqlite://file:/var/lib/cryptoutil/data/app.db"
			case cryptoutilSharedMagic.DeploymentPoolPG:
				// PG credentials and verified TLS are supplied through the exact
				// protected database manifest, not embedded plaintext in this plan.
				databaseURL = ""
			}

			framework := "bind-public-address: 0.0.0.0\n"
			if databaseURL != "" {
				framework += "database-url: " + databaseURL + "\n"
			}

			input.Runtime.Workloads[runtimeWorkloadInputKey(release.Key, workload.PSID, workload.Pool)] = RuntimeWorkloadInput{
				MaterializerConfig: string(raw), FrameworkConfigFile: workload.Name + "-framework.yml", FrameworkConfig: framework,
				DomainConfigFile: workload.Name + "-domain.yml", DomainConfig: "{}\n", Resources: resources,
			}
		}
	}

	for _, id := range identities.Identities {
		if id.Kind != "Job" {
			continue
		}

		name := strings.TrimSuffix(id.Name, "-"+model.Revision)

		args, source, err := db.JobArguments(name)

		secretNames := []string{source}
		if name == "infra-pki" || name == cryptoutilSharedMagic.PSIDPKIInit {
			source = verificationSources[id.Namespace]
			if source == "" {
				return result, fmt.Errorf("development PKI verification source absent")
			}

			args = []string{pki.VerifyCommand, "--input=/run/cryptoutil/secrets/" + source + "/" + pki.VerificationInputName}
			secretNames = []string{}

			for _, secret := range lifecycle.Secrets {
				if secret.Namespace == id.Namespace && (strings.HasSuffix(secret.Name, "-tls") || strings.HasSuffix(secret.Name, "-health-client") ||
					strings.HasSuffix(secret.Name, "-otlp-client") || strings.HasSuffix(secret.Name, "-workload-trust") ||
					strings.HasSuffix(secret.Name, "-ca")) {
					secretNames = append(secretNames, secret.Name)
				}
			}
		} else if err != nil {
			return result, fmt.Errorf("development database Job contract: %w", err)
		}

		lifecycle.Jobs = append(lifecycle.Jobs, LifecycleJobInput{
			Namespace: id.Namespace, Name: name, Image: image, Command: []string{pki.SuiteExecutable}, Args: args, SecretNames: secretNames,
			WriteSecretNames: []string{}, ConfigMapNames: []string{}, ActiveDeadlineSeconds: developmentJobDeadlineSeconds, BackoffLimit: 0, Resources: jobResources,
		})
	}

	model.Revision, err = LifecycleRevision(lifecycle)
	if err != nil {
		return result, err
	}

	inventory, err = Build(model, selector)
	if err != nil {
		return result, err
	}

	if _, err := BuildLifecycleChartValues(model, inventory, lifecycle); err != nil {
		return result, err
	}

	values := make([]RuntimeChartValues, 0, len(inventory.Releases))
	for _, release := range inventory.Releases {
		value, err := BuildRuntimeChartValues(model, inventory, release.Key, input.Images, input.PKIMaterializer, input.Storage, input.Runtime)
		if err != nil {
			return result, err
		}

		values = append(values, value)
	}

	return DevelopmentRuntimePlan{Model: model, Selector: selector, RuntimeInput: input, Lifecycle: lifecycle, Cluster: cluster, inventory: inventory, runtimeValues: values}, nil
}
