import json
import os
from pathlib import Path
import stat
import subprocess
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[3]
ACTIONS = ("start", "stop", "status", "update", "redeploy")


class BridgeLifecycleTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name)
        self.repo = self.base / "bridge"
        self.runtime = self.base / "runtime"
        self.bin = self.base / "bin"
        self.log = self.base / "commands.jsonl"
        (self.repo / ".git").mkdir(parents=True)
        (self.repo / "deploy/compose").mkdir(parents=True)
        (self.repo / "tests/deployment").mkdir(parents=True)
        (self.repo / "deploy/compose/compose.yml").write_text("services: {}\n")
        verify = self.repo / "tests/deployment/verify_compose.sh"
        verify.write_text("#!/bin/sh\nexit 0\n")
        verify.chmod(0o755)
        (self.runtime / "secrets").mkdir(parents=True)
        (self.runtime / "compose.env").write_text("SECRETS_DIR=/runtime/secrets\n")
        self.bin.mkdir()
        self.fake_command("docker", self.docker_script())
        self.fake_command("git", self.git_script())
        self.env = {
            **os.environ,
            "PATH": str(self.bin) + os.pathsep + os.environ["PATH"],
            "OPAI_BRIDGE_REPO_DIR": str(self.repo),
            "OPAI_BRIDGE_RUNTIME_DIR": str(self.runtime),
            "OPAI_BRIDGE_LOCK_FILE": str(self.base / "lifecycle.lock"),
            "OPAI_BRIDGE_HEALTH_TIMEOUT": "1",
            "FAKE_COMMAND_LOG": str(self.log),
        }

    def fake_command(self, name, body):
        path = self.bin / name
        path.write_text("#!/usr/bin/env python3\n" + textwrap.dedent(body))
        path.chmod(path.stat().st_mode | stat.S_IXUSR)

    @staticmethod
    def docker_script():
        return r'''
import json, os, sys
with open(os.environ["FAKE_COMMAND_LOG"], "a") as output:
    output.write(json.dumps(["docker", *sys.argv[1:]]) + "\n")
args = sys.argv[1:]
if args and args[0] == "inspect":
    print("healthy")
elif "ps" in args and "-q" in args:
    print(args[-1] + "-container")
'''

    @staticmethod
    def git_script():
        return r'''
import json, os, sys
with open(os.environ["FAKE_COMMAND_LOG"], "a") as output:
    output.write(json.dumps(["git", *sys.argv[1:]]) + "\n")
args = sys.argv[1:]
if args[-2:] == ["status", "--porcelain"] and os.environ.get("FAKE_GIT_DIRTY") == "1":
    print(" M local-change")
'''

    def run_action(self, action, extra_env=None):
        env = {**self.env, **(extra_env or {})}
        return subprocess.run(
            [str(ROOT / f"scripts/opai-bridge-{action}.sh")],
            cwd=self.base,
            env=env,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )

    def commands(self):
        if not self.log.exists():
            return []
        return [json.loads(line) for line in self.log.read_text().splitlines()]

    def test_all_entrypoints_offer_help_from_any_working_directory(self):
        for action in ACTIONS:
            result = subprocess.run(
                [str(ROOT / f"scripts/opai-bridge-{action}.sh"), "--help"],
                cwd=self.base,
                env=self.env,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            self.assertEqual(result.returncode, 0, (action, result.stderr))
            self.assertIn(f"opai-bridge-{action}.sh", result.stdout)

    def test_start_validates_then_starts_and_checks_every_service(self):
        result = self.run_action("start")
        self.assertEqual(result.returncode, 0, result.stderr)
        commands = self.commands()
        rendered = [" ".join(command) for command in commands]
        self.assertTrue(any("compose.yml config --quiet" in command for command in rendered))
        self.assertTrue(any("compose.yml up -d" in command for command in rendered))
        for service in ("bridge", "tunnel-client-cody", "tunnel-client-chad"):
            self.assertTrue(any(f"ps -q {service}" in command for command in rendered), service)

    def test_update_refuses_dirty_checkout_before_fetch_or_deploy(self):
        result = self.run_action("update", {"FAKE_GIT_DIRTY": "1"})
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("lokale Aenderungen", result.stderr)
        rendered = [" ".join(command) for command in self.commands()]
        self.assertFalse(any(" fetch " in f" {command} " for command in rendered))
        self.assertFalse(any(" up -d" in command for command in rendered))

    def test_redeploy_builds_before_recreating_services(self):
        result = self.run_action("redeploy")
        self.assertEqual(result.returncode, 0, result.stderr)
        rendered = [" ".join(command) for command in self.commands()]
        build = next(i for i, command in enumerate(rendered) if " build --pull bridge" in command)
        deploy = next(i for i, command in enumerate(rendered) if " up -d" in command)
        self.assertLess(build, deploy)


if __name__ == "__main__":
    unittest.main()
