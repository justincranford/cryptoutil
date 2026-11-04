// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const infrastructureValuesYAMLDecodeError = "decode infrastructure values YAML"

// InfrastructureChartValues is the typed input contract for the run-scoped infrastructure chart.
type InfrastructureChartValues struct {
	Run           string                     `json:"run"`
	HelmRelease   string                     `json:"helmRelease"`
	ReleaseKey    string                     `json:"releaseKey"`
	Revision      string                     `json:"revision"`
	Namespace     string                     `json:"namespace"`
	ClusterDomain string                     `json:"clusterDomain"`
	APIServerIP   string                     `json:"apiServerIP"`
	Storage       InfrastructureChartStorage `json:"storage"`
	Components    InfrastructureComponents   `json:"components"`
	Clients       []InfrastructureClient     `json:"clients"`
	Databases     []Database                 `json:"databases"`
	Quota         InfrastructureQuota        `json:"quota"`
}

// InfrastructureChartStorage configures the retain-on-delete storage class for infrastructure.
type InfrastructureChartStorage struct {
	ClassName string `json:"className"`
}

// InfrastructureComponents contains the fixed infrastructure component identities.
type InfrastructureComponents struct {
	PostgresLeader   InfrastructureComponent `json:"postgresLeader"`
	PostgresFollower InfrastructureComponent `json:"postgresFollower"`
	Collector        InfrastructureComponent `json:"collector"`
	LGTM             InfrastructureComponent `json:"lgtm"`
}

// InfrastructureComponent is the strict typed shape accepted by the infrastructure chart.
type InfrastructureComponent struct {
	Enabled              bool                           `json:"enabled"`
	Image                string                         `json:"image"`
	Replicas             int                            `json:"replicas"`
	ConfigMapName        string                         `json:"configMapName"`
	ConfigMountPath      string                         `json:"configMountPath"`
	ConfigFiles          map[string]string              `json:"configFiles"`
	ContainerArgs        []string                       `json:"containerArgs"`
	ServicePorts         []InfrastructureServicePort    `json:"servicePorts"`
	Probes               InfrastructureProbeSet         `json:"probes"`
	Storage              InfrastructureComponentStorage `json:"storage"`
	SecretMounts         []InfrastructureSecretMount    `json:"secretMounts"`
	Scratch              []InfrastructureScratch        `json:"scratch"`
	RunAsUser            int                            `json:"runAsUser"`
	RunAsGroup           int                            `json:"runAsGroup"`
	FSGroup              int                            `json:"fsGroup"`
	Resources            ResourceRequirements           `json:"resources"`
	BootstrapSecretName  string                         `json:"bootstrapSecretName,omitempty"`
	BootstrapUsernameKey string                         `json:"bootstrapUsernameKey,omitempty"`
	BootstrapPasswordKey string                         `json:"bootstrapPasswordKey,omitempty"`
	AdminSecretName      string                         `json:"adminSecretName,omitempty"`
	AdminPasswordKey     string                         `json:"adminPasswordKey,omitempty"`
}

// InfrastructureServicePort describes one component service and container port.
type InfrastructureServicePort struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
}

// InfrastructureProbeSet configures bounded startup, liveness, and readiness probes.
type InfrastructureProbeSet struct {
	Startup   InfrastructureProbe `json:"startup"`
	Liveness  InfrastructureProbe `json:"liveness"`
	Readiness InfrastructureProbe `json:"readiness"`
}

// InfrastructureProbe is an HTTP or TCP kubelet probe definition.
type InfrastructureProbe struct {
	Type             string `json:"type"`
	Path             string `json:"path,omitempty"`
	Scheme           string `json:"scheme,omitempty"`
	Port             int    `json:"port"`
	PeriodSeconds    int    `json:"periodSeconds"`
	TimeoutSeconds   int    `json:"timeoutSeconds"`
	FailureThreshold int    `json:"failureThreshold"`
}

// InfrastructureComponentStorage is one component's retained data mount and requested size.
type InfrastructureComponentStorage struct {
	Size      string `json:"size"`
	MountPath string `json:"mountPath"`
}

// InfrastructureSecretMount references one externally materialized, read-only Secret.
type InfrastructureSecretMount struct {
	SecretName  string `json:"secretName"`
	MountPath   string `json:"mountPath"`
	DefaultMode int    `json:"defaultMode"`
}

// InfrastructureScratch is a bounded in-memory scratch mount.
type InfrastructureScratch struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SizeLimit string `json:"sizeLimit"`
}

// ResourceRequirements are container CPU and memory requests/limits.
type ResourceRequirements struct {
	Requests ResourceSet `json:"requests"`
	Limits   ResourceSet `json:"limits"`
}

// ResourceSet contains CPU and memory quantities.
type ResourceSet struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// InfrastructureClient describes one release namespace allowed to reach shared infrastructure.
type InfrastructureClient struct {
	Namespace string `json:"namespace"`
	Release   string `json:"release"`
	Postgres  bool   `json:"postgres"`
}

// InfrastructureQuota bounds resources in the run-scoped infrastructure namespace.
type InfrastructureQuota struct {
	Pods           string `json:"pods"`
	RequestsCPU    string `json:"requestsCPU"`
	RequestsMemory string `json:"requestsMemory"`
	LimitsCPU      string `json:"limitsCPU"`
	LimitsMemory   string `json:"limitsMemory"`
}

// DecodeInfrastructureChartValues decodes canonical Helm YAML values into the typed chart contract.
func DecodeInfrastructureChartValues(raw []byte) (InfrastructureChartValues, error) {
	var yamlValue any
	if err := yaml.Unmarshal(raw, &yamlValue); err != nil {
		return InfrastructureChartValues{}, fmt.Errorf("%s: %w", infrastructureValuesYAMLDecodeError, err)
	}

	encoded, err := json.Marshal(yamlValue)
	if err != nil {
		return InfrastructureChartValues{}, fmt.Errorf("encode infrastructure values JSON: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()

	var values InfrastructureChartValues
	if err := decoder.Decode(&values); err != nil {
		return InfrastructureChartValues{}, fmt.Errorf("decode typed infrastructure values: %w", err)
	}

	return values, nil
}

// BuildInfrastructureChartValues binds canonical component settings to one independent inventory.
func BuildInfrastructureChartValues(
	model Model,
	inventory Inventory,
	defaults InfrastructureChartValues,
) (InfrastructureChartValues, error) {
	if inventory.Run != model.Run || inventory.Version != model.Version {
		return InfrastructureChartValues{}, fmt.Errorf("model and inventory identities differ")
	}

	if err := validateBootstrap(model.Bootstrap); err != nil {
		return InfrastructureChartValues{}, err
	}

	if err := validateInventory(inventory); err != nil {
		return InfrastructureChartValues{}, fmt.Errorf("invalid inventory: %w", err)
	}

	if defaults.Storage.ClassName == "" {
		return InfrastructureChartValues{}, fmt.Errorf("infrastructure storage class is required")
	}

	if model.Infrastructure.Collector != defaults.Components.Collector.Replicas {
		return InfrastructureChartValues{}, fmt.Errorf(
			"model collector replicas %d do not match canonical infrastructure chart replicas %d",
			model.Infrastructure.Collector, defaults.Components.Collector.Replicas,
		)
	}

	result := defaults
	result.Run = model.Run
	result.HelmRelease = "cu-" + model.Run + "-infra"
	result.ReleaseKey = "infra:" + model.Run
	result.Revision = model.Revision
	result.Namespace = inventory.InfraNamespace
	result.ClusterDomain = model.ClusterDomain
	result.APIServerIP = model.Bootstrap.APIServiceIP

	hasPostgres := false
	clients := make([]InfrastructureClient, 0, len(inventory.Releases))
	databases := make([]Database, 0)

	for _, release := range inventory.Releases {
		postgres := len(release.Databases) > 0
		hasPostgres = hasPostgres || postgres

		databases = append(databases, release.Databases...)
		clients = append(clients, InfrastructureClient{
			Namespace: release.Namespace,
			Release:   release.Helm,
			Postgres:  postgres,
		})
	}

	sort.Slice(clients, func(left, right int) bool {
		leftKey := clients[left].Namespace + "/" + clients[left].Release
		rightKey := clients[right].Namespace + "/" + clients[right].Release

		return leftKey < rightKey
	})
	result.Clients = clients

	sort.Slice(databases, func(left, right int) bool {
		return databases[left].Name < databases[right].Name
	})
	result.Databases = databases

	graphHasLeader := inventoryContainsObject(inventory.Infra, "StatefulSet", cryptoutilSharedMagic.PKIInitPostgresLeaderService)
	graphHasFollower := inventoryContainsObject(inventory.Infra, "StatefulSet", cryptoutilSharedMagic.PKIInitPostgresFollowerService)
	graphHasCollector := inventoryContainsObject(inventory.Infra, "StatefulSet", "collector")

	graphHasLGTM := inventoryContainsObject(inventory.Infra, "StatefulSet", "lgtm")
	if graphHasLeader != hasPostgres || graphHasFollower != hasPostgres || !graphHasCollector || !graphHasLGTM {
		return InfrastructureChartValues{}, fmt.Errorf("infrastructure inventory and selected release dependencies disagree")
	}

	result.Components.PostgresLeader.Enabled = hasPostgres
	result.Components.PostgresFollower.Enabled = hasPostgres
	result.Components.Collector.Enabled = true
	result.Components.LGTM.Enabled = true
	result.Components.PostgresLeader.Replicas = model.Infrastructure.Leader
	result.Components.PostgresFollower.Replicas = model.Infrastructure.Follower
	result.Components.Collector.Replicas = model.Infrastructure.Collector
	result.Components.LGTM.Replicas = model.Infrastructure.LGTM

	if countInventoryPodsWithPrefix(inventory.InfraPods, "lgtm-") != model.Infrastructure.LGTM ||
		countInventoryPodsWithPrefix(inventory.InfraPods, "collector-") != model.Infrastructure.Collector {
		return InfrastructureChartValues{}, fmt.Errorf("infrastructure pod identities disagree with model component replicas")
	}

	leaderPods := countInventoryPodsWithPrefix(inventory.InfraPods, "postgres-leader-")

	followerPods := countInventoryPodsWithPrefix(inventory.InfraPods, "postgres-follower-")
	if (hasPostgres && (leaderPods != model.Infrastructure.Leader || followerPods != model.Infrastructure.Follower)) ||
		(!hasPostgres && (leaderPods != 0 || followerPods != 0)) {
		return InfrastructureChartValues{}, fmt.Errorf("PostgreSQL pod identities disagree with model component replicas")
	}

	return result, nil
}

// MarshalInfrastructureChartValues encodes chart values as canonical JSON (valid YAML) with a final LF.
func MarshalInfrastructureChartValues(values InfrastructureChartValues) ([]byte, error) {
	return marshalCanonicalJSON(values)
}

func inventoryContainsObject(objects []Object, kind, name string) bool {
	return slices.Contains(objects, Object{Kind: kind, Name: name})
}

func countInventoryPodsWithPrefix(pods []Object, prefix string) int {
	count := 0

	for _, pod := range pods {
		if pod.Kind == "Pod" && strings.HasPrefix(pod.Name, prefix) {
			count++
		}
	}

	return count
}
