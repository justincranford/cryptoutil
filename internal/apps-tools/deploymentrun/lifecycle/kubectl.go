// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	cryptoutilAppsToolsDeploymentgen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// KubectlConfig requires an explicit cluster boundary and trusted executable/
// repository paths. The parent certifies tool versions, chart provenance,
// kubeconfig permissions, admission and quotas; this adapter does not install
// or discover them. PollInterval has no default.
type KubectlConfig struct {
	Kubectl        string
	Helm           string
	RepositoryRoot string
	Kubeconfig     string
	Context        string
	PollInterval   time.Duration
}

// KubectlClient uses the canonical Helm chart for bootstrap/Job manifests.
// Existing objects are verify-only; absent objects are atomically created.
// No apply/patch/replace/delete, implicit context, or shared volume is used.
type KubectlClient struct {
	config    KubectlConfig
	runner    Runner
	resources map[string]map[string]any
	bootstrap []Identity
	uids      map[string]string
	imports   map[string]cryptoutilAppsToolsDeploymentgen.LifecycleSecretSource
	run       string
	prepared  bool
}

var _ Client = (*KubectlClient)(nil)

var errNotFound = errors.New("resource not found")

// NewKubectlClient does not contact a cluster. Use one client per Execute call.
func NewKubectlClient(config KubectlConfig, runner Runner) (*KubectlClient, error) {
	for _, path := range []string{config.Kubectl, config.Helm, config.RepositoryRoot, config.Kubeconfig} {
		if !filepath.IsAbs(path) {
			return nil, errors.New("lifecycle requires absolute trusted paths")
		}
	}

	if runner == nil || strings.TrimSpace(config.Context) == "" || strings.HasPrefix(config.Context, dash) ||
		config.PollInterval <= 0 || config.PollInterval > time.Minute {
		return nil, errors.New("lifecycle requires runner, explicit cluster context and bounded polling")
	}

	return &KubectlClient{config: config, runner: runner, uids: make(map[string]string)}, nil
}

func identityKey(id Identity) string { return id.Kind + "/" + id.Namespace + "/" + id.Name }

func (c *KubectlClient) Prepare(ctx context.Context, values cryptoutilAppsToolsDeploymentgen.LifecycleChartValues) error {
	if c.prepared {
		return errors.New("lifecycle client already prepared")
	}

	c.prepared = true

	input, _ := json.Marshal(values) // Typed contract has no unsupported JSON values.

	chart := filepath.Join(c.config.RepositoryRoot, "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-lifecycle")

	raw, err := c.runner.Run(ctx, c.config.Helm, []string{
		cryptoutilSharedMagic.HelmCommandTemplate, "cu-lifecycle", chart, "--values", dash, "--kube-version", "1.37.1",
	}, input)
	if err != nil {
		return safeError("render", Identity{}, err)
	}

	expected := make(map[string]Identity)

	for _, boundary := range values.Namespaces {
		for _, kind := range []string{"ServiceAccount", "Role", "RoleBinding"} {
			id := Identity{Namespace: boundary.Name, Kind: kind, Name: bootstrapName, Run: values.Run, Release: boundary.Release}
			expected[identityKey(id)] = id
			c.bootstrap = append(c.bootstrap, id)
		}
	}

	for _, job := range values.Jobs {
		id := Identity{Namespace: job.Namespace, Kind: jobKind, Name: job.Name, Run: values.Run, Release: job.Release, Revision: values.Revision}
		expected[identityKey(id)] = id
	}

	resources, err := decodeRendered(raw, expected)
	if err != nil {
		c.bootstrap = nil

		return err
	}

	for _, job := range values.Jobs {
		document := resources[identityKey(Identity{Namespace: job.Namespace, Kind: jobKind, Name: job.Name})]
		spec := objectMap(document["spec"])
		pod := objectMap(objectMap(spec[cryptoutilSharedMagic.KubernetesPodTemplateField])["spec"])

		containers, _ := pod["containers"].([]any)
		if spec["activeDeadlineSeconds"] != float64(job.ActiveDeadlineSeconds) ||
			spec["backoffLimit"] != float64(job.BackoffLimit) || spec["completions"] != float64(1) ||
			spec["parallelism"] != float64(1) || pod["serviceAccountName"] != bootstrapName ||
			pod["automountServiceAccountToken"] != false || pod["restartPolicy"] != "Never" || len(containers) != 1 {
			return ErrCollision
		}

		container := objectMap(containers[0])
		wanted, _ := json.Marshal(map[string]any{"image": job.Image, "command": job.Command, "args": job.Args, "resources": job.Resources})

		var wantedContainer map[string]any

		_ = json.Unmarshal(wanted, &wantedContainer)
		if !containsDesired(container, wantedContainer) {
			return ErrCollision
		}
	}

	c.resources = resources
	c.run = values.Run

	c.imports = make(map[string]cryptoutilAppsToolsDeploymentgen.LifecycleSecretSource, len(values.Secrets))
	for _, source := range values.Secrets {
		c.imports[source.Namespace+"/"+source.Name] = source
	}

	return nil
}

func decodeRendered(raw []byte, expected map[string]Identity) (map[string]map[string]any, error) {
	if len(raw) > maxOutputBytes {
		return nil, errors.New("rendered lifecycle exceeds bound")
	}

	result := make(map[string]map[string]any)
	decoder := yaml.NewDecoder(bytes.NewReader(raw))

	for {
		var document map[string]any
		if err := decoder.Decode(&document); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, errors.New("invalid rendered lifecycle YAML")
		}

		if len(document) == 0 {
			continue
		}

		// Normalize YAML integers to the same JSON representation as kubectl.
		encoded, err := json.Marshal(document)
		if err != nil {
			return nil, errors.New("invalid rendered lifecycle document")
		}

		_ = json.Unmarshal(encoded, &document) // json.Marshal above produced valid JSON.

		id := documentIdentity(document)
		key := identityKey(id)

		owned, exists := expected[key]
		if !exists || result[key] != nil || !owns(document, owned) {
			return nil, errors.New("rendered lifecycle ownership or coverage mismatch")
		}

		result[key] = document
	}

	if len(result) != len(expected) {
		return nil, errors.New("incomplete rendered lifecycle")
	}

	return result, nil
}

func documentIdentity(document map[string]any) Identity {
	meta := objectMap(document["metadata"])

	return Identity{Kind: text(document["kind"]), Namespace: text(meta["namespace"]), Name: text(meta[cryptoutilSharedMagic.ClaimName])}
}

func objectMap(value any) map[string]any {
	result, _ := value.(map[string]any)

	return result
}

func text(value any) string {
	result, _ := value.(string)

	return result
}

func owns(document map[string]any, id Identity) bool {
	if documentIdentity(document) != (Identity{Kind: id.Kind, Namespace: id.Namespace, Name: id.Name}) {
		return false
	}

	meta := objectMap(document["metadata"])
	labels := objectMap(meta["labels"])
	annotations := objectMap(meta["annotations"])

	return labels[runLabel] == id.Run &&
		(id.Release == "" || labels[cryptoutilSharedMagic.KubernetesLabelRelease] == id.Release) &&
		(id.Revision == "" || annotations["cu.io/revision"] == id.Revision) &&
		meta["deletionTimestamp"] == nil
}

func (c *KubectlClient) command(ctx context.Context, args []string, input []byte) ([]byte, error) {
	prefix := []string{"--kubeconfig", c.config.Kubeconfig, "--context", c.config.Context}

	output, err := c.runner.Run(ctx, c.config.Kubectl, append(prefix, args...), input)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, safeError("transport", Identity{}, err)
		}

		return nil, ErrTransport
	}

	if len(output) > maxOutputBytes {
		return nil, errors.New("lifecycle transport output exceeds bound")
	}

	return output, nil
}

func (c *KubectlClient) get(ctx context.Context, id Identity) (map[string]any, error) {
	args := []string{"get", id.Kind, id.Name, "--ignore-not-found", outputFlag, jsonOutput}
	if id.Namespace != "" {
		args = append(args, "--namespace", id.Namespace)
	}

	raw, err := c.command(ctx, args, nil)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, errNotFound
	}

	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil || len(document) == 0 {
		return nil, errors.New("invalid Kubernetes response")
	}

	return document, nil
}

func (c *KubectlClient) Require(ctx context.Context, id Identity) error {
	document, err := c.get(ctx, id)
	if errors.Is(err, errNotFound) {
		return ErrPrerequisite
	}

	if err != nil {
		return err
	}

	if !owns(document, id) {
		return ErrPrerequisite
	}

	if id.Kind == namespaceKind && objectMap(document[cryptoutilSharedMagic.StringStatus])["phase"] != "Active" {
		return ErrPrerequisite
	}

	return nil
}

func (c *KubectlClient) EnsureSecret(ctx context.Context, secret cryptoutilAppsToolsDeploymentgen.LifecycleSecret) error {
	source, exists := c.imports[secret.Metadata.Namespace+"/"+secret.Metadata.Name]
	if !exists || source.Type != secret.Type || len(source.Files) != len(secret.Data) ||
		secret.Kind != secretKind || secret.APIVersion != apiVersion || secret.Metadata.Labels["cu.io/owner"] != importOwner ||
		secret.Metadata.Labels[runLabel] != c.run {
		return ErrCollision
	}

	if _, err := validateSecrets(cryptoutilAppsToolsDeploymentgen.LifecycleChartValues{
		Run: c.run, Secrets: []cryptoutilAppsToolsDeploymentgen.LifecycleSecretSource{source},
	}, []cryptoutilAppsToolsDeploymentgen.LifecycleSecret{secret}); err != nil {
		return ErrCollision
	}

	raw, _ := json.Marshal(secret) // Typed import has no unsupported JSON values.

	var desired map[string]any

	_ = json.Unmarshal(raw, &desired)

	id := Identity{Namespace: secret.Metadata.Namespace, Kind: secretKind, Name: secret.Metadata.Name, Run: secret.Metadata.Labels[runLabel]}

	return c.ensure(ctx, id, desired)
}

func (c *KubectlClient) EnsureBootstrap(ctx context.Context) error {
	if c.resources == nil {
		return errors.New("lifecycle is not prepared")
	}

	for _, id := range c.bootstrap {
		if err := c.ensure(ctx, id, c.resources[identityKey(id)]); err != nil {
			return err
		}
	}

	return nil
}

func (c *KubectlClient) EnsureJob(ctx context.Context, job cryptoutilAppsToolsDeploymentgen.LifecycleJobValues) error {
	id := Identity{Namespace: job.Namespace, Kind: jobKind, Name: job.Name}

	desired := c.resources[identityKey(id)]
	if desired == nil {
		return errors.New("job not in prepared contract")
	}

	labels := objectMap(objectMap(desired["metadata"])["labels"])
	id.Run, id.Release = text(labels[runLabel]), text(labels[cryptoutilSharedMagic.KubernetesLabelRelease])
	id.Revision = text(objectMap(objectMap(desired["metadata"])["annotations"])["cu.io/revision"])

	return c.ensure(ctx, id, desired)
}

func (c *KubectlClient) ensure(ctx context.Context, id Identity, desired map[string]any) error {
	existing, err := c.get(ctx, id)
	if err != nil && !errors.Is(err, errNotFound) {
		return err
	}

	if existing == nil {
		raw, err := json.Marshal(desired)
		if err != nil {
			return errors.New("encode resource")
		}

		// Atomic create rejects a race with another owner; never retry as apply.
		output, err := c.command(ctx, []string{"create", "-f", dash, outputFlag, jsonOutput}, raw)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(output, &existing); err != nil {
			return errors.New("invalid create response")
		}
	}

	if !owns(existing, id) || !containsDesired(existing, desired) || !safeExistingPod(existing, desired) {
		return ErrCollision
	}

	if id.Kind == secretKind {
		if !reflect.DeepEqual(existing["data"], desired["data"]) || existing["stringData"] != nil {
			return ErrCollision
		}
	}

	uid := text(objectMap(existing["metadata"])["uid"])
	if uid == "" {
		return errors.New("resource UID missing")
	}

	c.uids[identityKey(id)] = uid

	return nil
}

// Reject executable/security fields injected into an existing Pod template.
// Only known inert API-server Pod/container defaults may be additional fields.
// Admission mutations not represented by the certified chart fail closed.
func safeExistingPod(actual, desired map[string]any) bool {
	if desired["kind"] != jobKind {
		return true
	}

	have := objectMap(objectMap(objectMap(actual["spec"])[cryptoutilSharedMagic.KubernetesPodTemplateField])["spec"])
	want := objectMap(objectMap(objectMap(desired["spec"])[cryptoutilSharedMagic.KubernetesPodTemplateField])["spec"])

	defaults := map[string]any{
		"dnsPolicy": "ClusterFirst", "schedulerName": "default-scheduler", "enableServiceLinks": true,
		"preemptionPolicy": "PreemptLowerPriority", "priority": float64(0),
	}
	if !onlyKnownFields(have, want, defaults) {
		return false
	}

	haveContainers, _ := have["containers"].([]any)

	wantContainers, _ := want["containers"].([]any)
	if len(haveContainers) != len(wantContainers) {
		return false
	}

	for index, container := range haveContainers {
		haveContainer, wantContainer := objectMap(container), objectMap(wantContainers[index])
		if !onlyKnownFields(haveContainer, wantContainer, map[string]any{
			"terminationMessagePath": "/dev/termination-log", "terminationMessagePolicy": "File",
		}) || !reflect.DeepEqual(haveContainer["securityContext"], wantContainer["securityContext"]) {
			return false
		}
	}

	return reflect.DeepEqual(have["securityContext"], want["securityContext"])
}

func onlyKnownFields(actual, desired, defaults map[string]any) bool {
	for key, value := range actual {
		if _, exists := desired[key]; exists {
			continue
		}

		if defaultValue, exists := defaults[key]; !exists || !reflect.DeepEqual(value, defaultValue) {
			return false
		}
	}

	return true
}

// containsDesired tolerates server-generated defaults/metadata, but not changes
// to any explicitly rendered field. Lists (commands, RBAC, mounts) are exact.
func containsDesired(actual, desired any) bool {
	switch wanted := desired.(type) {
	case map[string]any:
		have, ok := actual.(map[string]any)
		if !ok {
			return false
		}

		for key, value := range wanted {
			if !containsDesired(have[key], value) {
				return false
			}
		}

		return true
	case []any:
		have, ok := actual.([]any)
		if !ok || len(have) != len(wanted) {
			return false
		}

		for index, value := range wanted {
			if !containsDesired(have[index], value) {
				return false
			}
		}

		return true
	default:
		return reflect.DeepEqual(actual, desired)
	}
}

// String deliberately does not expose kubeconfig, command input or resources.
func (c *KubectlClient) String() string {
	return fmt.Sprintf("lifecycle Kubernetes client (%d ensured identities)", len(c.uids))
}
