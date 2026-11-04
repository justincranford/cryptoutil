// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const (
	Command              = "pki-materialize"
	PodNameEnvironment   = "POD_NAME"
	NamespaceEnvironment = "POD_NAMESPACE"
	InstancePath         = "/etc/cryptoutil/instance/instance.yaml"
	InputPath            = "/etc/cryptoutil/instance/materializer.json"
	TrustDirectory       = "/run/cryptoutil/trust"
	rolePublic           = "public-server"
	roleAdmin            = "admin-server"
	rolePostgres         = "postgres-client"
	SuiteExecutable      = cryptoutilSharedMagic.KubernetesSuiteImageBinary
)

// CommandInput is non-secret. Trust files refer only to the explicit issuer
// mount; identities are operator inputs, not discovered from tokens or leaves.
type CommandInput struct {
	Namespace     string            `json:"namespace"`
	APIEndpoint   string            `json:"apiEndpoint"`
	APIServerName string            `json:"apiServerName"`
	TrustFiles    map[string]string `json:"trustFiles"`
	Pods          []CommandPod      `json:"pods"`
}

type CommandPod struct {
	PodName                  string                             `json:"podName"`
	ExpectedAdminDNS         string                             `json:"expectedAdminDNS"`
	ExpectedPublicDNSNames   []string                           `json:"expectedPublicDNSNames"`
	PostgreSQLClientBindings map[string]PostgreSQLClientBinding `json:"postgresqlClientBindings,omitempty"`
}

// DecodeCommandInput rejects unknown, duplicate and trailing properties.
func DecodeCommandInput(raw []byte) (CommandInput, error) {
	var input CommandInput
	if len(raw) == 0 || len(raw) > cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes {
		return input, errors.New("materializer input size rejected")
	}

	var unique map[string]any
	if err := yaml.Unmarshal(raw, &unique); err != nil {
		return input, errors.New("materializer input duplicate or malformed properties")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return CommandInput{}, errors.New("materializer input invalid")
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return CommandInput{}, errors.New("materializer input trailing content")
	}

	return input, nil
}

// ValidateCommandInput checks all reserved pods before any Secret/token reads.
func ValidateCommandInput(input CommandInput, config InstanceConfig) error {
	if err := config.validate(); err != nil {
		return err
	}

	if !validDNSLabel(input.Namespace) {
		return errors.New("materializer input namespace invalid")
	}

	if _, err := validateAPIEndpoint(input.APIEndpoint, input.APIServerName); err != nil {
		return err
	}

	required := []string{rolePublic, roleAdmin, roleHealth, roleOTLP}
	if config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
		required = append(required, rolePostgres)
	}

	if len(input.TrustFiles) != len(required) {
		return errors.New("materializer trust files must exactly cover pool roles")
	}

	for _, role := range required {
		file := input.TrustFiles[role]
		if !strings.HasPrefix(file, TrustDirectory+"/") || path.Clean(file) != file ||
			!safeConfigPath(strings.TrimPrefix(file, TrustDirectory+"/")) {
			return errors.New("materializer trust file must be explicit under issuer mount")
		}
	}

	if len(input.Pods) != len(config.InstanceServices) {
		return errors.New("materializer input must cover every reserved pod")
	}

	seen := make(map[string]bool, len(input.Pods))
	for _, pod := range input.Pods {
		if seen[pod.PodName] {
			return errors.New("materializer input duplicate pod")
		}

		seen[pod.PodName] = true

		service, err := SelectPodService(config, pod.PodName)
		if err != nil {
			return err
		}

		admin, err := expectedAdminServerDNS(config, service, input.Namespace)
		if err != nil || admin != pod.ExpectedAdminDNS {
			return errors.New("materializer input admin identity mismatch")
		}

		if _, err := validateExpectedPublicDNSNames(config, service, input.Namespace, pod.ExpectedPublicDNSNames); err != nil {
			return err
		}

		if config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
			if err := validatePostgreSQLClientBindings(pod.PostgreSQLClientBindings, config.InstanceServices); err != nil {
				return err
			}
		} else if len(pod.PostgreSQLClientBindings) != 0 {
			return errors.New("SQLite materializer must not have PostgreSQL bindings")
		}
	}

	return nil
}

// Main is routed by the existing cryptoutil suite; no standalone binary.
func Main(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return MainWithDependencies(args, stdout, stderr, os.Getenv, boundedCommandRead,
		func(ctx context.Context, raw []byte, inputs RuntimeInputs, trust TrustInputs) error {
			return New(Dependencies{}).Materialize(ctx, raw, inputs, trust)
		})
}

// MainWithDependencies exposes parallel-safe process boundaries for tests.
func MainWithDependencies(args []string, stdout, stderr io.Writer, getenv func(string) string,
	read func(string) ([]byte, error), materialize func(context.Context, []byte, RuntimeInputs, TrustInputs) error,
) int {
	flags := flag.NewFlagSet(Command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	instance := flags.String("instance", "", "explicit instance.yaml path")

	inputPath := flags.String("input", "", "explicit non-secret materializer input path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(stdout, "cryptoutil pki-materialize --instance="+InstancePath+" --input="+InputPath)
			if err == nil {
				return 0
			}
		}

		return commandFailure(stderr, "materializer flags rejected")
	}

	if flags.NArg() != 0 || *instance != InstancePath || *inputPath != InputPath {
		return commandFailure(stderr, "materializer requires fixed instance and input paths")
	}

	raw, err := read(*instance)
	if err != nil {
		return commandFailure(stderr, "materializer instance read rejected")
	}

	config, err := DecodeInstanceConfig(raw)
	if err != nil {
		return commandFailure(stderr, "materializer instance rejected")
	}

	inputRaw, err := read(*inputPath)
	if err != nil {
		return commandFailure(stderr, "materializer input read rejected")
	}

	input, err := DecodeCommandInput(inputRaw)
	if err != nil {
		return commandFailure(stderr, "materializer input rejected")
	}

	if err := ValidateCommandInput(input, config); err != nil {
		return commandFailure(stderr, "materializer explicit contract rejected")
	}

	podName, namespace := getenv(PodNameEnvironment), getenv(NamespaceEnvironment)
	if namespace != input.Namespace {
		return commandFailure(stderr, "materializer Downward API namespace mismatch")
	}

	var selected *CommandPod

	for index := range input.Pods {
		if input.Pods[index].PodName == podName {
			selected = &input.Pods[index]
		}
	}

	if selected == nil {
		return commandFailure(stderr, "materializer Downward API pod mismatch")
	}

	trust := TrustInputs{PostgreSQLClientBindings: selected.PostgreSQLClientBindings}
	for role, destination := range map[string]*[]byte{
		rolePublic: &trust.PublicServerPEM, roleAdmin: &trust.AdminServerPEM,
		roleHealth: &trust.HealthClientPEM, roleOTLP: &trust.OTLPClientPEM, rolePostgres: &trust.PostgreSQLClientPEM,
	} {
		file, exists := input.TrustFiles[role]
		if !exists {
			continue
		}

		*destination, err = read(file)
		if err != nil {
			return commandFailure(stderr, "materializer issuer trust read rejected")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DeploymentMaterializerOverallTimeout)
	defer cancel()

	if err := materialize(ctx, raw, RuntimeInputs{
		PodName: podName, Namespace: namespace,
		APIEndpoint: input.APIEndpoint, APIServerName: input.APIServerName,
		ExpectedAdminDNS: selected.ExpectedAdminDNS, ExpectedPublicDNSNames: selected.ExpectedPublicDNSNames,
	}, trust); err != nil {
		return commandFailure(stderr, "materializer validation or publication rejected")
	}

	if _, err := fmt.Fprintln(stdout, "PASS: materialized verified pod PKI"); err != nil {
		return 1
	}

	return 0
}

func boundedCommandRead(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.New("materializer source unavailable")
	}

	data, readErr := io.ReadAll(io.LimitReader(file, cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes+1))

	closeErr := file.Close()
	if readErr != nil || closeErr != nil || len(data) > cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes {
		return nil, errors.New("materializer source read rejected")
	}

	return data, nil
}

func commandFailure(stderr io.Writer, message string) int {
	_, _ = fmt.Fprintln(stderr, message)

	return 1
}
