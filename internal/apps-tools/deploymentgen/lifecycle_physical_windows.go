// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

const lifecycleWindowsAllowACE = "A"

func validatePhysicalLifecycleHandle(file *os.File, _ fs.FileInfo, directory bool) error {
	handle := windows.Handle(file.Fd())

	var info windows.ByHandleFileInformation

	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return fmt.Errorf("inspect protected Windows handle: unavailable")
	}

	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (!directory && info.NumberOfLinks != 1) {
		return fmt.Errorf("protected Windows source must not be a reparse point or hard-linked file")
	}

	descriptor, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil || descriptor == nil {
		return fmt.Errorf("read protected Windows ownership and ACL: unavailable")
	}

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("read protected Windows process identity: unavailable")
	}

	return validateLifecycleWindowsDescriptor(descriptor, user.User.Sid)
}

func validateLifecycleWindowsDescriptor(descriptor *windows.SECURITY_DESCRIPTOR, user *windows.SID) error {
	if descriptor == nil || !descriptor.IsValid() || user == nil {
		return fmt.Errorf("protected Windows source requires a valid security descriptor and process identity")
	}

	owner, defaulted, err := descriptor.Owner()
	if err != nil || owner == nil || defaulted {
		return fmt.Errorf("protected Windows source requires an explicit owner")
	}

	if !windows.EqualSid(owner, user) {
		return fmt.Errorf("protected Windows source must be owned by the current process user")
	}

	acl, _, err := descriptor.DACL()
	if err != nil || acl == nil {
		return fmt.Errorf("protected Windows source requires a non-null restrictive DACL")
	}

	// Native SDDL conversion avoids manually walking variable-length ACE/SID
	// memory. Accept only ordinary six-field allow/deny entries, and verify
	// enumeration completeness against the native ACL's ACE count.
	return validateLifecycleWindowsGrants(descriptor.String(), owner, int(acl.AceCount))
}

func validateLifecycleWindowsGrants(sddl string, owner *windows.SID, count int) error {
	start := strings.Index(sddl, "D:")
	if start < 0 {
		return fmt.Errorf("protected Windows DACL conversion is unavailable")
	}

	dacl := sddl[start+len("D:"):]
	if end := strings.Index(dacl, "S:"); end >= 0 {
		dacl = dacl[:end]
	}

	seen := 0

	for {
		begin := strings.IndexByte(dacl, '(')
		if begin < 0 {
			break
		}

		end := strings.IndexByte(dacl[begin:], ')')
		if end < 0 {
			return fmt.Errorf("protected Windows DACL entry is incomplete")
		}

		fields := strings.Split(dacl[begin+1:begin+end], ";")
		if len(fields) != 6 || fields[3] != "" || fields[4] != "" || (fields[0] != lifecycleWindowsAllowACE && fields[0] != "D") {
			return fmt.Errorf("protected Windows ACL contains an unsupported entry type")
		}

		seen++

		if fields[0] == lifecycleWindowsAllowACE {
			principal, err := windows.SecurityDescriptorFromString("O:" + fields[5])
			if err != nil {
				return fmt.Errorf("protected Windows ACL principal is invalid")
			}

			sid, _, err := principal.Owner()
			if err != nil || sid == nil || (!windows.EqualSid(owner, sid) && !sid.IsWellKnown(windows.WinLocalSystemSid)) {
				return fmt.Errorf("protected Windows ACL grants access outside owner and SYSTEM")
			}
		}

		dacl = dacl[begin+end+1:]
	}

	if seen != count {
		return fmt.Errorf("protected Windows DACL enumeration is incomplete")
	}

	return nil
}
