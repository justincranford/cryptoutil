package rotate_secret

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityRotation "cryptoutil/internal/identity/rotation"
)

func TestParseFlags_Success(t *testing.T) {
	t.Parallel()

	args := []string{
		"--client-id", "019ac131-1234-5678-9abc-def012345678",
		"--grace-period", "48h",
		"--reason", "Scheduled rotation",
		"--output", "json",
	}

	cfg, err := parseFlags(args)

	require.NoError(t, err)
	require.Equal(t, "019ac131-1234-5678-9abc-def012345678", cfg.ClientID)
	require.Equal(t, 48*time.Hour, cfg.GracePeriod)
	require.Equal(t, "Scheduled rotation", cfg.Reason)
	require.Equal(t, "json", cfg.OutputFormat)
}

func TestParseFlags_Defaults(t *testing.T) {
	t.Parallel()

	args := []string{
		"--client-id", "019ac131-1234-5678-9abc-def012345678",
	}

	cfg, err := parseFlags(args)

	require.NoError(t, err)
	require.Equal(t, "019ac131-1234-5678-9abc-def012345678", cfg.ClientID)
	require.Equal(t, 24*time.Hour, cfg.GracePeriod) // Default.
	require.Equal(t, "", cfg.Reason)
	require.Equal(t, "text", cfg.OutputFormat) // Default.
}

func TestParseFlags_InvalidDuration(t *testing.T) {
	t.Parallel()

	args := []string{
		"--client-id", "019ac131-1234-5678-9abc-def012345678",
		"--grace-period", "invalid",
	}

	_, err := parseFlags(args)

	require.Error(t, err)
	require.Contains(t, err.Error(), "parse error")
}

func TestValidateConfig_Success(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		ClientID:     "019ac131-1234-5678-9abc-def012345678",
		GracePeriod:  24 * time.Hour,
		Reason:       "Test rotation",
		OutputFormat: "text",
	}

	err := validateConfig(cfg)

	require.NoError(t, err)
}

func TestValidateConfig_MissingClientID(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		ClientID:     "",
		GracePeriod:  24 * time.Hour,
		OutputFormat: "text",
	}

	err := validateConfig(cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "--client-id is required")
}

func TestValidateConfig_InvalidUUID(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		ClientID:     "not-a-uuid",
		GracePeriod:  24 * time.Hour,
		OutputFormat: "text",
	}

	err := validateConfig(cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid client ID format")
}

func TestValidateConfig_NegativeGracePeriod(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		ClientID:     "019ac131-1234-5678-9abc-def012345678",
		GracePeriod:  -1 * time.Hour,
		OutputFormat: "text",
	}

	err := validateConfig(cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "grace period must be non-negative")
}

func TestValidateConfig_InvalidOutputFormat(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		ClientID:     "019ac131-1234-5678-9abc-def012345678",
		GracePeriod:  24 * time.Hour,
		OutputFormat: "xml",
	}

	err := validateConfig(cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "output format must be 'text' or 'json'")
}

func TestOutputJSON(t *testing.T) {
	t.Parallel()

	eventID := googleUuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0f10")
	result := &cryptoutilIdentityRotation.RotateClientSecretResult{
		OldVersion:         1,
		NewVersion:         2,
		NewSecretPlaintext: "test-secret-base64",
		GracePeriodEnd:     time.Date(2025, 11, 27, 12, 0, 0, 0, time.UTC),
		EventID:            eventID,
	}

	// Output to discard (test structure, not actual output).
	err := outputJSON(result)

	require.NoError(t, err)
}

func TestOutputText(t *testing.T) {
	t.Parallel()

	eventID := googleUuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0f10")
	result := &cryptoutilIdentityRotation.RotateClientSecretResult{
		OldVersion:         1,
		NewVersion:         2,
		NewSecretPlaintext: "test-secret-base64",
		GracePeriodEnd:     time.Date(2025, 11, 27, 12, 0, 0, 0, time.UTC),
		EventID:            eventID,
	}

	// Output to discard (test structure, not actual output).
	err := outputText(result)

	require.NoError(t, err)
}
