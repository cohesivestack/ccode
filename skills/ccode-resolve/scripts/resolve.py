#!/usr/bin/env python3
"""Collect pending ccode artifacts without generation; edits and resolution belong to the agent."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys


class WorkflowError(Exception):
    pass


class Resolver:
    def __init__(self, args):
        self.args = args
        self.config = Path(args.config).resolve()
        self.source = Path(args.ccode_path).resolve()
        hidden = Path(args.hidden_path)
        self.hidden = (hidden if hidden.is_absolute() else self.source / hidden).resolve()
        self.output = Path(args.output_path).resolve()
        self.prefix = [
            "ccode", "--config", str(self.config),
            "--ccode-path", str(self.source), "--output-path", str(self.output),
        ]
        self.env = dict(os.environ, CCODE_HIDDEN_PATH=str(self.hidden))

    def command(self, *arguments):
        command = self.prefix + list(arguments)
        print(json.dumps(command), file=sys.stderr, flush=True)
        result = subprocess.run(
            command, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
            stderr=subprocess.PIPE, encoding="utf-8", errors="replace", env=self.env,
        )
        if result.stderr:
            print(result.stderr, file=sys.stderr, end="", flush=True)
        if result.returncode:
            if result.stdout:
                print(result.stdout, file=sys.stderr, end="", flush=True)
            raise WorkflowError(
                "{} failed (exit {}): {}".format(
                    " ".join(arguments), result.returncode,
                    result.stderr.strip() or result.stdout.strip(),
                )
            )
        return result.stdout

    def json_command(self, *arguments):
        try:
            return json.loads(self.command(*arguments))
        except json.JSONDecodeError as error:
            raise WorkflowError("Invalid JSON from {}: {}".format(" ".join(arguments), error))

    def preflight(self):
        if not self.config.is_file():
            raise WorkflowError("Project config is missing: " + str(self.config))

    def collect(self):
        metadata = self.json_command("acc", "list", "--json")
        if not isinstance(metadata, list) or any(
            not isinstance(item, dict)
            or not isinstance(item.get("artifact_id"), str)
            or not item["artifact_id"].strip()
            or not isinstance(item.get("state"), str)
            or (item.get("guide_path") is not None
                and not isinstance(item["guide_path"], str))
            for item in metadata
        ):
            raise WorkflowError("Expected artifact metadata array from ccode acc list")
        artifacts = []
        for item in sorted(metadata, key=lambda entry: entry["artifact_id"]):
            artifact = dict(item)
            artifact_id = item["artifact_id"]
            state = item["state"]
            if state == "resolved":
                continue
            if state == "missing_guide" or (
                state == "pending" and not (item.get("guide_path") or "").strip()
            ):
                artifact.update(status="needs_guide", reason="No readable guide is attached.")
            elif state != "pending":
                artifact.update(status="blocked", reason=item.get("message") or state)
            else:
                try:
                    guide = self.command("acc", "guide", artifact_id)
                    if not guide.strip():
                        artifact.update(status="needs_guide", reason="The attached guide is empty.")
                    else:
                        artifact.update(
                            guide=guide,
                            diff=self.command("acc", "diff", artifact_id),
                            generated_content=self.command("acc", "content", artifact_id),
                            status="ready",
                        )
                except (WorkflowError, OSError) as error:
                    artifact.update(status="blocked", reason=str(error))
            artifacts.append(artifact)
        return artifacts

    def run(self):
        self.preflight()
        artifacts = self.collect()
        return {
            "config": str(self.config),
            "ccode_path": str(self.source),
            "hidden_path": str(self.hidden),
            "output_path": str(self.output),
            "command_prefix": self.prefix,
            "environment": {"CCODE_HIDDEN_PATH": str(self.hidden)},
            "artifacts": artifacts,
            "status": "artifacts_pending" if artifacts else "no_pending_artifacts",
        }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", default="ccode.yaml")
    parser.add_argument("--ccode-path", required=True, help="Effective source directory")
    parser.add_argument("--output-path", default=".")
    parser.add_argument("--hidden-path", default=".ccode", help="Absolute or relative to ccode-path")
    args = parser.parse_args()
    try:
        report = Resolver(args).run()
    except (WorkflowError, OSError) as error:
        print(json.dumps({"status": "error", "message": str(error)}, ensure_ascii=False))
        return 1
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
