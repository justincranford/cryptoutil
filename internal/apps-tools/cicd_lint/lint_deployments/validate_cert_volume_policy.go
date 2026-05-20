package lint_deployments

import (
	"fmt"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	composeFileName                 = "compose.yml"
	certBindMountPattern            = "./certs:/certs"
	templateCertVolumeName          = "__PS_ID__-certs"
	certVolumeMountPathPrefix       = ":/certs"
	certVolumeInfraMountPathPrefix  = ":/mnt/ps-certs-src"
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
			if strings.Contains(volume, certBindMountPattern) {
				result.Errors = append(result.Errors,
					fmt.Sprintf("[ValidateCertVolumePolicy] service '%s' uses forbidden cert bind mount '%s' | See: %s",
						serviceName, volume, certVolumePolicyHandbookSection))
				result.Valid = false
			}
		}
	}

	if !hasCertVolumeMountReference(compose, expectedVolumeName) {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateCertVolumePolicy] missing named cert volume mount reference '%s%s' | See: %s",
				expectedVolumeName, certVolumeMountPathPrefix, certVolumePolicyHandbookSection))
		result.Valid = false
	}

	if !hasTopLevelCertVolumeDeclaration(compose, expectedVolumeName) {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateCertVolumePolicy] missing top-level named volume declaration '%s:' | See: %s",
				expectedVolumeName, certVolumePolicyHandbookSection))
		result.Valid = false
	}

	return result, nil
}

func expectedCertVolumeName(deploymentName string) string {
	if deploymentName == cryptoutilSharedMagic.SkeletonTemplateServiceName {
		return templateCertVolumeName
	}

	return deploymentName + "-certs"
}

func hasCertVolumeMountReference(compose *composeFile, expectedVolumeName string) bool {
	expectedPrefixes := []string{
		expectedVolumeName + certVolumeMountPathPrefix,
		expectedVolumeName + certVolumeInfraMountPathPrefix,
	}

	for _, serviceName := range sortedServiceNames(compose) {
		svc := compose.Services[serviceName]
		for _, volume := range svc.Volumes {
			for _, expectedPrefix := range expectedPrefixes {
				if strings.Contains(volume, expectedPrefix) {
					return true
				}
			}
		}
	}

	return false
}

func hasTopLevelCertVolumeDeclaration(compose *composeFile, expectedVolumeName string) bool {
	if compose.Volumes == nil {
		return false
	}

	_, ok := compose.Volumes[expectedVolumeName]

	return ok
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
