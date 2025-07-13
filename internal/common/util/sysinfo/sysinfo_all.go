package sysinfo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	cryptoutilUtil "cryptoutil/internal/common/util"
)

func GetAllInfoWithTimeout(sysInfoProvider SysInfoProvider, timeout time.Duration) ([][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		runtimeGoArch = sysInfoProvider.RuntimeGoArch()
		runtimeGoOS   = sysInfoProvider.RuntimeGoOS()
		runtimeNumCPU = fmt.Sprintf("%d", sysInfoProvider.RuntimeNumCPU())

		cpuVendorID, cpuFamily, cpuPhysicalID, cpuModelName string
		ramSize                                             string
		osHostname                                          string
		hostID                                              string
		userID, groupID, username                           string

		cpuErr, ramErr, osErr, hostIDErr, userErr error
	)

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		done := make(chan struct{})

		go func() {
			cpuVendorID, cpuFamily, cpuPhysicalID, cpuModelName, cpuErr = sysInfoProvider.CPUInfo()
			close(done)
		}()

		select {
		case <-done:
			// Completed normally
		case <-ctx.Done():
			cpuErr = fmt.Errorf("CPU info timeout: %w", ctx.Err())
		}
	}()

	// RAM info with context check
	go func() {
		defer wg.Done()
		done := make(chan struct{})

		go func() {
			var ram uint64
			ram, ramErr = sysInfoProvider.RAMSize()
			ramSize = fmt.Sprintf("%d", ram)
			close(done)
		}()

		select {
		case <-done:
			// Completed normally
		case <-ctx.Done():
			ramErr = fmt.Errorf("RAM info timeout: %w", ctx.Err())
		}
	}()

	// Hostname info with context check
	go func() {
		defer wg.Done()
		done := make(chan struct{})

		go func() {
			osHostname, osErr = sysInfoProvider.OSHostname()
			close(done)
		}()

		select {
		case <-done:
			// Completed normally
		case <-ctx.Done():
			osErr = fmt.Errorf("hostname info timeout: %w", ctx.Err())
		}
	}()

	// Host ID with context check
	go func() {
		defer wg.Done()
		done := make(chan struct{})

		go func() {
			hostID, hostIDErr = sysInfoProvider.HostID()
			close(done)
		}()

		select {
		case <-done:
			// Completed normally
		case <-ctx.Done():
			hostIDErr = fmt.Errorf("host ID info timeout: %w", ctx.Err())
		}
	}()

	// User info with context check
	go func() {
		defer wg.Done()
		done := make(chan struct{})

		go func() {
			userID, groupID, username, userErr = sysInfoProvider.UserInfo()
			close(done)
		}()

		select {
		case <-done:
			// Completed normally
		case <-ctx.Done():
			userErr = fmt.Errorf("user info timeout: %w", ctx.Err())
		}
	}()

	wg.Wait()

	// Collect all errors
	var errs []error
	if cpuErr != nil {
		errs = append(errs, cpuErr)
	}
	if ramErr != nil {
		errs = append(errs, ramErr)
	}
	if osErr != nil {
		errs = append(errs, osErr)
	}
	if hostIDErr != nil {
		errs = append(errs, hostIDErr)
	}
	if userErr != nil {
		errs = append(errs, userErr)
	}

	// If there are errors, return them
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cryptoutilUtil.StringPointersToBytes(&hostID, &userID, &groupID, &runtimeGoArch, &runtimeGoOS, &runtimeNumCPU, &cpuVendorID, &cpuFamily, &cpuPhysicalID, &cpuModelName, &ramSize, &osHostname, &username), nil
}
