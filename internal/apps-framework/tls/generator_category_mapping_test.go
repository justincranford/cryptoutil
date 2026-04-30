// Copyright (c) 2025-2026 Justin Cranford.
//

package tls_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestGenerate_CategoryMapping verifies that Generate creates all expected
// directories for each of the 14 TLS certificate categories (skeleton-template PS-ID,
// 2 realms: file and db). Each category runs as a parallel subtest.
func TestGenerate_CategoryMapping(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSkeletonTemplate
	realms := []string{"file", "db"}

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		os.MkdirAll,
		stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()
	require.NoError(t, gen.Generate(psID, tmpDir))

	basePath := filepath.Join(tmpDir, psID)
	categories := expectedDirsForPSID(psID, realms)

	for catName, dirs := range categories {
		t.Run(catName, func(t *testing.T) {
			t.Parallel()

			for _, dir := range dirs {
				require.DirExists(t, filepath.Join(basePath, filepath.FromSlash(dir)),
					"%s: expected dir %q not generated", catName, dir)
			}
		})
	}
}

// TestGenerate_CategoryUniqueness verifies that every generated directory maps to
// exactly one certificate category: no orphan dirs and no duplicates in the spec.
func TestGenerate_CategoryUniqueness(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSkeletonTemplate
	realms := []string{"file", "db"}

	categories := expectedDirsForPSID(psID, realms)

	// Build dir→category inverted map; detect spec duplicates.
	dirToCategory := make(map[string]string)

	var specErrs []string

	for catName, dirs := range categories {
		for _, dir := range dirs {
			if existing, seen := dirToCategory[dir]; seen {
				specErrs = append(specErrs, fmt.Sprintf("spec error: dir %q appears in both %q and %q", dir, existing, catName))
			}

			dirToCategory[dir] = catName
		}
	}

	require.Empty(t, specErrs, "spec errors found")

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		os.MkdirAll,
		stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()
	require.NoError(t, gen.Generate(psID, tmpDir))

	basePath := filepath.Join(tmpDir, psID)

	var orphans []string

	require.NoError(t, filepath.WalkDir(basePath, func(p string, d fs.DirEntry, _ error) error {
		if !d.IsDir() || p == basePath {
			return nil
		}

		rel, err := filepath.Rel(basePath, p)
		if err != nil {
			return err
		}

		if _, ok := dirToCategory[filepath.ToSlash(rel)]; !ok {
			orphans = append(orphans, filepath.ToSlash(rel))
		}

		return nil
	}))

	require.Empty(t, orphans, "generated dirs not present in any category spec")
}

// expectedDirsForPSID returns the expected directory paths (relative to basePath,
// using forward slashes) grouped by TLS certificate category name.
// Used by both the category mapping test (1-to-many: cat→dirs) and the
// uniqueness test (1-to-1: dir→cat).
func expectedDirsForPSID(psID string, realms []string) map[string][]string {
	grafana := cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm // "grafana-otel-lgtm"
	otel := cryptoutilSharedMagic.PKIInitOtelCollectorContrib     // "otel-collector-contrib"
	infra := cryptoutilSharedMagic.PKIInitEntityInfra             // "infra"
	leader := cryptoutilSharedMagic.PKIInitPostgresLeader         // "leader"
	follower := cryptoutilSharedMagic.PKIInitPostgresFollower     // "follower"

	appSuffixes := cryptoutilAppsFrameworkTls.PKIInitAppInstanceSuffixes()              // sqlite-1, sqlite-2, postgres-1, postgres-2
	adminSuffixes := cryptoutilAppsFrameworkTls.PKIInitAdminInstanceSuffixes()          // sqlite-1, sqlite-2, postgres-1, postgres-2
	clientDomains := cryptoutilAppsFrameworkTls.PKIInitClientPKIDomains()               // sqlite-1, sqlite-2, postgres
	postgresSuffixes := cryptoutilAppsFrameworkTls.PKIInitPostgresAppInstanceSuffixes() // postgres-1, postgres-2
	userTypes := cryptoutilAppsFrameworkTls.PKIInitUserTypes()                          // browseruser, serviceuser

	// caChain returns the 4 dirs for a root+issuing CA pair (keystore + truststore each).
	caChain := func(prefix, suffix string) []string {
		root := prefix + "-root-" + suffix
		issuing := prefix + "-issuing-" + suffix

		return []string{
			root,
			root + "/truststore",
			issuing,
			issuing + "/truststore",
		}
	}

	// Cat 1: Global HTTPS Server CAs (4 dirs: root + root/truststore + issuing + issuing/truststore).
	cat1 := caChain("public-https-server", "ca")

	// Cat 2: Grafana LGTM + OTel Collector server leaf certs (2 dirs).
	cat2 := []string{
		"public-https-server-entity-" + grafana,
		"public-https-server-entity-" + otel,
	}

	// Cat 3: PS-ID app server leaf certs — 1 dir per instance suffix (4 total).
	cat3 := make([]string, 0, len(appSuffixes))
	for _, suffix := range appSuffixes {
		cat3 = append(cat3, "public-https-server-entity-"+psID+"-"+suffix)
	}

	// Cat 4: PS-ID HTTPS client CA chains — 4 dirs per PKI domain (12 total for 3 domains).
	cat4 := make([]string, 0, len(clientDomains)*4)
	for _, domain := range clientDomains {
		cat4 = append(cat4, caChain("public-https-client", "ca-"+psID+"-"+domain)...)
	}

	// Cat 5: PS-ID client leaf certs — 1 dir per domain × userType × realm.
	// Formula: 3 domains × 2 userTypes × len(realms) = 12 for 2 realms.
	cat5 := make([]string, 0, len(clientDomains)*len(userTypes)*len(realms))

	for _, domain := range clientDomains {
		for _, userType := range userTypes {
			for _, realm := range realms {
				cat5 = append(cat5, "public-https-client-entity-"+psID+"-"+domain+"-"+userType+"-"+realm)
			}
		}
	}

	// Cat 6: PS-ID admin mTLS CA chains — 4 dirs per instance suffix (16 total for 4 suffixes).
	cat6 := make([]string, 0, len(adminSuffixes)*4)
	for _, suffix := range adminSuffixes {
		cat6 = append(cat6, caChain("private-https-mutual", "ca-"+psID+"-"+suffix)...)
	}

	// Cat 7: PS-ID admin mTLS leaf certs — 1 dir per instance suffix (4 total).
	cat7 := make([]string, 0, len(adminSuffixes))
	for _, suffix := range adminSuffixes {
		cat7 = append(cat7, "private-https-mutual-entity-"+psID+"-"+suffix)
	}

	// Cat 8: Grafana LGTM + OTel Collector client CA chains (8 dirs = 2 chains × 4 each).
	cat8 := append(
		caChain(grafana+"-https-client", "ca"),
		caChain(otel+"-https-client", "ca")...,
	)

	// Cat 9: Global admin client certs (2) + global infra client certs (2)
	// + per-PS-ID client certs (4 suffixes × 2 = 8) = 12 dirs total.
	cat9 := make([]string, 0, 4+2*len(appSuffixes))
	cat9 = append(cat9,
		// Cat 9 (admin): global admin client certs for grafana and otel.
		grafana+"-https-client-entity-admin",
		otel+"-https-client-entity-admin",
		// Cat 9 (infra): OTel→Grafana forwarding service client certs.
		grafana+"-https-client-entity-"+infra,
		otel+"-https-client-entity-"+infra,
	)
	// Cat 9 (per-PS-ID): one grafana + one otel client cert per instance suffix.
	for _, suffix := range appSuffixes {
		cat9 = append(cat9,
			grafana+"-https-client-entity-"+psID+"-"+suffix,
			otel+"-https-client-entity-"+psID+"-"+suffix,
		)
	}

	// Cat 10: PostgreSQL Server CA chain (4 dirs).
	cat10 := caChain("postgres-tls-server", "ca")

	// Cat 11: PostgreSQL server leaf certs — leader and follower (2 dirs).
	cat11 := []string{
		"postgres-tls-server-entity-" + leader,
		"postgres-tls-server-entity-" + follower,
	}

	// Cat 12: PostgreSQL Client CA chain (4 dirs).
	cat12 := caChain("postgres-tls-client", "ca")

	// Cat 13: PostgreSQL replication client certs — leader and follower (2 dirs).
	cat13 := []string{
		"postgres-tls-client-entity-" + leader + "-replication",
		"postgres-tls-client-entity-" + follower + "-replication",
	}

	// Cat 14: PS-ID PostgreSQL app client certs — 1 dir per role × postgres instance suffix (4 total).
	cat14 := make([]string, 0, 2*len(postgresSuffixes))

	for _, role := range []string{leader, follower} {
		for _, suffix := range postgresSuffixes {
			cat14 = append(cat14, "postgres-tls-client-entity-"+role+"-"+psID+"-"+suffix)
		}
	}

	return map[string][]string{
		"Cat 1: Global HTTPS Server CAs":            cat1,
		"Cat 2: Grafana and OTel Server Certs":      cat2,
		"Cat 3: PS-ID App Server Certs":             cat3,
		"Cat 4: PS-ID HTTPS Client CAs":             cat4,
		"Cat 5: PS-ID HTTPS Client Leaf Certs":      cat5,
		"Cat 6: PS-ID Admin mTLS CAs":               cat6,
		"Cat 7: PS-ID Admin mTLS Leaf Certs":        cat7,
		"Cat 8: Grafana and OTel Client CAs":        cat8,
		"Cat 9: Global and PS-ID Client Leaf Certs": cat9,
		"Cat 10: PostgreSQL Server CAs":             cat10,
		"Cat 11: PostgreSQL Server Certs":           cat11,
		"Cat 12: PostgreSQL Client CAs":             cat12,
		"Cat 13: PostgreSQL Replication Certs":      cat13,
		"Cat 14: PS-ID PostgreSQL App Client Certs": cat14,
	}
}
