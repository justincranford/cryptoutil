// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	sha256 "crypto/sha256"
	json "encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// LifecycleInput contains only operator-owned image pins, execution policy, and
// protected file references. It never contains Secret values.
type LifecycleInput struct {
	Version int                     `json:"version"`
	Secrets []LifecycleSecretSource `json:"secrets"`
	Jobs    []LifecycleJobInput     `json:"jobs"`
}

// LifecycleSecretSource imports a complete Secret from protected relative files.
// There is deliberately no ExternalSecrets operator or Helm-managed Secret data.
type LifecycleSecretSource struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Files     map[string]string `json:"files"`
}

// LifecycleJobInput is explicit because the registry does not pin a deployable
// suite image or define Kubernetes PKI/database adapter commands.
// Name is the graph Job's stable base, without its revision suffix.
type LifecycleJobInput struct {
	Namespace             string               `json:"namespace"`
	Name                  string               `json:"name"`
	Image                 string               `json:"image"`
	Command               []string             `json:"command"`
	Args                  []string             `json:"args"`
	Resources             ResourceRequirements `json:"resources"`
	ActiveDeadlineSeconds int                  `json:"activeDeadlineSeconds"`
	BackoffLimit          int                  `json:"backoffLimit"`
	SecretNames           []string             `json:"secretNames"`
	WriteSecretNames      []string             `json:"writeSecretNames"`
	ConfigMapNames        []string             `json:"configMapNames"`
}

// LifecycleNamespace is the exact get/update-only bootstrap permission boundary.
type LifecycleNamespace struct {
	Name             string   `json:"name"`
	Release          string   `json:"release"`
	ReadSecretNames  []string `json:"readSecretNames"`
	WriteSecretNames []string `json:"writeSecretNames"`
}

// LifecycleJobValues adds graph ownership and an explicit PKI-only API token.
type LifecycleJobValues struct {
	LifecycleJobInput
	Release  string `json:"release"`
	APIToken bool   `json:"apiToken"`
}

// LifecycleChartValues renders Jobs and restricted bootstrap RBAC, not Secret
// material. Import Secrets before submitting Jobs; an orchestrator must order
// PKI, database initialization, migration, and replication, never Helm hooks.
type LifecycleChartValues struct {
	Run         string                  `json:"run"`
	Revision    string                  `json:"revision"`
	APIAudience string                  `json:"apiAudience"`
	Namespaces  []LifecycleNamespace    `json:"namespaces"`
	Jobs        []LifecycleJobValues    `json:"jobs"`
	Secrets     []LifecycleSecretSource `json:"externalSecrets"`
}

// LifecycleRevision hashes the whole non-secret operator contract. Changing an
// immutable Job specification requires a new graph revision, not a Job patch.
func LifecycleRevision(input LifecycleInput) (string, error) {
	raw, err := marshalCanonicalJSON(input)
	if err != nil {
		return "", fmt.Errorf("encode lifecycle revision: %w", err)
	}

	sum := sha256.Sum256(raw)

	return fmt.Sprintf("r%x", sum[:8]), nil
}

// DecodeLifecycleInput rejects duplicate fields, unknown fields, and case drift.
func DecodeLifecycleInput(raw []byte) (LifecycleInput, error) {
	var input LifecycleInput
	if err := validateJSONDocument(raw); err != nil {
		return input, fmt.Errorf("lifecycle input: %w", err)
	}

	if err := json.Unmarshal(raw, &input); err != nil {
		return input, fmt.Errorf("decode lifecycle input: %w", err)
	}

	// Compare the incoming field structure with the exact typed encoding. This
	// rejects unknown/case-insensitive keys at every nesting level.
	typed, err := json.Marshal(input)
	if err != nil {
		return input, fmt.Errorf("encode typed lifecycle input: %w", err)
	}

	var original, normalized any
	if err := json.Unmarshal(raw, &original); err != nil {
		return input, fmt.Errorf("decode lifecycle fields: %w", err)
	}

	if err := json.Unmarshal(typed, &normalized); err != nil {
		return input, fmt.Errorf("decode typed lifecycle fields: %w", err)
	}

	if !lifecycleFieldsMatch(original, normalized) {
		return input, fmt.Errorf("lifecycle input has unknown, missing, or case-mismatched fields")
	}

	return input, nil
}

func lifecycleFieldsMatch(original, normalized any) bool {
	switch originalValue := original.(type) {
	case map[string]any:
		normalizedValue, ok := normalized.(map[string]any)
		if !ok || len(originalValue) != len(normalizedValue) {
			return false
		}

		for key, value := range originalValue {
			other, exists := normalizedValue[key]
			if !exists || !lifecycleFieldsMatch(value, other) {
				return false
			}
		}
	case []any:
		normalizedValue, ok := normalized.([]any)
		if !ok || len(originalValue) != len(normalizedValue) {
			return false
		}

		for index, value := range originalValue {
			if !lifecycleFieldsMatch(value, normalizedValue[index]) {
				return false
			}
		}
	default:
		switch normalized.(type) {
		case map[string]any, []any:
			return false
		default:
			return original == normalized
		}
	}

	return true
}

// BuildLifecycleChartValues requires exact graph coverage and fails closed on
// missing image pins, file bindings, Job resources, deadlines, and permissions.
func BuildLifecycleChartValues(model Model, inventory Inventory, input LifecycleInput) (LifecycleChartValues, error) {
	result := LifecycleChartValues{}
	if inventory.Run != model.Run || inventory.Version != model.Version {
		return result, fmt.Errorf("model and inventory identities differ")
	}

	identities, err := BuildLifecycleInventory(inventory)
	if err != nil {
		return result, err
	}

	revision, err := LifecycleRevision(input)
	if err != nil {
		return result, err
	}

	if input.Version != 1 || model.Revision != revision {
		return result, fmt.Errorf("lifecycle requires version 1 and content-addressed model revision %s", revision)
	}

	if err := validateBootstrap(model.Bootstrap); err != nil {
		return result, err
	}

	result.Run, result.Revision = model.Run, revision

	result.APIAudience = model.Bootstrap.APIServerAudience
	if result.APIAudience == "" {
		result.APIAudience = "https://kubernetes.default.svc"
	}

	releases := map[string]string{inventory.InfraNamespace: "cu-" + model.Run + "-infra"}
	configs := make(map[string]bool)
	addConfigs := func(namespace string, objects []Object) {
		for _, object := range objects {
			if object.Kind == "ConfigMap" {
				configs[namespace+"/"+object.Name] = true
			}
		}
	}
	addConfigs(inventory.InfraNamespace, inventory.Infra)

	for _, release := range inventory.Releases {
		releases[release.Namespace] = release.Helm
		addConfigs(release.Namespace, release.Objects)

		for _, workload := range release.Workloads {
			addConfigs(release.Namespace, workload.Objects)
		}
	}

	secrets, jobs := make(map[string]bool), make(map[string]bool)

	for _, identity := range identities.Identities {
		key := identity.Namespace + "/" + identity.Name

		switch identity.Kind {
		case cryptoutilSharedMagic.KubernetesSecretKind:
			secrets[key] = true
		case lifecycleJobKind:
			jobs[key] = true
		}
	}

	if len(input.Secrets) != len(secrets) || len(input.Jobs) != len(jobs) {
		return result, fmt.Errorf("lifecycle inputs must cover every graph Secret and Job exactly")
	}

	for _, secret := range input.Secrets {
		key := secret.Namespace + "/" + secret.Name
		if !secrets[key] {
			return LifecycleChartValues{}, fmt.Errorf("unknown or duplicate external Secret: %s", key)
		}

		secrets[key] = false

		if err := validateLifecycleSecretSource(secret); err != nil {
			return LifecycleChartValues{}, err
		}

		result.Secrets = append(result.Secrets, secret)
	}

	for _, job := range input.Jobs {
		key := job.Namespace + "/" + job.Name + "-" + revision
		if !jobs[key] {
			return LifecycleChartValues{}, fmt.Errorf("unknown or duplicate lifecycle Job: %s", key)
		}

		jobs[key] = false
		pki := strings.HasPrefix(job.Name, lifecyclePKICommand) || job.Name == lifecycleInfraPKIName

		if err := validateLifecycleJob(job, secrets, configs, pki); err != nil {
			return LifecycleChartValues{}, err
		}

		if !pki {
			if err := validateDatabaseJobCommand(job, result.Secrets); err != nil {
				return LifecycleChartValues{}, err
			}
		}

		if pki && job.Args[0] == lifecycleVerifyCommand {
			if err := validatePKIVerificationCommand(job, result.Secrets); err != nil {
				return result, err
			}
		}

		job.Name += "-" + revision
		result.Jobs = append(result.Jobs, LifecycleJobValues{
			LifecycleJobInput: job, Release: releases[job.Namespace], APIToken: pki && job.Args[0] == lifecyclePKICommand,
		})
	}

	for namespace, release := range releases {
		boundary := LifecycleNamespace{
			Name: namespace, Release: release,
			ReadSecretNames: []string{}, WriteSecretNames: []string{},
		}

		for _, job := range result.Jobs {
			if job.Namespace != namespace || !job.APIToken {
				continue
			}

			boundary.ReadSecretNames = append(boundary.ReadSecretNames, job.SecretNames...)
			boundary.WriteSecretNames = append(boundary.WriteSecretNames, job.WriteSecretNames...)
		}

		slices.Sort(boundary.ReadSecretNames)
		slices.Sort(boundary.WriteSecretNames)
		boundary.ReadSecretNames = slices.Compact(boundary.ReadSecretNames)
		boundary.WriteSecretNames = slices.Compact(boundary.WriteSecretNames)
		result.Namespaces = append(result.Namespaces, boundary)
	}

	slices.SortFunc(result.Namespaces, func(a, b LifecycleNamespace) int { return strings.Compare(a.Name, b.Name) })
	slices.SortFunc(result.Jobs, func(a, b LifecycleJobValues) int {
		return strings.Compare(a.Namespace+"/"+a.Name, b.Namespace+"/"+b.Name)
	})
	slices.SortFunc(result.Secrets, func(a, b LifecycleSecretSource) int {
		return strings.Compare(a.Namespace+"/"+a.Name, b.Namespace+"/"+b.Name)
	})

	return result, nil
}

func validateLifecycleSecretSource(secret LifecycleSecretSource) error {
	if (secret.Type != lifecycleOpaqueSecretType && secret.Type != lifecycleTLSSecretType) || len(secret.Files) == 0 {
		return fmt.Errorf("external Secret %s requires a supported type and protected files", secret.Name)
	}

	if secret.Type == lifecycleTLSSecretType && (secret.Files[cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate] == "" || secret.Files[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey] == "") {
		return fmt.Errorf("TLS Secret %s requires tls.crt and tls.key", secret.Name)
	}

	requiredKeys := []string{}

	switch {
	case strings.HasSuffix(secret.Name, "-tls") && secret.Type == lifecycleOpaqueSecretType && secret.Files[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate] != "":
		requiredKeys = []string{cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey, cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate, cryptoutilSharedMagic.AdminPrivateKeyFileName}
	case strings.HasSuffix(secret.Name, "-tls"), strings.HasSuffix(secret.Name, "-health-client"), strings.HasSuffix(secret.Name, "-otlp-client"):
		requiredKeys = []string{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey}
	case strings.HasSuffix(secret.Name, "-ca"):
		requiredKeys = []string{"ca.crt"}
	case secret.Name == lifecyclePostgresBootstrap || secret.Name == lifecyclePostgresFollowerBootstrap:
		requiredKeys = []string{lifecycleUsernameKey, lifecyclePasswordKey}
	case secret.Name == "lgtm-admin":
		requiredKeys = []string{lifecyclePasswordKey}
	}

	for _, key := range requiredKeys {
		if secret.Files[key] == "" {
			return fmt.Errorf("external Secret %s requires consumer key %s", secret.Name, key)
		}
	}

	for key, file := range secret.Files {
		if !regexp.MustCompile(`^[a-zA-Z0-9._-]{1,253}$`).MatchString(key) ||
			!fs.ValidPath(file) || strings.ContainsAny(file, `:\`) || file == lifecycleRootPath {
			return fmt.Errorf("external Secret %s requires safe data keys and relative file paths", secret.Name)
		}
	}

	return nil
}

func validateLifecycleJob(job LifecycleJobInput, secrets, configs map[string]bool, pki bool) error {
	if !imageDigestPattern.MatchString(job.Image) || !slices.Equal(job.Command, []string{cryptoutilSharedMagic.KubernetesSuiteImageBinary}) ||
		len(job.Args) == 0 || job.ActiveDeadlineSeconds < 1 || job.ActiveDeadlineSeconds > lifecycleMaxDeadlineSeconds ||
		job.BackoffLimit < 0 || job.BackoffLimit > lifecycleMaxBackoffLimit {
		return fmt.Errorf("job %s requires a pinned suite image, explicit command/args and bounded execution", job.Name)
	}

	for _, arg := range job.Args {
		if strings.TrimSpace(arg) == "" {
			return fmt.Errorf("job %s has an empty argument", job.Name)
		}
	}

	if err := validateRuntimeResources(job.Resources); err != nil {
		return fmt.Errorf("job %s resources: %w", job.Name, err)
	}

	if !pki && len(job.WriteSecretNames) != 0 {
		return fmt.Errorf("database Job %s must not have Secret write permissions", job.Name)
	}

	if pki && job.Args[0] != lifecyclePKICommand && job.Args[0] != lifecycleVerifyCommand {
		return fmt.Errorf("pki job %s must use the suite pki-init command", job.Name)
	}

	if pki && job.Args[0] == lifecycleVerifyCommand && (len(job.Args) != 2 || len(job.WriteSecretNames) != 0) {
		return fmt.Errorf("verify-only PKI Job requires a protected manifest and no write permissions")
	}

	if !pki && job.Args[0] == lifecyclePKICommand {
		return fmt.Errorf("database Job %s must not invoke PKI generation", job.Name)
	}

	for _, names := range [][]string{job.SecretNames, job.WriteSecretNames, job.ConfigMapNames} {
		if len(names) != len(slices.Compact(slices.Sorted(slices.Values(names)))) {
			return fmt.Errorf("job %s has duplicate resource references", job.Name)
		}
	}

	for _, name := range append(append([]string(nil), job.SecretNames...), job.WriteSecretNames...) {
		if _, exists := secrets[job.Namespace+"/"+name]; !exists {
			return fmt.Errorf("job %s references an unowned Secret %s", job.Name, name)
		}

		if !pki && (name == lifecycleAuthoritySecret || name == lifecycleUnsealSecret || name == lifecyclePepperSecret ||
			strings.HasSuffix(name, "-e2e-tls") || strings.HasSuffix(name, "-credentials")) {
			return fmt.Errorf("database Job %s cannot consume signing or application credentials", job.Name)
		}

		if !pki && strings.HasSuffix(job.Name, "-migrate") && name != strings.TrimSuffix(job.Name, "-migrate")+"-migration-db" {
			return fmt.Errorf("migration Job %s may consume only its member migration Secret", job.Name)
		}
	}

	for _, name := range job.WriteSecretNames {
		if name == lifecycleAuthoritySecret || name == lifecycleUnsealSecret || name == lifecyclePepperSecret ||
			strings.HasSuffix(name, "-credentials") || strings.HasSuffix(name, "-db") ||
			strings.HasSuffix(name, "-bootstrap") || strings.HasSuffix(name, "-admin") {
			return fmt.Errorf("pki job %s cannot overwrite imported authorities or credentials", job.Name)
		}
	}

	for _, name := range job.ConfigMapNames {
		if !configs[job.Namespace+"/"+name] {
			return fmt.Errorf("job %s references an unowned ConfigMap %s", job.Name, name)
		}
	}

	return nil
}
