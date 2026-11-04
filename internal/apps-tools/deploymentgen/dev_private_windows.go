// Copyright (c) 2025-2026 Justin Cranford.

//go:build windows

package deploymentgen

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

func secureDevelopmentDirectory(path string) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return errors.New("development owner unavailable")
	}

	descriptor, err := windows.SecurityDescriptorFromString("O:" + user.User.Sid.String() +
		"D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")(A;OICI;FA;;;SY)")
	if err != nil {
		return errors.New("development DACL construction failed")
	}

	acl, _, err := descriptor.DACL()
	if err != nil {
		return errors.New("development DACL unavailable")
	}

	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil); err != nil {
		return fmt.Errorf("development protected DACL application: %w", err)
	}

	return nil
}
