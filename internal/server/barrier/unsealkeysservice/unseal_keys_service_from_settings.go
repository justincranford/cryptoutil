package unsealkeysservice

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeysServiceFromSettings struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealKeysServiceFromSettings) EncryptKey(clearRootKey joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks, clearRootKey)
}

func (u *UnsealKeysServiceFromSettings) DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks, encryptedRootKeyBytes)
}

func (u *UnsealKeysServiceFromSettings) Shutdown() {
	u.unsealJwks = nil
}

// readJWKFile reads a JWK from a file
func readJWKFile(filePath string) (joseJwk.Key, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWK file %s: %w", filePath, err)
	}

	key, err := joseJwk.ParseKey(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWK from file %s: %w", filePath, err)
	}

	return key, nil
}

func NewUnsealKeysServiceFromSettings(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) (UnsealKeysService, error) {
	telemetryService.Slogger.Info("Creating UnsealKeysService from settings", "mode", settings.UnsealMode, "files", settings.UnsealFiles)
	// Parse mode - could be "N", "M-of-N", or "sysinfo"
	switch {
	case settings.UnsealMode == "":
		return NewUnsealKeysServiceFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
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

		filesContents, err := readFilesContents(&settings.UnsealFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to read shared secrets files: %w", err)
		}
		if len(filesContents) != n {
			return nil, fmt.Errorf("expected %d shared secret files, got %d", n, len(filesContents))
		}

		return NewUnsealKeysServiceSharedSecrets(filesContents, m)

	default:
		n, err := strconv.Atoi(settings.UnsealMode) // Try to parse as a number (N mode)
		if err != nil {
			return nil, fmt.Errorf("invalid unseal mode %s: %w", settings.UnsealMode, err)
		}

		// Split the file list
		fileList := strings.Split(settings.UnsealFiles, ",")
		if len(fileList) == 0 {
			return nil, fmt.Errorf("no unseal files specified for mode %s", settings.UnsealMode)
		}
		if len(fileList) != n {
			return nil, fmt.Errorf("expected %d files for N mode, got %d", n, len(fileList))
		}

		// Read all JWKs
		unsealJwks := make([]joseJwk.Key, 0, len(fileList))
		for _, filePath := range fileList {
			filePath = strings.TrimSpace(filePath)
			if filePath == "" {
				continue
			}

			jwk, err := readJWKFile(filePath)
			if err != nil {
				return nil, err
			}
			unsealJwks = append(unsealJwks, jwk)
		}

		if len(unsealJwks) == 0 {
			return nil, fmt.Errorf("no valid JWK files found")
		}

		// Verify we have enough keys
		if n > 0 && len(unsealJwks) < n {
			return nil, fmt.Errorf("insufficient JWKs: required %d, found %d", n, len(unsealJwks))
		}

		// Use all the JWKs we found
		return NewUnsealKeysServiceSimple(unsealJwks)
	}
}

func readFilesContents(filePaths *string) ([][]byte, error) {
	fileList := strings.Split(*filePaths, ",")
	if len(fileList) == 0 {
		return nil, fmt.Errorf("no files specified")
	}

	for i, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return nil, fmt.Errorf("empty file path %d of %d in list", i+1, len(fileList))
		}
	}

	filesContents := make([][]byte, 0, len(fileList))
	for i, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		fileContents, err := readFileContents(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %d of %d (%s): %w", i+1, len(fileList), filePath, err)
		}
		filesContents = append(filesContents, fileContents)
	}

	return filesContents, nil
}

func readFileContents(filePath string) ([]byte, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return fileData, nil
}
