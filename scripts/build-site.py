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


def load_readings() -> dict[str, dict]:
    """Every reading.yaml with its lesson body and quiz, keyed by id."""
    readings: dict[str, dict] = {}
    for path in sorted(CONTENT.rglob("reading.yaml")):
        data = yaml.safe_load(path.read_text())
        data["lesson"] = (path.parent / "lesson.md").read_text()
        data.setdefault("prerequisites", [])
        # Same shape as the API: the YAML key `q` is served as `prompt`.
        data["quiz"] = [{"prompt": q["q"], **{k: v for k, v in q.items() if k != "q"}} for q in data.get("quiz") or []]
        readings[data["id"]] = data
    return readings


def build_catalog(lab_meta: dict[str, dict], readings: dict[str, dict]) -> dict:
    path = yaml.safe_load(PATH_FILE.read_text())
    phases = []
    total_labs = 0
    total_readings = 0
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
            module_readings = [
                {k: readings[rid][k] for k in ("id", "title", "summary", "estimatedMinutes")}
                for rid in module.get("readings") or []
            ]
            total_readings += len(module_readings)
            total_minutes += sum(int(r["estimatedMinutes"]) for r in module_readings)
            modules.append({
                "id": module["id"],
                "title": module["title"],
                "summary": module.get("summary") or "",
                "source": module.get("source") or "",
                "readings": module_readings,
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
            "readings": total_readings,
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
    readings = load_readings()
    catalog = build_catalog(load_lab_meta(), readings)
    (DIST / "catalog.json").write_text(json.dumps(catalog, indent=2) + "\n")
    (DIST / "readings").mkdir()
    for reading_id, reading in readings.items():
        (DIST / "readings" / f"{reading_id}.json").write_text(json.dumps(reading, indent=2) + "\n")
    (DIST / ".nojekyll").write_text("")
    print(
        f"Built {DIST} with {catalog['stats']['labs']} labs, "
        f"{catalog['stats']['readings']} readings ({catalog['stats']['hours']}h)"
    )


if __name__ == "__main__":
    main()
