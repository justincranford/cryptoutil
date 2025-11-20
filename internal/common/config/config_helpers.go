// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet("", pflag.ExitOnError)

	viper.Reset()
}

func registerSetting(setting *Setting) *Setting {
	allRegisteredSettings = append(allRegisteredSettings, setting)

	return setting
}

// Helper functions for safe type assertions in configuration.
func registerAsBoolSetting(s *Setting) bool {
	if v, ok := s.value.(bool); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not bool", s.name))
}

func registerAsStringSetting(s *Setting) string {
	if v, ok := s.value.(string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not string", s.name))
}

func registerAsUint16Setting(s *Setting) uint16 {
	if v, ok := s.value.(uint16); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not uint16", s.name))
}

func registerAsStringSliceSetting(s *Setting) []string {
	if v, ok := s.value.([]string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not []string", s.name))
}

func registerAsStringArraySetting(s *Setting) []string {
	if v, ok := s.value.([]string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not []string for array", s.name))
}

func registerAsDurationSetting(s *Setting) time.Duration {
	if v, ok := s.value.(time.Duration); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not time.Duration", s.name))
}

func registerAsIntSetting(s *Setting) int {
	if v, ok := s.value.(int); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not int", s.name))
}

func formatDefault(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case []string:
		if len(v) == 0 {
			return "[]"
		}

		return fmt.Sprintf("[%s]", strings.Join(v, ","))
	case bool:
		return fmt.Sprintf("%t", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func analyzeSettings(settings []*Setting) analysisResult {
	result := analysisResult{
		SettingsByNames:      make(map[string][]*Setting),
		SettingsByShorthands: make(map[string][]*Setting),
	}
	for _, setting := range settings {
		result.SettingsByNames[setting.name] = append(result.SettingsByNames[setting.name], setting)
		result.SettingsByShorthands[setting.shorthand] = append(result.SettingsByShorthands[setting.shorthand], setting)
	}

	for _, setting := range settings {
		if len(result.SettingsByNames[setting.name]) > 1 {
			result.DuplicateNames = append(result.DuplicateNames, setting.name)
		}

		if setting.shorthand != "" && len(result.SettingsByShorthands[setting.shorthand]) > 1 {
			result.DuplicateShorthands = append(result.DuplicateShorthands, setting.shorthand)
		}
	}

	return result
}
