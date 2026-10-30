"""Behavioral checks with an isolated project and a mocked ccode executable."""

import argparse
from contextlib import redirect_stderr
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from produce import Producer, WorkflowError


class ProducerTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        root = Path(self.temp.name).resolve()
        source = root / "custom source"
        hidden = root / "custom hidden"
        source.mkdir()
        (source / "tsconfig.json").write_text("{}", encoding="utf-8")
        (hidden / "lib").mkdir(parents=True)
        (hidden / "lib" / "context.ts").write_text("", encoding="utf-8")
        config = root / "custom.yaml"
        config.write_text("version: v1.0.0\n", encoding="utf-8")
        self.args = argparse.Namespace(
            config=str(config), ccode_path=str(source), output_path=str(root / "output"),
            hidden_path=str(hidden), process=None,
        )
        self.calls = []
        self.responses = {("proc", "list"): ["main"], ("run", "main"): "produced\n"}
        self.addCleanup(patch.stopall)
        patch("produce.subprocess.run", side_effect=self.ccode).start()
        self.stderr = io.StringIO()
        self.redirect = redirect_stderr(self.stderr)
        self.redirect.__enter__()
        self.addCleanup(self.redirect.__exit__, None, None, None)

    def ccode(self, command, **kwargs):
        # Every call must carry the same config, paths, and hidden directory.
        self.assertEqual(command[:7], [
            "ccode", "--config", self.args.config, "--ccode-path", self.args.ccode_path,
            "--output-path", self.args.output_path,
        ])
        self.assertEqual(kwargs["env"]["CCODE_HIDDEN_PATH"], self.args.hidden_path)
        self.assertNotIn("shell", kwargs)
        self.assertEqual(kwargs["stdin"], subprocess.DEVNULL)
        arguments = tuple(command[7:])
        self.calls.append(arguments)
        response = self.responses[arguments]
        if isinstance(response, Exception):
            raise response
        if isinstance(response, subprocess.CompletedProcess):
            return response
        output = json.dumps(response) if isinstance(response, (list, dict)) else response
        return subprocess.CompletedProcess(command, 0, stdout=output, stderr="")

    def test_missing_initialization_stops_before_cli(self):
        for filename in (self.args.config,
                         str(Path(self.args.hidden_path) / "lib" / "context.ts")):
            with self.subTest(missing=filename):
                path = Path(filename)
                contents = path.read_bytes()
                path.unlink()
                with self.assertRaisesRegex(WorkflowError, "initialization"):
                    Producer(self.args).run()
                path.write_bytes(contents)
        self.assertEqual(self.calls, [])

    def test_main_is_selected_and_output_is_logged(self):
        self.responses[("proc", "list")] = ["nested/main", "main", "other"]
        report = Producer(self.args).run()
        self.assertEqual(report["process"], "main")
        self.assertEqual(report["status"], "generated")
        self.assertEqual(report["next_skill"], "ccode-resolve")
        for name in ("config", "ccode_path", "hidden_path", "output_path"):
            self.assertEqual(report[name], getattr(self.args, name))
        self.assertEqual(self.calls, [("proc", "list"), ("run", "main")])
        self.assertIn("produced", self.stderr.getvalue())

    def test_explicit_process_skips_discovery(self):
        self.args.process = "nested/chosen.proc.ts"
        self.responses[("run", self.args.process)] = ""
        Producer(self.args).run()
        self.assertEqual(self.calls, [("run", self.args.process)])

    def test_selection_is_required_even_for_one_process(self):
        self.responses[("proc", "list")] = ["nested/main"]
        report = Producer(self.args).run()
        self.assertEqual(report["status"], "selection_required")
        self.assertNotIn("next_skill", report)
        self.assertEqual(report["available_processes"], ["nested/main"])
        self.assertEqual(self.calls, [("proc", "list")])

    def test_no_processes_does_not_run(self):
        self.responses[("proc", "list")] = []
        report = Producer(self.args).run()
        self.assertEqual(report["status"], "no_processes")
        self.assertNotIn("next_skill", report)
        self.assertEqual(self.calls, [("proc", "list")])

    def test_failed_run_never_lists_accelerators(self):
        self.responses[("run", "main")] = subprocess.CompletedProcess(
            [], 1, stdout="", stderr="process failed"
        )
        with self.assertRaisesRegex(WorkflowError, "process failed"):
            Producer(self.args).run()
        self.assertEqual(self.calls, [("proc", "list"), ("run", "main")])

    def test_invalid_json_stops_before_running(self):
        self.responses[("proc", "list")] = "not JSON"
        with self.assertRaisesRegex(WorkflowError, "Invalid JSON"):
            Producer(self.args).run()
        self.assertEqual(self.calls, [("proc", "list")])


if __name__ == "__main__":
    unittest.main()
