"""Online read-only source backup; publish completed generations atomically.

No payloads, SQL errors or filenames are emitted. SQLite backup includes live WAL.
"""
import argparse
import math
import contextlib
import datetime
import json
import os
import pathlib
import sqlite3
import sys
import tempfile
import time


def publish(root, source, history, timeout=120):
    root = pathlib.Path(root)
    root.mkdir(parents=True, exist_ok=True)
    # OS file locking serializes manual and scheduled refresh, without a daemon.
    with (root / "snapshot.lock").open("a+b") as lock:
        if sys.platform == "win32":
            import msvcrt
            lock.seek(0)
            if not lock.read(1):
                lock.write(b"0")
                lock.flush()
            lock.seek(0)
            msvcrt.locking(lock.fileno(), msvcrt.LK_NBLCK, 1)
        else:
            import fcntl
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        return _publish(root, source, history, timeout)


def _publish(root, source, history, timeout=120):
    root, source, history = map(pathlib.Path, (root, source, history))
    source = source.resolve(strict=True)
    # Source and destination must have disjoint ownership.
    if source.is_relative_to(root.resolve()):
        raise ValueError("invalid source")
    generations = root / "snapshots"
    generations.mkdir(parents=True, exist_ok=True)
    generation = pathlib.Path(tempfile.mkdtemp(prefix="snapshot-", dir=generations))
    deadline = time.monotonic() + timeout

    def progress(status, remaining, total):
        if time.monotonic() > deadline:
            raise TimeoutError("backup deadline")

    try:
        # Pin a read transaction: concurrent WAL writers proceed while backup
        # copies one coherent view, rather than indefinitely restarting.
        with contextlib.closing(sqlite3.connect(source.as_uri() + "?mode=ro", uri=True, timeout=2)) as src:
            src.execute("PRAGMA query_only=ON")
            src.execute("BEGIN")
            src.execute("SELECT count(*) FROM sqlite_master").fetchone()
            with contextlib.closing(sqlite3.connect(generation / "snapshot.sqlite")) as dst:
                src.backup(dst, pages=256, progress=progress, sleep=0.01)
                if dst.execute("PRAGMA quick_check").fetchone() != ("ok",):
                    raise ValueError("invalid backup")
                # Only the owned copy changes journal mode.
                dst.execute("PRAGMA journal_mode=DELETE")
                dst.execute("SELECT count(*) FROM attention_engine_events").fetchone()
        safe = json.loads(history.read_text(encoding="utf-8-sig"))
        projected = {}
        for topic, points in safe.items():
            projected[topic] = []
            for point in points:
                ts, attention = point["ts"], point["attention"]
                if not isinstance(ts, int) or not isinstance(attention, (float, int)) or not math.isfinite(attention) or not 0 <= attention <= 1:
                    raise ValueError("invalid history")
                projected[topic].append({"ts": ts, "attention": attention})
        # Retain previously confirmed snapshot state points. Timestamp comes
        # from node.updated_at, never from the wall clock of this backup.
        current = root / "current.json"
        if current.exists():
            previous = json.loads(current.read_text(encoding="utf-8"))["generation"]
            if pathlib.Path(previous).name != previous or previous in (".", ".."):
                raise ValueError("invalid generation")
            old = json.loads((generations / previous / "history.json").read_text(encoding="utf-8"))
            for topic, points in old.items():
                known = {p["ts"] for p in projected.setdefault(topic, [])}
                projected[topic].extend(p for p in points if p["ts"] not in known)
        with contextlib.closing(sqlite3.connect((generation / "snapshot.sqlite").as_uri() + "?mode=ro", uri=True)) as db:
            for topic, ts, attention in db.execute("SELECT topic,updated_at,attention FROM attention_engine_nodes ORDER BY topic"):
                if ts is None or ts <= 0:
                    continue
                points = projected.setdefault(topic, [])
                if not any(p["ts"] == ts for p in points):
                    points.append({"ts": ts, "attention": attention})
        for points in projected.values():
            points.sort(key=lambda p: p["ts"])
        (generation / "history.json").write_text(json.dumps(projected, allow_nan=False), encoding="utf-8")
        publication = {"generation": generation.name, "snapshot_at": datetime.datetime.now(datetime.timezone.utc).isoformat()}
        # Same-volume replace publishes DB and history together; old readers
        # keep their immutable generation, subsequent requests reopen the pointer.
        fd, tmp = tempfile.mkstemp(prefix="publication-", suffix=".json", dir=root)
        with os.fdopen(fd, "w", encoding="utf-8") as stream:
            json.dump(publication, stream)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(tmp, root / "current.json")
        return publication
    except Exception:
        # Incomplete generations are never published; retained for local diagnosis.
        raise


SOURCE = pathlib.Path(r'C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w1/formal-20261013-first/state.sqlite')
HISTORY = pathlib.Path(r'C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1/replay-002/topic-timeseries-safe.json')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", required=True)
    args = parser.parse_args()
    try:
        config = json.loads(pathlib.Path(args.config).read_text(encoding="utf-8-sig"))
        root = pathlib.Path(os.environ["LOCALAPPDATA"]) / "astrocyte-workbench" / "v0.1"
        publication = publish(root, SOURCE, HISTORY)
        print(json.dumps({"status": "ok", "snapshot_at": publication["snapshot_at"]}))
    except Exception:
        print('{"status":"error","code":"SNAPSHOT_UNAVAILABLE"}')
        raise SystemExit(1)


if __name__ == "__main__":
    main()
