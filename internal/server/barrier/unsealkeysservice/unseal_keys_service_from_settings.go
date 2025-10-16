package unsealkeysservice

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const (
	MaxFiles        = 9
	MaxBytesPerFile = 4096
)

type UnsealKeysServiceFromSettings struct {
	unsealJWKs []joseJwk.Key
}

func (u *UnsealKeysServiceFromSettings) EncryptKey(clearRootKey joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJWKs, clearRootKey)
}

func (u *UnsealKeysServiceFromSettings) DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJWKs, encryptedRootKeyBytes)
}

func (u *UnsealKeysServiceFromSettings) Shutdown() {
	u.unsealJWKs = nil
}

func NewUnsealKeysServiceFromSettings(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) (UnsealKeysService, error) {
	if settings.DevMode { // Generate random unseal key for dev mode
		randomBytes, err := cryptoutilUtil.GenerateBytes(64)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes for dev mode: %w", err)
		}

		telemetryService.Slogger.Debug("Generated random unseal secret for dev mode", "length", len(randomBytes))
		sharedSecretsM := [][]byte{randomBytes}

		return NewUnsealKeysServiceSharedSecrets(sharedSecretsM, len(sharedSecretsM))
	}

	if telemetryService.VerboseMode {
		telemetryService.Slogger.Info("Creating UnsealKeysService from settings", "mode", settings.UnsealMode, "files", settings.UnsealFiles)
	}
	// Parse mode - could be "N", "M-of-N", or "sysinfo"
	switch {
	case settings.UnsealMode == "sysinfo":
		return NewUnsealKeysServiceFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})

	case strings.Contains(settings.UnsealMode, "-of-"):
		parts := strings.Split(settings.UnsealMode, "-of-") // M-of-N mode - shared secrets
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid unseal mode format: %s, expected M-of-N", settings.UnsealMode)
		}

		m, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid M value in unseal mode %s: %w", settings.UnsealMode, err)
		}

		n, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid N value in unseal mode %s: %w", settings.UnsealMode, err)
		}

		if m <= 0 || n <= 0 || m > n {
			return nil, fmt.Errorf("invalid M-of-N values in unseal mode %s: M must be > 0, N must be >= M", settings.UnsealMode)
		}

		filesContents, err := cryptoutilUtil.ReadFilesBytesLimit(settings.UnsealFiles, MaxFiles, MaxBytesPerFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read shared secrets files: %w", err)
		} else if len(filesContents) != n {
			return nil, fmt.Errorf("expected %d shared secret files, got %d", n, len(filesContents))
		}

		return NewUnsealKeysServiceSharedSecrets(filesContents, m)
	default:
		n, err := strconv.Atoi(settings.UnsealMode) // Try to parse as a number (N mode)
		if err != nil {
			return nil, fmt.Errorf("invalid unseal mode %s: %w", settings.UnsealMode, err)
		}

		if n <= 0 {
			return nil, fmt.Errorf("invalid unseal mode %s: N must be > 0", settings.UnsealMode)
		}

		filesContents, err := cryptoutilUtil.ReadFilesBytesLimit(settings.UnsealFiles, MaxFiles, MaxBytesPerFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read shared secrets files: %w", err)
		} else if len(filesContents) != n {
			return nil, fmt.Errorf("expected %d shared secret files, got %d", n, len(filesContents))
		}

		unsealJWKs := make([]joseJwk.Key, 0, len(filesContents))

		for _, fileContents := range filesContents {
			jwk, err := joseJwk.ParseKey(fileContents)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JWK from file contents: %w", err)
			}

			unsealJWKs = append(unsealJWKs, jwk)
		}

		return NewUnsealKeysServiceSimple(unsealJWKs)
	}
}
