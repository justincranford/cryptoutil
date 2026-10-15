//go:build windows

package tls

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSecurePrivateStagingDirectory_RestrictsInheritedACLs(t *testing.T) {
	t.Parallel()

	stagingDir := filepath.Join(t.TempDir(), "staging")
	require.NoError(t, os.Mkdir(stagingDir, cryptoutilSharedMagic.PKIInitCertsDirMode))
	require.NoError(t, setWorldReadableACL(stagingDir))
	require.ErrorContains(t, validatePrivateStagingDirectoryACL(stagingDir), "DACL")

	require.NoError(t, securePrivateStagingDirectory(stagingDir))
	require.NoError(t, validatePrivateStagingDirectoryACL(stagingDir))

	privateFile := filepath.Join(stagingDir, "private.key")
	require.NoError(t, os.WriteFile(privateFile, []byte("private"), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode))

	descriptor, err := windows.GetNamedSecurityInfo(privateFile, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	require.NoError(t, err)
	dacl, _, err := descriptor.DACL()
	require.NoError(t, err)
	require.Equal(t, uint16(2), dacl.AceCount)

	tokenUser, err := windows.GetCurrentProcessToken().GetTokenUser()
	require.NoError(t, err)
	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	require.NoError(t, err)
	worldSID, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	require.NoError(t, err)

	descriptorText := descriptor.String()
	require.Contains(t, descriptorText, tokenUser.User.Sid.String())
	require.True(t, strings.Contains(descriptorText, systemSID.String()) || strings.Contains(descriptorText, ";;;SY)"))
	require.NotContains(t, descriptorText, worldSID.String())

	runtime.KeepAlive(descriptor)
}

func TestValidateStagedFileMode_RestrictsBroadPrivateFileACL(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "private.key")
	require.NoError(t, os.WriteFile(path, []byte("private"), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode))

	require.NoError(t, setWorldReadableACL(path))
	t.Cleanup(func() {
		require.NoError(t, setWorldReadableACL(path))
		require.NoError(t, os.Chmod(path, 0o600))
	})
	require.ErrorContains(t, validatePrivateFileACL(path), "DACL")

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Zero(t, info.Mode().Perm()&0o222)

	stagedErr := validateStagedFileMode(path, info.Mode())
	require.NoError(t, stagedErr)
	require.NoError(t, validatePrivateFileACL(path))
}

func setWorldReadableACL(path string) error {
	tokenUser, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}

	worldSID, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	if err != nil {
		return err
	}

	var pinner runtime.Pinner
	defer pinner.Unpin()

	pinner.Pin(tokenUser.User.Sid)
	pinner.Pin(worldSID)

	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{
		{
			AccessPermissions: windows.GENERIC_ALL,
			AccessMode:        windows.GRANT_ACCESS,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  windows.TRUSTEE_IS_USER,
				TrusteeValue: windows.TrusteeValueFromSID(tokenUser.User.Sid),
			},
		},
		{
			AccessPermissions: windows.GENERIC_READ,
			AccessMode:        windows.GRANT_ACCESS,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
				TrusteeValue: windows.TrusteeValueFromSID(worldSID),
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	return windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
}
