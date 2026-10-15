package lint_deployments

import (
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

type composeVolume struct {
	Raw    string
	Type   string
	Source string
	Target string
}

func (volume *composeVolume) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		if err := node.Decode(&volume.Raw); err != nil {
			return fmt.Errorf("decode volume mount: %w", err)
		}

		volume.Source, volume.Target = splitShortVolumeMount(volume.Raw)
		volume.Type = inferVolumeMountType(volume.Source)
	case yaml.MappingNode:
		var mount struct {
			Type   string `yaml:"type"`
			Source string `yaml:"source"`
			Target string `yaml:"target"`
		}
		if err := node.Decode(&mount); err != nil {
			return fmt.Errorf("decode long volume mount: %w", err)
		}

		volume.Type = mount.Type
		volume.Source = mount.Source
		volume.Target = mount.Target

		volume.Raw = fmt.Sprintf("type=%s source=%s target=%s", mount.Type, mount.Source, mount.Target)
		if volume.Type == "" {
			volume.Type = inferVolumeMountType(volume.Source)
		}
	default:
		return fmt.Errorf("volume mount must be a string or mapping")
	}

	return nil
}

func splitShortVolumeMount(raw string) (string, string) {
	parts := strings.Split(raw, ":")
	for targetIndex := 1; targetIndex < len(parts); targetIndex++ {
		if strings.HasPrefix(parts[targetIndex], "/") {
			return strings.Join(parts[:targetIndex], ":"), parts[targetIndex]
		}
	}

	return "", ""
}

func inferVolumeMountType(source string) string {
	isWindowsPath := len(source) > 1 && source[1] == ':'

	isBindMount := source == cryptoutilSharedMagic.CurrentDirectory ||
		strings.HasPrefix(source, "./") ||
		strings.HasPrefix(source, "../") ||
		strings.HasPrefix(source, "/") ||
		strings.HasPrefix(source, "~") ||
		strings.HasPrefix(source, `\`) ||
		isWindowsPath
	if isBindMount {
		return cryptoutilSharedMagic.DockerComposeVolumeMountTypeBind
	}

	return cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume
}
