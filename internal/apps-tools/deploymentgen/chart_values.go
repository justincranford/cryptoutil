// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	cpuMilliUnitsPerCore      int64 = 1000
	binaryQuantityPrefixShift uint  = 10
	binaryQuantityBase        int64 = 1 << binaryQuantityPrefixShift
	binaryQuantityKibi              = binaryQuantityBase
	binaryQuantityMebi              = binaryQuantityKibi * binaryQuantityKibi
	binaryQuantityGibi              = binaryQuantityMebi * binaryQuantityKibi
	binaryQuantityTebi              = binaryQuantityGibi * binaryQuantityKibi
	runtimeMemorySuffixKibi         = "Ki"
	runtimeMemorySuffixMebi         = "Mi"
	runtimeMemorySuffixGibi         = "Gi"
	runtimeMemorySuffixTebi         = "Ti"
)

var (
	imageDigestPattern    = regexp.MustCompile(`^.+@sha256:[0-9a-f]{64}$`)
	storageSizePattern    = regexp.MustCompile(`^[1-9][0-9]*(Mi|Gi|Ti)$`)
	storageClassPattern   = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	configFilePattern     = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,253}$`)
	cpuResourcePattern    = regexp.MustCompile(`^[1-9][0-9]*m?$`)
	memoryResourcePattern = regexp.MustCompile(`^[1-9][0-9]*(Ki|Mi|Gi|Ti)?$`)
)

// ContainerImage is a pinned image and the commands supplied by that image.
type ContainerImage struct {
	Reference string        `json:"image"`
	Command   []string      `json:"command"`
	Args      []string      `json:"args,omitempty"`
	Drain     []string      `json:"drain,omitempty"`
	Probes    ProbeCommands `json:"probes,omitempty"`
}

// PKIMaterializerImage contains only the fields supported by the Helm init image schema.
type PKIMaterializerImage struct {
	Reference string               `json:"image"`
	Command   []string             `json:"command"`
	Args      []string             `json:"args,omitempty"`
	Resources ResourceRequirements `json:"resources"`
}

// ProbeCommands are in-container commands for the framework's loopback mTLS probes.
type ProbeCommands struct {
	Startup   []string `json:"startup"`
	Liveness  []string `json:"liveness"`
	Readiness []string `json:"readiness"`
}

// ChartStorage describes retained per-pod claim provisioning.
type ChartStorage struct {
	ClassName string `json:"className"`
	Size      string `json:"size"`
}

// RuntimeChartInputs supplies non-secret per-release workload configuration.
// Credential values must remain in Kubernetes Secrets and never appear here.
type RuntimeChartInputs struct {
	Workloads map[string]RuntimeWorkloadInput `json:"workloads"`
}

// RuntimeWorkloadInput contains backend-specific arguments and ConfigMap files.
type RuntimeWorkloadInput struct {
	MaterializerConfig  string               `json:"materializerConfig"`
	FrameworkConfigFile string               `json:"frameworkConfigFile"`
	FrameworkConfig     string               `json:"frameworkConfig"`
	DomainConfigFile    string               `json:"domainConfigFile"`
	DomainConfig        string               `json:"domainConfig"`
	Args                []string             `json:"args,omitempty"`
	Resources           ResourceRequirements `json:"resources"`
}

// RuntimeChartConfigurationMap is a generated framework or domain ConfigMap.
type RuntimeChartConfigurationMap struct {
	Name  string            `json:"name"`
	PSID  string            `json:"psid"`
	Kind  string            `json:"kind"`
	Files map[string]string `json:"files"`
}

// RuntimeAPIEdge is one registry-approved same-release caller-to-target API path.
type RuntimeAPIEdge struct {
	SourcePSID string `json:"sourcePsid"`
	TargetPSID string `json:"targetPsid"`
}

// RuntimeChartValues is the strict input contract for the canonical runtime Helm chart.
type RuntimeChartValues struct {
	Run               string                         `json:"run"`
	HelmRelease       string                         `json:"helmRelease"`
	ReleaseKey        string                         `json:"releaseKey"`
	Revision          string                         `json:"revision"`
	Namespace         string                         `json:"namespace"`
	ClusterDomain     string                         `json:"clusterDomain"`
	InfraNamespace    string                         `json:"infraNamespace"`
	APIServerIP       string                         `json:"apiServerIP"`
	APIServerAudience string                         `json:"apiServerAudience,omitempty"`
	HasPostgres       bool                           `json:"hasPostgres"`
	Quota             RuntimeChartQuota              `json:"quota"`
	PKI               PKIMaterializerImage           `json:"pkiMaterializer"`
	Storage           ChartStorage                   `json:"storage"`
	ConfigurationMaps []RuntimeChartConfigurationMap `json:"configurationMaps"`
	ApprovedAPIEdges  []RuntimeAPIEdge               `json:"approvedAPIEdges"`
	Workloads         []RuntimeChartWorkload         `json:"workloads"`
}

// RuntimeChartWorkload is one registry-derived PS-ID/backend pool.
type RuntimeChartWorkload struct {
	MaterializerConfig     string               `json:"materializerConfig"`
	Name                   string               `json:"name"`
	PSID                   string               `json:"psid"`
	Pool                   string               `json:"pool"`
	MinReplicas            int                  `json:"minReplicas"`
	MaxReplicas            int                  `json:"maxReplicas"`
	Image                  ContainerImage       `json:"container"`
	FrameworkConfigMapName string               `json:"frameworkConfigMapName"`
	FrameworkConfigFile    string               `json:"frameworkConfigFile"`
	DomainConfigMapName    string               `json:"domainConfigMapName"`
	DomainConfigFile       string               `json:"domainConfigFile"`
	Resources              ResourceRequirements `json:"resources"`
	CredentialSecretNames  []string             `json:"credentialSecretNames"`
	TrustSecretName        string               `json:"trustSecretName"`
	DatabaseSecretName     string               `json:"databaseSecretName,omitempty"`
	TLSLeafSecretNames     []string             `json:"tlsLeafSecretNames"`
	InstanceServices       []InstanceService    `json:"instanceServices"`
}

// InstanceService maps a stable ordinal-specific name to its StatefulSet pod.
type InstanceService struct {
	Name                   string `json:"name"`
	PodName                string `json:"podName"`
	TLSLeafSecretName      string `json:"tlsLeafSecretName"`
	HealthClientSecretName string `json:"healthClientSecretName"`
	OTLPClientSecretName   string `json:"otlpClientSecretName"`
}

// BuildRuntimeChartValues projects one independently selected release into the Helm values
// contract. It requires digest-pinned images and explicit app health commands; the graph
// remains useful without images, but cannot be deployed without them.
func BuildRuntimeChartValues(
	model Model,
	inventory Inventory,
	releaseKey string,
	images map[string]ContainerImage,
	pkiMaterializer PKIMaterializerImage,
	storage ChartStorage,
	inputs RuntimeChartInputs,
) (RuntimeChartValues, error) {
	if !imageDigestPattern.MatchString(pkiMaterializer.Reference) ||
		!slices.Equal(pkiMaterializer.Command, []string{materializerSuiteExecutable}) ||
		!slices.Equal(pkiMaterializer.Args, []string{"pki-materialize", "--instance=/etc/cryptoutil/instance/instance.yaml", "--input=/etc/cryptoutil/instance/materializer.json"}) {
		return RuntimeChartValues{}, fmt.Errorf("PKI materializer requires a digest-pinned image and command")
	}

	if err := validateRuntimeResources(pkiMaterializer.Resources); err != nil {
		return RuntimeChartValues{}, fmt.Errorf("PKI materializer has invalid resources: %w", err)
	}

	if !storageClassPattern.MatchString(storage.ClassName) || len(storage.ClassName) > 63 ||
		!storageSizePattern.MatchString(storage.Size) {
		return RuntimeChartValues{}, fmt.Errorf("retained storage class and size must be valid Kubernetes values")
	}

	if err := validateBootstrap(model.Bootstrap); err != nil {
		return RuntimeChartValues{}, err
	}

	if inventory.Run != model.Run || inventory.Version != model.Version {
		return RuntimeChartValues{}, fmt.Errorf("model and inventory identities differ")
	}

	if err := validateInventory(inventory); err != nil {
		return RuntimeChartValues{}, fmt.Errorf("invalid inventory: %w", err)
	}

	var selected *Release

	for index := range inventory.Releases {
		if inventory.Releases[index].Key == releaseKey {
			selected = &inventory.Releases[index]

			break
		}
	}

	if selected == nil {
		return RuntimeChartValues{}, fmt.Errorf("release not found in inventory: %s", releaseKey)
	}

	result := RuntimeChartValues{
		Run: model.Run, HelmRelease: selected.Helm, ReleaseKey: selected.Key, Revision: model.Revision, Namespace: selected.Namespace,
		ClusterDomain: model.ClusterDomain, InfraNamespace: inventory.InfraNamespace, APIServerIP: model.Bootstrap.APIServiceIP,
		HasPostgres:       len(selected.Databases) > 0,
		APIServerAudience: model.Bootstrap.APIServerAudience,
		PKI:               pkiMaterializer, Storage: storage,
		ConfigurationMaps: make([]RuntimeChartConfigurationMap, 0, len(selected.Workloads)*2),
		ApprovedAPIEdges:  make([]RuntimeAPIEdge, 0),
		Workloads:         make([]RuntimeChartWorkload, 0, len(selected.Workloads)),
	}

	releaseMembers := make([]string, 0, len(selected.Workloads))

	releaseMemberSet := make(map[string]bool, len(selected.Workloads))
	for _, workload := range selected.Workloads {
		if !releaseMemberSet[workload.PSID] {
			releaseMembers = append(releaseMembers, workload.PSID)
			releaseMemberSet[workload.PSID] = true
		}
	}

	for _, dependency := range registeredAPIDependenciesForMembers(releaseMembers) {
		result.ApprovedAPIEdges = append(result.ApprovedAPIEdges, RuntimeAPIEdge(dependency))
	}

	configurationMaps := make(map[string]RuntimeChartConfigurationMap, len(selected.Workloads)*2)

	for _, workload := range selected.Workloads {
		image, ok := images[workload.PSID]
		if !ok {
			return RuntimeChartValues{}, fmt.Errorf("image missing for PS-ID %s", workload.PSID)
		}

		if err := validateContainerImage(workload.PSID, image); err != nil {
			return RuntimeChartValues{}, err
		}

		inputKey := runtimeWorkloadInputKey(selected.Key, workload.PSID, workload.Pool)

		input, ok := inputs.Workloads[inputKey]
		if !ok {
			return RuntimeChartValues{}, fmt.Errorf("runtime workload configuration missing: %s", inputKey)
		}

		if err := validateRuntimeWorkloadInput(workload, input); err != nil {
			return RuntimeChartValues{}, err
		}

		frameworkConfigMapName := workload.PSID + "-framework"
		domainConfigMapName := workload.PSID + "-domain"
		frameworkConfigArg := "--config=/etc/cryptoutil/framework/" + input.FrameworkConfigFile

		domainConfigArg := "--config=/etc/cryptoutil/domain/" + input.DomainConfigFile
		if slices.Contains(image.Args, frameworkConfigArg) || slices.Contains(image.Args, domainConfigArg) ||
			slices.Contains(input.Args, frameworkConfigArg) || slices.Contains(input.Args, domainConfigArg) {
			return RuntimeChartValues{}, fmt.Errorf("runtime workload %s must not duplicate generated config arguments", workload.Name)
		}

		if err := addRuntimeConfigurationFile(
			configurationMaps, workload.PSID, cryptoutilSharedMagic.FrameworkProductName,
			input.FrameworkConfigFile, input.FrameworkConfig,
		); err != nil {
			return RuntimeChartValues{}, err
		}

		if err := addRuntimeConfigurationFile(
			configurationMaps, workload.PSID, "domain", input.DomainConfigFile, input.DomainConfig,
		); err != nil {
			return RuntimeChartValues{}, err
		}

		args := slices.Clone(image.Args)
		args = append(args, input.Args...)
		args = append(args, frameworkConfigArg, domainConfigArg)
		image.Args = args

		value := RuntimeChartWorkload{
			MaterializerConfig: input.MaterializerConfig,
			Name:               workload.Name, PSID: workload.PSID, Pool: workload.Pool,
			MinReplicas: workload.Bounds.Min, MaxReplicas: workload.Bounds.Max,
			Image:                  image,
			FrameworkConfigMapName: frameworkConfigMapName, FrameworkConfigFile: input.FrameworkConfigFile,
			DomainConfigMapName: domainConfigMapName, DomainConfigFile: input.DomainConfigFile,
			Resources:             input.Resources,
			CredentialSecretNames: []string{lifecycleUnsealSecret, lifecyclePepperSecret, workload.PSID + "-credentials"},
			TrustSecretName:       workload.PSID + "-workload-trust",
			TLSLeafSecretNames:    make([]string, 0, len(workload.Pods)),
			InstanceServices:      make([]InstanceService, 0, len(workload.Pods)),
		}

		if workload.Pool == cryptoutilSharedMagic.DeploymentPoolPG {
			value.DatabaseSecretName = workload.PSID + "-db"
		}

		for _, pod := range workload.Pods {
			value.TLSLeafSecretNames = append(value.TLSLeafSecretNames, pod.TLS)
			value.InstanceServices = append(value.InstanceServices, InstanceService{
				Name: pod.Service, PodName: pod.Name, TLSLeafSecretName: pod.TLS,
				HealthClientSecretName: pod.Name + "-health-client",
				OTLPClientSecretName:   pod.Name + "-otlp-client",
			})
		}

		if err := validateMaterializerProjection(result, value); err != nil {
			return RuntimeChartValues{}, err
		}

		result.Workloads = append(result.Workloads, value)
	}

	quota, err := buildRuntimeChartQuota(result.Workloads, pkiMaterializer)
	if err != nil {
		return RuntimeChartValues{}, err
	}

	result.Quota = quota

	for _, configurationMap := range configurationMaps {
		result.ConfigurationMaps = append(result.ConfigurationMaps, configurationMap)
	}

	slices.SortFunc(result.ConfigurationMaps, func(left, right RuntimeChartConfigurationMap) int {
		return strings.Compare(left.Name, right.Name)
	})

	return result, nil
}

// MarshalRuntimeChartValues encodes chart values as canonical JSON (valid YAML) with a final LF.
func MarshalRuntimeChartValues(values RuntimeChartValues) ([]byte, error) {
	return marshalCanonicalJSON(values)
}

// RuntimeChartQuota is the exact maximum-capacity budget for one runtime release. It accounts
// for max replicas and Kubernetes' max(app containers, sequential init container) pod request.
type RuntimeChartQuota struct {
	Pods           string `json:"pods"`
	RequestsCPU    string `json:"requestsCPU"`
	RequestsMemory string `json:"requestsMemory"`
	LimitsCPU      string `json:"limitsCPU"`
	LimitsMemory   string `json:"limitsMemory"`
}

func buildRuntimeChartQuota(
	workloads []RuntimeChartWorkload,
	pkiMaterializer PKIMaterializerImage,
) (RuntimeChartQuota, error) {
	initRequests, initLimits, err := parseResourceRequirements(pkiMaterializer.Resources)
	if err != nil {
		return RuntimeChartQuota{}, fmt.Errorf("invalid PKI init resources for runtime quota: %w", err)
	}

	var (
		pods     int
		requests ResourceCapacity
		limits   ResourceCapacity
	)

	for _, workload := range workloads {
		if workload.MaxReplicas < 1 || workload.MaxReplicas > maxExpandedPods-pods {
			return RuntimeChartQuota{}, fmt.Errorf("runtime quota pod count exceeds limit of %d", maxExpandedPods)
		}

		appRequests, appLimits, err := parseResourceRequirements(workload.Resources)
		if err != nil {
			return RuntimeChartQuota{}, fmt.Errorf("invalid runtime quota resources for %s: %w", workload.Name, err)
		}

		podRequests := ResourceCapacity{
			CPUMilli:    max(appRequests.CPUMilli, initRequests.CPUMilli),
			MemoryBytes: max(appRequests.MemoryBytes, initRequests.MemoryBytes),
		}
		podLimits := ResourceCapacity{
			CPUMilli:    max(appLimits.CPUMilli, initLimits.CPUMilli),
			MemoryBytes: max(appLimits.MemoryBytes, initLimits.MemoryBytes),
		}

		if err := addResourceCapacity(&requests, podRequests, workload.MaxReplicas); err != nil {
			return RuntimeChartQuota{}, fmt.Errorf("sum runtime quota requests: %w", err)
		}

		if err := addResourceCapacity(&limits, podLimits, workload.MaxReplicas); err != nil {
			return RuntimeChartQuota{}, fmt.Errorf("sum runtime quota limits: %w", err)
		}

		pods += workload.MaxReplicas
	}

	if pods == 0 {
		return RuntimeChartQuota{}, fmt.Errorf("runtime quota requires at least one workload")
	}

	return RuntimeChartQuota{
		Pods:           strconv.Itoa(pods),
		RequestsCPU:    formatRuntimeCPUMilli(requests.CPUMilli),
		RequestsMemory: formatRuntimeMemoryBytes(requests.MemoryBytes),
		LimitsCPU:      formatRuntimeCPUMilli(limits.CPUMilli),
		LimitsMemory:   formatRuntimeMemoryBytes(limits.MemoryBytes),
	}, nil
}

func formatRuntimeCPUMilli(value int64) string {
	return fmt.Sprintf("%dm", value)
}

func formatRuntimeMemoryBytes(value int64) string {
	for _, unit := range []struct {
		suffix string
		bytes  int64
	}{
		{suffix: runtimeMemorySuffixTebi, bytes: binaryQuantityTebi},
		{suffix: runtimeMemorySuffixGibi, bytes: binaryQuantityGibi},
		{suffix: runtimeMemorySuffixMebi, bytes: binaryQuantityMebi},
		{suffix: runtimeMemorySuffixKibi, bytes: binaryQuantityKibi},
	} {
		if value%unit.bytes == 0 {
			return strconv.FormatInt(value/unit.bytes, 10) + unit.suffix
		}
	}

	return strconv.FormatInt(value, 10)
}

func validateContainerImage(name string, image ContainerImage) error {
	if !imageDigestPattern.MatchString(image.Reference) || strings.ContainsAny(image.Reference, " \t\r\n") {
		return fmt.Errorf("%s image must include a lowercase sha256 digest", name)
	}

	if len(image.Command) == 0 || len(image.Drain) == 0 || len(image.Probes.Startup) == 0 ||
		len(image.Probes.Liveness) == 0 || len(image.Probes.Readiness) == 0 {
		return fmt.Errorf("%s image requires command, drain, and startup/liveness/readiness probe commands", name)
	}

	return nil
}

func runtimeWorkloadInputKey(releaseKey, psid, pool string) string {
	return strings.Join([]string{releaseKey, psid, pool}, "/")
}

func validateRuntimeWorkloadInput(workload Workload, input RuntimeWorkloadInput) error {
	if !configFilePattern.MatchString(input.FrameworkConfigFile) || !configFilePattern.MatchString(input.DomainConfigFile) {
		return fmt.Errorf("runtime workload %s config file names must be safe Kubernetes data keys", workload.Name)
	}

	if err := validateRuntimeResources(input.Resources); err != nil {
		return fmt.Errorf("runtime workload %s has invalid resources: %w", workload.Name, err)
	}

	return nil
}

func validateRuntimeResources(resources ResourceRequirements) error {
	_, _, err := parseResourceRequirements(resources)

	return err
}

func addRuntimeConfigurationFile(
	configurationMaps map[string]RuntimeChartConfigurationMap,
	psid, kind, fileName, content string,
) error {
	name := psid + "-" + kind

	configurationMap, ok := configurationMaps[name]
	if !ok {
		configurationMap = RuntimeChartConfigurationMap{
			Name: name, PSID: psid, Kind: kind, Files: make(map[string]string),
		}
	}

	if existingContent, exists := configurationMap.Files[fileName]; exists && existingContent != content {
		return fmt.Errorf("runtime ConfigMap %s contains conflicting file %s", name, fileName)
	}

	configurationMap.Files[fileName] = content
	configurationMaps[name] = configurationMap

	return nil
}
