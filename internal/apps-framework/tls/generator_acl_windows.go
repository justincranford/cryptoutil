//go:build windows

package tls

import (
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

func securePrivateFileACL(path string) error {
	tokenUser, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("get current user SID: %w", err)
	}

	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return fmt.Errorf("create LocalSystem SID: %w", err)
	}

	userSID := tokenUser.User.Sid

	var pinner runtime.Pinner
	defer pinner.Unpin()

	pinner.Pin(userSID)
	pinner.Pin(systemSID)

	entries := []windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.GRANT_ACCESS,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(userSID),
		},
	}}
	if !userSID.Equals(systemSID) {
		entries = append(entries, windows.EXPLICIT_ACCESS{
			AccessPermissions: windows.GENERIC_ALL,
			AccessMode:        windows.GRANT_ACCESS,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
				TrusteeValue: windows.TrusteeValueFromSID(systemSID),
			},
		})
	}

	acl, err := windows.ACLFromEntries(entries, nil)
	if err != nil {
		return fmt.Errorf("build private file DACL: %w", err)
	}

	if err := windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	); err != nil {
		return fmt.Errorf("set protected private file DACL: %w", err)
	}

	return validatePrivateFileACL(path)
}

func validatePrivateFileACL(path string) error {
	tokenUser, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("get current user SID: %w", err)
	}

	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return fmt.Errorf("create LocalSystem SID: %w", err)
	}

	descriptor, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return fmt.Errorf("read private file security descriptor: %w", err)
	}

	dacl, _, err := descriptor.DACL()
	if err != nil {
		return fmt.Errorf("read private file DACL: %w", err)
	}

	control, _, err := descriptor.Control()
	if err != nil {
		return fmt.Errorf("read private file DACL control flags: %w", err)
	}

	expectedSIDs := []*windows.SID{tokenUser.User.Sid}
	if !tokenUser.User.Sid.Equals(systemSID) {
		expectedSIDs = append(expectedSIDs, systemSID)
	}

	if dacl == nil || int(dacl.AceCount) != len(expectedSIDs) || control&windows.SE_DACL_PROTECTED == 0 {
		return fmt.Errorf("private file DACL is not restricted and protected")
	}

	descriptorText := descriptor.String()

	userACE := "(A;;FA;;;" + tokenUser.User.Sid.String() + ")"
	if tokenUser.User.Sid.Equals(systemSID) {
		userACE = "(A;;FA;;;SY)"
	} else if !strings.Contains(descriptorText, "(A;;FA;;;SY)") {
		return fmt.Errorf("private file DACL does not grant access to LocalSystem")
	}

	if !strings.Contains(descriptorText, userACE) {
		return fmt.Errorf("private file DACL does not grant full access to the current user")
	}

	runtime.KeepAlive(descriptor)

	return nil
}
