package spa

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testDefaultConfigPath = "default-config.yml"
	testConfigPath        = "/etc/cryptoutil/config.yml"
	testConfigFileURI     = cryptoutilSharedMagic.FileURIScheme + testConfigPath
	testUnrelatedFlag     = "--other"
	testAdminPortFlag     = "--admin-port"
	testAdminPortValue    = "12345"
	testInvalidAdminPort  = "not-a-port"
	testOverflowAdminPort = "999999999999999999999999999999"
	testDefaultAdminPort  = 9091
	testExpectedAdminPort = 12345
)

func TestParseConfigFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parameters []string
		want       string
	}{
		{name: "default config path when absent", want: testDefaultConfigPath},
		{name: "long equals flag", parameters: []string{cryptoutilSharedMagic.IdentityCLIFlagConfig + "=" + testConfigPath}, want: testConfigPath},
		{name: "long equals file URI", parameters: []string{cryptoutilSharedMagic.IdentityCLIFlagConfig + "=" + testConfigFileURI}, want: testConfigPath},
		{name: "short equals flag", parameters: []string{configFlagShort + "=" + testConfigPath}, want: testConfigPath},
		{name: "short equals file URI", parameters: []string{configFlagShort + "=" + testConfigFileURI}, want: testConfigPath},
		{name: "long separate flag", parameters: []string{cryptoutilSharedMagic.IdentityCLIFlagConfig, testConfigPath}, want: testConfigPath},
		{name: "long separate file URI", parameters: []string{cryptoutilSharedMagic.IdentityCLIFlagConfig, testConfigFileURI}, want: testConfigPath},
		{name: "short separate flag", parameters: []string{configFlagShort, testConfigPath}, want: testConfigPath},
		{name: "short separate file URI", parameters: []string{configFlagShort, testConfigFileURI}, want: testConfigPath},
		{name: "flag after unrelated argument", parameters: []string{testUnrelatedFlag, cryptoutilSharedMagic.IdentityCLIFlagConfig, testConfigPath}, want: testConfigPath},
		{name: "unmatched config flag", parameters: []string{testUnrelatedFlag}, want: testDefaultConfigPath},
		{name: "long flag without value", parameters: []string{cryptoutilSharedMagic.IdentityCLIFlagConfig}, want: testDefaultConfigPath},
		{name: "short flag without value", parameters: []string{configFlagShort}, want: testDefaultConfigPath},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, parseConfigFlag(test.parameters, testDefaultConfigPath))
		})
	}
}

func TestParseAdminPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parameters []string
		want       int
	}{
		{name: "default admin port when absent", want: testDefaultAdminPort},
		{name: "valid port", parameters: []string{testAdminPortFlag, testAdminPortValue}, want: testExpectedAdminPort},
		{name: "valid port after unrelated argument", parameters: []string{testUnrelatedFlag, testAdminPortFlag, testAdminPortValue}, want: testExpectedAdminPort},
		{name: "invalid port", parameters: []string{testAdminPortFlag, testInvalidAdminPort}, want: testDefaultAdminPort},
		{name: "overflowing port", parameters: []string{testAdminPortFlag, testOverflowAdminPort}, want: testDefaultAdminPort},
		{name: "flag without value", parameters: []string{testAdminPortFlag}, want: testDefaultAdminPort},
		{name: "unmatched admin port flag", parameters: []string{testUnrelatedFlag, testAdminPortValue}, want: testDefaultAdminPort},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, parseAdminPort(test.parameters, testDefaultAdminPort))
		})
	}
}
