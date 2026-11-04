// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const (
	testLifecyclePhysicalMaterial  = "nonproduction fixture"
	testLifecycleOutsideOwnerError = "outside owner"
)

func protectLifecycleTestPath(t *testing.T, name string, extraAccess string) {
	t.Helper()

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	require.NoError(t, err)

	sddl := "D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")" + extraAccess
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	require.NoError(t, err)
	acl, _, err := descriptor.DACL()
	require.NoError(t, err)
	require.NoError(t, windows.SetNamedSecurityInfo(name, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil))
}

func TestLifecyclePhysicalWindowsProtectedImport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*testing.T, string, string)
		want string
	}{
		{name: "protected owner and SYSTEM", edit: func(t *testing.T, root, file string) {
			protectLifecycleTestPath(t, root, "(A;OICI;FA;;;SY)")
			protectLifecycleTestPath(t, file, "(A;;FA;;;SY)")
		}},
		{name: "readable by everyone", edit: func(t *testing.T, _, file string) {
			protectLifecycleTestPath(t, file, "(A;;FR;;;WD)")
		}, want: testLifecycleOutsideOwnerError},
		{name: "writable parent", edit: func(t *testing.T, root, _ string) {
			protectLifecycleTestPath(t, root, "(A;OICI;FW;;;WD)")
		}, want: testLifecycleOutsideOwnerError},
		{name: "hard linked", edit: func(t *testing.T, root, file string) {
			require.NoError(t, os.Link(file, filepath.Join(root, "alias")))
		}, want: "hard-linked"},
		{name: cryptoutilSharedMagic.EMPTY, edit: func(t *testing.T, _, file string) {
			require.NoError(t, os.WriteFile(file, nil, 0o600))
		}, want: testNonemptyError},
		{name: "removed source file", edit: func(t *testing.T, _, file string) {
			require.NoError(t, os.Remove(file))
		}, want: testLifecycleMissingFileError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			root := t.TempDir()
			protectLifecycleTestPath(t, root, "")
			// All consumers use one explicit protected fixture file; no real material.
			for index := range input.Secrets {
				for key := range input.Secrets[index].Files {
					input.Secrets[index].Files[key] = testLifecycleMaterialFile
				}
			}

			var err error

			model.Revision, err = LifecycleRevision(input)
			require.NoError(t, err)
			inventory, err := build(model, RegisteredScopes(), testServiceSMKMSMemSelector)
			require.NoError(t, err)

			file := filepath.Join(root, testLifecycleMaterialFile)
			require.NoError(t, os.WriteFile(file, []byte(testLifecyclePhysicalMaterial), 0o600))
			protectLifecycleTestPath(t, file, "")
			test.edit(t, root, file)

			secrets, err := ImportPhysicalLifecycleSecrets(model, inventory, input, root)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, secrets)

				return
			}

			require.NoError(t, err)
			require.Len(t, secrets, len(input.Secrets))
		})
	}
}

func TestLifecycleWindowsDescriptorBoundaries(t *testing.T) {
	t.Parallel()

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	require.NoError(t, err)

	sid := user.User.Sid.String()

	tests := []struct {
		name, sddl, want string
	}{
		{"owner exclusive", "O:" + sid + "D:P(A;;FA;;;" + sid + ")", ""},
		{"SYSTEM privileged", "O:" + sid + "D:P(A;;FA;;;" + sid + ")(A;;FA;;;SY)", ""},
		{"deny entry", "O:" + sid + "D:P(D;;FR;;;WD)(A;;FA;;;" + sid + ")", ""},
		{"no owner", "D:P(A;;FA;;;" + sid + ")", "explicit owner"},
		{"wrong owner", "O:SYD:P(A;;FA;;;SY)", "current process"},
		{"null DACL", "O:" + sid + "D:NO_ACCESS_CONTROL", "non-null"},
		{"Everyone read", "O:" + sid + "D:P(A;;FR;;;WD)", testLifecycleOutsideOwnerError},
		{"admin write", "O:" + sid + "D:P(A;;FW;;;BA)", testLifecycleOutsideOwnerError},
		{"inherited Everyone", "O:" + sid + "D:P(A;OICI;FR;;;WD)", testLifecycleOutsideOwnerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			descriptor, err := windows.SecurityDescriptorFromString(test.sddl)
			require.NoError(t, err)

			err = validateLifecycleWindowsDescriptor(descriptor, user.User.Sid)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)

				return
			}

			require.NoError(t, err)
		})
	}

	require.Error(t, validateLifecycleWindowsDescriptor(nil, user.User.Sid))

	descriptor, err := windows.SecurityDescriptorFromString("O:" + sid + "D:P")
	require.NoError(t, err)
	require.Error(t, validateLifecycleWindowsDescriptor(descriptor, nil))
}

func TestLifecycleCLIPhysicalWindowsEndToEnd(t *testing.T) {
	t.Parallel()

	model, _, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	directory := t.TempDir()
	protectLifecycleTestPath(t, directory, "")

	for index := range input.Secrets {
		for key := range input.Secrets[index].Files {
			input.Secrets[index].Files[key] = testLifecycleMaterialFile
		}
	}

	var err error

	model.Revision, err = LifecycleRevision(input)
	require.NoError(t, err)

	file := filepath.Join(directory, testLifecycleMaterialFile)
	require.NoError(t, os.WriteFile(file, []byte(testLifecyclePhysicalMaterial), 0o600))
	protectLifecycleTestPath(t, file, "")

	inputJSON, err := json.Marshal(input)
	require.NoError(t, err)
	modelJSON, err := json.Marshal(model)
	require.NoError(t, err)

	read := func(name string) ([]byte, error) {
		switch name {
		case testLifecycleOperatorFile:
			return inputJSON, nil
		case defaultPlannerPaths().model:
			return modelJSON, nil
		default:
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "import must not use ordinary file reader for material: %s", name)

			return nil, nil
		}
	}

	var stdout, stderr bytes.Buffer

	args := []string{
		testLifecycleSubcommand, testLifecycleInputFlag, testLifecycleOperatorFile, cliSelectFlag, testServiceSMKMSMemSelector,
		cliFormatFlag, lifecycleImportCheckFormat, testLifecycleSecretRootFlag, directory,
	}
	require.Zero(t, Main(args, &stdout, &stderr, read, func(string, []byte, os.FileMode) error {
		t.Fatal("import must not persist material")

		return nil
	}), stderr.String())
	require.NotContains(t, stdout.String()+stderr.String(), testLifecyclePhysicalMaterial)
	require.Empty(t, stderr.String())
	require.Contains(t, stdout.String(), "protected lifecycle Secrets in memory")

	stdout.Reset()
	stderr.Reset()
	protectLifecycleTestPath(t, file, "(A;;FR;;;WD)")
	require.Equal(t, 1, Main(args, &stdout, &stderr, read, nil))
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), testLifecycleOutsideOwnerError)
	require.NotContains(t, stderr.String(), testLifecyclePhysicalMaterial)
}
