// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	cryptoutilMaterializer "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// LifecycleSecret is an import request, never a Helm value or committed artifact.
// A runtime owner must create only absent Secrets or verify existing data before
// reconciling. This pure importer performs no Kubernetes mutation.
type LifecycleSecret = cryptoutilMaterializer.PublicationSecret

// LifecycleMetadata records the run boundary used by the importing owner.
type LifecycleMetadata = cryptoutilMaterializer.PublicationMetadata

// ImportLifecycleSecrets validates the complete contract before opening any
// protected files. All file paths must be rooted in the supplied filesystem;
// symlink components, world/group permissions, empty files, and oversized Secrets
// are rejected. Material is returned in memory only, never logged or persisted.
func ImportLifecycleSecrets(
	model Model, inventory Inventory, input LifecycleInput, source fs.FS,
) ([]LifecycleSecret, error) {
	values, err := BuildLifecycleChartValues(model, inventory, input)
	if err != nil {
		return nil, err
	}

	result := make([]LifecycleSecret, 0, len(values.Secrets))
	for _, binding := range values.Secrets {
		secret := LifecycleSecret{
			APIVersion: "v1", Kind: cryptoutilSharedMagic.KubernetesSecretKind, Type: binding.Type,
			Metadata: LifecycleMetadata{
				Name: binding.Name, Namespace: binding.Namespace,
				Labels: map[string]string{cryptoutilSharedMagic.KubernetesLabelRun: model.Run, "cu.io/owner": "external-secret-import"},
			},
			Data: make(map[string][]byte, len(binding.Files)),
		}
		total := 0

		for key, file := range binding.Files {
			data, err := readProtectedLifecycleFile(source, file)
			if err != nil {
				return nil, fmt.Errorf("import Secret %s/%s key %s: %w", binding.Namespace, binding.Name, key, err)
			}

			total += len(data)
			if total > lifecycleMaxSecretBytes {
				return nil, fmt.Errorf("secret %s/%s exceeds 1 MiB", binding.Namespace, binding.Name)
			}

			secret.Data[key] = data
		}

		result = append(result, secret)
	}

	return result, nil
}

func readProtectedLifecycleFile(source fs.FS, file string) (data []byte, returnErr error) {
	parent := lifecycleRootPath
	parts := strings.Split(file, "/")

	for index, part := range parts {
		entries, err := fs.ReadDir(source, parent)
		if err != nil {
			return nil, fmt.Errorf("inspect protected file parent: %w", err)
		}

		var selected fs.DirEntry

		for _, entry := range entries {
			if entry.Name() == part {
				selected = entry

				break
			}
		}

		if selected == nil || selected.Type()&fs.ModeSymlink != 0 ||
			(index < len(parts)-1 && !selected.IsDir()) {
			return nil, fmt.Errorf("protected file is missing or has a symlink/non-directory component")
		}

		parent = path.Join(parent, part)
	}

	handle, err := source.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open protected file: %w", err)
	}

	defer func() {
		if err := handle.Close(); err != nil && returnErr == nil {
			data = nil
			returnErr = fmt.Errorf("close protected file: %w", err)
		}
	}()

	info, err := handle.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect protected file: %w", err)
	}

	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() < 1 || info.Size() > lifecycleMaxSecretBytes {
		return nil, fmt.Errorf("protected file requires a nonempty regular owner-only file of at most 1 MiB")
	}

	data, err = io.ReadAll(io.LimitReader(handle, lifecycleMaxSecretBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read protected file: %w", err)
	}

	if len(data) == 0 || len(data) > lifecycleMaxSecretBytes {
		return nil, fmt.Errorf("protected file size changed outside supported bounds")
	}

	return data, nil
}
