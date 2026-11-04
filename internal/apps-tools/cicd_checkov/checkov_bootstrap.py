# Copyright (c) 2025-2026 Justin Cranford.
"""Backport the pinned scanner's Terraform exclusion typo without changing scans.

Checkov 3.3.29 character-joins the repr of the exclusion list instead of joining
the patterns. The correction preserves every pattern and every scanner framework.
It is process-local: installed packages and caches are never modified. An
unexpected version or source fails closed instead of applying a speculative fix.
"""

import importlib.metadata
import inspect
import json
import re
from collections import Counter
from contextvars import ContextVar
from functools import wraps
from pathlib import Path

EXPECTED_VERSION = "3.3.29"
BROKEN_EXPRESSION = """re.compile('|'.join(f"({excluded_paths})"))"""
CORRECT_EXPRESSION = """re.compile('|'.join(f"({pattern})" for pattern in excluded_paths))"""

INVENTORY_PATH = "api/cryptosuite-registry/k8s/k8s-expected-inventory.json"
LIFECYCLE_PATHS = {
    "api/cryptosuite-registry/templates/kubernetes/cryptoutil-lifecycle/templates/lifecycle.yaml",
    "deployments/k8s/charts/cryptoutil-lifecycle/templates/lifecycle.yaml",
}
POD_KEYS = {"active", "fqdn", "host", "legacySlot", "name", "service", "tlsSecret"}
POD_NAME = re.compile(r"(?:sm-kms|identity-(?:authz|idp|rp|rs|spa)|skeleton-template)-(?:mem|disk|pg)-[0-9]+")
MANIFEST_REFERENCES = {
    '{{- $operation = "init" -}}{{- $manifestSecret = "postgres-bootstrap" -}}': "postgres-bootstrap",
    '{{- $operation = "replicate" -}}{{- $manifestSecret = "postgres-follower-bootstrap" -}}': "postgres-follower-bootstrap",
    '{{- $manifestSecret := "postgres-follower-bootstrap" -}}': "postgres-follower-bootstrap",
    '{{- if eq $base "db-init" -}}{{- $manifestSecret = "postgres-bootstrap" -}}': "postgres-bootstrap",
}


def only_reference_occurrences(text, recognized):
    """The native collection deduplicates values across lines. Keep a candidate
    if any identical value also occurs outside its verified public references.
    """
    counts = Counter(recognized.values())
    return {line: value for line, value in recognized.items() if text.count(value) == counts[value]}


def unique_json_object(pairs):
    """Ambiguous/duplicate fields must never authorize reference classification."""
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("Duplicate inventory field")
        result[key] = value
    return result


def public_reference_lines(root_folder, filename):
    """Recognize only public names at their certified artifact fields/locations.

    A TLS reference must equal its declared Pod name plus the TLS resource suffix.
    Arbitrary strings, credential fields, changed shapes, and other files are not
    references. Helm operation names are restricted to the exact canonical
    assignment statements, not every occurrence of those public strings.
    """
    root = Path(root_folder).resolve()
    path = Path(filename)
    path = (path if path.is_absolute() else root / path).resolve()
    try:
        relative = path.relative_to(root).as_posix()
    except ValueError:
        return {}
    if relative != INVENTORY_PATH and relative not in LIFECYCLE_PATHS:
        return {}
    try:
        text = path.read_text(encoding="utf-8")
        lines = text.splitlines()
        if relative in LIFECYCLE_PATHS:
            recognized = {number: MANIFEST_REFERENCES[line] for number, line in enumerate(lines, 1) if line in MANIFEST_REFERENCES}
            return only_reference_occurrences(text, recognized)
        json.loads(text, object_pairs_hook=unique_json_object)
    except OSError, UnicodeError, ValueError:
        return {}

    recognized = {}
    for number, line in enumerate(lines, 1):
        match = re.fullmatch(r'\s*"tlsSecret":\s*("(?:[^"\\]|\\.)*")\s*,?\s*', line)
        if match:
            value = json.loads(match[1])
            # Classify this exact enclosing object, not an equal value elsewhere.
            # The certified generated inventory uses two-space JSON indentation.
            indent = len(line) - len(line.lstrip())
            opening = " " * max(0, indent - 2) + "{"
            start = number - 2
            while start >= 0 and lines[start] != opening:
                start -= 1
            end = number
            closing = " " * max(0, indent - 2) + "}"
            while end < len(lines) and lines[end] not in (closing, closing + ","):
                end += 1
            if start < 0 or end == len(lines):
                continue
            try:
                pod = json.loads("\n".join(lines[start : end + 1]).rstrip(","), object_pairs_hook=unique_json_object)
            except ValueError:
                continue
            name = pod.get("name") if isinstance(pod, dict) else None
            if (
                isinstance(pod, dict)
                and isinstance(name, str)
                and (set(pod) == POD_KEYS or (set(pod) == POD_KEYS | {"pvc"} and "-disk-" in name and pod["pvc"] == "data-" + name))
                and POD_NAME.fullmatch(name)
                and isinstance(pod["active"], bool)
                and all(isinstance(pod[key], str) for key in POD_KEYS - {"active"})
                and value == pod["tlsSecret"] == name + "-tls"
            ):
                recognized[number] = value
    return only_reference_occurrences(text, recognized)


def classify_public_references(secrets, root_folder):
    """Leave every detector/check unchanged; distinguish public resource names.

    Only CKV_SECRET_6 candidates matching a qualified reference's exact value
    and line are non-secret metadata. Genuine credentials and all other checks
    stay in the original scanner pipeline, including identical strings elsewhere.
    """
    from checkov.secrets.runner import SECRET_TYPE_TO_ID
    from detect_secrets.core.secrets_collection import SecretsCollection

    classified = SecretsCollection()
    locations = {}
    for filename, secret in secrets:
        check_id = getattr(secret, "check_id", None) or SECRET_TYPE_TO_ID.get(secret.type)
        if root_folder and check_id == "CKV_SECRET_6":
            if filename not in locations:
                locations[filename] = public_reference_lines(root_folder, filename)
            if secret.secret_value is not None and locations[filename].get(secret.line_number) == secret.secret_value:
                continue
        classified[filename].add(secret)
    return classified


def apply_public_reference_classification():
    """Attach field-aware input classification, never remove failed reports."""
    if importlib.metadata.version("checkov") != EXPECTED_VERSION:
        raise RuntimeError("Pinned Checkov version does not match")

    from checkov.secrets.runner import Runner

    original = Runner.get_report
    signature = inspect.signature(original)
    expected = ("self", "secrets", "runner_filter", "history_store", "root_folder", "secret_suppressions_ids", "cleanupFn", "use_secret_filename")
    if tuple(signature.parameters) != expected:
        raise RuntimeError("Pinned Checkov secrets report interface does not match")

    @wraps(original)
    def get_report(*args, **kwargs):
        bound = signature.bind(*args, **kwargs)
        bound.apply_defaults()
        bound.arguments["secrets"] = classify_public_references(bound.arguments["secrets"], bound.arguments["root_folder"])
        return original(*bound.args, **bound.kwargs)

    Runner.get_report = get_report


def corrected_module_source(source):
    """Apply exactly the known one-expression backport, otherwise fail closed."""
    if source.count(BROKEN_EXPRESSION) != 1:
        raise RuntimeError("Pinned Checkov Terraform exclusion source does not match")
    return source.replace(BROKEN_EXPRESSION, CORRECT_EXPRESSION)


def apply_pinned_terraform_backport():
    """Correct module discovery before loading the scanner's normal CLI."""
    if importlib.metadata.version("checkov") != EXPECTED_VERSION:
        raise RuntimeError("Pinned Checkov version does not match")

    from checkov.terraform.module_loading import module_finder

    original = module_finder.find_modules
    source = inspect.getsource(original)
    corrected_module_source(source)
    if source.count("re.compile(") != 1:
        raise RuntimeError("Pinned Checkov Terraform compile site does not match")

    # Interpose only this module's compile site, without executing rewritten code.
    # Context-local state preserves independent concurrent module discoveries.
    original_re = module_finder.re
    current_regex = ContextVar("cryptoutil_terraform_exclusion_regex", default=None)

    class ModuleFinderRegex:
        def compile(self, pattern):
            corrected = current_regex.get()
            return corrected if corrected is not None else original_re.compile(pattern)

        def __getattr__(self, name):
            return getattr(original_re, name)

    @wraps(original)
    def find_modules(*args, **kwargs):
        excluded_paths = args[3] if len(args) > 3 else kwargs.get("excluded_paths")
        compiled = original_re.compile("|".join(f"({pattern})" for pattern in excluded_paths)) if excluded_paths else None
        token = current_regex.set(compiled)
        try:
            return original(*args, **kwargs)
        finally:
            current_regex.reset(token)

    module_finder.re = ModuleFinderRegex()
    module_finder.find_modules = find_modules


if __name__ == "__main__":
    apply_pinned_terraform_backport()
    apply_public_reference_classification()

    from checkov.main import Checkov

    raise SystemExit(Checkov().run())
