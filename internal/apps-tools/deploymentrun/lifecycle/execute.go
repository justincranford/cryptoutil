// Copyright (c) 2025-2026 Justin Cranford.

// Package lifecycle executes the deploymentgen lifecycle contract without
// creating clusters, installing runtime charts, or deleting resources.
package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	cryptoutilAppsToolsDeploymentgen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Client is the guarded Kubernetes boundary. Implementations must honor context,
// reject ownership/spec collisions, never overwrite existing Secret data, and
// never expose payloads or transport diagnostics in errors. Prepare must not
// submit Jobs. EnsureJob submits only the supplied Job; WaitJob must observe
// Complete (not merely Pod readiness) and fail on Failed or identity replacement.
// Clients and request material must not be shared with concurrent executions.
type Client interface {
	Prepare(context.Context, cryptoutilAppsToolsDeploymentgen.LifecycleChartValues) error
	Require(context.Context, Identity) error
	EnsureSecret(context.Context, cryptoutilAppsToolsDeploymentgen.LifecycleSecret) error
	EnsureBootstrap(context.Context) error
	EnsureJob(context.Context, cryptoutilAppsToolsDeploymentgen.LifecycleJobValues) error
	WaitJob(context.Context, Identity) error
	WaitDatabase(context.Context, Identity) error
}

// Identity is a non-secret, namespaced ownership boundary.
type Identity struct {
	Namespace string
	Kind      string
	Name      string
	Run       string
	Release   string
	Revision  string
}

// Request references the already-selected inventory and complete imported
// material. Callers must keep these inputs stable until Execute returns.
type Request struct {
	Model     cryptoutilAppsToolsDeploymentgen.Model
	Inventory cryptoutilAppsToolsDeploymentgen.Inventory
	Input     cryptoutilAppsToolsDeploymentgen.LifecycleInput
	Secrets   []cryptoutilAppsToolsDeploymentgen.LifecycleSecret
	// VerifyPKI must recheck published bytes/certificate policy through read-only
	// APIs after an owned verify-only PKI Job completes. Complete alone is not
	// sufficient. Production callers must validate the whole plan before Execute.
	VerifyPKI func(context.Context, cryptoutilAppsToolsDeploymentgen.LifecycleJobValues, Client) error
}

// Policy has no implicit deadlines. MaxDuration bounds the entire invocation;
// OperationTimeout bounds each non-Job operation. Job apply+wait share the
// operator's activeDeadlineSeconds and the enclosing invocation deadline.
type Policy struct {
	MaxDuration      time.Duration
	OperationTimeout time.Duration
}

// Result counts confirmed completions only. A failure returns partial progress;
// no rollback or cleanup is attempted. Retry is verify-or-create, not replacement.
type Result struct {
	Secrets int
	Jobs    int
}

// Safe failure classes are stable and contain no transport diagnostics.
var (
	ErrCollision      = errors.New("ownership or specification collision")
	ErrPrerequisite   = errors.New("runtime prerequisite unavailable")
	ErrJobFailed      = errors.New("job failed")
	ErrReplaced       = errors.New("resource identity replaced")
	ErrTransport      = errors.New("kubernetes transport failed")
	ErrUnsupportedPKI = errors.New("PKI generation cannot attest prepublished material; verify-only contract required")
)

// ExecutionError reports only the failing operation and owned identity.
// Transport errors are intentionally not wrapped, because stderr can contain
// Secret values. Cancellation/deadline identity is preserved via Unwrap.
type ExecutionError struct {
	Operation string
	Identity  Identity
	cause     error
}

func (e *ExecutionError) Error() string {
	reason := "operation rejected"
	if e.cause != nil {
		reason = e.cause.Error()
	}

	return fmt.Sprintf("lifecycle %s failed: %s %s/%s (%s)", e.Operation, e.Identity.Kind, e.Identity.Namespace, e.Identity.Name, reason)
}

func (e *ExecutionError) Unwrap() error { return e.cause }

func safeError(operation string, identity Identity, err error) error {
	var cause error

	switch {
	case errors.Is(err, context.Canceled):
		cause = context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		cause = context.DeadlineExceeded
	default:
		for _, known := range []error{ErrCollision, ErrPrerequisite, ErrJobFailed, ErrReplaced, ErrTransport} {
			if errors.Is(err, known) {
				cause = known

				break
			}
		}
	}

	return &ExecutionError{Operation: operation, Identity: identity, cause: cause}
}

// Execute validates everything before contacting Kubernetes. Its deterministic
// DAG has barriers: imported Secrets/bootstrap -> infra PKI -> all release PKI
// -> ready PostgreSQL -> union db-init -> all migrations -> replication.
// SQLite-only selections omit every PostgreSQL stage. Sibling Jobs execute
// serially (bounded concurrency=1); no dependent Job is submitted early.
func Execute(ctx context.Context, request Request, client Client, policy Policy) (Result, error) {
	result := Result{}
	if ctx == nil || client == nil || policy.MaxDuration <= 0 || policy.MaxDuration > 24*time.Hour ||
		policy.OperationTimeout <= 0 || policy.OperationTimeout > time.Hour {
		return result, errors.New("lifecycle requires client, context and explicit bounded deadlines")
	}

	runCtx, cancel := context.WithTimeout(ctx, policy.MaxDuration)
	defer cancel()

	if err := runCtx.Err(); err != nil {
		return result, safeError("validate-contract", Identity{}, err)
	}

	values, err := cryptoutilAppsToolsDeploymentgen.BuildLifecycleChartValues(request.Model, request.Inventory, request.Input)
	if err != nil {
		// Even validation diagnostics supplied by another package are not logged.
		return result, safeError("validate-contract", Identity{}, err)
	}

	secrets, err := validateSecrets(values, request.Secrets)
	if err != nil {
		return result, safeError(cryptoutilSharedMagic.CLIValidateSecretsCommand, Identity{}, err)
	}

	stages, err := jobStages(values, request.Inventory.InfraNamespace)
	if err != nil {
		return result, safeError("validate-dag", Identity{}, err)
	}

	for _, stage := range stages[:2] {
		for _, job := range stage {
			if request.VerifyPKI == nil || len(job.Args) != 2 || job.Args[0] != "pki-verify" || len(job.WriteSecretNames) != 0 {
				return result, ErrUnsupportedPKI
			}
		}
	}

	operation := func(name string, identity Identity, fn func(context.Context) error) error {
		callCtx, callCancel := context.WithTimeout(runCtx, policy.OperationTimeout)
		defer callCancel()

		if err := callCtx.Err(); err != nil {
			return safeError(name, identity, err)
		}

		err := fn(callCtx)
		if callCtx.Err() != nil {
			err = callCtx.Err()
		}

		if err != nil {
			return safeError(name, identity, err)
		}

		return nil
	}
	if err := operation("prepare", Identity{}, func(ctx context.Context) error { return client.Prepare(ctx, values) }); err != nil {
		return result, err
	}

	// Namespaces and explicit ConfigMaps must already exist under this run.
	// Installing infrastructure/application charts remains a parent-owned step.
	required := make(map[string]Identity)

	for _, boundary := range values.Namespaces {
		id := Identity{Kind: namespaceKind, Name: boundary.Name, Run: values.Run, Release: boundary.Release}
		required[id.Kind+"/"+id.Name] = id
	}

	for _, job := range values.Jobs {
		for _, name := range job.ConfigMapNames {
			id := Identity{Namespace: job.Namespace, Kind: configMapKind, Name: name, Run: values.Run, Release: job.Release}
			required[id.Kind+"/"+id.Namespace+"/"+id.Name] = id
		}
	}

	keys := make([]string, 0, len(required))
	for key := range required {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		id := required[key]
		if err := operation("require", id, func(ctx context.Context) error { return client.Require(ctx, id) }); err != nil {
			return result, err
		}
	}

	for _, secret := range secrets {
		id := Identity{Namespace: secret.Metadata.Namespace, Kind: secretKind, Name: secret.Metadata.Name, Run: values.Run}
		if err := operation("import", id, func(ctx context.Context) error { return client.EnsureSecret(ctx, secret) }); err != nil {
			return result, err
		}

		result.Secrets++
	}

	if err := operation(bootstrapName, Identity{}, client.EnsureBootstrap); err != nil {
		return result, err
	}

	for stageIndex, stage := range stages {
		if stageIndex == 2 && len(stage) > 0 {
			for _, name := range []string{cryptoutilSharedMagic.PKIInitPostgresLeaderService, cryptoutilSharedMagic.PKIInitPostgresFollowerService} {
				id := Identity{
					Namespace: request.Inventory.InfraNamespace, Kind: statefulSetKind, Name: name,
					Run: values.Run, Release: "cu-" + values.Run + "-infra",
				}
				if err := operation("database-ready", id, func(ctx context.Context) error { return client.WaitDatabase(ctx, id) }); err != nil {
					return result, err
				}
			}
		}

		for _, job := range stage {
			id := Identity{Namespace: job.Namespace, Kind: jobKind, Name: job.Name, Run: values.Run, Release: job.Release, Revision: values.Revision}
			jobCtx, jobCancel := context.WithTimeout(runCtx, time.Duration(job.ActiveDeadlineSeconds)*time.Second)
			err := jobCtx.Err()
			op := "apply-job"

			if err == nil {
				err = client.EnsureJob(jobCtx, job)
			}

			if err == nil && jobCtx.Err() == nil {
				op = "wait-job"
				err = client.WaitJob(jobCtx, id)
			}

			if jobCtx.Err() != nil {
				err = jobCtx.Err()
			}

			jobCancel()

			if err != nil {
				return result, safeError(op, id, err)
			}

			if stageIndex < 2 {
				if err := operation("verify-published-pki", id, func(ctx context.Context) error { return request.VerifyPKI(ctx, job, client) }); err != nil {
					return result, err
				}
			}

			result.Jobs++
		}
	}

	return result, nil
}

func jobStages(values cryptoutilAppsToolsDeploymentgen.LifecycleChartValues, infraNamespace string) ([][]cryptoutilAppsToolsDeploymentgen.LifecycleJobValues, error) {
	stages := make([][]cryptoutilAppsToolsDeploymentgen.LifecycleJobValues, 5)

	for _, job := range values.Jobs {
		base := strings.TrimSuffix(job.Name, dash+values.Revision)
		switch {
		case base == infraPKIName && job.Namespace == infraNamespace:
			stages[0] = append(stages[0], job)
		case base == cryptoutilSharedMagic.PSIDPKIInit && job.Namespace != infraNamespace:
			stages[1] = append(stages[1], job)
		case base == databaseInitName && job.Namespace == infraNamespace:
			stages[2] = append(stages[2], job)
		case strings.HasSuffix(base, "-migrate") && job.Namespace != infraNamespace:
			stages[3] = append(stages[3], job)
		case base == "replication" && job.Namespace == infraNamespace:
			stages[4] = append(stages[4], job)
		default:
			return nil, errors.New("unrecognized lifecycle stage")
		}
	}

	if len(stages[0]) != 1 || len(stages[1]) != len(values.Namespaces)-1 ||
		len(stages[2]) > 1 || len(stages[4]) != len(stages[2]) ||
		(len(stages[3]) > 0) != (len(stages[2]) > 0) {
		return nil, errors.New("incomplete lifecycle stage graph")
	}

	return stages, nil
}

func validateSecrets(
	values cryptoutilAppsToolsDeploymentgen.LifecycleChartValues,
	secrets []cryptoutilAppsToolsDeploymentgen.LifecycleSecret,
) ([]cryptoutilAppsToolsDeploymentgen.LifecycleSecret, error) {
	if len(secrets) != len(values.Secrets) {
		return nil, errors.New("incomplete imports")
	}

	expected := make(map[string]cryptoutilAppsToolsDeploymentgen.LifecycleSecretSource, len(values.Secrets))
	for _, source := range values.Secrets {
		expected[source.Namespace+"/"+source.Name] = source
	}

	for _, secret := range secrets {
		key := secret.Metadata.Namespace + "/" + secret.Metadata.Name

		source, exists := expected[key]
		if !exists || secret.APIVersion != apiVersion || secret.Kind != secretKind || secret.Type != source.Type ||
			secret.Metadata.Labels[runLabel] != values.Run || secret.Metadata.Labels["cu.io/owner"] != importOwner ||
			len(secret.Data) != len(source.Files) {
			return nil, errors.New("invalid imported identity")
		}

		delete(expected, key)

		total := 0

		for name := range source.Files {
			data := secret.Data[name]
			if len(data) == 0 || len(data) > 1<<20-total {
				return nil, errors.New("invalid imported data")
			}

			total += len(data)
		}
	}

	// Snapshot bytes/maps so an injected client cannot change a later import.
	snapshot := make([]cryptoutilAppsToolsDeploymentgen.LifecycleSecret, len(secrets))
	for index, secret := range secrets {
		snapshot[index] = secret

		snapshot[index].Metadata.Labels = make(map[string]string, len(secret.Metadata.Labels))
		for key, value := range secret.Metadata.Labels {
			snapshot[index].Metadata.Labels[key] = value
		}

		snapshot[index].Data = make(map[string][]byte, len(secret.Data))
		for key, value := range secret.Data {
			snapshot[index].Data[key] = bytes.Clone(value)
		}
	}

	slices.SortFunc(snapshot, func(a, b cryptoutilAppsToolsDeploymentgen.LifecycleSecret) int {
		return strings.Compare(a.Metadata.Namespace+"/"+a.Metadata.Name, b.Metadata.Namespace+"/"+b.Metadata.Name)
	})

	return snapshot, nil
}
