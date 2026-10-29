"""Behavior checks for isolated authoring projects, without invoking a live generator."""

import argparse
from contextlib import redirect_stderr
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from author import Author, WorkflowError


class AuthorTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.source = self.root / "custom source"
        self.hidden = self.source / "generated support"
        self.config = self.root / "settings" / "project.yaml"
        self.args = argparse.Namespace(
            config=str(self.config), ccode_path=str(self.source),
            hidden_path=str(self.hidden), output_path=str(self.root / "output"),
            action="inspect", version="v1.2.3", name=None, content_file=None,
        )
        self.calls = []
        self.processes = []
        self.fail = None
        mock = patch("author.subprocess.run", side_effect=self.ccode)
        mock.start()
        self.addCleanup(mock.stop)
        redirect = redirect_stderr(io.StringIO())
        redirect.__enter__()
        self.addCleanup(redirect.__exit__, None, None, None)

    def initialized(self):
        self.config.parent.mkdir(parents=True, exist_ok=True)
        if not self.config.exists():
            self.config.write_text("# Preserve my config\nversion: v1.2.3\n", encoding="utf-8")
        (self.hidden / "lib").mkdir(parents=True, exist_ok=True)
        (self.hidden / "lib" / "context.ts").write_text("export interface Context {}\n", encoding="utf-8")
        (self.source / "tsconfig.json").write_text("{}\n", encoding="utf-8")

    def ccode(self, command, **kwargs):
        self.assertEqual(command[:3], ["ccode", "--config", str(self.config)])
        self.assertNotIn("shell", kwargs)
        self.assertEqual(kwargs["stdin"], subprocess.DEVNULL)
        self.assertEqual(kwargs["env"]["CCODE_HIDDEN_PATH"], str(self.hidden))
        if command[3] == "init":
            arguments = tuple(command[3:])
        else:
            self.assertEqual(command[3:7], [
                "--ccode-path", str(self.source), "--output-path", self.args.output_path,
            ])
            arguments = tuple(command[7:])
        self.calls.append(arguments)
        if self.fail:
            return subprocess.CompletedProcess(command, 1, stdout="details", stderr=self.fail)
        if arguments[0] == "init":
            self.initialized()
            output = "initialized\n"
        elif arguments[:2] == ("proc", "add"):
            path = self.source / (arguments[-1] + ".proc.ts")
            path.parent.mkdir(parents=True, exist_ok=True)
            with path.open("x", encoding="utf-8") as stream:
                stream.write("// CLI scaffold\n")
            output = ""
        else:
            self.assertEqual(arguments, ("proc", "list"))
            output = json.dumps(self.processes)
        return subprocess.CompletedProcess(command, 0, stdout=output, stderr="")

    def test_inspect_missing_project_does_not_invoke_cli(self):
        report = Author(self.args).run()
        self.assertEqual(report["status"], "initialization_required")
        self.assertIn(str(self.config), report["missing"])
        self.assertEqual(self.calls, [])
        self.assertFalse(self.source.exists())

    def test_inspect_lists_authored_files_and_support_separately(self):
        self.initialized()
        self.processes = ["nested/generate", "main"]
        for name in ("templates/model.tpl", "guides/model.md", "node_modules/skip.ts", ".git/skip"):
            path = self.source / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text("content", encoding="utf-8")
        report = Author(self.args).run()
        self.assertEqual(report["status"], "ready")
        self.assertEqual(report["processes"], ["main", "nested/generate"])
        self.assertEqual(report["files"], ["guides/model.md", "templates/model.tpl", "tsconfig.json"])
        self.assertEqual(report["support_modules"], [str(self.hidden / "lib" / "context.ts")])
        self.assertEqual(self.calls, [("proc", "list")])

    def test_init_writes_portable_quoted_config_and_checks_result(self):
        self.args.action = "init"
        report = Author(self.args).run()
        self.assertEqual(report["status"], "ready")
        self.assertTrue(report["created_config"])
        values = dict(line.split(": ", 1) for line in self.config.read_text().splitlines())
        self.assertEqual(json.loads(values["ccode_path"]), "../custom source")
        self.assertEqual(json.loads(values["hidden_path"]), "generated support")
        self.assertEqual(json.loads(values["version"]), "v1.2.3")
        self.assertEqual(self.calls, [
            ("init", str(self.source), "--version", "v1.2.3"), ("proc", "list"),
        ])

    def test_init_leaves_existing_yaml_to_cli(self):
        self.initialized()
        original = self.config.read_bytes()
        self.args.action = "init"
        report = Author(self.args).run()
        self.assertFalse(report["created_config"])
        self.assertEqual(self.config.read_bytes(), original)

    def test_init_rejects_unpinned_or_unsafe_version_before_writing(self):
        self.args.action = "init"
        for version in ("", "latest", " v1.0.0", "v1.0.0\n", "#bad", "--help"):
            with self.subTest(version=version):
                self.args.version = version
                with self.assertRaisesRegex(WorkflowError, "pinned version"):
                    Author(self.args).run()
                self.assertFalse(self.config.exists())
        self.assertEqual(self.calls, [])

    def test_failed_init_does_not_claim_success_or_continue(self):
        self.args.action = "init"
        self.fail = "initialization failed"
        with self.assertRaisesRegex(WorkflowError, "initialization failed"):
            Author(self.args).run()
        self.assertTrue(self.config.is_file())
        self.assertFalse(self.source.exists())
        self.assertEqual(len(self.calls), 1)

    def test_add_process_uses_cli_and_normalizes_optional_suffix(self):
        self.initialized()
        self.args.action = "add-process"
        self.args.name = "api\\generate.proc.ts"
        report = Author(self.args).run()
        self.assertEqual(report["process"], "api/generate")
        self.assertTrue(Path(report["path"]).is_file())
        self.assertEqual(self.calls, [("proc", "add", "--", "api/generate")])

    def test_add_process_preserves_existing_file(self):
        self.initialized()
        self.args.action = "add-process"
        self.args.name = "main"
        path = self.source / "main.proc.ts"
        path.write_text("existing work", encoding="utf-8")
        with self.assertRaisesRegex(WorkflowError, "already exists"):
            Author(self.args).run()
        self.assertEqual(path.read_text(), "existing work")
        self.assertEqual(self.calls, [])

    def test_add_process_requires_process_extension(self):
        self.initialized()
        self.args.action = "add-process"
        self.args.name = "helper.ts"
        with self.assertRaisesRegex(WorkflowError, "suffix"):
            Author(self.args).run()
        self.assertEqual(self.calls, [])

    def test_template_and_guide_preserve_utf8_and_line_endings(self):
        self.initialized()
        content = "Hello {{ data.name }} — café\r\n".encode("utf-8")
        staging = self.root / "staged.txt"
        staging.write_bytes(content)
        self.args.content_file = str(staging)
        for action, name in (("add-template", "templates/test.tpl"), ("add-guide", "guides/test.md")):
            with self.subTest(action=action):
                self.args.action, self.args.name = action, name
                report = Author(self.args).run()
                self.assertEqual(report["status"], "created")
                self.assertEqual((self.source / name).read_bytes(), content)
                with self.assertRaises(FileExistsError):
                    Author(self.args).run()
                self.assertEqual((self.source / name).read_bytes(), content)
        self.assertEqual(self.calls, [])

    def test_invalid_content_is_rejected_before_creating_target(self):
        self.initialized()
        self.args.action, self.args.name = "add-guide", "guides/new.md"
        staging = self.root / "staged.txt"
        self.args.content_file = str(staging)
        for content in (b" \n", b"\xff"):
            staging.write_bytes(content)
            with self.assertRaises((WorkflowError, UnicodeError)):
                Author(self.args).run()
            self.assertFalse((self.source / "guides").exists())

    def test_destinations_cannot_escape_source_or_edit_hidden_files(self):
        author = Author(self.args)
        for name in ("", "../escape", "/absolute", "C:\\outside", "a/../escape",
                     ".git/config", "generated support/lib/context.ts"):
            with self.subTest(name=name), self.assertRaises(WorkflowError):
                author.destination(name)

    def test_symlink_escape_is_rejected(self):
        self.initialized()
        try:
            (self.source / "linked").symlink_to(self.root, target_is_directory=True)
        except OSError:
            self.skipTest("Creating symlinks is unavailable on this platform")
        with self.assertRaisesRegex(WorkflowError, "outside hidden_path"):
            Author(self.args).destination("linked/outside.txt")

    def test_creation_requires_initialization(self):
        for action in ("add-process", "add-template", "add-guide"):
            self.args.action = action
            with self.subTest(action=action), self.assertRaisesRegex(WorkflowError, "initialization"):
                Author(self.args).run()
        self.assertEqual(self.calls, [])

    def test_inspection_rejects_unexpected_process_schema(self):
        self.initialized()
        self.processes = [{"name": "main"}]
        with self.assertRaisesRegex(WorkflowError, "process name array"):
            Author(self.args).run()


if __name__ == "__main__":
    unittest.main()
