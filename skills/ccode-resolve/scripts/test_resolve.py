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

from resolve import Resolver, WorkflowError


class ResolverTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        root = Path(self.temp.name).resolve()
        source = root / "custom source"
        hidden = root / "custom hidden"
        source.mkdir()
        config = root / "custom.yaml"
        config.write_text("version: v1.0.0\n", encoding="utf-8")
        self.args = argparse.Namespace(
            config=str(config), ccode_path=str(source), output_path=str(root / "output"),
            hidden_path=str(hidden),
        )
        self.calls = []
        self.responses = {("acc", "list", "--json"): []}
        self.addCleanup(patch.stopall)
        patch("resolve.subprocess.run", side_effect=self.ccode).start()
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

    def test_missing_config_stops_before_cli(self):
        Path(self.args.config).unlink()
        with self.assertRaisesRegex(WorkflowError, "config is missing"):
            Resolver(self.args).run()
        self.assertEqual(self.calls, [])

    def test_no_pending_work_does_not_discover_or_run_processes(self):
        report = Resolver(self.args).run()
        self.assertEqual(report["status"], "no_pending_artifacts")
        self.assertEqual(report["artifacts"], [])
        self.assertEqual(report["config"], self.args.config)
        self.assertEqual(report["ccode_path"], self.args.ccode_path)
        self.assertEqual(report["hidden_path"], self.args.hidden_path)
        self.assertEqual(report["output_path"], self.args.output_path)
        self.assertEqual(self.calls, [("acc", "list", "--json")])
        # Resolution needs no TypeScript runtime or generation support files.
        self.assertFalse((Path(self.args.ccode_path) / "tsconfig.json").exists())
        self.assertFalse(Path(self.args.hidden_path).exists())

    def test_invalid_json_stops_collection(self):
        self.responses[("acc", "list", "--json")] = "not JSON"
        with self.assertRaisesRegex(WorkflowError, "Invalid JSON"):
            Resolver(self.args).run()
        self.assertEqual(self.calls, [("acc", "list", "--json")])

    def test_failed_list_is_not_reported_as_no_pending_work(self):
        self.responses[("acc", "list", "--json")] = subprocess.CompletedProcess(
            [], 1, stdout="", stderr="cannot read accelerator state"
        )
        with self.assertRaisesRegex(WorkflowError, "cannot read accelerator state"):
            Resolver(self.args).run()
        self.assertEqual(self.calls, [("acc", "list", "--json")])

    def test_invalid_metadata_is_rejected_before_reading_artifacts(self):
        for metadata in ({}, ["artifact"], [{"artifact_id": "x"}],
                         [{"artifact_id": "x", "state": "pending", "guide_path": 42}]):
            with self.subTest(metadata=metadata):
                self.responses[("acc", "list", "--json")] = metadata
                with self.assertRaisesRegex(WorkflowError, "artifact metadata array"):
                    Resolver(self.args).run()
        self.assertTrue(all(call == ("acc", "list", "--json") for call in self.calls))

    def test_refresh_never_runs_generation(self):
        for _ in range(2):
            self.assertEqual(Resolver(self.args).run()["status"], "no_pending_artifacts")
        self.assertEqual(self.calls, [("acc", "list", "--json")] * 2)

    def test_collects_ready_and_skips_missing_guides_and_problem_states(self):
        self.responses[("acc", "list", "--json")] = [
            {"artifact_id": "ready.ts", "state": "pending", "guide_path": "guide.md"},
            {"artifact_id": "no-guide.ts", "state": "pending", "guide_path": None},
            {"artifact_id": "lost-guide.ts", "state": "missing_guide", "guide_path": "lost.md"},
            {"artifact_id": "conflict.ts", "state": "ambiguous", "message": "repair required"},
            {"artifact_id": "done.ts", "state": "resolved"},
        ]
        self.responses[("acc", "guide", "ready.ts")] = "Preserve edits; add logging.\n"
        self.responses[("acc", "diff", "ready.ts")] = "-generated\n+current\n"
        self.responses[("acc", "content", "ready.ts")] = "generated\n"
        report = Resolver(self.args).run()
        items = {item["artifact_id"]: item for item in report["artifacts"]}
        self.assertEqual(items["ready.ts"]["status"], "ready")
        self.assertEqual(items["ready.ts"]["diff"], "-generated\n+current\n")
        self.assertEqual(items["ready.ts"]["generated_content"], "generated\n")
        self.assertEqual(items["no-guide.ts"]["status"], "needs_guide")
        self.assertEqual(items["lost-guide.ts"]["status"], "needs_guide")
        self.assertEqual(items["conflict.ts"]["status"], "blocked")
        self.assertNotIn("done.ts", items)
        self.assertEqual(self.calls, [
            ("acc", "list", "--json"), ("acc", "guide", "ready.ts"),
            ("acc", "diff", "ready.ts"), ("acc", "content", "ready.ts"),
        ])

    def test_inspection_failure_does_not_prevent_other_artifacts(self):
        self.responses[("acc", "list", "--json")] = [
            {"artifact_id": "a.ts", "state": "pending", "guide_path": "guide.md"},
            {"artifact_id": "b.ts", "state": "pending", "guide_path": "guide.md"},
        ]
        for artifact in ("a.ts", "b.ts"):
            self.responses[("acc", "guide", artifact)] = "guide"
            self.responses[("acc", "diff", artifact)] = ""
            self.responses[("acc", "content", artifact)] = ""
        self.responses[("acc", "diff", "a.ts")] = subprocess.CompletedProcess(
            [], 1, stdout="", stderr="cannot read output"
        )
        items = Resolver(self.args).run()["artifacts"]
        self.assertEqual([item["status"] for item in items], ["blocked", "ready"])
        self.assertEqual(items[1]["diff"], "")

    def test_empty_guide_stays_pending(self):
        self.responses[("acc", "list", "--json")] = [
            {"artifact_id": "a.ts", "state": "pending", "guide_path": "empty.md"},
        ]
        self.responses[("acc", "guide", "a.ts")] = " \n"
        self.assertEqual(Resolver(self.args).run()["artifacts"][0]["status"], "needs_guide")
        self.assertEqual(self.calls, [("acc", "list", "--json"), ("acc", "guide", "a.ts")])


if __name__ == "__main__":
    unittest.main()
