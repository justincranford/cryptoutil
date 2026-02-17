package lint_deployments

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestValidateNaming_Simple(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.MkdirAll(filepath.Join(dir, "service-one"), 0o755))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}
