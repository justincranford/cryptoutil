// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cli_test

import (
"testing"

"github.com/stretchr/testify/require"

cryptoutilAppsTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

func TestIsHelpRequest(t *testing.T) {
t.Parallel()

tests := []struct {
name   string
args   []string
expect bool
}{
{name: "help_word", args: []string{"help"}, expect: true},
{name: "help_long_flag", args: []string{"--help"}, expect: true},
{name: "help_short_flag", args: []string{"-h"}, expect: true},
{name: "empty_args", args: []string{}, expect: false},
{name: "non_help_arg", args: []string{"server"}, expect: false},
{name: "help_not_first", args: []string{"server", "help"}, expect: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

result := cryptoutilAppsTemplateCli.IsHelpRequest(tc.args)
require.Equal(t, tc.expect, result)
})
}
}
