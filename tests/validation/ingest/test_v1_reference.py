import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location("verify_v1", Path(__file__).resolve().parents[3] / "scripts/validation/verify_v1.py")
reference = importlib.util.module_from_spec(spec)
spec.loader.exec_module(reference)


class V1IndependentReferenceTests(unittest.TestCase):
    def floor_trace(self):
        start = 1789142400000
        def points(timestamp, phase):
            return [{"topic": topic, "ts": timestamp, "updated_at": timestamp,
                     "attention": 0.05, "momentum": 0.0, "confidence": 0.5,
                     "phase": phase} for topic in sorted(reference.SEEDS)]
        trace = [{"phase": "seed", "ts": start, "points": points(start, "seed")}]
        for i in range(30):
            timestamp = start + (i + 1) * reference.DAY_MS - 1
            date = reference.dt.date(2026, 9, 12) + reference.dt.timedelta(days=i)
            trace.append({"phase": "sleep", "ts": timestamp, "day": str(date),
                          "signal_ids": [], "input_events": 0,
                          "points": points(timestamp, "sleep")})
        return trace

    def test_corrupt_actual_does_not_change_expected(self):
        trace = self.floor_trace()
        result, _ = reference.verify(trace, 0)
        self.assertEqual(result["result"], "PASS")
        trace[15]["points"][0]["attention"] = 0.9
        result, errors = reference.verify(trace, 0)
        self.assertEqual(result["result"], "RED")
        self.assertEqual(result["failures"][0]["operation"], 15)
        self.assertEqual(errors[15 * 6]["expected_attention"], 0.05)

    def test_sleep_reoffering_unknown_or_duplicate_signal_is_rejected(self):
        trace = self.floor_trace()
        trace[1]["signal_ids"] = ["fabricated", "fabricated"]
        with self.assertRaises(ValueError):
            reference.verify(trace)

    def test_equal_ts_each_call_changes_momentum_after_clamp(self):
        state = {"a": 0.05, "m": 0.0, "c": 0.5, "ts": 0}
        reference.advance(state, 0, 0.6)
        self.assertAlmostEqual(state["a"], 0.65)
        self.assertAlmostEqual(state["m"], 0.42)
        reference.advance(state, 0, 0.6)
        self.assertAlmostEqual(state["a"], 1)
        self.assertAlmostEqual(state["m"], 0.371)
        reference.advance(state, 0, 0.6)
        self.assertAlmostEqual(state["m"], 0.1113)

    def test_empty_day_decays_then_floor_and_same_time_noop(self):
        state = {"a": 1.0, "m": 0.0, "c": 0.5, "ts": 0}
        reference.advance(state, 86400000, 0)
        self.assertAlmostEqual(state["a"], 0.9048374180359595)
        self.assertAlmostEqual(state["m"], -0.06661380737482834)
        snapshot = dict(state)
        reference.advance(state, 86400000, 0)
        self.assertEqual(state, snapshot)
        reference.advance(state, 100 * 86400000, 0)
        self.assertEqual(state["a"], 0.05)

    def test_backward_clock_rejected(self):
        with self.assertRaises(ValueError):
            reference.advance({"a": 1, "m": 0, "ts": 2}, 1, 0)


if __name__ == "__main__":
    unittest.main()
