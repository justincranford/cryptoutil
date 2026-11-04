// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"text/template"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const (
	k8sToolchainLockFileName = "k8s-toolchain.lock.json"
	localK3sComposeTemplate  = "api/cryptosuite-registry/templates/kubernetes/local-k3s/compose.yml.tmpl"
)

var (
	k3sVersionPattern  = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+-k3s[0-9]+$`)
	kubeVersionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	k3sRunPattern      = regexp.MustCompile(`^[a-z][a-z0-9]{0,7}$`)
	sha256Pattern      = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ToolchainLock contains the immutable versions and image references used by local K3s.
type ToolchainLock struct {
	SchemaVersion int    `json:"schemaVersion"`
	Platform      string `json:"platform"`
	Kubernetes    struct {
		K3S struct {
			Version           string `json:"version"`
			Image             string `json:"image"`
			OCIManifestDigest string `json:"ociManifestDigest"`
		} `json:"k3s"`
		Kubectl struct {
			Version            string `json:"version"`
			LinuxAMD64SHA256   string `json:"linuxAmd64Sha256"`
			WindowsAMD64SHA256 string `json:"windowsAmd64Sha256"`
		} `json:"kubectl"`
		Helm struct {
			Version               string `json:"version"`
			LinuxAMD64TarGZSHA256 string `json:"linuxAmd64TarGzSha256"`
			WindowsAMD64ZipSHA256 string `json:"windowsAmd64ZipSha256"`
		} `json:"helm"`
	} `json:"kubernetes"`
	Images struct {
		Postgres struct {
			Reference string `json:"reference"`
		} `json:"postgres"`
		OTelCollector struct {
			Reference string `json:"reference"`
		} `json:"otelCollector"`
		LGTM struct {
			Reference string `json:"reference"`
		} `json:"lgtm"`
	} `json:"images"`
}

// LocalK3sNode identifies one explicitly named Compose-hosted cluster node.
type LocalK3sNode struct {
	Service   string
	Container string
	Hostname  string
	Volume    string
	Server    bool
}

// LocalK3sPlan is a non-secret deterministic input to the canonical Compose template.
type LocalK3sPlan struct {
	Run             string
	Project         string
	Network         string
	Image           string
	APIHost         string
	APIPort         int
	GatewayHost     string
	GatewayPort     int
	APIServiceIP    string
	ServiceCIDR     string
	PodCIDR         string
	TokenSecretFile string
	Server          LocalK3sNode
	Agents          []LocalK3sNode
}

// DecodeToolchainLock strictly decodes and validates the checked-in toolchain pin document.
func DecodeToolchainLock(data []byte) (ToolchainLock, error) {
	var lock ToolchainLock
	if err := validateJSONDocument(data); err != nil {
		return lock, fmt.Errorf("toolchain lock: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&lock); err != nil {
		return lock, fmt.Errorf("toolchain lock: %w", err)
	}

	if err := validateToolchainLock(lock); err != nil {
		return lock, err
	}

	return lock, nil
}

func validateToolchainLock(lock ToolchainLock) error {
	if lock.SchemaVersion != 1 || lock.Platform != "linux/amd64" ||
		!k3sVersionPattern.MatchString(lock.Kubernetes.K3S.Version) ||
		!kubeVersionPattern.MatchString(lock.Kubernetes.Kubectl.Version) ||
		!kubeVersionPattern.MatchString(lock.Kubernetes.Helm.Version) {
		return fmt.Errorf("toolchain lock has invalid schema version, platform, or Kubernetes tool version")
	}

	for _, digest := range []string{
		lock.Kubernetes.K3S.OCIManifestDigest,
		lock.Kubernetes.Kubectl.LinuxAMD64SHA256,
		lock.Kubernetes.Kubectl.WindowsAMD64SHA256,
		lock.Kubernetes.Helm.LinuxAMD64TarGZSHA256,
		lock.Kubernetes.Helm.WindowsAMD64ZipSHA256,
	} {
		if !sha256Pattern.MatchString(strings.TrimPrefix(digest, "sha256:")) {
			return fmt.Errorf("toolchain lock contains invalid sha256 digest")
		}
	}

	images := []string{
		lock.Kubernetes.K3S.Image, lock.Images.Postgres.Reference,
		lock.Images.OTelCollector.Reference, lock.Images.LGTM.Reference,
	}
	for _, image := range images {
		if !imageDigestPattern.MatchString(image) || strings.ContainsAny(image, " \t\r\n") {
			return fmt.Errorf("toolchain lock image reference must be digest-pinned")
		}
	}

	return nil
}

// BuildLocalK3sPlan derives unique server/agent identities and loopback endpoints.
func BuildLocalK3sPlan(model Model, lock ToolchainLock) (LocalK3sPlan, error) {
	if !k3sRunPattern.MatchString(model.Run) {
		return LocalK3sPlan{}, fmt.Errorf("local K3s run identity must be 1-8 lowercase alphanumeric characters")
	}

	if err := validateBootstrap(model.Bootstrap); err != nil {
		return LocalK3sPlan{}, err
	}

	if err := validateToolchainLock(lock); err != nil {
		return LocalK3sPlan{}, err
	}

	project := "cu-" + model.Run + "-cluster"
	plan := LocalK3sPlan{
		Run: model.Run, Project: project, Network: "cu-" + model.Run + "-nodes", Image: lock.Kubernetes.K3S.Image,
		APIHost: model.Bootstrap.APIHost, APIPort: model.Bootstrap.APIPort,
		GatewayHost: model.Bootstrap.GatewayHost, GatewayPort: model.Bootstrap.GatewayPort,
		APIServiceIP: model.Bootstrap.APIServiceIP, ServiceCIDR: model.Bootstrap.ServiceCIDR, PodCIDR: model.Bootstrap.PodCIDR,
		TokenSecretFile: fmt.Sprintf(cryptoutilSharedMagic.K3SBootstrapTokenSecretFileFormat, model.Run),
		Agents:          make([]LocalK3sNode, 0, model.Bootstrap.AgentNodes),
	}
	plan.Server = localK3sNode(plan, cryptoutilSharedMagic.CMD_SERVER, 0, true)

	for index := range model.Bootstrap.AgentNodes {
		plan.Agents = append(plan.Agents, localK3sNode(plan, "agent", index, false))
	}

	return plan, nil
}

func localK3sNode(plan LocalK3sPlan, role string, ordinal int, server bool) LocalK3sNode {
	id := fmt.Sprintf("%s-%s-%d", plan.Project, role, ordinal)

	return LocalK3sNode{
		Service: fmt.Sprintf("%s-%d", role, ordinal), Container: id, Hostname: id, Volume: id + "-data", Server: server,
	}
}

// RenderLocalK3sCompose executes the canonical template and validates its Compose topology.
func RenderLocalK3sCompose(source []byte, plan LocalK3sPlan) ([]byte, error) {
	if len(source) == 0 {
		return nil, fmt.Errorf("local K3s Compose template is empty")
	}

	renderer, err := template.New("local-k3s-compose").Option("missingkey=error").Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse local K3s Compose template: %w", err)
	}

	var output bytes.Buffer
	if err := renderer.Execute(&output, plan); err != nil {
		return nil, fmt.Errorf("render local K3s Compose template: %w", err)
	}

	rendered := output.Bytes()
	if !bytes.HasSuffix(rendered, []byte{'\n'}) || bytes.Contains(rendered, []byte{'\r'}) {
		return nil, fmt.Errorf("rendered local K3s Compose must use LF and end with LF")
	}

	if err := validateLocalK3sCompose(rendered, plan); err != nil {
		return nil, err
	}

	return append([]byte(nil), rendered...), nil
}

func validateLocalK3sCompose(data []byte, plan LocalK3sPlan) error {
	var document struct {
		Name     string `yaml:"name"`
		Services map[string]struct {
			ContainerName string            `yaml:"container_name"`
			Hostname      string            `yaml:"hostname"`
			Image         string            `yaml:"image"`
			Command       []string          `yaml:"command"`
			Environment   map[string]string `yaml:"environment"`
			Secrets       []struct {
				Source string `yaml:"source"`
				Target string `yaml:"target"`
			} `yaml:"secrets"`
			Ports []struct {
				Target    int    `yaml:"target"`
				Published string `yaml:"published"`
				HostIP    string `yaml:"host_ip"`
			} `yaml:"ports"`
		} `yaml:"services"`
		Secrets map[string]struct {
			File string `yaml:"file"`
		} `yaml:"secrets"`
	}

	if err := yaml.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("parse rendered local K3s Compose YAML: %w", err)
	}

	if document.Name != plan.Project || len(document.Services) != 1+len(plan.Agents) {
		return fmt.Errorf("rendered Compose project or node count does not match the bootstrap plan")
	}

	joinTokenSecret, exists := document.Secrets[cryptoutilSharedMagic.K3SBootstrapTokenSecretName]
	if len(document.Secrets) != 1 || !exists ||
		joinTokenSecret.File != plan.TokenSecretFile {
		return fmt.Errorf("rendered Compose must use the protected relative join-token Docker Secret file")
	}

	nodes := append([]LocalK3sNode{plan.Server}, plan.Agents...)
	for _, node := range nodes {
		service, exists := document.Services[node.Service]
		if !exists || service.ContainerName != node.Container || service.Hostname != node.Hostname ||
			service.Image != plan.Image || len(service.Command) == 0 {
			return fmt.Errorf("rendered Compose node does not match plan: %s", node.Service)
		}

		if len(service.Secrets) != 1 ||
			service.Secrets[0].Source != cryptoutilSharedMagic.K3SBootstrapTokenSecretName ||
			service.Secrets[0].Target != cryptoutilSharedMagic.K3SBootstrapTokenSecretTarget {
			return fmt.Errorf("rendered Compose node must mount the join-token Docker Secret read-only at the protected path: %s", node.Service)
		}

		if _, exists := service.Environment[cryptoutilSharedMagic.K3SBootstrapTokenEnvironment]; exists {
			return fmt.Errorf("rendered Compose node must not expose the join token through its environment: %s", node.Service)
		}

		if _, exists := service.Environment[cryptoutilSharedMagic.K3SBootstrapURLEnvironment]; exists {
			return fmt.Errorf("rendered Compose agent URL must use the CLI argument, not its environment: %s", node.Service)
		}

		for _, argument := range service.Command {
			if argument == cryptoutilSharedMagic.K3SBootstrapTokenArgument ||
				strings.HasPrefix(argument, cryptoutilSharedMagic.K3SBootstrapTokenAssignmentPrefix) {
				return fmt.Errorf("rendered Compose node must not embed a join token in its command: %s", node.Service)
			}
		}

		if node.Server {
			if service.Command[0] != cryptoutilSharedMagic.CMD_SERVER || len(service.Ports) != 2 ||
				service.Ports[0].Target != 6443 || service.Ports[0].Published != fmt.Sprint(plan.APIPort) ||
				service.Ports[0].HostIP != plan.APIHost || service.Ports[1].Target != 30443 ||
				service.Ports[1].Published != fmt.Sprint(plan.GatewayPort) ||
				service.Ports[1].HostIP != plan.GatewayHost {
				return fmt.Errorf("rendered server API/gateway ports do not match loopback plan")
			}

			if !containsCommandArgument(service.Command, cryptoutilSharedMagic.K3SBootstrapTokenFileArgument) {
				return fmt.Errorf("rendered server must use the protected token-file CLI argument: %s", node.Service)
			}

			if !containsCommandArgument(service.Command, cryptoutilSharedMagic.K3SBootstrapServerEncryptionArgument) ||
				!containsCommandArgument(service.Command, cryptoutilSharedMagic.K3SBootstrapServerEncryptionProviderArgument) {
				return fmt.Errorf("rendered server must enable K3s Secret encryption with aescbc: %s", node.Service)
			}
		} else {
			if service.Command[0] != "agent" || len(service.Ports) != 0 {
				return fmt.Errorf("rendered agent must be a non-published K3s agent: %s", node.Service)
			}

			if !containsCommandArgument(service.Command, cryptoutilSharedMagic.K3SBootstrapAgentServerArgument) ||
				!containsCommandArgument(service.Command, cryptoutilSharedMagic.K3SBootstrapTokenFileArgument) {
				return fmt.Errorf("rendered agent must use the explicit server URL and protected token file CLI arguments: %s", node.Service)
			}
		}
	}

	return nil
}

func containsCommandArgument(arguments []string, expected string) bool {
	for _, argument := range arguments {
		if argument == expected {
			return true
		}
	}

	return false
}

// CheckBootstrapPorts confirms both host endpoints are bindable and releases every probe socket.
func CheckBootstrapPorts(
	bootstrap Bootstrap,
	listen func(string, string) (net.Listener, error),
) error {
	if listen == nil {
		return fmt.Errorf("bootstrap host-port listener is required")
	}

	if err := validateBootstrap(bootstrap); err != nil {
		return err
	}

	endpoints := []string{
		net.JoinHostPort(bootstrap.APIHost, fmt.Sprint(bootstrap.APIPort)),
		net.JoinHostPort(bootstrap.GatewayHost, fmt.Sprint(bootstrap.GatewayPort)),
	}
	listeners := make([]net.Listener, 0, len(endpoints))

	var resultErr error

	for _, endpoint := range endpoints {
		listener, err := listen("tcp", endpoint)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("bootstrap host endpoint unavailable %s: %w", endpoint, err))

			break
		}

		if listener == nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("bootstrap host-port listener returned nil listener for %s", endpoint))

			break
		}

		listeners = append(listeners, listener)
	}

	for _, listener := range listeners {
		if err := listener.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close bootstrap port probe: %w", err))
		}
	}

	return resultErr
}
