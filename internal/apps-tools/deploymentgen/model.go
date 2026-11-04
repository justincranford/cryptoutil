// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"net"
	"regexp"
	"slices"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Bounds is the configured lower and upper replica capacity for a workload pool.
type Bounds struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type bounds = Bounds

// Mode selects backend pools for one deployment release.
type Mode struct {
	Name  string   `json:"name"`
	Pools []string `json:"pools"`
}

type mode = Mode

// Infrastructure describes the shared run-level dependency replica counts.
type Infrastructure struct {
	Leader    int `json:"postgresLeader"`
	Follower  int `json:"postgresFollower"`
	Collector int `json:"collector"`
	LGTM      int `json:"lgtm"`
}

// Bootstrap configures the loopback-only local K3s cluster and in-cluster address plan.
type Bootstrap struct {
	APIHost           string `json:"apiHost"`
	APIPort           int    `json:"apiPort"`
	GatewayHost       string `json:"gatewayHost"`
	GatewayPort       int    `json:"gatewayPort"`
	APIServiceIP      string `json:"apiServiceIP"`
	APIServerAudience string `json:"apiServerAudience,omitempty"`
	ServiceCIDR       string `json:"serviceCIDR"`
	PodCIDR           string `json:"podCIDR"`
	ServerNodes       int    `json:"serverNodes"`
	AgentNodes        int    `json:"agentNodes"`
}

// Model is the versioned, non-secret deployment input document.
type Model struct {
	Version        int               `json:"version"`
	Run            string            `json:"run"`
	Revision       string            `json:"revision"`
	ClusterDomain  string            `json:"clusterDomain"`
	LocalDomain    string            `json:"localDomain"`
	Replicas       bounds            `json:"replicas"`
	Overrides      map[string]bounds `json:"overrides"`
	Modes          []mode            `json:"modes"`
	Infrastructure Infrastructure    `json:"infrastructure"`
	Bootstrap      Bootstrap         `json:"bootstrap"`
}

type model = Model

const (
	tierService       = string(TierService)
	tierProduct       = string(TierProduct)
	tierSuite         = string(TierSuite)
	poolBackendSQLite = "sqlite"
)

func label(value string) bool {
	return len(value) <= cryptoutilSharedMagic.FQDNLabelMaxLength &&
		regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`).MatchString(value)
}

func dnsLabel(value string) bool {
	if len(value) == 0 || len(value) > cryptoutilSharedMagic.FQDNLabelMaxLength || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}

	for i := 0; i < len(value); i++ {
		c := value[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}

	return true
}

func dnsName(value string) bool {
	if len(value) == 0 || len(value) > cryptoutilSharedMagic.FQDNMaxLength {
		return false
	}

	for _, part := range strings.Split(value, ".") {
		if !dnsLabel(part) {
			return false
		}
	}

	return true
}

func validate(m model, scopes []Scope) error {
	if m.Version != 2 || !regexp.MustCompile(`^[a-z][a-z0-9]{0,7}$`).MatchString(m.Run) || !label(m.Revision) {
		return fmt.Errorf("invalid version, run (1-8 lowercase alphanumeric), or revision")
	}

	for _, domain := range []string{m.ClusterDomain, m.LocalDomain} {
		if len(domain) > cryptoutilSharedMagic.FQDNMaxLength {
			return fmt.Errorf("domain too long: %s", domain)
		}

		for _, part := range strings.Split(domain, ".") {
			if !dnsLabel(part) {
				return fmt.Errorf("invalid domain: %s", domain)
			}
		}
	}

	if len(scopes) == 0 || len(m.Modes) == 0 {
		return fmt.Errorf("registry and modes must be nonempty")
	}

	members := make(map[string]bool)
	seenScopes := make(map[Tier]map[string]bool)

	for _, scope := range scopes {
		if !label(scope.ID) {
			return fmt.Errorf("invalid scope ID for tier %q: %s", scope.Tier, scope.ID)
		}

		seenForTier, tierExists := seenScopes[scope.Tier]
		if !tierExists {
			seenForTier = make(map[string]bool)
			seenScopes[scope.Tier] = seenForTier
		}

		if seenForTier[scope.ID] {
			return fmt.Errorf("duplicate scope: %s:%s", scope.Tier, scope.ID)
		}

		seenForTier[scope.ID] = true

		switch scope.Tier {
		case TierService:
			if len(scope.Members) != 1 || scope.Members[0] != scope.ID || members[scope.ID] {
				return fmt.Errorf("invalid or duplicate PS-ID scope: %s", scope.ID)
			}

			members[scope.ID] = true
		case TierProduct, TierSuite:
			if len(scope.Members) == 0 {
				return fmt.Errorf("empty registry scope: %s", scope.ID)
			}
		default:
			return fmt.Errorf("unknown registry tier: %s", scope.Tier)
		}
	}

	if len(seenScopes[TierService]) == 0 || len(seenScopes[TierProduct]) == 0 || len(seenScopes[TierSuite]) == 0 {
		return fmt.Errorf("registry tiers must be nonempty")
	}

	for _, scope := range scopes {
		if scope.Tier == TierService {
			continue
		}

		for _, member := range scope.Members {
			if !members[member] {
				return fmt.Errorf("unknown PS-ID %q in registry scope %q", member, scope.ID)
			}
		}
	}

	allBounds := []bounds{m.Replicas}

	overrideKeys := make([]string, 0, len(m.Overrides))
	for key := range m.Overrides {
		overrideKeys = append(overrideKeys, key)
	}

	slices.Sort(overrideKeys)

	for _, key := range overrideKeys {
		value := m.Overrides[key]

		parts := strings.Split(key, "/")
		if len(parts) != 2 || !members[parts[0]] || !slices.Contains([]string{cryptoutilSharedMagic.DeploymentPoolPG, cryptoutilSharedMagic.DeploymentPoolMem, cryptoutilSharedMagic.DeploymentPoolDisk}, parts[1]) {
			return fmt.Errorf("unknown override (expected PSID/pool): %s", key)
		}

		allBounds = append(allBounds, value)
	}

	for _, b := range allBounds {
		if b.Min < 1 || b.Max < b.Min {
			return fmt.Errorf("replicas require 1 <= min <= max")
		}
	}

	seenModes := map[string]bool{}
	for _, v := range m.Modes {
		if !label(v.Name) || seenModes[v.Name] || len(v.Pools) == 0 {
			return fmt.Errorf("invalid, duplicate, or empty mode: %s", v.Name)
		}

		seenModes[v.Name] = true

		seenPools := map[string]bool{}
		for _, pool := range v.Pools {
			if !slices.Contains([]string{cryptoutilSharedMagic.DeploymentPoolPG, cryptoutilSharedMagic.DeploymentPoolMem, cryptoutilSharedMagic.DeploymentPoolDisk}, pool) || seenPools[pool] {
				return fmt.Errorf("unknown or duplicate pool: %s", pool)
			}

			seenPools[pool] = true
		}
	}

	if m.Infrastructure.Leader != 1 || m.Infrastructure.Follower != 1 ||
		m.Infrastructure.LGTM != 1 || m.Infrastructure.Collector < 1 {
		return fmt.Errorf("leader, logical follower, and LGTM are singletons; collector must be positive")
	}

	if err := validateBootstrap(m.Bootstrap); err != nil {
		return err
	}

	return nil
}

func validateBootstrap(bootstrap Bootstrap) error {
	apiHost := net.ParseIP(bootstrap.APIHost)
	gatewayHost := net.ParseIP(bootstrap.GatewayHost)
	apiServiceIP := net.ParseIP(bootstrap.APIServiceIP)
	_, serviceNetwork, serviceErr := net.ParseCIDR(bootstrap.ServiceCIDR)
	_, podNetwork, podErr := net.ParseCIDR(bootstrap.PodCIDR)

	if apiHost == nil || apiHost.To4() == nil || !apiHost.IsLoopback() ||
		gatewayHost == nil || gatewayHost.To4() == nil || !gatewayHost.IsLoopback() ||
		bootstrap.APIPort < 1 || bootstrap.APIPort > 65535 ||
		bootstrap.GatewayPort < 1 || bootstrap.GatewayPort > 65535 ||
		(bootstrap.APIHost == bootstrap.GatewayHost && bootstrap.APIPort == bootstrap.GatewayPort) {
		return fmt.Errorf("bootstrap API and gateway endpoints must use distinct loopback IPv4 host ports")
	}

	if serviceErr != nil || podErr != nil || serviceNetwork.IP.To4() == nil || podNetwork.IP.To4() == nil ||
		apiServiceIP == nil || apiServiceIP.To4() == nil || !serviceNetwork.Contains(apiServiceIP) ||
		serviceNetwork.Contains(podNetwork.IP) || podNetwork.Contains(serviceNetwork.IP) {
		return fmt.Errorf("bootstrap API service IP and non-overlapping IPv4 service/pod CIDRs are required")
	}

	if bootstrap.ServerNodes != 1 || bootstrap.AgentNodes < 1 || bootstrap.AgentNodes > 9 {
		return fmt.Errorf("local K3s requires one server and 1-9 agents")
	}

	if bootstrap.APIServerAudience != "" && strings.TrimSpace(bootstrap.APIServerAudience) == "" {
		return fmt.Errorf("API server audience must be omitted or nonempty")
	}

	return nil
}
