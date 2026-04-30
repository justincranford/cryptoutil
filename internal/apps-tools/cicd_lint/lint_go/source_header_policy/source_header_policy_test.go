// Copyright (c) 2025-2026 Justin Cranford.

package source_header_policy

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckInDirWithYear_ValidHeaders(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	currentYear := time.Now().UTC().Year()

	goodFile := filepath.Join(tmpDir, "good.go")
	content := strings.Join([]string{
		"// Copyright (c) 2025-" + intToString(currentYear) + " Justin Cranford.",
		"// SPDX-License-Identifier: AGPL-3.0-only",
		"",
		"package example",
		"",
	}, "\n")

	require.NoError(t, os.WriteFile(goodFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")
	err := checkInDirWithYear(logger, tmpDir, currentYear)
	require.NoError(t, err)
}

func TestCheckInDirWithYear_DetectsSPDXMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	currentYear := time.Now().UTC().Year()

	badFile := filepath.Join(tmpDir, "bad_spdx.go")
	content := strings.Join([]string{
		"// Copyright (c) 2025-" + intToString(currentYear) + " Justin Cranford.",
		"// SPDX-License-Identifier: MIT",
		"",
		"package example",
		"",
	}, "\n")

	require.NoError(t, os.WriteFile(badFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")
	err := checkInDirWithYear(logger, tmpDir, currentYear)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SPDX-License-Identifier")
	require.Contains(t, err.Error(), "AGPL-3.0-only")
}

func TestCheckInDirWithYear_StaleYearAllowed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	currentYear := time.Now().UTC().Year()
	staleYear := currentYear - 1

	filePath := filepath.Join(tmpDir, "stale_year.go")
	content := strings.Join([]string{
		"// Copyright (c) " + intToString(staleYear) + " Justin Cranford.",
		"// SPDX-License-Identifier: AGPL-3.0-only",
		"",
		"package example",
		"",
	}, "\n")

	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")
	err := checkInDirWithYear(logger, tmpDir, currentYear)
	require.NoError(t, err)
}

func TestCheckInDirWithYear_FutureYearRejected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	currentYear := time.Now().UTC().Year()
	futureYear := currentYear + 1

	filePath := filepath.Join(tmpDir, "future_year.go")
	content := strings.Join([]string{
		"// Copyright (c) 2025-" + intToString(futureYear) + " Justin Cranford.",
		"// SPDX-License-Identifier: AGPL-3.0-only",
		"",
		"package example",
		"",
	}, "\n")

	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")
	err := checkInDirWithYear(logger, tmpDir, currentYear)
	require.Error(t, err)
	require.Contains(t, err.Error(), "copyright year range ends")
}

func TestCheckInDirWithYear_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	currentYear := time.Now().UTC().Year()

	nonGo := filepath.Join(tmpDir, "note.txt")
	content := "// SPDX-License-Identifier: MIT\n"
	require.NoError(t, os.WriteFile(nonGo, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")
	err := checkInDirWithYear(logger, tmpDir, currentYear)
	require.NoError(t, err)
}

func TestCheckInDirWithYear_IgnoresBodyStringLiterals(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	currentYear := time.Now().UTC().Year()

	filePath := filepath.Join(tmpDir, "fixture.go")
	content := strings.Join([]string{
		"// Copyright (c) 2025-" + intToString(currentYear) + " Justin Cranford.",
		"// SPDX-License-Identifier: AGPL-3.0-only",
		"",
		"package example",
		"",
		"const fixture = `// SPDX-License-Identifier: MIT`",
		"",
	}, "\n")

	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("source-header-policy-test")
	err := checkInDirWithYear(logger, tmpDir, currentYear)
	require.NoError(t, err)
}

func TestParseCopyrightYears(t *testing.T) {
	t.Parallel()

	startYear, endYear, err := parseCopyrightYears("2025", "2026")
	require.NoError(t, err)
	require.Equal(t, 2025, startYear)
	require.Equal(t, 2026, endYear)

	startYear, endYear, err = parseCopyrightYears("2026", "")
	require.NoError(t, err)
	require.Equal(t, 2026, startYear)
	require.Equal(t, 2026, endYear)
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
