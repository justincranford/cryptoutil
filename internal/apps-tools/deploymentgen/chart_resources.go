// Copyright (c) 2025-2026 Justin Cranford.
package deploymentgen

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func parseResourceRequirements(resources ResourceRequirements) (ResourceCapacity, ResourceCapacity, error) {
	requestCPU, err := parseCPUResource(resources.Requests.CPU)
	if err != nil {
		return ResourceCapacity{}, ResourceCapacity{}, fmt.Errorf("requests.cpu: %w", err)
	}

	limitCPU, err := parseCPUResource(resources.Limits.CPU)
	if err != nil {
		return ResourceCapacity{}, ResourceCapacity{}, fmt.Errorf("limits.cpu: %w", err)
	}

	requestMemory, err := parseMemoryResource(resources.Requests.Memory)
	if err != nil {
		return ResourceCapacity{}, ResourceCapacity{}, fmt.Errorf("requests.memory: %w", err)
	}

	limitMemory, err := parseMemoryResource(resources.Limits.Memory)
	if err != nil {
		return ResourceCapacity{}, ResourceCapacity{}, fmt.Errorf("limits.memory: %w", err)
	}

	if requestCPU > limitCPU || requestMemory > limitMemory {
		return ResourceCapacity{}, ResourceCapacity{}, fmt.Errorf("resource requests must not exceed limits")
	}

	return ResourceCapacity{CPUMilli: requestCPU, MemoryBytes: requestMemory},
		ResourceCapacity{CPUMilli: limitCPU, MemoryBytes: limitMemory}, nil
}

func parseCPUResource(value string) (int64, error) {
	if !cpuResourcePattern.MatchString(value) {
		return 0, fmt.Errorf("invalid positive CPU quantity %q", value)
	}

	multiplier := cpuMilliUnitsPerCore
	if strings.HasSuffix(value, "m") {
		multiplier = 1
		value = strings.TrimSuffix(value, "m")
	}

	quantity, err := strconv.ParseInt(value, 10, 64)
	if err != nil || quantity <= 0 || quantity > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("CPU quantity %q is out of range", value)
	}

	return quantity * multiplier, nil
}

func parseMemoryResource(value string) (int64, error) {
	if !memoryResourcePattern.MatchString(value) {
		return 0, fmt.Errorf("invalid positive memory quantity %q", value)
	}

	multiplier := int64(1)

	for suffix, valueMultiplier := range map[string]int64{
		runtimeMemorySuffixKibi: binaryQuantityKibi, runtimeMemorySuffixMebi: binaryQuantityMebi,
		runtimeMemorySuffixGibi: binaryQuantityGibi, runtimeMemorySuffixTebi: binaryQuantityTebi,
	} {
		if strings.HasSuffix(value, suffix) {
			multiplier = valueMultiplier
			value = strings.TrimSuffix(value, suffix)

			break
		}
	}

	quantity, err := strconv.ParseInt(value, 10, 64)
	if err != nil || quantity <= 0 || quantity > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("memory quantity %q is out of range", value)
	}

	return quantity * multiplier, nil
}
