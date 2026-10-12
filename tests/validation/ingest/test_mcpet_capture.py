import importlib.util
from pathlib import Path
import sys
import unittest

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts/validation"))
spec = importlib.util.spec_from_file_location("inventory_mcpet", ROOT / "scripts/validation/inventory_mcpet.py")
capture = importlib.util.module_from_spec(spec)
spec.loader.exec_module(capture)


class McPetCaptureTests(unittest.TestCase):
    def test_committer_time_window_exact_and_duplicate_excluded(self):
        start, end = int(capture.START), int(capture.END)
        lines = []
        for i, timestamp in [(1, start - 1000), (2, start), (3, end - 1000), (4, end), (2, start)]:
            lines.append(f"{i:040x}\0{timestamp // 1000}\0synthetic subject".encode())
        rows = capture.selected_records("synthetic-root", b"\n".join(lines))
        self.assertEqual([row["record"]["ts"] for row in rows], [start, end - 1000])
        self.assertEqual(len({row["record"]["commit"] for row in rows}), 2)

    def test_source_or_output_outside_bounds_rejected_before_git(self):
        with self.assertRaises(ValueError):
            capture.capture("C:/personal", "https://github.com/fixture/private", "", "C:/personal/output")


if __name__ == "__main__":
    unittest.main()
