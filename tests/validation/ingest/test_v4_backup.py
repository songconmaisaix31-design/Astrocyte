import hashlib
from contextlib import closing
import importlib.util
from pathlib import Path
import sqlite3
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts/validation"))
spec = importlib.util.spec_from_file_location("bootstrap_v4", ROOT / "scripts/validation/bootstrap_v4.py")
bootstrap = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bootstrap)


class V4BackupTests(unittest.TestCase):
    def test_day_snapshot_stays_immutable_when_live_source_advances(self):
        with tempfile.TemporaryDirectory() as temporary:
            source = Path(temporary) / "live.sqlite"
            snapshot = Path(temporary) / "day-state.sqlite"
            with closing(sqlite3.connect(source)) as connection:
                connection.execute("CREATE TABLE synthetic(value INTEGER)")
                connection.execute("INSERT INTO synthetic VALUES(7)")
                connection.commit()
            bootstrap.backup(source, snapshot)
            with closing(sqlite3.connect(source)) as connection:
                connection.execute("UPDATE synthetic SET value=8")
                connection.commit()
            with closing(sqlite3.connect(snapshot)) as connection:
                self.assertEqual(connection.execute("SELECT value FROM synthetic").fetchone()[0], 7)

    def test_sqlite_backup_preserves_original_bytes_and_rejects_wrong_checksum(self):
        with tempfile.TemporaryDirectory() as temporary:
            source = Path(temporary) / "source.sqlite"
            destination = Path(temporary) / "destination.sqlite"
            with closing(sqlite3.connect(source)) as connection:
                connection.execute("CREATE TABLE synthetic(value INTEGER)")
                connection.execute("INSERT INTO synthetic VALUES(7)")
                connection.commit()
            original = source.read_bytes()
            expected = hashlib.sha256(original).hexdigest()
            bootstrap.backup(source, destination, expected)
            self.assertEqual(source.read_bytes(), original)
            with closing(sqlite3.connect(destination)) as connection:
                self.assertEqual(connection.execute("SELECT value FROM synthetic").fetchone()[0], 7)
            with self.assertRaises(ValueError):
                bootstrap.backup(source, Path(temporary) / "bad.sqlite", "0" * 64)
            self.assertFalse((Path(temporary) / "bad.sqlite").exists())


if __name__ == "__main__":
    unittest.main()
