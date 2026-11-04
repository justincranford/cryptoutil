# Copyright (c) 2025-2026 Justin Cranford.
"""Executable regressions for the process-local pinned Checkov backport."""

import tempfile
import unittest
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
from unittest.mock import patch

import checkov_bootstrap


class PinnedTerraformBackportTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        checkov_bootstrap.apply_pinned_terraform_backport()

    def test_corrects_only_known_expression(self):
        source = "before\n" + checkov_bootstrap.BROKEN_EXPRESSION + "\nafter\n"
        expected = "before\n" + checkov_bootstrap.CORRECT_EXPRESSION + "\nafter\n"
        self.assertEqual(expected, checkov_bootstrap.corrected_module_source(source))

    def test_unexpected_source_fails_closed(self):
        for source in ("", checkov_bootstrap.CORRECT_EXPRESSION, checkov_bootstrap.BROKEN_EXPRESSION * 2):
            with self.subTest(source=source), self.assertRaisesRegex(RuntimeError, "source does not match"):
                checkov_bootstrap.corrected_module_source(source)

    def test_wrong_version_fails_closed(self):
        with patch("importlib.metadata.version", return_value="unsupported"), self.assertRaisesRegex(RuntimeError, "version does not match"):
            checkov_bootstrap.apply_pinned_terraform_backport()

    def test_changed_installed_source_fails_closed(self):
        with patch("inspect.getsource", return_value="unexpected source"), self.assertRaisesRegex(RuntimeError, "source does not match"):
            checkov_bootstrap.apply_pinned_terraform_backport()
        source = checkov_bootstrap.BROKEN_EXPRESSION + "\nre.compile('unexpected second site')"
        with patch("inspect.getsource", return_value=source), self.assertRaisesRegex(RuntimeError, "compile site does not match"):
            checkov_bootstrap.apply_pinned_terraform_backport()

    def test_real_module_discovery_preserves_exact_exclusion_semantics(self):
        from checkov.terraform.module_loading.module_finder import find_modules

        with tempfile.TemporaryDirectory() as directory:
            for name in ("allowed", "ignored"):
                path = Path(directory, name)
                path.mkdir()
                Path(path, "main.tf").write_text(f'module "fixture" {{\n  source = "./local-{name}"\n}}\n', encoding="utf-8")
            self.assertEqual(2, len(find_modules(directory)))
            self.assertEqual(1, len(find_modules(directory, excluded_paths=[r"(^|[\\/])ignored([\\/]|$)"])))
            self.assertEqual(2, len(find_modules(directory, excluded_paths=[r"(^|[\\/])missing([\\/]|$)"])))
            self.assertEqual(0, len(find_modules(directory, excluded_paths=[r".*\.tf$"])))
            self.assertEqual(1, len(find_modules(directory, None, None, [r"(^|[\\/])ignored([\\/]|$)"])))
            from checkov.terraform.module_loading import module_finder

            self.assertIsNotNone(module_finder.re.compile("plain").fullmatch("plain"))
            self.assertTrue(module_finder.re.IGNORECASE)

            def discover(pattern):
                return [module.address for module in find_modules(directory, excluded_paths=[pattern])]

            patterns = [r"(^|[\\/])ignored([\\/]|$)", r"(^|[\\/])allowed([\\/]|$)"] * 10
            with ThreadPoolExecutor(max_workers=4) as executor:
                results = list(executor.map(discover, patterns))
            for index, addresses in enumerate(results):
                self.assertEqual(1, len(addresses))
                expected = "local-allowed" if index % 2 == 0 else "local-ignored"
                self.assertIn(expected, addresses[0], "concurrent exclusions must remain context-local")


if __name__ == "__main__":
    unittest.main()
