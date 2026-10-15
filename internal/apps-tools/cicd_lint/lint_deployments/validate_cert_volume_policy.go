package lint_deployments

import (
	"fmt"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	composeFileName                 = "compose.yml"
	templateCertVolumeName          = "__PS_ID__-certs"
	certVolumeMountTarget           = "/certs"
	certVolumeInfraMountTarget      = "/mnt/ps-certs-src"
	certVolumePolicyHandbookSection = "ENG-HANDBOOK.md Section 12.1"
)

// CertVolumePolicyResult contains cert volume policy validation output.
type CertVolumePolicyResult struct {
	Path   string
	Valid  bool
	Errors []string
}

// ValidateCertVolumePolicy enforces CO-21/CO-22 cert volume requirements.
func ValidateCertVolumePolicy(deploymentPath, deploymentName string) (*CertVolumePolicyResult, error) {
	result := &CertVolumePolicyResult{
		Path:  deploymentPath,
		Valid: true,
	}

	composePath := filepath.Join(deploymentPath, composeFileName)

	compose, err := parseComposeWithIncludes(composePath)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateCertVolumePolicy] Failed to parse compose file: %v", err))
		result.Valid = false

		return result, nil
	}

	expectedVolumeName := expectedCertVolumeName(deploymentName)

	for _, serviceName := range sortedServiceNames(compose) {
		svc := compose.Services[serviceName]
		for _, volume := range svc.Volumes {
			if isCertVolumeTarget(volume.Target) &&
				(volume.Type == cryptoutilSharedMagic.DockerComposeVolumeMountTypeBind || volume.Source != expectedVolumeName) {
				result.Errors = append(result.Errors,
					fmt.Sprintf("[ValidateCertVolumePolicy] service '%s' uses forbidden cert bind mount or unexpected cert volume '%s' | See: %s",
						serviceName, volume.Raw, certVolumePolicyHandbookSection))
				result.Valid = false
			}
		}
	}

	if !hasCertVolumeMountReference(compose, expectedVolumeName) {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateCertVolumePolicy] missing named cert volume mount reference '%s%s' | See: %s",
				expectedVolumeName, ":"+certVolumeMountTarget, certVolumePolicyHandbookSection))
		result.Valid = false
	}

	if !hasTopLevelCertVolumeDeclaration(compose, expectedVolumeName) {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateCertVolumePolicy] missing top-level named volume declaration '%s:' | See: %s",
				expectedVolumeName, certVolumePolicyHandbookSection))
		result.Valid = false
	}

	validateCertVolumeActivationBarrier(compose, expectedVolumeName, result)

	return result, nil
}

func expectedCertVolumeName(deploymentName string) string {
	if deploymentName == cryptoutilSharedMagic.SkeletonTemplateServiceName {
		return templateCertVolumeName
	}

	return deploymentName + "-certs"
}

func hasCertVolumeMountReference(compose *composeFile, expectedVolumeName string) bool {
	for _, serviceName := range sortedServiceNames(compose) {
		svc := compose.Services[serviceName]
		for _, volume := range svc.Volumes {
			if volume.Source == expectedVolumeName &&
				volume.Type == cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume &&
				isCertVolumeTarget(volume.Target) {
				return true
			}
		}
	}

	return false
}

func isCertVolumeTarget(target string) bool {
	return target == certVolumeMountTarget || target == certVolumeInfraMountTarget
}

func hasTopLevelCertVolumeDeclaration(compose *composeFile, expectedVolumeName string) bool {
	if compose.Volumes == nil {
		return false
	}

	_, ok := compose.Volumes[expectedVolumeName]

	return ok
}

func validateCertVolumeActivationBarrier(compose *composeFile, expectedVolumeName string, result *CertVolumePolicyResult) {
	for _, serviceName := range sortedServiceNames(compose) {
		if serviceName == cryptoutilSharedMagic.PSIDPKIInit {
			continue
		}

		svc := compose.Services[serviceName]
		if !usesNamedCertVolume(&svc, expectedVolumeName) {
			continue
		}

		if !dependsOnSuccessfulPKIInit(compose, serviceName, make(map[string]struct{})) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("[ValidateCertVolumePolicy] service '%s' mounts cert volume '%s' without depending on '%s' with condition 'service_completed_successfully' | See: %s",
					serviceName, expectedVolumeName, cryptoutilSharedMagic.PSIDPKIInit, certVolumePolicyHandbookSection))
			result.Valid = false
		}
	}
}

func usesNamedCertVolume(svc *composeService, expectedVolumeName string) bool {
	for _, volume := range svc.Volumes {
		if volume.Source == expectedVolumeName && volume.Type == cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume {
			return true
		}
	}

	return false
}

func dependsOnSuccessfulPKIInit(compose *composeFile, serviceName string, visited map[string]struct{}) bool {
	if serviceName == cryptoutilSharedMagic.PSIDPKIInit {
		return true
	}

	if _, exists := visited[serviceName]; exists {
		return false
	}

	visited[serviceName] = struct{}{}

	svc := compose.Services[serviceName]

	dependencies, ok := svc.DependsOn.(map[string]any)
	if !ok {
		return false
	}

	for dependencyName, dependencyConfig := range dependencies {
		dependency, ok := dependencyConfig.(map[string]any)
		if !ok {
			continue
		}

		condition, ok := dependency["condition"].(string)
		if ok && condition == "service_completed_successfully" &&
			dependsOnSuccessfulPKIInit(compose, dependencyName, visited) {
			return true
		}
	}

	return false
}

// FormatCertVolumePolicyResult renders cert volume policy validation output.
func FormatCertVolumePolicyResult(result *CertVolumePolicyResult) string {
	if result == nil {
		return "[ValidateCertVolumePolicy] nil result"
	}

	if result.Valid {
		return fmt.Sprintf("[ValidateCertVolumePolicy] PASS: %s", result.Path)
	}

	if len(result.Errors) == 0 {
		return fmt.Sprintf("[ValidateCertVolumePolicy] FAIL: %s", result.Path)
	}

	return strings.Join(result.Errors, "\n")
}
