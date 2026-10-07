#!/usr/bin/env python3
"""Prove each exercise lab is solvable and not already solved, against the real engine.

For every lab with a directory under tests/labs/solutions/:
  1. start the lab; validate; every check must FAIL on the untouched starter
  2. copy the reference solution into /workspace; validate; every check must PASS

Needs Docker, the built binary (bin/platformforge) and network for the first image pull.
Uses a throwaway HOME so your real progress database is untouched.

    python3 tests/labs/verify.py [lab-id ...]
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
import time
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOLUTIONS = ROOT / "tests" / "labs" / "solutions"
PORT = 18099
BASE = f"http://127.0.0.1:{PORT}/api"


def call(method: str, path: str):
    req = urllib.request.Request(BASE + path, method=method, data=b"" if method == "POST" else None)
    with urllib.request.urlopen(req, timeout=300) as resp:
        body = resp.read()
    return json.loads(body) if body else None


def main() -> int:
    wanted = sys.argv[1:] or sorted(p.name for p in SOLUTIONS.iterdir() if p.is_dir())
    home = tempfile.mkdtemp(prefix="pf-verify-")
    server = subprocess.Popen(
        [str(ROOT / "bin" / "platformforge"), "serve", "--addr", f"127.0.0.1:{PORT}"],
        cwd=ROOT, env={**os.environ, "HOME": home}, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
    )
    failures = 0
    try:
        for _ in range(50):
            try:
                call("GET", "/health")
                break
            except OSError:
                time.sleep(0.2)
        for lab in wanted:
            print(f"== {lab}")
            call("POST", f"/labs/{lab}/start")
            container = call("GET", f"/labs/{lab}/status")["container"]
            before = call("POST", f"/labs/{lab}/validate")
            passed_before = [c["name"] for c in before["checks"] if c["passed"]]
            if passed_before:
                failures += 1
                print(f"   FAIL starter already passes: {passed_before}")
            subprocess.run(["docker", "cp", f"{SOLUTIONS / lab}/.", f"{container}:/workspace/"], check=True)
            after = call("POST", f"/labs/{lab}/validate")
            bad = [(c["name"], c["message"][-200:]) for c in after["checks"] if not c["passed"]]
            if bad or after["status"] != "passed":
                failures += 1
                print(f"   FAIL solution does not pass: {bad}")
            else:
                print(f"   ok: starter fails {len(before['checks'])}/{len(before['checks'])}, solution passes {after['passed']}/{len(after['checks'])}")
            call("POST", f"/labs/{lab}/stop")
    finally:
        server.terminate()
        server.wait(timeout=10)
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
