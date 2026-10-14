import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
(ROOT / ".verification").mkdir(exist_ok=True)
spec = importlib.util.spec_from_file_location("render_v1r", ROOT / "scripts/validation/render_v1r.py")
renderer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(renderer)


class V1RArtifactTests(unittest.TestCase):
    def test_standalone_chart_tables_and_no_overwrite(self):
        with tempfile.TemporaryDirectory(dir=ROOT / ".verification") as temporary:
            run = Path(temporary)
            topics = ["Morphogenesis", "Astrocyte/Attention Engine", "McPet"]
            series = {topic: [{"topic": topic, "ts": 1789142400000, "attention": 0.05, "phase": "seed"},
                              {"topic": topic, "ts": 1789228799999, "attention": 0.65, "phase": "sleep"}] for topic in topics}
            fixture = {"topic-timeseries-safe.json": series,
                       "coverage-safe.json": {"assigned": 1, "events": 1, "unassigned": 0, "topic_coverage": 1, "by_topic": {"McPet": 1}, "mcpet_git": 1, "mcpet_native": 0, "by_kind": {"commit": 1}, "by_source_kind": {"git_commit": 1}},
                       "weights-50-safe.json": {"seed": 2026101201, "samples": [{"event": {"event_id": "fixture", "ts": 1789142400000, "kind": "commit", "topic": "McPet"}, "s": 1, "w": 0.6, "contribution": 0.6}]},
                       "mcpet-supplement-safe.json": {"samples": []}}
            for filename, value in fixture.items():
                (run / filename).write_text(json.dumps(value), encoding="utf-8")
            renderer.render(run)
            self.assertGreater((run / "three-topic-attention.png").stat().st_size, 1000)
            self.assertIn("McPet", (run / "three-topic-attention.svg").read_text(encoding="utf-8"))
            table = (run / "weights-human-safe.md").read_text(encoding="utf-8")
            self.assertIn("PENDING", table)
            self.assertNotIn("raw_base64", table)
            with self.assertRaises(FileExistsError):
                renderer.render(run)


if __name__ == "__main__":
    unittest.main()
