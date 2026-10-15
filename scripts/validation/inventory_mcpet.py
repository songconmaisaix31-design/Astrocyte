"""Capture only the separately confirmed McPet HEAD; all subjects stay LOCAL.

The worker may invoke this only after root's explicit source confirmation.
Original frozen two-repository facts and native sessions are never recaptured.
"""
import argparse
import base64
from collections import Counter
from datetime import datetime
import json
import os
from pathlib import Path
import subprocess
import sys

from diagnose_native import START, END, unredirected

V1R_ROOT = Path("C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1")
CANDIDATE_ROOT = Path("C:/Users/DW/orca/mcpet")


def git(root, *args):
    return subprocess.run(["git", "-C", str(root), *args], capture_output=True, check=True).stdout


def selected_records(root, log):
    seen, rows = set(), []
    for line in log.splitlines():
        commit, seconds, subject = line.split(b"\0", 2)
        commit, timestamp = commit.decode("ascii"), int(seconds) * 1000
        if not START <= timestamp < END or commit in seen:
            continue
        seen.add(commit)
        rows.append({"project": str(root), "git_line_base64": base64.b64encode(line).decode("ascii"),
                     "record": {"commit": commit, "ts": timestamp, "subject": subject.decode("utf-8", errors="replace")}})
    return sorted(rows, key=lambda r: (r["record"]["ts"], r["record"]["commit"]))


def capture(confirmed_root, confirmed_remote, confirmation_reference, out):
    root, out = Path(confirmed_root).absolute(), Path(out).absolute()
    if root != CANDIDATE_ROOT or out.parent != V1R_ROOT or not confirmation_reference:
        raise ValueError("Require explicit fixed root confirmation and new V1R capture directory")
    if not root.is_dir() or not unredirected(root) or not unredirected(out):
        raise ValueError("Confirmed source/evidence root unavailable or redirected")
    if confirmed_remote != "https://github.com/songconmaisaix31-design/mcpet":
        raise ValueError("Confirmed remote differs from bounded source")
    # Reserve before reading any Git metadata or payload. Never overwrite a
    # failed capture, and never print subprocess stderr or commit subjects.
    out.mkdir(exist_ok=False)
    observed_root = git(root, "rev-parse", "--show-toplevel").decode().strip()
    remote = git(root, "config", "--get", "remote.origin.url").decode().strip()
    common = Path(git(root, "rev-parse", "--git-common-dir").decode().strip())
    if not common.is_absolute():
        common = root / common
    common = common.absolute()
    if Path(observed_root) != root or remote != confirmed_remote or not unredirected(common):
        raise ValueError("Confirmed Git identity did not match actual source")
    head = git(root, "rev-parse", "HEAD").decode("ascii").strip()
    log = git(root, "log", head, "--since-as-filter=2026-09-12T00:00:00+08:00",
              "--until=2026-10-12T00:00:00+08:00", "--format=%H%x00%ct%x00%s")
    rows = selected_records(root, log)
    zone = datetime.fromisoformat("2026-10-12T00:00:00+08:00").tzinfo
    days = Counter(datetime.fromtimestamp(row["record"]["ts"] / 1000, zone).date().isoformat() for row in rows)
    with (out / "git-log-local.jsonl").open("x", encoding="utf-8") as target:
        for row in rows:
            target.write(json.dumps(row, ensure_ascii=False) + "\n")
    source = {"approved_root": str(root), "remote": remote, "common_dir": str(common), "head": head,
              "confirmation_reference": confirmation_reference, "captured_at": datetime.now().astimezone().isoformat(),
              "window_start": int(START), "window_end_exclusive": int(END), "unique_commits": len(rows), "days": dict(sorted(days.items()))}
    with (out / "source-local.json").open("x", encoding="utf-8") as target:
        json.dump(source, target, indent=2)
    print(json.dumps({"source": "confirmed McPet", "head": head, "unique_commits": len(rows), "days": source["days"]}))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--confirmed-root", required=True)
    parser.add_argument("--confirmed-remote", required=True)
    parser.add_argument("--confirmation-reference", required=True)
    parser.add_argument("--out", required=True)
    args = parser.parse_args()
    try:
        capture(args.confirmed_root, args.confirmed_remote, args.confirmation_reference, args.out)
        return 0
    except Exception as error:
        print("McPet capture failed: " + type(error).__name__, file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
