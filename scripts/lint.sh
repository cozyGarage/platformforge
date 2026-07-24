#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> go vet"
go vet ./...

echo "==> validate lab manifests"
python3 - <<'PY'
import json, pathlib, sys
import yaml

schema_path = pathlib.Path("schemas/lab.schema.json")
schema = json.loads(schema_path.read_text())

required_top = set(schema["required"])
for path in sorted(pathlib.Path("content").rglob("lab.yaml")):
    data = yaml.safe_load(path.read_text())
    missing = required_top - set(data.keys())
    if missing:
        print(f"FAIL {path}: missing {missing}", file=sys.stderr)
        sys.exit(1)
    if data.get("version") != 1:
        print(f"FAIL {path}: version must be 1", file=sys.stderr)
        sys.exit(1)
    if not data.get("tasks"):
        print(f"FAIL {path}: requires tasks", file=sys.stderr)
        sys.exit(1)
    for task in data.get("tasks", []):
        for hint in task.get("hints", []) or []:
            if not isinstance(hint, str):
                print(f"FAIL {path}: hints must be strings (quote values with ':')", file=sys.stderr)
                sys.exit(1)
    if "content/90days/" in str(path) and not data.get("attribution"):
        print(f"FAIL {path}: 90days labs require attribution", file=sys.stderr)
        sys.exit(1)
    print(f"OK   {path}")
PY

echo "==> web tests"
cd web && npm test

echo "lint passed"
