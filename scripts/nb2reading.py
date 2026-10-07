#!/usr/bin/env python3
"""Convert a Jupyter notebook into a PlatformForge reading unit (reading.yaml + lesson.md).

Markdown cells are kept as-is, code cells become fenced blocks, outputs are dropped.
Optionally lifts selected multiple-choice questions from a notebook-style MCQ bank
(markdown cells with **Qn.** and - **a)** options, answer_key dict in the last code cell).

    nb2reading.py NOTEBOOK.ipynb --id helm-charts --summary "..." --minutes 25 \
        --out content/theory/helm-charts [--source "course/notebook"] \
        [--mcq mcq.ipynb --questions 12,13]
"""

from __future__ import annotations

import argparse
import ast
import json
import re
import sys
from pathlib import Path

import yaml


def cells(path: Path) -> list[dict]:
    return json.loads(path.read_text())["cells"]


def estimate_minutes(body: str) -> int:
    code = body.count("```python") + body.count("```sql")
    words = len(body.split())
    return max(10, 5 * round((words / 160 + 4 * code) / 5))


def clean_markdown(src: str) -> str:
    """Remove notebook-sequence navigation and links to files that do not travel with a reading."""
    lines = []
    for line in src.split("\n"):
        if re.match(r"^\*\*Next\*\*:\s*Notebook \d+", line):
            continue
        if re.match(r"^- Notebook \d+ [—-] ", line) or line.strip() == "**Connections back to this course**":
            continue
        lines.append(line)
    text = "\n".join(lines)
    text = re.sub(r"\[([^\]]+)\]\(\.{1,2}/[^)]*\)", r"\1", text)  # relative links -> plain text
    text = text.replace("this notebook", "this lesson").replace("This notebook", "This lesson")
    return text.strip()


def lesson_markdown(path: Path) -> tuple[str, str]:
    title, parts = "", []
    for i, cell in enumerate(cells(path)):
        src = "".join(cell["source"]).rstrip()
        if not src:
            continue
        if cell["cell_type"] == "markdown":
            if src.startswith("## Vault Insights"):
                continue  # book-derived summaries: not first-party, do not republish
            if src.startswith("## After This Notebook"):
                continue  # notebook-sequence navigation; the path orders units now
            if src.startswith("## Going Deeper") and "Find a well-regarded" in src:
                continue  # generated filler, not course content
            if src.startswith("## Hands-on Lab") and "](./lab_" in src:
                continue  # points at a lab folder that is not part of this unit
            src = clean_markdown(src)
            if not src:
                continue
            src = "\n".join(l for l in src.split("\n") if "Vault Insights" not in l)
            if i == 0 and src.startswith("# "):
                heading, _, rest = src.partition("\n")
                title = re.sub(r"^\d+:\s*", "", heading[2:].strip())
                # drop notebook-numbering lines such as "## Notebook 05: Alerting & On-Call"
                rest = "\n".join(l for l in rest.split("\n") if not re.match(r"^## Notebook \d+", l))
                src = rest.strip()
                if not src:
                    continue
            parts.append(src)
        else:
            parts.append(f"```python\n{src}\n```")
    return title, "\n\n".join(parts) + "\n"


def questions(path: Path, wanted: set[int]) -> list[dict]:
    md = "\n".join("".join(c["source"]) for c in cells(path) if c["cell_type"] == "markdown")
    code = "\n".join("".join(c["source"]) for c in cells(path) if c["cell_type"] == "code")
    match = re.search(r"answer_key\s*=\s*(\{.*?\}|\[.*?\])", code, re.S)
    if not match:
        sys.exit("no answer_key found in MCQ notebook")
    key = ast.literal_eval(match.group(1))
    if isinstance(key, list):  # some banks use a list indexed from question 1
        key = {i + 1: letter for i, letter in enumerate(key)}
    out = []
    for m in re.finditer(r"\*\*Q(\d+)\.\*\*\s*(.*?)(?=\n\*\*Q\d+\.\*\*|\Z)", md, re.S):
        n = int(m.group(1))
        if n not in wanted:
            continue
        body = m.group(2).strip()
        prompt, _, opts = body.partition("\n\n")
        # option styles seen: "- **a)** text", "- A) text" and "- **A.** text"
        options = re.findall(r"^- (?:\*\*)?([a-dA-D])[.)](?:\*\*)?\s*(.*)$", opts, re.M)
        letters = [letter.lower() for letter, _ in options]
        out.append(
            {
                "q": prompt.strip(),
                "options": [text.strip() for _, text in options],
                "answer": letters.index(str(key[n]).lower()),
            }
        )
    if len(out) != len(wanted):
        sys.exit(f"found {len(out)} of {len(wanted)} requested questions")
    return out


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("notebook", type=Path)
    ap.add_argument("--id", required=True)
    ap.add_argument("--title")
    ap.add_argument("--summary", required=True)
    ap.add_argument("--minutes", type=int, help="default: estimated from length and code cells")
    ap.add_argument("--out", type=Path, required=True)
    ap.add_argument("--source")
    ap.add_argument("--prereq", action="append", default=[])
    ap.add_argument("--mcq", type=Path)
    ap.add_argument("--questions", default="")
    args = ap.parse_args()

    title, body = lesson_markdown(args.notebook)
    meta = {
        "version": 1,
        "id": args.id,
        "title": args.title or title,
        "summary": args.summary,
        "estimatedMinutes": args.minutes or estimate_minutes(body),
        "prerequisites": args.prereq,
    }
    if args.source:
        meta["source"] = args.source
    if args.mcq:
        wanted = {int(n) for n in args.questions.split(",") if n}
        meta["quiz"] = questions(args.mcq, wanted)
    args.out.mkdir(parents=True, exist_ok=True)
    (args.out / "lesson.md").write_text(body)
    (args.out / "reading.yaml").write_text(yaml.safe_dump(meta, sort_keys=False, allow_unicode=True, width=100))
    print(f"wrote {args.out}")


if __name__ == "__main__":
    main()
