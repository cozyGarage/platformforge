#!/usr/bin/env python3
"""Build the static GitHub Pages site with a generated curriculum catalog."""

from __future__ import annotations

import json
import shutil
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
SITE = ROOT / "site"
DIST = ROOT / "site-dist"
CONTENT = ROOT / "content"
PATH_FILE = CONTENT / "paths" / "devops-engineer.yaml"


def load_lab_meta() -> dict[str, dict]:
    labs: dict[str, dict] = {}
    for path in CONTENT.rglob("lab.yaml"):
        data = yaml.safe_load(path.read_text())
        labs[data["id"]] = {
            "id": data["id"],
            "title": data["title"],
            "summary": data["summary"],
            "difficulty": data["difficulty"],
            "estimatedMinutes": data["estimatedMinutes"],
            "prerequisites": data.get("prerequisites") or [],
        }
    return labs


def build_catalog(lab_meta: dict[str, dict]) -> dict:
    path = yaml.safe_load(PATH_FILE.read_text())
    phases = []
    total_labs = 0
    total_minutes = 0
    for phase in path.get("phases", []):
        modules = []
        for module in phase.get("modules", []):
            labs = []
            for lab_id in module.get("labs") or []:
                meta = lab_meta.get(lab_id, {
                    "id": lab_id,
                    "title": lab_id,
                    "summary": "",
                    "difficulty": "unknown",
                    "estimatedMinutes": 0,
                    "prerequisites": [],
                })
                labs.append(meta)
                total_labs += 1
                total_minutes += int(meta.get("estimatedMinutes") or 0)
            modules.append({
                "id": module["id"],
                "title": module["title"],
                "summary": module.get("summary") or "",
                "source": module.get("source") or "",
                "labs": labs,
                "comingSoon": module.get("comingSoon") or [],
                "unlock": module.get("unlock") or None,
            })
        phases.append({
            "id": phase["id"],
            "title": phase["title"],
            "summary": phase.get("summary") or "",
            "modules": modules,
        })
    return {
        "path": {
            "id": path["id"],
            "title": path["title"],
            "summary": path["summary"],
            "source": path.get("source") or "",
        },
        "stats": {
            "labs": total_labs,
            "minutes": total_minutes,
            "hours": round(total_minutes / 60, 1),
        },
        "phases": phases,
        "repo": "https://github.com/cozyGarage/platformforge",
        "companion": {
            "name": "PatchLab",
            "url": "https://cozygarage.github.io/patchlab/",
            "repo": "https://github.com/cozyGarage/patchlab",
        },
    }


def main() -> None:
    if DIST.exists():
        shutil.rmtree(DIST)
    shutil.copytree(SITE, DIST, ignore=shutil.ignore_patterns(".*"))
    catalog = build_catalog(load_lab_meta())
    (DIST / "catalog.json").write_text(json.dumps(catalog, indent=2) + "\n")
    (DIST / ".nojekyll").write_text("")
    print(f"Built {DIST} with {catalog['stats']['labs']} labs ({catalog['stats']['hours']}h)")


if __name__ == "__main__":
    main()
