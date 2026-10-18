//go:build windows

package tls

import (
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

func securePrivateFileACL(path string) error {
	return securePrivateACL(path, 0)
}

func securePrivateStagingDirectory(path string) error {
	return securePrivateACL(path, windows.OBJECT_INHERIT_ACE|windows.CONTAINER_INHERIT_ACE)
}

func securePrivateACL(path string, inheritance uint32) error {
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
		Inheritance:       inheritance,
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
			Inheritance:       inheritance,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
				TrusteeValue: windows.TrusteeValueFromSID(systemSID),
			},
		})
	}

	acl, err := windows.ACLFromEntries(entries, nil)
	if err != nil {
		return fmt.Errorf("build private material DACL: %w", err)
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
		return fmt.Errorf("set protected private material DACL: %w", err)
	}

	return validatePrivateACL(path, inheritance)
}

func validatePrivateFileACL(path string) error {
	return validatePrivateACL(path, 0)
}

func validatePrivateStagingDirectoryACL(path string) error {
	return validatePrivateACL(path, windows.OBJECT_INHERIT_ACE|windows.CONTAINER_INHERIT_ACE)
}

type privateACLValidationOperations struct {
	getTokenUserFn         func() (*windows.Tokenuser, error)
	createWellKnownSidFn   func(windows.WELL_KNOWN_SID_TYPE) (*windows.SID, error)
	getNamedSecurityInfoFn func(string, windows.SE_OBJECT_TYPE, windows.SECURITY_INFORMATION) (*windows.SECURITY_DESCRIPTOR, error)
	getDACLFn              func(*windows.SECURITY_DESCRIPTOR) (*windows.ACL, bool, error)
	getControlFn           func(*windows.SECURITY_DESCRIPTOR) (windows.SECURITY_DESCRIPTOR_CONTROL, uint32, error)
}

func validatePrivateACL(path string, inheritance uint32) error {
	return validatePrivateACLWithOperations(path, inheritance, defaultPrivateACLValidationOperations())
}

func validatePrivateACLWithOperations(path string, inheritance uint32, ops privateACLValidationOperations) error {
	tokenUser, err := ops.getTokenUserFn()
	if err != nil {
		return fmt.Errorf("get current user SID: %w", err)
	}

	systemSID, err := ops.createWellKnownSidFn(windows.WinLocalSystemSid)
	if err != nil {
		return fmt.Errorf("create LocalSystem SID: %w", err)
	}

	descriptor, err := ops.getNamedSecurityInfoFn(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return fmt.Errorf("read private material security descriptor: %w", err)
	}

	dacl, _, err := ops.getDACLFn(descriptor)
	if err != nil {
		return fmt.Errorf("read private material DACL: %w", err)
	}

	control, _, err := ops.getControlFn(descriptor)
	if err != nil {
		return fmt.Errorf("read private material DACL control flags: %w", err)
	}

	expectedSIDs := []*windows.SID{tokenUser.User.Sid}
	if !tokenUser.User.Sid.Equals(systemSID) {
		expectedSIDs = append(expectedSIDs, systemSID)
	}

	aceCount := 0
	if dacl != nil {
		aceCount = int(dacl.AceCount)
	}

	expectedACECount := len(expectedSIDs)
	if inheritance != 0 {
		expectedACECount *= 2
	}

	protected := control&windows.SE_DACL_PROTECTED != 0
	if dacl == nil || aceCount != expectedACECount || !protected {
		return fmt.Errorf("private material DACL is not restricted and protected: ACEs=%d, want=%d, protected=%t",
			aceCount, expectedACECount, protected)
	}

	descriptorText := descriptor.String()
	for _, sid := range expectedSIDs {
		sidText := sid.String()
		if sid.Equals(systemSID) {
			sidText = "SY"
		}

		if !strings.Contains(descriptorText, "(A;;FA;;;"+sidText+")") {
			return fmt.Errorf("private material DACL does not grant full access to an authorized principal")
		}

		if inheritance != 0 && !strings.Contains(descriptorText, "(A;OICIIO;GA;;;"+sidText+")") {
			return fmt.Errorf("private staging DACL does not propagate access to child files and directories")
		}
	}

	runtime.KeepAlive(descriptor)

	return nil
}

func defaultPrivateACLValidationOperations() privateACLValidationOperations {
	return privateACLValidationOperations{
		getTokenUserFn: func() (*windows.Tokenuser, error) {
			return windows.GetCurrentProcessToken().GetTokenUser()
		},
		createWellKnownSidFn:   windows.CreateWellKnownSid,
		getNamedSecurityInfoFn: windows.GetNamedSecurityInfo,
		getDACLFn: func(descriptor *windows.SECURITY_DESCRIPTOR) (*windows.ACL, bool, error) {
			return descriptor.DACL()
		},
		getControlFn: func(descriptor *windows.SECURITY_DESCRIPTOR) (windows.SECURITY_DESCRIPTOR_CONTROL, uint32, error) {
			return descriptor.Control()
		},
	}
}
