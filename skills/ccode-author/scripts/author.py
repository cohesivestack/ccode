#!/usr/bin/env python3
"""Initialize and inspect ccode projects; create processes, templates, and guides."""

import argparse
import json
import os
from pathlib import Path, PurePosixPath, PureWindowsPath
import re
import subprocess
import sys


class WorkflowError(Exception):
    pass


def inside(path, directory):
    try:
        path.relative_to(directory)
        return True
    except ValueError:
        return False


def relative_or_absolute(path, directory):
    try:
        return Path(os.path.relpath(path, directory)).as_posix()
    except ValueError:  # Different Windows drives cannot have a relative path.
        return path.as_posix()


class Author:
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

    def command(self, *arguments, initializing=False):
        prefix = ["ccode", "--config", str(self.config)] if initializing else self.prefix
        command = prefix + list(arguments)
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
            raise WorkflowError("{} failed (exit {}): {}".format(
                " ".join(arguments), result.returncode,
                result.stderr.strip() or result.stdout.strip(),
            ))
        return result.stdout

    def missing(self):
        required = [
            (self.config, False), (self.source, True),
            (self.source / "tsconfig.json", False),
            (self.hidden / "lib" / "context.ts", False),
        ]
        return [str(path) for path, directory in required
                if not (path.is_dir() if directory else path.is_file())]

    def preflight(self):
        missing = self.missing()
        if missing:
            raise WorkflowError("Project initialization is required. Missing: " + ", ".join(missing))

    def inspect(self):
        report = {
            "config": str(self.config), "ccode_path": str(self.source),
            "hidden_path": str(self.hidden), "output_path": str(self.output),
            "command_prefix": self.prefix,
            "environment": {"CCODE_HIDDEN_PATH": str(self.hidden)},
            "missing": self.missing(),
        }
        if report["missing"]:
            report["status"] = "initialization_required"
            return report
        try:
            processes = json.loads(self.command("proc", "list"))
        except json.JSONDecodeError as error:
            raise WorkflowError("Invalid JSON from ccode proc list: " + str(error))
        if not isinstance(processes, list) or any(
            not isinstance(name, str) or not name.strip() for name in processes
        ):
            raise WorkflowError("Expected process name array from ccode proc list")
        files = []
        for directory, folders, filenames in os.walk(self.source):
            root = Path(directory)
            folders[:] = sorted(folder for folder in folders
                                if not folder.startswith(".") and folder != "node_modules"
                                and not inside((root / folder).resolve(), self.hidden))
            for filename in sorted(filenames):
                path = root / filename
                if not filename.startswith(".") and not path.is_symlink():
                    files.append(path.relative_to(self.source).as_posix())
        report.update(
            status="ready", processes=sorted(processes), files=sorted(files),
            support_modules=sorted(str(path) for path in (self.hidden / "lib").rglob("*.ts")),
        )
        return report

    def initialize(self):
        version = self.args.version
        if version == "latest" or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._+-]*", version):
            raise WorkflowError("Supply an explicit pinned version, not latest or a blank value")
        # Emit quoted scalar values, not a handwritten YAML parser. Existing YAML
        # belongs to the user and is read/validated by the CLI, never reserialized.
        created_config = False
        if not self.config.exists():
            self.config.parent.mkdir(parents=True, exist_ok=True)
            values = {
                "ccode_path": relative_or_absolute(self.source, self.config.parent),
                "version": version,
                "output_path": relative_or_absolute(self.output, Path.cwd()),
                "hidden_path": relative_or_absolute(self.hidden, self.source),
            }
            with self.config.open("x", encoding="utf-8", newline="\n") as stream:
                for key, value in values.items():
                    stream.write("{}: {}\n".format(key, json.dumps(value, ensure_ascii=False)))
            created_config = True
        output = self.command("init", str(self.source), "--version", version, initializing=True)
        if output:
            print(output, file=sys.stderr, end="", flush=True)
        self.preflight()
        report = self.inspect()
        report.update(action="init", created_config=created_config, version=version)
        return report

    def destination(self, name):
        normalized = PurePosixPath(name.replace("\\", "/"))
        if (not normalized.parts or normalized.is_absolute() or PureWindowsPath(name).drive
                or any(part.startswith(".") for part in normalized.parts)):
            raise WorkflowError("Use a non-hidden path relative to ccode_path: " + name)
        path = (self.source / Path(*normalized.parts)).resolve()
        if not inside(path, self.source) or inside(path, self.hidden):
            raise WorkflowError("Destination must stay in authored source outside hidden_path: " + name)
        return path

    def add_process(self):
        self.preflight()
        name = self.args.name.replace("\\", "/")
        if name.endswith(".ts") and not name.endswith(".proc.ts"):
            raise WorkflowError("Process names must omit the suffix or end in .proc.ts")
        target = self.destination(name if name.endswith(".proc.ts") else name + ".proc.ts")
        if target.exists():
            raise WorkflowError("Process already exists; edit it instead: " + str(target))
        process = target.relative_to(self.source).as_posix()[:-len(".proc.ts")]
        self.command("proc", "add", "--", process)
        if not target.is_file():
            raise WorkflowError("ccode proc add did not create " + str(target))
        return {"status": "created", "action": "add-process", "process": process, "path": str(target)}

    def add_file(self):
        self.preflight()
        target = self.destination(self.args.name)
        content = Path(self.args.content_file).read_bytes()
        content.decode("utf-8")  # Verify before creating any destination directories.
        if not content.strip():
            raise WorkflowError("Templates and guides must have nonblank UTF-8 content")
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open("xb") as stream:
            stream.write(content)
        return {"status": "created", "action": self.args.action, "path": str(target)}

    def run(self):
        if self.args.action == "inspect":
            return self.inspect()
        if self.args.action == "init":
            return self.initialize()
        if self.args.action == "add-process":
            return self.add_process()
        return self.add_file()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", default="ccode.yaml")
    parser.add_argument("--ccode-path", required=True, help="Effective source directory")
    parser.add_argument("--output-path", default=".")
    parser.add_argument("--hidden-path", default=".ccode", help="Absolute or relative to ccode-path")
    actions = parser.add_subparsers(dest="action", required=True)
    actions.add_parser("inspect", help="Report initialization evidence, processes, and source files")
    init = actions.add_parser("init", help="Initialize or refresh using an explicit version")
    init.add_argument("--version", required=True)
    process = actions.add_parser("add-process", help="Create a process with ccode proc add")
    process.add_argument("name")
    for action in ("add-template", "add-guide"):
        child = actions.add_parser(action, help="Create a UTF-8 file without overwriting")
        child.add_argument("name", help="Destination relative to ccode_path")
        child.add_argument("--content-file", required=True, help="File containing the authored content")
    args = parser.parse_args()
    try:
        report = Author(args).run()
    except (WorkflowError, OSError, UnicodeError) as error:
        print(json.dumps({"status": "error", "message": str(error)}, ensure_ascii=False))
        return 1
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
