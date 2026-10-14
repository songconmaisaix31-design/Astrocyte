import contextlib
import importlib.util
import json
import pathlib
import sqlite3
import tempfile
import threading
import time
import unittest

spec = importlib.util.spec_from_file_location("snapshot", pathlib.Path(__file__).with_name("snapshot.py"))
snapshot = importlib.util.module_from_spec(spec)
spec.loader.exec_module(snapshot)


class SnapshotTests(unittest.TestCase):
    def fixture(self, root):
        source = root / "source.sqlite"
        with contextlib.closing(sqlite3.connect(source)) as db:
            db.executescript("PRAGMA journal_mode=WAL; CREATE TABLE attention_engine_events(id INTEGER PRIMARY KEY, payload TEXT); CREATE TABLE attention_engine_nodes(topic TEXT, updated_at INTEGER, attention REAL); INSERT INTO attention_engine_nodes VALUES ('topic',1000,.7);")
            db.executemany("INSERT INTO attention_engine_events(payload) VALUES (?)", [("x" * 4096,)] * 4096)
            db.commit()
        history = root / "safe.json"
        history.write_text(json.dumps({"topic": [{"ts": 500, "attention": .5}]}), encoding="utf-8")
        return source, history

    def test_concurrent_wal_writer_coherent_backup(self):
        with tempfile.TemporaryDirectory() as temp:
            root = pathlib.Path(temp)
            source, history = self.fixture(root)
            stop = threading.Event()
            writes, failures = [], []

            def writer():
                try:
                    with contextlib.closing(sqlite3.connect(source, timeout=.1)) as db:
                        while not stop.is_set():
                            db.execute("INSERT INTO attention_engine_events(payload) VALUES ('writer')")
                            db.commit()
                            writes.append(time.monotonic())
                            time.sleep(.001)
                except Exception:
                    failures.append(True)

            t = threading.Thread(target=writer)
            t.start()
            while not writes:
                time.sleep(.001)
            before = len(writes)
            try:
                result = snapshot.publish(root / "runtime", source, history)
            finally:
                stop.set()
                t.join(5)
            self.assertFalse(failures)
            self.assertGreater(len(writes), before)
            file = root / "runtime" / "snapshots" / result["generation"] / "snapshot.sqlite"
            with contextlib.closing(sqlite3.connect(file.as_uri() + "?mode=ro", uri=True)) as db:
                self.assertEqual(db.execute("PRAGMA quick_check").fetchone(), ("ok",))
                self.assertGreaterEqual(db.execute("SELECT count(*) FROM attention_engine_events").fetchone()[0], 4096)

    def test_atomic_failed_refresh_retains_history_and_pointer(self):
        with tempfile.TemporaryDirectory() as temp:
            root = pathlib.Path(temp)
            source, history = self.fixture(root)
            runtime = root / "runtime"
            first = snapshot.publish(runtime, source, history)
            pointer = (runtime / "current.json").read_text()
            history.write_text("invalid PRIVATE_CANARY", encoding="utf-8")
            with self.assertRaises(Exception):
                snapshot.publish(runtime, source, history)
            self.assertEqual((runtime / "current.json").read_text(), pointer)
            history.write_text(json.dumps({"topic": [{"ts": 500, "attention": .5}]}), encoding="utf-8")
            with contextlib.closing(sqlite3.connect(source)) as db:
                db.execute("UPDATE attention_engine_nodes SET updated_at=2000,attention=.8")
                db.commit()
            second = snapshot.publish(runtime, source, history)
            points = json.loads((runtime / "snapshots" / second["generation"] / "history.json").read_text())["topic"]
            self.assertEqual([p["ts"] for p in points], [500, 1000, 2000])
            self.assertEqual([p["attention"] for p in points], [.5, .7, .8])
            self.assertNotEqual(first["generation"], second["generation"])


if __name__ == "__main__":
    unittest.main()
