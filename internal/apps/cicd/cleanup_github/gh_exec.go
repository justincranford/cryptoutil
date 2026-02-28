// Copyright (c) 2025 Justin Cranford

package cleanup_github

import (
"context"
"errors"
"fmt"
"os/exec"
"strings"
)

// ghExec runs a gh CLI command and returns stdout.
func ghExec(args ...string) ([]byte, error) {
ctx, cancel := context.WithTimeout(context.Background(), ghCommandTimeout)
defer cancel()

cmd := exec.CommandContext(ctx, ghBinary, args...)

output, err := cmd.Output()
if err != nil {
var exitErr *exec.ExitError
if ok := isExitError(err, &exitErr); ok {
return nil, fmt.Errorf("gh %s failed: %s", strings.Join(args, " "), string(exitErr.Stderr))
}

return nil, fmt.Errorf("gh %s failed: %w", strings.Join(args, " "), err)
}

return output, nil
}

// isExitError checks if the error is an exec.ExitError and assigns it.
func isExitError(err error, target **exec.ExitError) bool {
exitErr := &exec.ExitError{}

ok := errors.As(err, &exitErr)
if ok {
*target = exitErr
}

return ok
}

// repoArgs returns the --repo flag if cfg.Repo is set.
func repoArgs(cfg *CleanupConfig) []string {
if cfg.Repo != "" {
return []string{"--repo", cfg.Repo}
}

return nil
}
