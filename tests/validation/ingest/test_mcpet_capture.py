import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import unittest
import tempfile

ROOT = Path(__file__).resolve().parents[3]
(ROOT / ".verification").mkdir(exist_ok=True)
sys.path.insert(0, str(ROOT / "scripts/validation"))
spec = importlib.util.spec_from_file_location("inventory_mcpet", ROOT / "scripts/validation/inventory_mcpet.py")
capture = importlib.util.module_from_spec(spec)
spec.loader.exec_module(capture)


class McPetCaptureTests(unittest.TestCase):
    def test_capture_serializes_exact_integer_millisecond_window(self):
        with tempfile.TemporaryDirectory(dir=ROOT / ".verification") as temporary:
            base = Path(temporary).absolute()
            root, evidence = base / "owned-repository", base / "evidence"
            root.mkdir()
            evidence.mkdir()
            remote = base / "owned-origin.git"
            subprocess.run(["git", "init", "--bare", str(remote)], capture_output=True, check=True)
            capture.git(root, "init")
            capture.git(root, "config", "user.name", "Python validation")
            capture.git(root, "config", "user.email", "validation@example.invalid")
            capture.git(root, "remote", "add", "origin", str(remote))
            (root / "owned-input.txt").write_text("self-owned actual Git input", encoding="utf-8")
            capture.git(root, "add", "owned-input.txt")
            environment = dict(os.environ, GIT_AUTHOR_DATE="2026-10-01T12:00:00+08:00", GIT_COMMITTER_DATE="2026-10-01T12:00:00+08:00")
            subprocess.run(["git", "-C", str(root), "commit", "-m", "Record self-owned input"], env=environment, capture_output=True, check=True)
            capture.capture_repository(root, str(remote), "self-owned repository validation", evidence / "new-source")
            metadata = json.loads((evidence / "new-source/source-local.json").read_text())
            self.assertIs(type(metadata["window_start"]), int)
            self.assertIs(type(metadata["window_end_exclusive"]), int)
            self.assertEqual(metadata["window_start"], capture.START)
            self.assertEqual(metadata["window_end_exclusive"], capture.END)
            self.assertEqual(metadata["head"], capture.git(root, "rev-parse", "HEAD").decode().strip())
            self.assertEqual(metadata["unique_commits"], 1)
            self.assertEqual(metadata["remote"], str(remote))
            captured = json.loads((evidence / "new-source/git-log-local.jsonl").read_text())
            self.assertEqual(captured["record"]["commit"], metadata["head"])
            self.assertEqual(captured["record"]["subject"], "Record self-owned input")
            with self.assertRaises(FileExistsError):
                capture.capture_repository(root, str(remote), "self-owned repository validation", evidence / "new-source")

    def test_committer_time_window_exact_and_duplicate_excluded(self):
        start, end = int(capture.START), int(capture.END)
        lines = []
        for i, timestamp in [(1, start - 1000), (2, start), (3, end - 1000), (4, end), (2, start)]:
            lines.append(f"{i:040x}\0{timestamp // 1000}\0pure function input".encode())
        rows = capture.selected_records("pure-function-input", b"\n".join(lines))
        self.assertEqual([row["record"]["ts"] for row in rows], [start, end - 1000])
        self.assertEqual(len({row["record"]["commit"] for row in rows}), 2)

    def test_source_or_output_outside_bounds_rejected_before_git(self):
        with self.assertRaises(ValueError):
            capture.capture(ROOT / ".verification/rejected-source", "unapproved", "", ROOT / ".verification/rejected-output")

    def test_invalid_utf8_subject_fails_without_replacement(self):
        line = b"a" * 40 + b"\0" + str(int(capture.START) // 1000).encode() + b"\0\xff"
        with self.assertRaises(UnicodeDecodeError):
            capture.selected_records("pure-function-input", line)


if __name__ == "__main__":
    unittest.main()
