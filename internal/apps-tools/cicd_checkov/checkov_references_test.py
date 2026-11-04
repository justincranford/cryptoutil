# Copyright (c) 2025-2026 Justin Cranford.
"""Public-reference classification must never hide credential material."""

import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from checkov.secrets.runner import Runner
from detect_secrets.core.potential_secret import PotentialSecret
from detect_secrets.core.secrets_collection import SecretsCollection

import checkov_bootstrap


class PublicReferenceTests(unittest.TestCase):
    def test_inventory_requires_exact_field_value_location_and_pod_relationship(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            path = root / checkov_bootstrap.INVENTORY_PATH
            path.parent.mkdir(parents=True)
            reference = "identity-authz-pg-0-tls"
            pod = {
                "active": True,
                "fqdn": "pod.svc.cluster.local",
                "host": "pod.localhost",
                "legacySlot": "slot",
                "name": "identity-authz-pg-0",
                "service": "service",
                "tlsSecret": reference,
            }
            path.write_text(json.dumps({"pods": [pod]}, indent=2), encoding="utf-8", newline="\n")
            locations = checkov_bootstrap.public_reference_lines(root, path)
            self.assertEqual([reference], list(locations.values()))
            line = next(iter(locations))

            secrets = SecretsCollection()
            public = PotentialSecret("Base64 High Entropy String", str(path), reference, line)
            credential = PotentialSecret("Base64 High Entropy String", str(path), reference, line + 3)
            different_value = PotentialSecret("Base64 High Entropy String", str(path), "opaque-credential", line)
            other_check = PotentialSecret("Other credential detector", str(path), reference, line)
            other_check.check_id = "CKV_SECRET_3"
            for secret in (public, different_value):
                secrets[str(path)].add(secret)
            remaining = list(checkov_bootstrap.classify_public_references(secrets, root))
            self.assertEqual(1, len(remaining))
            self.assertEqual(2, len(list(secrets)), "the scanner's original collection must not be mutated")
            self.assertEqual(2, len(list(checkov_bootstrap.classify_public_references(secrets, None))))
            credentials = SecretsCollection()
            credentials[str(path)].add(credential)
            self.assertEqual(1, len(list(checkov_bootstrap.classify_public_references(credentials, root))))
            others = SecretsCollection()
            others[str(path)].add(other_check)
            self.assertEqual(1, len(list(checkov_bootstrap.classify_public_references(others, root))))
            unknown_value = PotentialSecret("Base64 High Entropy String", str(path), "unknown-value", line + 1)
            unknown_value.secret_value = None
            others[str(path)].add(unknown_value)
            self.assertEqual(2, len(list(checkov_bootstrap.classify_public_references(others, root))))
            path.write_text(json.dumps({"pods": [pod], "password": reference}, indent=2), encoding="utf-8", newline="\n")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))
            self.assertEqual(2, len(list(checkov_bootstrap.classify_public_references(secrets, root))))

            other = root / "credentials.json"
            other.write_text(path.read_text(encoding="utf-8"), encoding="utf-8", newline="\n")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, other))
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root / "outside", path))
            pod["tlsSecret"] = "opaque-credential"
            path.write_text(json.dumps({"pods": [pod]}, indent=2), encoding="utf-8", newline="\n")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))
            pod["tlsSecret"] = reference
            pod["password"] = reference
            path.write_text(json.dumps({"pods": [pod]}, indent=2), encoding="utf-8", newline="\n")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))

    def test_helm_assignments_only_not_credentials_or_other_files(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            assignment, reference = next(iter(checkov_bootstrap.MANIFEST_REFERENCES.items()))
            for relative in checkov_bootstrap.LIFECYCLE_PATHS:
                path = root / relative
                path.parent.mkdir(parents=True)
                path.write_text(assignment + "\n", encoding="utf-8", newline="\n")
                self.assertEqual({1: reference}, checkov_bootstrap.public_reference_lines(root, path))
                path.write_text(f'{assignment}\npassword: "{reference}"\n{assignment} trailing\n', encoding="utf-8", newline="\n")
                self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))
            other = root / "credentials.yaml"
            other.write_text(assignment, encoding="utf-8", newline="\n")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, other))

    def test_invalid_or_ambiguous_inventory_fails_closed(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            path = root / checkov_bootstrap.INVENTORY_PATH
            path.parent.mkdir(parents=True)
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))
            for content in ("{", '{"pods": [], "pods": []}'):
                path.write_text(content, encoding="utf-8", newline="\n")
                self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))
            path.write_bytes(b"\xff")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))

    def test_disk_reference_requires_matching_public_pvc_name(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            path = root / checkov_bootstrap.INVENTORY_PATH
            path.parent.mkdir(parents=True)
            name = "identity-authz-disk-0"
            pod = {
                "active": True,
                "fqdn": "pod.svc.cluster.local",
                "host": "pod.localhost",
                "legacySlot": "slot",
                "name": name,
                "service": "service",
                "tlsSecret": name + "-tls",
                "pvc": "data-" + name,
            }
            path.write_text(json.dumps({"pods": [pod]}, indent=2), encoding="utf-8", newline="\n")
            self.assertEqual([name + "-tls"], list(checkov_bootstrap.public_reference_lines(root, path).values()))
            pod["pvc"] = "opaque-credential"
            path.write_text(json.dumps({"pods": [pod]}, indent=2), encoding="utf-8", newline="\n")
            self.assertEqual({}, checkov_bootstrap.public_reference_lines(root, path))

    def test_classifier_guard_rejects_version_or_interface_drift(self):
        with patch("importlib.metadata.version", return_value="unsupported"), self.assertRaisesRegex(RuntimeError, "version does not match"):
            checkov_bootstrap.apply_public_reference_classification()
        with patch.object(Runner, "get_report", lambda: None), self.assertRaisesRegex(RuntimeError, "interface does not match"):
            checkov_bootstrap.apply_public_reference_classification()


if __name__ == "__main__":
    unittest.main()
