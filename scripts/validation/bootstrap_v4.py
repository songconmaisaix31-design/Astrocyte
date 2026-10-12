"""One-time SQLite backup only: no model math or production-table writes."""
import argparse
from contextlib import closing
import hashlib
from pathlib import Path
import sqlite3
import sys

from diagnose_native import unredirected

SOURCE = Path("C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1/replay-002/state.sqlite")
SOURCE_SHA = "279ffd78ad8cf083b1e279fb8c4ed32ae53a662b52d8a0f220f96d98c9816d4e"
ROOT = SOURCE.parents[2] / "v4-w1"


def backup(source, destination, expected_sha):
    with source.open("rb") as handle:
        before = hashlib.file_digest(handle, "sha256").hexdigest()
    if before != expected_sha:
        raise ValueError("Frozen source checksum mismatch")
    # ro prohibits database changes, including journal mode changes. Backup
    # consumes the SQLite snapshot rather than copying possibly stale pages.
    with closing(sqlite3.connect(source.as_uri() + "?mode=ro", uri=True)) as reader:
        reader.execute("PRAGMA query_only=ON")
        with closing(sqlite3.connect(destination)) as writer:
            reader.backup(writer)
    with source.open("rb") as handle:
        after = hashlib.file_digest(handle, "sha256").hexdigest()
    if after != expected_sha:
        raise ValueError("Frozen source changed during backup")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    out = args.out.absolute()
    if out.name != "state.sqlite" or out.parent.parent != ROOT or out.exists() or not unredirected(out) or not unredirected(SOURCE):
        raise ValueError("Require new bounded live state and unredirected frozen source")
    out.parent.mkdir(exist_ok=False)
    # Exclusive reservation survives a failed backup; do not silently retry.
    with out.open("xb"):
        pass
    backup(SOURCE, out, SOURCE_SHA)


if __name__ == "__main__":
    try:
        main()
    except Exception:
        print("Frozen read-only bootstrap failed; inspect LOCAL state", file=sys.stderr)
        raise SystemExit(1)
