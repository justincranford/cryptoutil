package lint_deployments

import (
"testing"
)

func TestMapDeploymentToConfig(t *testing.T) {
t.Parallel()

tests := []struct {
name       string
deployment string
want       string
}{
{name: "identity service", deployment: "identity-authz", want: "identity"},
{name: "identity product", deployment: "identity", want: "identity"},
{name: "cipher service", deployment: "cipher-im", want: "cipher"},
{name: "cipher product", deployment: "cipher", want: "cipher"},
{name: "jose service", deployment: "jose-ja", want: "jose"},
{name: "jose product", deployment: "jose", want: "jose"},
{name: "pki explicit mapping", deployment: "pki", want: "ca"},
{name: "pki-ca explicit mapping", deployment: "pki-ca", want: "ca"},
{name: "sm explicit mapping", deployment: "sm", want: "sm"},
{name: "sm-kms explicit mapping", deployment: "sm-kms", want: "sm"},
{name: "single segment fallback", deployment: "newproduct", want: "newproduct"},
{name: "product-service fallback", deployment: "newproduct-service", want: "newproduct"},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

got := mapDeploymentToConfig(tc.deployment)
if got != tc.want {
t.Errorf("mapDeploymentToConfig(%q) = %q, want %q", tc.deployment, got, tc.want)
}
})
}
}

func TestGetSubdirectories(t *testing.T) {
t.Parallel()

t.Run("nonexistent directory", func(t *testing.T) {
t.Parallel()

dirs, err := getSubdirectories("/nonexistent/path")
if err == nil {
t.Error("expected error for nonexistent directory")
}

if dirs != nil {
t.Error("expected nil dirs on error")
}
})

t.Run("empty directory", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

dirs, err := getSubdirectories(tmpDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if len(dirs) != 0 {
t.Errorf("expected 0 subdirectories, got %d", len(dirs))
}
})

t.Run("directories and files", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
createTestDir(t, tmpDir, "subdir1")
createTestDir(t, tmpDir, "subdir2")
createTestFile(t, tmpDir, "file.txt", "")

dirs, err := getSubdirectories(tmpDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if len(dirs) != 2 {
t.Errorf("expected 2 subdirectories, got %d: %v", len(dirs), dirs)
}
})
}

func TestValidateStructuralMirror(t *testing.T) {
t.Parallel()

t.Run("both directories empty", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if !result.Valid {
t.Error("expected valid for empty directories")
}

if len(result.MissingMirrors) != 0 {
t.Errorf("expected 0 missing, got %d", len(result.MissingMirrors))
}
})

t.Run("nonexistent deployments directory", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "configs")

_, err := ValidateStructuralMirror("/nonexistent", configsDir)
if err == nil {
t.Error("expected error for nonexistent deployments dir")
}
})

t.Run("nonexistent configs directory", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
createTestDir(t, tmpDir, "deployments")

_, err := ValidateStructuralMirror(deploymentsDir, "/nonexistent")
if err == nil {
t.Error("expected error for nonexistent configs dir")
}
})

t.Run("excluded directories skipped", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

// Create excluded deployment directories.
createTestDir(t, deploymentsDir, "shared-postgres")
createTestDir(t, deploymentsDir, "shared-citus")
createTestDir(t, deploymentsDir, "shared-telemetry")
createTestDir(t, deploymentsDir, "compose")
createTestDir(t, deploymentsDir, "template")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if !result.Valid {
t.Error("expected valid when only excluded dirs exist")
}

if len(result.Excluded) != 5 {
t.Errorf("expected 5 excluded, got %d: %v", len(result.Excluded), result.Excluded)
}

if len(result.MissingMirrors) != 0 {
t.Errorf("expected 0 missing, got %d", len(result.MissingMirrors))
}
})

t.Run("missing config directory detected", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

// Create deployment dir without matching config dir.
createTestDir(t, deploymentsDir, "cipher-im")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if result.Valid {
t.Error("expected invalid when config mirror missing")
}

if len(result.MissingMirrors) != 1 {
t.Fatalf("expected 1 missing, got %d", len(result.MissingMirrors))
}

		if result.MissingMirrors[0] != "cipher-im" {
			t.Errorf("expected missing mirror 'cipher-im', got %q", result.MissingMirrors[0])
}
})

t.Run("matching config directory passes", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

// Create deployment and matching config.
createTestDir(t, deploymentsDir, "cipher-im")
createTestDir(t, configsDir, "cipher")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if !result.Valid {
t.Errorf("expected valid, got errors: %v, missing: %v", result.Errors, result.MissingMirrors)
}
})

t.Run("orphaned config directory warns", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

// Create config without matching deployment.
createTestDir(t, configsDir, "orphaned-service")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if !result.Valid {
t.Error("orphaned configs should not invalidate result")
}

if len(result.Orphans) != 1 {
t.Fatalf("expected 1 orphan, got %d", len(result.Orphans))
}

if result.Orphans[0] != "orphaned-service" {
t.Errorf("expected orphan 'orphaned-service', got %q", result.Orphans[0])
}
})

t.Run("deduplication of config names", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

// Both identity and identity-authz map to identity config.
createTestDir(t, deploymentsDir, "identity")
createTestDir(t, deploymentsDir, "identity-authz")
createTestDir(t, configsDir, "identity")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if !result.Valid {
t.Errorf("expected valid when both product and service map to same config, errors: %v, missing: %v", result.Errors, result.MissingMirrors)
}

if len(result.MissingMirrors) != 0 {
t.Errorf("expected 0 missing, got %d: %v", len(result.MissingMirrors), result.MissingMirrors)
}
})

t.Run("explicit mapping pki to ca", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

createTestDir(t, deploymentsDir, "pki-ca")
createTestDir(t, configsDir, "ca")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if !result.Valid {
t.Errorf("expected valid for pki-ca -> ca mapping, errors: %v, missing: %v", result.Errors, result.MissingMirrors)
}
})

t.Run("warnings for orphaned configs", func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
deploymentsDir := tmpDir + "/deployments"
configsDir := tmpDir + "/configs"
createTestDir(t, tmpDir, "deployments")
createTestDir(t, tmpDir, "configs")

createTestDir(t, configsDir, "orphan1")
createTestDir(t, configsDir, "orphan2")

result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if len(result.Warnings) < 2 {
t.Errorf("expected at least 2 warnings, got %d", len(result.Warnings))
}
})
}

func TestFormatMirrorResult(t *testing.T) {
t.Parallel()

t.Run("valid result", func(t *testing.T) {
t.Parallel()

result := &MirrorResult{
Valid:          true,
MissingMirrors: []string{},
Orphans:        []string{},
Excluded:       []string{"shared-postgres"},
}

output := FormatMirrorResult(result)

if output == "" {
t.Error("expected non-empty output")
}
})

t.Run("invalid result with missing mirrors", func(t *testing.T) {
t.Parallel()

result := &MirrorResult{
Valid:          false,
MissingMirrors: []string{"cipher", "jose"},
Orphans:        []string{"orphan1"},
Excluded:       []string{"template"},
Errors:         []string{"some error"},
Warnings:       []string{"orphaned: orphan1"},
}

output := FormatMirrorResult(result)

if output == "" {
t.Error("expected non-empty output")
}
})

t.Run("empty result", func(t *testing.T) {
t.Parallel()

result := &MirrorResult{
Valid: true,
}

output := FormatMirrorResult(result)

if output == "" {
t.Error("expected non-empty output for empty result")
}
})
}
