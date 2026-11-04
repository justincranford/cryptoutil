// Copyright (c) 2025-2026 Justin Cranford.

// Package deploymentgen builds deterministic Kubernetes deployment graphs from
// the canonical cryptoutil registry and deployment inputs.
package deploymentgen

import (
	"slices"
	"strings"

	entityregistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// Tier identifies the registry level that owns a deployment scope.
type Tier string

const (
	TierService Tier = "s"
	TierProduct Tier = "p"
	TierSuite   Tier = "u"
)

// Scope is a registry-backed deployment scope. Members contains the PS-IDs
// deployed by the scope and is sorted for deterministic graph construction.
type Scope struct {
	Tier    Tier
	ID      string
	Members []string
}

type apiServiceDependency struct {
	SourcePSID string
	TargetPSID string
}

// RegisteredScopes derives all deployment scopes from the validated typed
// registry accessors. It does not maintain a parallel product or service list.
func RegisteredScopes() []Scope {
	services := entityregistry.AllProductServices()
	products := entityregistry.AllProducts()
	suites := entityregistry.AllSuites()

	scopes := make([]Scope, 0, len(services)+len(products)+len(suites))
	for _, service := range services {
		scopes = append(scopes, Scope{
			Tier:    TierService,
			ID:      service.PSID,
			Members: []string{service.PSID},
		})
	}

	for _, product := range products {
		scope := Scope{Tier: TierProduct, ID: product.ID, Members: make([]string, 0)}
		for _, service := range services {
			if service.Product == product.ID {
				scope.Members = append(scope.Members, service.PSID)
			}
		}

		sortScopeMembers(&scope)
		scopes = append(scopes, scope)
	}

	for _, suite := range suites {
		scope := Scope{Tier: TierSuite, ID: suite.ID, Members: make([]string, 0, len(services))}
		for _, service := range services {
			scope.Members = append(scope.Members, service.PSID)
		}

		sortScopeMembers(&scope)
		scopes = append(scopes, scope)
	}

	slices.SortFunc(scopes, func(left, right Scope) int {
		if tierOrder := strings.Compare(string(left.Tier), string(right.Tier)); tierOrder != 0 {
			return tierOrder
		}

		return strings.Compare(left.ID, right.ID)
	})

	return scopes
}

func sortScopeMembers(scope *Scope) {
	slices.Sort(scope.Members)
}

func registeredAPIDependenciesForMembers(members []string) []apiServiceDependency {
	memberSet := make(map[string]bool, len(members))
	for _, member := range members {
		memberSet[member] = true
	}

	result := make([]apiServiceDependency, 0)

	for _, dependency := range entityregistry.AllAPIServiceDependencies() {
		if memberSet[dependency.SourcePSID] && memberSet[dependency.TargetPSID] {
			result = append(result, apiServiceDependency{
				SourcePSID: dependency.SourcePSID,
				TargetPSID: dependency.TargetPSID,
			})
		}
	}

	return result
}

func apiServiceDependencyPolicyNames(dependency apiServiceDependency) []string {
	return []string{
		"allow-api-egress-" + dependency.SourcePSID + "-to-" + dependency.TargetPSID,
		"allow-api-ingress-" + dependency.TargetPSID + "-from-" + dependency.SourcePSID,
	}
}
