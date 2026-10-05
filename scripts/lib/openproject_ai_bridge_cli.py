"""Lifecycle commands for the private OpenProject AI Bridge deployment."""

import argparse
import fcntl
import os
from pathlib import Path
import subprocess
import sys
import time


SERVICES = ("bridge", "tunnel-client-cody", "tunnel-client-chad")


class Lifecycle:
    def __init__(self):
        self.repo = Path(os.environ.get("OPAI_BRIDGE_REPO_DIR", "/home/andy/openproject-ai-bridge"))
        self.runtime = Path(os.environ.get("OPAI_BRIDGE_RUNTIME_DIR", "/srv/openproject-ai-bridge"))
        self.lock_file = Path(os.environ.get("OPAI_BRIDGE_LOCK_FILE", "/run/lock/openproject-ai-bridge.lock"))
        self.timeout = positive_int(os.environ.get("OPAI_BRIDGE_HEALTH_TIMEOUT", "180"), "OPAI_BRIDGE_HEALTH_TIMEOUT")
        self.env_file = self.runtime / "compose.env"
        self.compose_file = self.repo / "deploy/compose/compose.yml"
        self.secrets = self.runtime / "secrets"

    def compose(self, *args, capture=False):
        command = ["docker", "compose", "--env-file", str(self.env_file), "-f", str(self.compose_file), *args]
        return run(command, cwd=self.repo, capture=capture)

    def preflight(self):
        required = (
            self.repo / ".git",
            self.env_file,
            self.compose_file,
            self.secrets,
            self.repo / "tests/deployment/verify_compose.sh",
        )
        missing = [str(path) for path in required if not path.exists()]
        if missing:
            raise RuntimeError("Bridge-Pfade fehlen: " + ", ".join(missing))
        self.compose("config", "--quiet")
        run(
            [str(self.repo / "tests/deployment/verify_compose.sh"), "--check-secrets", str(self.secrets)],
            cwd=self.repo,
        )

    def start(self):
        self.preflight()
        self.compose("up", "-d", "--remove-orphans")
        self.wait_healthy()

    def stop(self):
        self.require_compose_files()
        self.compose("stop")

    def status(self):
        self.require_compose_files()
        self.compose("ps")
        unhealthy = []
        for service in SERVICES:
            state = self.service_state(service)
            print(f"{service}: {state}")
            if state != "healthy":
                unhealthy.append(service)
        if unhealthy:
            raise RuntimeError("Nicht bereit: " + ", ".join(unhealthy))

    def update(self):
        self.require_repo()
        dirty = run(["git", "-C", str(self.repo), "status", "--porcelain"], capture=True).strip()
        if dirty:
            raise RuntimeError("Bridge-Checkout enthaelt lokale Aenderungen; Update abgebrochen")
        run(["git", "-C", str(self.repo), "pull", "--ff-only"])
        self.redeploy()

    def redeploy(self):
        self.preflight()
        self.compose("build", "--pull", "bridge")
        self.compose("pull", "tunnel-client-cody", "tunnel-client-chad")
        self.compose("up", "-d", "--remove-orphans")
        self.wait_healthy()

    def wait_healthy(self):
        deadline = time.monotonic() + self.timeout
        pending = set(SERVICES)
        last = {}
        while pending:
            for service in tuple(pending):
                state = self.service_state(service)
                if last.get(service) != state:
                    print(f"{service}: {state}", flush=True)
                    last[service] = state
                if state == "healthy":
                    pending.remove(service)
                elif state in {"exited", "dead", "unhealthy", "missing"}:
                    raise RuntimeError(f"Dienst nicht bereit: {service} ({state})")
            if not pending:
                return
            if time.monotonic() >= deadline:
                detail = ", ".join(f"{name}={last.get(name, 'unknown')}" for name in sorted(pending))
                raise RuntimeError("Health-Timeout: " + detail)
            time.sleep(min(2, max(0, deadline - time.monotonic())))

    def service_state(self, service):
        container = self.compose("ps", "-q", service, capture=True).strip()
        if not container:
            return "missing"
        return run(
            [
                "docker",
                "inspect",
                "--format",
                "{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}",
                container,
            ],
            capture=True,
        ).strip()

    def require_repo(self):
        if not (self.repo / ".git").exists():
            raise RuntimeError(f"Bridge-Checkout fehlt: {self.repo}")

    def require_compose_files(self):
        if not self.env_file.is_file() or not self.compose_file.is_file():
            raise RuntimeError("Bridge-Compose-Konfiguration fehlt")

    def locked(self):
        self.lock_file.parent.mkdir(parents=True, exist_ok=True)
        handle = self.lock_file.open("a+")
        try:
            fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            handle.close()
            raise RuntimeError("Eine andere Bridge-Lifecycle-Aktion laeuft bereits") from None
        return handle


def positive_int(value, name):
    try:
        parsed = int(value)
    except ValueError:
        raise RuntimeError(f"{name} muss eine positive Ganzzahl sein") from None
    if parsed <= 0:
        raise RuntimeError(f"{name} muss eine positive Ganzzahl sein")
    return parsed


def run(command, cwd=None, capture=False):
    result = subprocess.run(
        command,
        cwd=cwd,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
    )
    if result.returncode != 0:
        raise RuntimeError(f"Befehl fehlgeschlagen: {Path(command[0]).name}")
    return result.stdout or ""


def main(action):
    parser = argparse.ArgumentParser(
        prog=f"opai-bridge-{action}.sh",
        description=f"OpenProject AI Bridge: {action}",
    )
    parser.parse_args()
    lifecycle = Lifecycle()
    with lifecycle.locked():
        getattr(lifecycle, action)()
    print(f"OpenProject AI Bridge {action} erfolgreich")


if __name__ == "__main__":
    try:
        main(sys.argv.pop(1))
    except (RuntimeError, KeyboardInterrupt) as error:
        print(f"Fehler: {error}", file=sys.stderr)
        raise SystemExit(1)
