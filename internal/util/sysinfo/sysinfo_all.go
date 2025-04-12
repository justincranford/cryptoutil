package sysinfo

import (
	"errors"
	"fmt"
	"sync"

	cryptoutilUtil "cryptoutil/internal/util"
)

func GetAllInfo(sysInfoProvider SysInfoProvider) ([][]byte, error) {
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
		cpuVendorID, cpuFamily, cpuPhysicalID, cpuModelName, cpuErr = sysInfoProvider.CPUInfo()
	}()
	go func() {
		defer wg.Done()
		var ram uint64
		ram, ramErr = sysInfoProvider.RAMSize()
		ramSize = fmt.Sprintf("%d", ram)
	}()
	go func() {
		defer wg.Done()
		osHostname, osErr = sysInfoProvider.OSHostname()
	}()
	go func() {
		defer wg.Done()
		hostID, hostIDErr = sysInfoProvider.HostID()
	}()
	go func() {
		defer wg.Done()
		userID, groupID, username, userErr = sysInfoProvider.UserInfo()
	}()
	wg.Wait()

	if cpuErr != nil || ramErr != nil || osErr != nil || hostIDErr != nil || userErr != nil {
		return nil, errors.Join(cpuErr, ramErr, osErr, hostIDErr, userErr)
	}

	return cryptoutilUtil.StringPointersToBytes(&hostID, &userID, &groupID, &runtimeGoArch, &runtimeGoOS, &runtimeNumCPU, &cpuVendorID, &cpuFamily, &cpuPhysicalID, &cpuModelName, &ramSize, &osHostname, &username), nil
}
