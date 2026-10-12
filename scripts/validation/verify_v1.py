"""Independent V1 reference: no production formulas, imports or output-as-expected.

Only event kind, topic, timestamp, call order and literal frozen parameters
drive expected A/M/C. Actual points and reported s/w are comparison targets.
Every persisted signal is seen once; daily re-offering must be a duplicate.
"""

import argparse
import datetime as dt
import json
import math
from pathlib import Path

DAY_MS = 86400000
WEIGHTS = {"commit": 0.6, "edit_file": 0.6, "search": 0.6}
SEEDS = {"Morphogenesis", "Astrocyte/Attention Engine", "Agent Swarm", "MCP", "Agent Memory", "McPet"}


def advance(state, timestamp, contribution):
    if timestamp < state["ts"]:
        raise ValueError("nonchronological reference input")
    elapsed = (timestamp - state["ts"]) / DAY_MS
    if elapsed == 0 and contribution == 0:
        return
    previous = state["a"]
    state["a"] = max(0.05, min(1.0, previous * math.exp(-0.1 * elapsed) + contribution))
    state["m"] = 0.3 * state["m"] + 0.7 * (state["a"] - previous)
    state["ts"] = timestamp


def verify(operations, expected_population=None):
    states, seen, stable_seen, daily_ids = {}, set(), set(), set()
    failures, errors = [], []
    previous_key = None
    started = None
    sleep_index = 0
    event_count = assigned_count = unassigned_count = compared = empty_days = equal_ts = 0
    previous_event_ts = None
    max_a = max_m = max_c = 0.0
    per_topic = {}
    for index, op in enumerate(operations):
        phase, timestamp = op["phase"], op["ts"]
        expected_topics = set()
        if phase == "seed":
            if index != 0 or states:
                raise ValueError("seed must be the first operation")
            started = timestamp
            fixed_start = int(dt.datetime(2026, 9, 12, tzinfo=dt.timezone(dt.timedelta(hours=8))).timestamp() * 1000)
            if started != fixed_start:
                raise ValueError("seed clock differs from frozen window start")
            # Expected state uses declared constants, never reported seed A/M/C.
            states = {topic: {"a": 0.05, "m": 0.0, "c": 0.5, "ts": timestamp} for topic in SEEDS}
            expected_topics = SEEDS
        elif phase == "event":
            event = op["event"]
            if set(event) != {"event_id", "kind", "topic", "tokens", "scope", "ts"}:
                raise ValueError("event exposure is not exactly six fields")
            if timestamp != event["ts"] or not started <= timestamp < started + 30 * DAY_MS:
                raise ValueError("event time/window mismatch")
            if not started + sleep_index * DAY_MS <= timestamp < started + (sleep_index + 1) * DAY_MS:
                raise ValueError("event fed before/after its simulated day")
            key = (timestamp, op["replay_id"])
            if previous_key is not None and key <= previous_key:
                raise ValueError("equal-ts stable identity ordering violated")
            previous_key = key
            if event["event_id"] in seen or op["replay_id"] in stable_seen:
                raise ValueError("duplicate event identity")
            seen.add(event["event_id"])
            stable_seen.add(op["replay_id"])
            daily_ids.add(event["event_id"])
            if previous_event_ts == timestamp:
                equal_ts += 1
            previous_event_ts = timestamp
            event_count += 1
            # Frozen V0 contains these three kinds only. Do not infer unknown
            # paper depth/outcome or supply an undecided confidence increment.
            weight = WEIGHTS[event["kind"]]
            if op["strength"] != 1 or op["weight"] != weight:
                failures.append({"operation": index, "reason": "s/w binding"})
            topic = event["topic"]
            if topic:
                advance(states[topic], timestamp, weight)
                assigned_count += 1
                expected_topics = {topic}
            else:
                unassigned_count += 1
        elif phase == "sleep":
            if timestamp != started + (sleep_index + 1) * DAY_MS - 1:
                raise ValueError("Sleep clock is not this simulated day end")
            date = dt.datetime.fromtimestamp((started + sleep_index * DAY_MS) / 1000, dt.timezone(dt.timedelta(hours=8))).date().isoformat()
            if op["day"] != date:
                raise ValueError("Sleep calendar day mismatch")
            offered = op.get("signal_ids", [])
            if len(offered) != len(set(offered)) or set(offered) != daily_ids or op.get("input_events", 0) != len(daily_ids):
                raise ValueError("daily signal receipts disagree with event stream")
            # All daily offered signals were applied on their event calls or
            # persist unassigned. No signal contributes again during Sleep.
            if not daily_ids:
                empty_days += 1
            for state in states.values():
                advance(state, timestamp, 0)
            expected_topics = SEEDS
            daily_ids = set()
            sleep_index += 1
        else:
            raise ValueError("unexpected operation")
        if {p["topic"] for p in op["points"]} != expected_topics or len(op["points"]) != len(expected_topics):
            raise ValueError("missing/duplicate state comparison points")
        for p in op["points"]:
            state = states[p["topic"]]
            da, dm, dc = abs(p["attention"] - state["a"]), abs(p["momentum"] - state["m"]), abs(p["confidence"] - state["c"])
            max_a, max_m, max_c = max(max_a, da), max(max_m, dm), max(max_c, dc)
            per_topic[p["topic"]] = per_topic.get(p["topic"], 0) + 1
            errors.append({"operation": index, "topic": p["topic"], "phase": phase, "ts": timestamp, "attention_error": da, "momentum_error": dm, "confidence_error": dc, "expected_attention": state["a"], "expected_momentum": state["m"], "expected_confidence": state["c"]})
            compared += 1
            if max(da, dm, dc) > 1e-12 or p["updated_at"] != state["ts"] or p["ts"] != timestamp:
                failures.append({"operation": index, "topic": p["topic"], "reason": "independent state mismatch", "a_error": da, "m_error": dm, "c_error": dc})
    if sleep_index != 30 or daily_ids:
        raise ValueError("replay must contain exactly 30 completed days")
    if expected_population is not None and event_count != expected_population:
        raise ValueError("population count changed")
    result = {"result": "PASS" if not failures else "RED", "reference": "literal independent Python exp/clamp/momentum recurrence from ordered event kind/time and daily duplicate signal offers", "events": event_count, "assigned_events": assigned_count, "unassigned_events": unassigned_count, "sleep_calls": sleep_index, "empty_days": empty_days, "equal_ts_adjacent_pairs": equal_ts, "compared_points": compared, "per_topic_points": per_topic, "max_attention_error": max_a, "max_momentum_error": max_m, "max_confidence_error": max_c, "failures": failures, "human_focus_confirmation": "PENDING", "human_50_weight_review": "PENDING"}
    return result, errors


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--trace", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--errors", required=True, type=Path)
    parser.add_argument("--expected-population", type=int)
    args = parser.parse_args()
    with args.trace.open(encoding="utf-8") as source:
        result, errors = verify((json.loads(line) for line in source), args.expected_population)
    with args.errors.open("x", encoding="utf-8") as target:
        json.dump(errors, target, indent=2)
    with args.output.open("x", encoding="utf-8") as target:
        json.dump(result, target, indent=2)
    # No raw or event command is printed, including on mismatch.
    print(json.dumps({k: result[k] for k in ("result", "events", "sleep_calls", "compared_points", "max_attention_error", "max_momentum_error")}))
    return 0 if result["result"] == "PASS" else 1


if __name__ == "__main__":
    raise SystemExit(main())
