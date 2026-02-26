package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"fmt"
"os"
"strings"

"gopkg.in/yaml.v3"
)

// SchemaValidationResult holds the outcome of schema validation.
type SchemaValidationResult struct {
Path     string
Valid    bool
Errors   []string
Warnings []string
}

// fieldType represents the expected YAML value type for a config field.
type fieldType int

const (
// fieldTypeString is a string value (e.g., "https", "127.0.0.1").
fieldTypeString fieldType = iota
// fieldTypeInt is an integer value (e.g., 8080, 9090).
fieldTypeInt
// fieldTypeBool is a boolean value (e.g., true, false).
fieldTypeBool
// fieldTypeStringArray is an array of strings (e.g., CORS origins).
fieldTypeStringArray
)

// schemaField defines a single config field's schema.
type schemaField struct {
// Type is the expected YAML value type.
Type fieldType
// Required indicates whether this field must be present.
Required bool
// ValidValues restricts string fields to specific allowed values (nil = any string).
ValidValues []string
// Description documents the field purpose (used in error messages and code documentation).
Description string
}

// configSchema defines the comprehensive schema for flat kebab-case config files.
// These map to viper/pflag definitions used by the service template.
//
// Config key format: flat kebab-case (e.g., bind-public-address)
// NOT nested YAML (e.g., server.bind_address)
//
// Schema groups:
// 1. Public Server: bind-public-* (protocol, address, port)
// 2. Admin Server: bind-private-* (protocol, address, port)
// 3. TLS: tls-public-mode, tls-private-mode
// 4. OTLP Telemetry: otlp, otlp-service, otlp-environment, otlp-endpoint
// 5. CORS: cors-max-age, cors-allowed-origins
// 6. Session: browser-session-*, service-session-*
// 7. Database: database-url.
var configSchema = map[string]schemaField{
// Public Server Configuration.
// bind-public-protocol MUST be "https" (TLS required for all services).
"bind-public-protocol": {
Type:        fieldTypeString,
Required:    true,
ValidValues: []string{cryptoutilSharedMagic.ProtocolHTTPS},
Description: "Public server protocol (MUST be https)",
},
// bind-public-address: IPv4 address to bind on (0.0.0.0 for containers, 127.0.0.1 for local dev).
"bind-public-address": {
Type:     fieldTypeString,
Required: true,
Description: "Public server bind address (0.0.0.0 for containers, 127.0.0.1 for local)",
},
// bind-public-port: TCP port for public API (range 1-65535, typically 8000-8999 for services).
"bind-public-port": {
Type:     fieldTypeInt,
Required: true,
Description: "Public server bind port (1-65535)",
},

// Admin Server Configuration.
// bind-private-protocol MUST be "https" (TLS required for admin endpoints).
"bind-private-protocol": {
Type:        fieldTypeString,
Required:    true,
ValidValues: []string{cryptoutilSharedMagic.ProtocolHTTPS},
Description: "Admin server protocol (MUST be https)",
},
// bind-private-address MUST be "127.0.0.1" (admin never exposed outside container).
"bind-private-address": {
Type:        fieldTypeString,
Required:    true,
ValidValues: []string{cryptoutilSharedMagic.IPv4Loopback},
Description: "Admin server bind address (MUST be 127.0.0.1)",
},
// bind-private-port: TCP port for admin API (typically 9090).
"bind-private-port": {
Type:     fieldTypeInt,
Required: true,
Description: "Admin server bind port (typically 9090)",
},

// TLS Configuration.
// tls-public-mode: Certificate provisioning mode for public endpoint.
"tls-public-mode": {
Type:        fieldTypeString,
Required:    true,
ValidValues: []string{cryptoutilSharedMagic.DefaultTLSPublicMode, "manual"},
Description: "TLS certificate mode for public endpoint",
},
// tls-private-mode: Certificate provisioning mode for admin endpoint.
"tls-private-mode": {
Type:        fieldTypeString,
Required:    true,
ValidValues: []string{cryptoutilSharedMagic.DefaultTLSPublicMode, "manual"},
Description: "TLS certificate mode for admin endpoint",
},

// OTLP Telemetry Configuration.
// otlp: Master switch for OTLP telemetry export.
"otlp": {
Type:     fieldTypeBool,
Required: true,
Description: "Enable OTLP telemetry export",
},
// otlp-service: Service name reported to telemetry collector.
"otlp-service": {
Type:     fieldTypeString,
Required: false,
Description: "OTLP service name (required when otlp: true)",
},
// otlp-environment: Deployment environment label.
"otlp-environment": {
Type:        fieldTypeString,
Required:    false,
ValidValues: []string{"development", "production", "ci"},
Description: "OTLP environment label",
},
// otlp-endpoint: Collector endpoint URL (gRPC or HTTP).
"otlp-endpoint": {
Type:     fieldTypeString,
Required: false,
Description: "OTLP collector endpoint (required when otlp: true)",
},

// CORS Configuration.
// cors-max-age: Preflight cache duration in seconds.
"cors-max-age": {
Type:     fieldTypeInt,
Required: false,
Description: "CORS preflight cache duration in seconds",
},
// cors-allowed-origins: List of allowed CORS origins (HTTP/HTTPS URLs).
"cors-allowed-origins": {
Type:     fieldTypeStringArray,
Required: false,
Description: "Allowed CORS origins",
},

// Session Configuration.
// browser-session-algorithm: Session token format for browser clients.
"browser-session-algorithm": {
Type:        fieldTypeString,
Required:    false,
ValidValues: []string{cryptoutilSharedMagic.DefaultServiceSessionAlgorithm, string(cryptoutilSharedMagic.SessionAlgorithmJWE), "Opaque"},
Description: "Browser session token format",
},
// browser-session-jws-algorithm: JWS signing algorithm (when browser-session-algorithm: JWS).
"browser-session-jws-algorithm": {
Type:        fieldTypeString,
Required:    false,
ValidValues: []string{cryptoutilSharedMagic.JoseAlgHS256, cryptoutilSharedMagic.JoseAlgHS384, cryptoutilSharedMagic.JoseAlgHS512},
Description: "Browser JWS signing algorithm",
},
// browser-session-jwe-algorithm: JWE encryption algorithm (when browser-session-algorithm: JWE).
"browser-session-jwe-algorithm": {
Type:     fieldTypeString,
Required: false,
Description: "Browser JWE encryption algorithm",
},
// service-session-algorithm: Session token format for service clients.
"service-session-algorithm": {
Type:        fieldTypeString,
Required:    false,
ValidValues: []string{cryptoutilSharedMagic.DefaultServiceSessionAlgorithm, string(cryptoutilSharedMagic.SessionAlgorithmJWE), "Opaque"},
Description: "Service session token format",
},
// service-session-jws-algorithm: JWS signing algorithm (when service-session-algorithm: JWS).
"service-session-jws-algorithm": {
Type:        fieldTypeString,
Required:    false,
ValidValues: []string{cryptoutilSharedMagic.JoseAlgHS256, cryptoutilSharedMagic.JoseAlgHS384, cryptoutilSharedMagic.JoseAlgHS512},
Description: "Service JWS signing algorithm",
},
// service-session-jwe-algorithm: JWE encryption algorithm (when service-session-algorithm: JWE).
"service-session-jwe-algorithm": {
Type:     fieldTypeString,
Required: false,
Description: "Service JWE encryption algorithm",
},

// Database Configuration.
// database-url: Connection string or Docker secret reference.
// SHOULD use file:///run/secrets/ pattern (not inline credentials).
"database-url": {
Type:     fieldTypeString,
Required: false,
Description: "Database connection string (prefer file:///run/secrets/ reference)",
},
}

// ValidateSchema validates a flat kebab-case config file against the hardcoded schema.
// Checks: required fields present, value types correct, valid values match.
// For policy-level validation (bind addresses, ports, admin policy), see ValidateConfigFile.
func ValidateSchema(configPath string) (*SchemaValidationResult, error) {
result := &SchemaValidationResult{
Path:  configPath,
Valid: true,
}

data, err := os.ReadFile(configPath)
if err != nil {
result.Valid = false
result.Errors = append(result.Errors, fmt.Sprintf("cannot read file: %s", err))

return result, nil
}

var config map[string]any
if err := yaml.Unmarshal(data, &config); err != nil {
result.Valid = false
result.Errors = append(result.Errors, fmt.Sprintf("YAML parse error: %s", err))

return result, nil
}

if len(config) == 0 {
result.Warnings = append(result.Warnings, "config file is empty")

return result, nil
}

// Check required fields.
validateRequiredFields(config, result)

// Check field types and valid values.
validateFieldTypes(config, result)

// Check for unknown fields.
validateUnknownFields(config, result)

return result, nil
}

// validateRequiredFields checks that all required schema fields are present.
func validateRequiredFields(config map[string]any, result *SchemaValidationResult) {
for fieldName, schema := range configSchema {
if !schema.Required {
continue
}

if _, exists := config[fieldName]; !exists {
result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Required field '%s' missing (%s)", fieldName, schema.Description))
result.Valid = false
}
}
}

// validateFieldTypes checks that present fields have correct types and valid values.
func validateFieldTypes(config map[string]any, result *SchemaValidationResult) {
for fieldName, value := range config {
schema, known := configSchema[fieldName]
if !known {
continue // Unknown fields handled by validateUnknownFields.
}

switch schema.Type {
case fieldTypeString:
s, ok := value.(string)
if !ok {
result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Field '%s' must be a string, got %T", fieldName, value))
result.Valid = false

continue
}

validateStringEnum(fieldName, s, schema, result)
case fieldTypeInt:
if !isIntLike(value) {
result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Field '%s' must be an integer, got %T", fieldName, value))
result.Valid = false
}
case fieldTypeBool:
if _, ok := value.(bool); !ok {
result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Field '%s' must be a boolean, got %T", fieldName, value))
result.Valid = false
}
case fieldTypeStringArray:
validateStringArray(fieldName, value, result)
}
}
}

// validateStringEnum checks if a string value is in the allowed set.
func validateStringEnum(fieldName, value string, schema schemaField, result *SchemaValidationResult) {
if len(schema.ValidValues) == 0 {
return // No restriction on string values.
}

for _, valid := range schema.ValidValues {
if value == valid {
return
}
}

result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Field '%s' value '%s' not in allowed values: [%s]",
fieldName, value, strings.Join(schema.ValidValues, ", ")))
result.Valid = false
}

// validateStringArray verifies a field value is a slice of strings.
func validateStringArray(fieldName string, value any, result *SchemaValidationResult) {
arr, ok := value.([]any)
if !ok {
result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Field '%s' must be a string array, got %T", fieldName, value))
result.Valid = false

return
}

for i, item := range arr {
if _, ok := item.(string); !ok {
result.Errors = append(result.Errors, fmt.Sprintf(
"[ValidateSchema] Field '%s[%d]' must be a string, got %T", fieldName, i, item))
result.Valid = false
}
}
}

// validateUnknownFields warns about fields not in the schema.
func validateUnknownFields(config map[string]any, result *SchemaValidationResult) {
for fieldName := range config {
if _, known := configSchema[fieldName]; !known {
result.Warnings = append(result.Warnings, fmt.Sprintf(
"[ValidateSchema] Unknown field '%s' (not in schema)", fieldName))
}
}
}

// isIntLike checks if a YAML-parsed value is int-compatible.
// YAML parsers may decode integers as int, int64, or float64.
func isIntLike(v any) bool {
switch v.(type) {
case int, int64, float64:
return true
default:
return false
}
}

// FormatSchemaValidationResult formats a SchemaValidationResult for display.
func FormatSchemaValidationResult(result *SchemaValidationResult) string {
var sb strings.Builder

_, _ = fmt.Fprintf(&sb, "Schema Validation: %s\n", result.Path)

if result.Valid {
sb.WriteString("  Status: PASS\n")
} else {
sb.WriteString("  Status: FAIL\n")
}

for _, err := range result.Errors {
_, _ = fmt.Fprintf(&sb, "  ERROR: %s\n", err)
}

for _, warn := range result.Warnings {
_, _ = fmt.Fprintf(&sb, "  WARNING: %s\n", warn)
}

return sb.String()
}
