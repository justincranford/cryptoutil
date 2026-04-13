// Copyright (c) 2025 Justin Cranford

package template_drift

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// templateComplianceFn is the function signature for template compliance checking.
// Production code uses defaultComplianceFn; tests can inject alternatives.
type templateComplianceFn func(projectRoot string) error

// CheckTemplateCompliance verifies all deployment artifacts match their canonical templates.
func CheckTemplateCompliance(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkTemplateComplianceInDir(logger, ".", defaultComplianceFn)
}

// checkTemplateComplianceInDir is the seam-injectable private function for testing.
func checkTemplateComplianceInDir(logger *cryptoutilCmdCicdCommon.Logger, projectRoot string, fn templateComplianceFn) error {
	logger.Log("Checking template compliance...")

	if err := fn(projectRoot); err != nil {
		return err
	}

	logger.Log("template-compliance: all deployment artifacts match canonical templates")

	return nil
}

// defaultComplianceFn loads templates, builds expected FS, and compares against actual files.
func defaultComplianceFn(projectRoot string) error {
	templates, err := LoadTemplatesDir(projectRoot)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	expected, err := BuildExpectedFS(templates)
	if err != nil {
		return fmt.Errorf("build expected FS: %w", err)
	}

	if err := CompareExpectedFS(expected, projectRoot); err != nil {
		return err
	}

	return nil
}
