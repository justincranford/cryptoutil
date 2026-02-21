// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/require"
)

// TestAuthenticator_Reload tests the Reload function with various scenarios.
func TestAuthenticator_Reload(t *testing.T) {
t.Parallel()

tests := []struct {
name      string
setup     func(t *testing.T) string
wantErr   bool
errContains string
}{
{
name: "success with no config file",
setup: func(t *testing.T) string {
t.Helper()
// Return empty temp dir - no realms.yml file.
return t.TempDir()
},
wantErr: false,
},
{
name: "success with valid config",
setup: func(t *testing.T) string {
t.Helper()
tmpDir := t.TempDir()
// Write minimal valid realms.yml.
content := "version: \"1.0\"\nrealms: []\ndefaults:\n  password_policy:\n    algorithm: sha256\n    iterations: 600000\n    salt_bytes: 32\n    hash_bytes: 32\n"
err := os.WriteFile(filepath.Join(tmpDir, "realms.yml"), []byte(content), 0o600)
require.NoError(t, err)

return tmpDir
},
wantErr: false,
},
{
name: "error with invalid yaml",
setup: func(t *testing.T) string {
t.Helper()
tmpDir := t.TempDir()
// Write invalid YAML content.
err := os.WriteFile(filepath.Join(tmpDir, "realms.yml"), []byte("{\tinvalid yaml content"), 0o600)
require.NoError(t, err)

return tmpDir
},
wantErr:     true,
errContains: "failed to reload config",
},
{
name: "error with invalid realm id",
setup: func(t *testing.T) string {
t.Helper()
tmpDir := t.TempDir()
// Write valid YAML that fails validation (bad UUID).
content := "version: \"1.0\"\nrealms:\n  - id: \"not-a-uuid\"\n    name: \"test\"\n    type: \"file\"\n    enabled: true\n"
err := os.WriteFile(filepath.Join(tmpDir, "realms.yml"), []byte(content), 0o600)
require.NoError(t, err)

return tmpDir
},
wantErr:     true,
errContains: "failed to reload config",
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

auth, err := NewAuthenticator(DefaultConfig())
require.NoError(t, err)
require.NotNil(t, auth)

configDir := tc.setup(t)
err = auth.Reload(configDir)

if tc.wantErr {
require.Error(t, err)

if tc.errContains != "" {
require.Contains(t, err.Error(), tc.errContains)
}
} else {
require.NoError(t, err)
}
})
}
}
