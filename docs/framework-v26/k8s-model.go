//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

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

type entity struct {
	ID string `yaml:"id"`
}

type member struct {
	PSID    string `yaml:"ps_id"`
	Product string `yaml:"product"`
}

type registry struct {
	Suites   []entity `yaml:"suites"`
	Products []entity `yaml:"products"`
	Members  []member `yaml:"product_services"`
}

func decodeModel(data []byte) (model, error) {
	var m model
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()

	if err := d.Decode(&m); err != nil {
		return m, fmt.Errorf("model: %w", err)
	}

	if err := d.Decode(new(any)); err != io.EOF {
		return m, fmt.Errorf("model: expected exactly one JSON document")
	}

	return m, nil
}

func decodeRegistry(data []byte) (registry, error) {
	var r registry
	if err := yaml.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("registry: %w", err)
	}

	slices.SortFunc(r.Members, func(a, b member) int { return strings.Compare(a.PSID, b.PSID) })
	slices.SortFunc(r.Products, func(a, b entity) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(r.Suites, func(a, b entity) int { return strings.Compare(a.ID, b.ID) })

	return r, nil
}

func label(value string) bool {
	return len(value) <= 63 && regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`).MatchString(value)
}

func validate(m model, r registry) error {
	if m.Version != 1 || !regexp.MustCompile(`^[a-z][a-z0-9]{0,7}$`).MatchString(m.Run) || !label(m.Revision) {
		return fmt.Errorf("invalid version, run (1-8 lowercase alphanumeric), or revision")
	}

	for _, domain := range []string{m.ClusterDomain, m.LocalDomain} {
		if len(domain) > 200 {
			return fmt.Errorf("domain too long: %s", domain)
		}

		for _, part := range strings.Split(domain, ".") {
			if !label(part) {
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

	for product := range products {
		if !slices.ContainsFunc(r.Members, func(s member) bool { return s.Product == product }) {
			return fmt.Errorf("empty product: %s", product)
		}
	}

	allBounds := []bounds{m.Replicas}
	for key, value := range m.Overrides {
		parts := strings.Split(key, "/")
		if len(parts) != 2 || !members[parts[0]] || !slices.Contains([]string{"pg", "mem", "disk"}, parts[1]) {
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
			if !slices.Contains([]string{"pg", "mem", "disk"}, pool) || seenPools[pool] {
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
