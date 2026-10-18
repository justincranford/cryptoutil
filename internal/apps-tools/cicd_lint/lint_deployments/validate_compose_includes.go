package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// composeInclude represents an include directive.
type composeInclude struct {
	Path []string `yaml:"path"`
}

func (include *composeInclude) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("include must be a mapping")
	}

	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value != "path" {
			continue
		}

		pathNode := node.Content[index+1]
		switch pathNode.Kind {
		case yaml.ScalarNode:
			var path string
			if err := pathNode.Decode(&path); err != nil {
				return fmt.Errorf("decode include path: %w", err)
			}

			include.Path = []string{path}
		case yaml.SequenceNode:
			for _, includedPathNode := range pathNode.Content {
				var path string
				if err := includedPathNode.Decode(&path); err != nil {
					return fmt.Errorf("decode include path: %w", err)
				}

				include.Path = append(include.Path, path)
			}
		default:
			return fmt.Errorf("include path must be a string or sequence of strings")
		}

		return nil
	}

	return fmt.Errorf("include mapping is missing a path")
}

// parseComposeWithIncludes parses a compose file and merges resources from included files.
func parseComposeWithIncludes(composePath string) (*composeFile, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file %s: %w", composePath, err)
	}

	var root composeFile
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("YAML parse error in %s: %w", composePath, err)
	}

	baseDir := filepath.Dir(composePath)
	setComposeServiceDir(&root, baseDir)

	compose := &composeFile{}
	includeStack := make(map[string]struct{})

	for _, inc := range root.Include {
		for _, includePath := range inc.Path {
			mergeIncludedFileRecursive(baseDir, includePath, compose, includeStack)
		}
	}

	mergeComposeResources(compose, &root)

	return compose, nil
}

// mergeIncludedFile reads an included compose file and merges its secrets/services.
func mergeIncludedFile(baseDir string, includePath string, compose *composeFile) {
	mergeIncludedFileRecursive(baseDir, includePath, compose, make(map[string]struct{}))
}

func mergeIncludedFileRecursive(baseDir string, includePath string, compose *composeFile, includeStack map[string]struct{}) {
	if includePath == "" {
		return
	}

	fullPath := filepath.Clean(filepath.Join(baseDir, includePath))
	if _, exists := includeStack[fullPath]; exists {
		return
	}

	includeData, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	var source composeFile
	if err := yaml.Unmarshal(includeData, &source); err != nil {
		return
	}

	includedBaseDir := filepath.Dir(fullPath)
	setComposeServiceDir(&source, includedBaseDir)

	includeStack[fullPath] = struct{}{}
	defer delete(includeStack, fullPath)

	included := &composeFile{}

	for _, nestedInclude := range source.Include {
		for _, nestedIncludePath := range nestedInclude.Path {
			mergeIncludedFileRecursive(includedBaseDir, nestedIncludePath, included, includeStack)
		}
	}

	mergeComposeResources(included, &source)
	mergeComposeResources(compose, included)
}

func mergeComposeResources(destination, source *composeFile) {
	if len(source.Secrets) > 0 {
		if destination.Secrets == nil {
			destination.Secrets = make(map[string]composeSecret)
		}

		for name, secret := range source.Secrets {
			destination.Secrets[name] = secret
		}
	}

	if len(source.Volumes) > 0 {
		if destination.Volumes == nil {
			destination.Volumes = make(map[string]any)
		}

		for name, volume := range source.Volumes {
			destination.Volumes[name] = volume
		}
	}

	if len(source.Networks) > 0 {
		if destination.Networks == nil {
			destination.Networks = make(map[string]any)
		}

		for name, network := range source.Networks {
			destination.Networks[name] = network
		}
	}

	if len(source.Services) > 0 {
		if destination.Services == nil {
			destination.Services = make(map[string]composeService)
		}

		for name, service := range source.Services {
			if existing, exists := destination.Services[name]; exists {
				destination.Services[name] = mergeComposeService(existing, service)
			} else {
				destination.Services[name] = service
			}
		}
	}
}

func mergeComposeService(existing, override composeService) composeService {
	if override.Image != "" {
		existing.Image = override.Image
	}

	if override.Build != nil {
		existing.Build = override.Build
	}

	if len(override.Ports) > 0 {
		existing.Ports = override.Ports
	}

	existing.Volumes = mergeComposeVolumes(existing.Volumes, override.Volumes)
	existing.Environment = mergeComposeEnvironment(existing.Environment, override.Environment)

	if override.EnvFile != nil {
		existing.EnvFile = override.EnvFile
		existing.composeDir = override.composeDir
	}

	existing.Secrets = mergeComposeSecretRefs(existing.Secrets, override.Secrets)
	existing.DependsOn = mergeComposeDependsOn(existing.DependsOn, override.DependsOn)

	if override.Healthcheck != nil {
		existing.Healthcheck = override.Healthcheck
	}

	if override.Networks != nil {
		existing.Networks = override.Networks
	}

	if override.Command != nil {
		existing.Command = override.Command
	}

	if override.Deploy != nil {
		existing.Deploy = override.Deploy
	}

	if len(override.Profiles) > 0 {
		existing.Profiles = override.Profiles
	}

	if override.WorkingDir != "" {
		existing.WorkingDir = override.WorkingDir
	}

	if override.Entrypoint != nil {
		existing.Entrypoint = override.Entrypoint
	}

	if override.ShmSize != "" {
		existing.ShmSize = override.ShmSize
	}

	return existing
}

func mergeComposeVolumes(existing, override []composeVolume) []composeVolume {
	merged := append([]composeVolume(nil), existing...)

	for _, volume := range override {
		replaced := false

		for index, existingVolume := range merged {
			if existingVolume.Target != "" && existingVolume.Target == volume.Target {
				merged[index] = volume
				replaced = true

				break
			}
		}

		if !replaced {
			merged = append(merged, volume)
		}
	}

	return merged
}

func mergeComposeEnvironment(existing, override any) any {
	if override == nil {
		return existing
	}

	existingMap, existingIsMap := existing.(map[string]any)

	overrideMap, overrideIsMap := override.(map[string]any)
	if !existingIsMap || !overrideIsMap {
		return override
	}

	merged := make(map[string]any, len(existingMap)+len(overrideMap))
	for key, value := range existingMap {
		merged[key] = value
	}

	for key, value := range overrideMap {
		merged[key] = value
	}

	return merged
}

func mergeComposeSecretRefs(existing, override []any) []any {
	merged := append([]any(nil), existing...)

	for _, secret := range override {
		duplicate := false

		for _, existingSecret := range merged {
			if fmt.Sprintf("%v", existingSecret) == fmt.Sprintf("%v", secret) {
				duplicate = true

				break
			}
		}

		if !duplicate {
			merged = append(merged, secret)
		}
	}

	return merged
}

func setComposeServiceDir(compose *composeFile, dir string) {
	for name, service := range compose.Services {
		service.composeDir = dir
		compose.Services[name] = service
	}
}

func mergeComposeDependsOn(existing, included any) any {
	if existing == nil {
		return included
	}

	existingMap, existingIsMap := existing.(map[string]any)

	includedMap, includedIsMap := included.(map[string]any)
	if !existingIsMap || !includedIsMap {
		return existing
	}

	merged := make(map[string]any, len(existingMap)+len(includedMap))
	for dependency, configuration := range existingMap {
		merged[dependency] = configuration
	}

	for dependency, configuration := range includedMap {
		merged[dependency] = configuration
	}

	return merged
}
