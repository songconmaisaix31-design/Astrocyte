"""Acceptance checks for selection and authorized-root accounting, using fixtures."""
import importlib.util
from pathlib import Path
import tempfile
import unittest

MODULE = Path(__file__).resolve().parents[3] / "scripts/validation/diagnose_native.py"
SPEC = importlib.util.spec_from_file_location("diagnose_native", MODULE)
native = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(native)


class SamplingTests(unittest.TestCase):
    def test_kind_blind_deterministic_time_and_session_coverage(self):
        rows = [(i, i // 10, i % 10 + 1, i // 10, "fixture", True) for i in range(1000)]
        selected = native.select(rows)
        self.assertEqual(selected, native.select(list(reversed(rows))))
        self.assertEqual(len(selected), 50)
        self.assertEqual(len({r[:3] for r in selected}), 50)
        self.assertEqual(len({r[3] for r in selected}), 50)
        self.assertEqual(selected, sorted(selected))

    def test_half_open_calendar_boundary(self):
        self.assertEqual(native.timestamp({"timestamp": "2026-09-12T00:00:00+08:00"}), native.START)
        self.assertEqual(native.timestamp({"timestamp": "2026-10-12T00:00:00+08:00"}), native.END)
        self.assertIsNone(native.timestamp({"timestamp": "unknown"}))

    def test_layout_parse_and_header_inventory(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            dated = root / "2026/10/01/session.jsonl"
            dated.parent.mkdir(parents=True)
            dated.write_text('{"type":"session_meta","timestamp":"2026-10-01T00:00:00Z","payload":{"id":"fixture","cwd":"fixture"}}\ninvalid\n{"type":"event_msg","timestamp":"2026-10-12T00:00:00+08:00"}\n', encoding="utf-8")
            (root / "legacy.jsonl").write_text('{"type":"response_item","timestamp":"2026-10-01T00:00:00Z"}\n', encoding="utf-8")
            counts, layouts, _, _, _, rows = native.inventory(root, "codex")
            self.assertEqual(layouts["dated_files"], 1)
            self.assertEqual(layouts["legacy_other_files"], 1)
            self.assertEqual(counts["parse_error_all"], 1)
            self.assertEqual(counts["outside_window_all"], 1)
            self.assertEqual(counts["header_missing_window"], 1)
            self.assertEqual(len(rows), 2)

    def test_subagent_layout_and_small_source_not_duplicated(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            path = root / "project/session/subagents/agent.jsonl"
            path.parent.mkdir(parents=True)
            path.write_text('{"type":"user","timestamp":"2026-10-01T00:00:00Z","sessionId":"fixture","cwd":"fixture"}\n', encoding="utf-8")
            counts, layouts, _, _, _, rows = native.inventory(root, "claude_code")
            self.assertEqual(layouts["subagent_files"], 1)
            self.assertEqual(counts["window_records"], 1)
            self.assertEqual(len(native.select(rows)), 1)


if __name__ == "__main__":
    unittest.main()
