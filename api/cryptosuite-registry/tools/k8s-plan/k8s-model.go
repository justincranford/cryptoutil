//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	yaml "gopkg.in/yaml.v3"
)

type bounds struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type mode struct {
	Name  string   `json:"name"`
	Pools []string `json:"pools"`
}

type model struct {
	Version        int               `json:"version"`
	Run            string            `json:"run"`
	Revision       string            `json:"revision"`
	ClusterDomain  string            `json:"clusterDomain"`
	LocalDomain    string            `json:"localDomain"`
	Replicas       bounds            `json:"replicas"`
	Overrides      map[string]bounds `json:"overrides"`
	Modes          []mode            `json:"modes"`
	Infrastructure struct {
		Leader    int `json:"postgresLeader"`
		Follower  int `json:"postgresFollower"`
		Collector int `json:"collector"`
		LGTM      int `json:"lgtm"`
	} `json:"infrastructure"`
}

const (
	tierService       = "s"
	tierProduct       = "p"
	tierSuite         = "u"
	poolPG            = "pg"
	poolMem           = "mem"
	poolDisk          = "disk"
	poolBackendSQLite = "sqlite"
)

type suite struct {
	ID          string `yaml:"id"`
	DisplayName string `yaml:"display_name"`
	CmdDir      string `yaml:"cmd_dir"`
	InitPSID    string `yaml:"init_ps_id"`
}

type product struct {
	ID              string `yaml:"id"`
	DisplayName     string `yaml:"display_name"`
	InternalAppsDir string `yaml:"internal_apps_dir"`
	CmdDir          string `yaml:"cmd_dir"`
	InitPSID        string `yaml:"init_ps_id"`
}

type member struct {
	PSID                string         `yaml:"ps_id"`
	Product             string         `yaml:"product"`
	Service             string         `yaml:"service"`
	DisplayName         string         `yaml:"display_name"`
	InternalAppsDir     string         `yaml:"internal_apps_dir"`
	MagicFile           string         `yaml:"magic_file"`
	BasePort            int            `yaml:"base_port"`
	Entrypoint          []string       `yaml:"entrypoint"`
	PGHostPort          int            `yaml:"pg_host_port"`
	MigrationRangeStart int            `yaml:"migration_range_start"`
	MigrationRangeEnd   int            `yaml:"migration_range_end"`
	GoTemplateParams    templateParams `yaml:"go_template_params"`
	Realms              []realm        `yaml:"realms"`
	APIResources        []string       `yaml:"api_resources"`
}

type registry struct {
	Suites         []suite         `yaml:"suites"`
	Products       []product       `yaml:"products"`
	Members        []member        `yaml:"product_services"`
	SharedServices []sharedService `yaml:"shared_services"`
	InfraTools     []infraTool     `yaml:"infra_tools"`
}

type templateParams struct {
	UsagePrefix             string `yaml:"usage_prefix"`
	ProductNameConst        string `yaml:"product_name_const"`
	ServiceNameConst        string `yaml:"service_name_const"`
	ServiceIDConst          string `yaml:"service_id_const"`
	ServicePortConst        string `yaml:"service_port_const"`
	ServiceDisplayNameConst string `yaml:"service_display_name_const"`
}

type realm struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Type     string `yaml:"type"`
}

type sharedService struct {
	ID                string `yaml:"id"`
	DisplayName       string `yaml:"display_name"`
	LeaderPort        int    `yaml:"leader_port"`
	FollowerPort      int    `yaml:"follower_port"`
	LeaderDatabases   int    `yaml:"leader_databases"`
	FollowerDatabases int    `yaml:"follower_databases"`
	Replication       string `yaml:"replication"`
}

type infraTool struct {
	ID                string `yaml:"id"`
	DisplayName       string `yaml:"display_name"`
	CollectorGRPCPort int    `yaml:"collector_grpc_port"`
	CollectorHTTPPort int    `yaml:"collector_http_port"`
	GrafanaPort       int    `yaml:"grafana_port"`
}

func decodeRegistry(data []byte) (registry, error) {
	var r registry

	shapeDecoder := yaml.NewDecoder(bytes.NewReader(data))

	var document yaml.Node
	if err := shapeDecoder.Decode(&document); err != nil {
		return r, fmt.Errorf("registry: %w", err)
	}

	if err := requireSingleYAMLDocument(shapeDecoder); err != nil {
		return r, fmt.Errorf("registry: %w", err)
	}

	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return r, fmt.Errorf("registry: expected exactly one non-null YAML mapping document")
	}

	if key, duplicate := duplicateYAMLMappingKey(document.Content[0]); duplicate {
		return r, fmt.Errorf("registry: duplicate YAML mapping key: %q", key)
	}

	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)

	if err := d.Decode(&r); err != nil {
		return r, fmt.Errorf("registry: %w", err)
	}

	slices.SortFunc(r.Members, func(a, b member) int { return strings.Compare(a.PSID, b.PSID) })
	slices.SortFunc(r.Products, func(a, b product) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(r.Suites, func(a, b suite) int { return strings.Compare(a.ID, b.ID) })

	return r, nil
}

func requireSingleYAMLDocument(d *yaml.Decoder) error {
	err := d.Decode(new(yaml.Node))
	if errors.Is(err, io.EOF) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("expected exactly one YAML document: %w", err)
	}

	return fmt.Errorf("expected exactly one YAML document")
}

func duplicateYAMLMappingKey(node *yaml.Node) (string, bool) {
	if node.Kind == yaml.MappingNode {
		seen := map[string]struct{}{}

		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind == yaml.ScalarNode {
				identity := key.Tag + "\x00" + key.Value
				if _, duplicate := seen[identity]; duplicate {
					return key.Value, true
				}

				seen[identity] = struct{}{}
			}
		}
	}

	for _, child := range node.Content {
		if key, duplicate := duplicateYAMLMappingKey(child); duplicate {
			return key, true
		}
	}

	return "", false
}

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

func validate(m model, r registry) error {
	if m.Version != 1 || !regexp.MustCompile(`^[a-z][a-z0-9]{0,7}$`).MatchString(m.Run) || !label(m.Revision) {
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

	if len(r.Suites) == 0 || len(r.Products) == 0 || len(r.Members) == 0 || len(m.Modes) == 0 {
		return fmt.Errorf("registry and modes must be nonempty")
	}

	products := map[string]bool{}
	for _, p := range r.Products {
		if !label(p.ID) || products[p.ID] {
			return fmt.Errorf("invalid or duplicate product: %s", p.ID)
		}

		products[p.ID] = true
	}

	suites := map[string]bool{}
	for _, s := range r.Suites {
		if !label(s.ID) || suites[s.ID] {
			return fmt.Errorf("invalid or duplicate suite: %s", s.ID)
		}

		suites[s.ID] = true
	}

	members := map[string]bool{}
	for _, s := range r.Members {
		if !label(s.PSID) || members[s.PSID] || !products[s.Product] {
			return fmt.Errorf("invalid, duplicate, or orphan PS-ID: %s", s.PSID)
		}

		members[s.PSID] = true
	}

	productIDs := make([]string, 0, len(products))
	for product := range products {
		productIDs = append(productIDs, product)
	}

	slices.Sort(productIDs)

	for _, product := range productIDs {
		if !slices.ContainsFunc(r.Members, func(s member) bool { return s.Product == product }) {
			return fmt.Errorf("empty product: %s", product)
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
		if len(parts) != 2 || !members[parts[0]] || !slices.Contains([]string{poolPG, poolMem, poolDisk}, parts[1]) {
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
			if !slices.Contains([]string{poolPG, poolMem, poolDisk}, pool) || seenPools[pool] {
				return fmt.Errorf("unknown or duplicate pool: %s", pool)
			}

			seenPools[pool] = true
		}
	}

	if m.Infrastructure.Leader != 1 || m.Infrastructure.Follower != 1 ||
		m.Infrastructure.LGTM != 1 || m.Infrastructure.Collector < 1 {
		return fmt.Errorf("leader, logical follower, and LGTM are singletons; collector must be positive")
	}

	return nil
}
