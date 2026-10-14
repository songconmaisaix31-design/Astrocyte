# 只读历史回放；派生证据通过审计后导出，原始资料留在本地。
from __future__ import annotations

import argparse
import base64
from collections import Counter
from contextlib import closing
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import time


TOPICS = ("Morphogenesis", "Astrocyte/Attention Engine", "McPet")
ZERO_TOPICS = ("Agent Swarm", "Agent Memory", "MCP")
EVENT_FIELDS = {"event_id", "kind", "topic", "tokens", "scope", "ts"}
EXPLANATION_FIELDS = {"topic", "attention", "momentum", "evidence", "related",
                      "related_projects", "updated_at"}
SIGNAL_FIELDS = {"event", "event_id", "kind", "class", "weight", "ts", "strength",
                 "elapsed_days", "decay", "decay_factor", "attention_contribution",
                 "confidence_contribution", "confidence_only"}
FILENAMES = dict(zip(TOPICS, ("morphogenesis.md", "astrocyte-attention-engine.md", "mcpet.md")))
REPLAY_SOURCE = "7d57e06eaeaa1ddc8ddf132af47840acf8f777cc"
ROOT_BASE = "fbad72cf5257aebabe171af67098d41c3739c6d6"
REPLAY_DATABASE_SHA256 = "279ffd78ad8cf083b1e279fb8c4ed32ae53a662b52d8a0f220f96d98c9816d4e"


class AuditError(Exception):
    """Deliberately fixed public errors: never interpolate source contents."""


def require(condition, message):
    if not condition:
        raise AuditError(message)


def readonly(path):
    connection = sqlite3.connect(path.resolve().as_uri() + "?mode=ro", uri=True)
    connection.execute("PRAGMA query_only=ON")
    connection.row_factory = sqlite3.Row
    return connection


def database_snapshot(path):
    """Compare all persisted tables, including attention, boosts and history."""
    result = {}
    with closing(readonly(path)) as connection:
        connection.execute("BEGIN")
        for name, sql in connection.execute("SELECT name,sql FROM sqlite_master WHERE type='table' ORDER BY name"):
            quoted = '"' + name.replace('"', '""') + '"'
            digest = hashlib.sha256(sql.encode())
            count = 0
            for row in connection.execute("SELECT * FROM " + quoted + " ORDER BY rowid"):
                encoded = json.dumps([{"blob": base64.b64encode(v).decode()} if isinstance(v, bytes) else v
                                      for v in row], ensure_ascii=False, separators=(",", ":")).encode()
                digest.update(len(encoded).to_bytes(8, "big"))
                digest.update(encoded)
                count += 1
            result[name] = {"rows": count, "sha256": digest.hexdigest()}
    return result


def physical_snapshot(path):
    result = {}
    for suffix in ("", "-wal"):
        candidate = Path(str(path) + suffix)
        if candidate.exists():
            result[candidate.name] = hashlib.sha256(candidate.read_bytes()).hexdigest()
    return result


def json_strings(value, keys=False, nested=False):
    if isinstance(value, str):
        yield value
        if nested:
            try:
                decoded = json.loads(value)
            except json.JSONDecodeError:
                return
            if not isinstance(decoded, str) or decoded != value:
                yield from json_strings(decoded, keys, nested)
    elif isinstance(value, dict):
        for key, child in value.items():
            if keys:
                yield key
            yield from json_strings(child, keys, nested)
    elif isinstance(value, list):
        for child in value:
            yield from json_strings(child, keys, nested)


def derived_event(row):
    data = json.loads(row["data"])
    return {"event_id": row["id"], **{key: data.get(key) for key in EVENT_FIELDS - {"event_id"}}}


def raw_unique(raw, derived):
    # 使用 v2.1 Go 审计的 UTF-8 字节阈值，逐事件扣除六个派生标量。
    allowed = {str(value) for value in derived.values() if value is not None}
    return {value for value in json_strings(json.loads(raw))
            if len(value.encode("utf-8")) >= 8 and value not in allowed}


def audit_raw(rows, responses, markdowns):
    targets = []
    for response in responses:
        targets.append(json.dumps(response, ensure_ascii=False, separators=(",", ":")))
        targets.extend(json_strings(response, keys=True, nested=True))
    targets.extend(markdowns)
    # 对完整传输字节、解码键和值执行字面子串检查，合并预筛保留逐目标规则。
    # 冲突记录只包含数字行号。
    all_values = set()
    distinct_raw = set()
    per_event = 0
    zero = 0
    owners = {}
    for index, row in enumerate(rows):
        raw = row["raw_bytes"]
        require(isinstance(raw, bytes) and len(raw) > 0, "Original BLOB unavailable")
        values = raw_unique(raw, derived_event(row))
        distinct_raw.update(values)
        per_event += len(values)
        zero += not values
        values.add(base64.b64encode(raw).decode())
        for value in values:
            owners.setdefault(value, index)
        all_values.update(values)
    require(per_event > 0, "No measurable original strings")
    joined = "\0".join(targets)
    collisions = []
    for value in all_values:
        if value in joined:
            for index, target in enumerate(targets):
                if value in target:
                    collisions.append({"source_row_index": owners[value], "target_index": index})
    return {"events": len(rows), "raw_unique_strings_per_event": per_event,
            "raw_unique_distinct_strings": len(distinct_raw),
            "raw_unique_distinct_and_full_blob_base64": len(all_values),
            "zero_raw_unique_events": zero, "responses": len(responses),
            "markdowns": len(markdowns), "violations": len(collisions), "collision_indexes": collisions}


def tool_result(response):
    require("error" not in response, "Unexpected MCP RPC error")
    result = response.get("result", {})
    require(result.get("isError") is False, "Unexpected MCP tool failure")
    content = result.get("content", [])
    require(len(content) == 1 and content[0].get("type") == "text", "Unexpected MCP content")
    decoded = json.loads(content[0]["text"])
    if "structuredContent" in result:
        require(result["structuredContent"] == decoded, "MCP structured/text disagreement")
    return decoded


def validate_explanation(explanation, topic, nodes, rows):
    require(set(explanation) == EXPLANATION_FIELDS, "Explanation projection changed")
    require(explanation["topic"] == topic, "MCP topic mismatch")
    node = nodes[topic]
    for key in ("attention", "momentum", "updated_at"):
        require(explanation[key] == node[key], "MCP node value mismatch")
    data = json.loads(node["data"])
    applied = data["history"]["applied"] or {}
    expected = [(used["signal"]["event_id"], used["signal"]["class"]) for used in applied.values()]
    used_ids = {event_id for event_id, _ in expected}
    expected.extend((ref["event_id"], "") for ref in data["node"]["evidence"] or []
                    if ref["event_id"] not in used_ids)
    evidence = explanation["evidence"]
    require(Counter((item["event_id"], item["class"]) for item in evidence) == Counter(expected),
            "Evidence history incomplete or duplicated")
    require(evidence == sorted(evidence, key=lambda item: (item["ts"], item["event_id"], item["class"])),
            "Evidence order mismatch")
    for item in evidence:
        require(set(item) == SIGNAL_FIELDS and set(item["event"]) == EVENT_FIELDS, "Evidence projection changed")
        row = rows.get(item["event_id"])
        require(row is not None, "Evidence original receipt missing")
        require(item["event"] == derived_event(row), "Derived event traceback mismatch")
        require(item["ts"] == row["ts"] and item["kind"] == row["kind"] and row["topic"] == topic,
                "Signal traceback mismatch")
        require(isinstance(row["raw_bytes"], bytes) and len(row["raw_bytes"]) > 0, "Original receipt BLOB missing")
        matches = [used for used in applied.values() if used["signal"]["event_id"] == item["event_id"]
                   and used["signal"]["class"] == item["class"]]
        if matches:
            require(len(matches) == 1, "Ambiguous signal history")
            used = matches[0]
            for key in ("weight", "strength", "ts", "class"):
                require(item[key] == used["signal"][key], "Signal weight traceback mismatch")
            for key in ("elapsed_days", "decay", "decay_factor", "attention_contribution", "confidence_contribution"):
                require(item[key] == used[key], "Stored contribution traceback mismatch")
        else:
            refs = [ref for ref in data["node"]["evidence"] if ref["event_id"] == item["event_id"]]
            require(len(refs) == 1 and item["weight"] == refs[0]["weight_at_use"], "Clustering weight mismatch")
    expected_related = data["node"]["related_topics"] or {}
    require({item["topic"]: item["strength"] for item in explanation["related"]} == expected_related,
            "Related node completeness mismatch")
    for related in explanation["related"]:
        require(set(related) == {"topic", "strength", "ts"}, "Related projection changed")
        require(related["ts"] == (data["history"].get("related_at") or {}).get(related["topic"], 0),
                "Related timestamp mismatch")
    require(explanation["related_projects"] == data["node"]["related_projects"], "Related projects mismatch")


def cell(value):
    return str(value).replace("|", "\\|").replace("\n", " ").replace("\r", " ")


def timestamp(ts):
    return datetime.fromtimestamp(ts / 1000, timezone.utc).isoformat(timespec="milliseconds")


def render_markdown(explanation, confidence, source_sha, replay_path, cutoff):
    evidence = explanation["evidence"]
    kinds = Counter(item["kind"] for item in evidence)
    lines = [f'# {explanation["topic"]} — V2 真实库 MCP 证据', "",
             f'导出源码 SHA：`{source_sha}`；精确 root 基线：`{ROOT_BASE}`。',
             f'回放源码：`{REPLAY_SOURCE}`；历史回放库：`{replay_path.as_posix()}`。',
             f'截止 ts：`{cutoff}`（{timestamp(cutoff)}）；节点 updated_at：`{explanation["updated_at"]}`。', "",
             "数据来自该库备份副本的生产 `cmd/mcp` stdio 调用；没有新回放、Sleep、信号或时间推进。",
             "`get_attention_state` 与 `explain_focus` 的 A/M 已互相核对；接口均无 confidence 字段。", "",
             "| 指标 | 值 | 来源 |", "| --- | ---: | --- |",
             f'| attention | {explanation["attention"]!r} | 生产 MCP |',
             f'| momentum | {explanation["momentum"]!r} | 生产 MCP |',
             f'| confidence | {confidence if confidence is not None else "未导出（等待来源裁决）"} | {"原库 query_only 只读节点值；不是 MCP 返回" if confidence is not None else "共享接口缺项"} |',
             f'| 全部证据 | {len(evidence)} | 生产 explain_focus |',
             f'| 关联节点 | {len(explanation["related"])} | 生产 explain_focus |', "",
             "## 分 kind 计数", "", "| kind | 数量 |", "| --- | ---: |"]
    lines.extend(f"| {cell(kind)} | {count} |" for kind, count in sorted(kinds.items()))
    if not kinds:
        lines.append("| 无 | 0 |")
    lines.extend(["", "## 全部证据（ts/event_id 排序）", "",
                  "w 为生产 explanation 的 weight；未重算、未补信号。event_id 在原库对应非空原始 BLOB，完整本地回溯索引仅 LOCAL。",
                  "", "| ts (毫秒) | UTC | kind | w | event_id |", "| ---: | --- | --- | ---: | --- |"])
    lines.extend(f'| {item["ts"]} | {timestamp(item["ts"])} | {cell(item["kind"])} | {item["weight"]!r} | {cell(item["event_id"])} |'
                 for item in evidence)
    lines.extend(["", "## 关联节点与强度", ""])
    if explanation["related"]:
        lines.extend(["| 节点 | 强度 | ts |", "| --- | ---: | ---: |"])
        lines.extend(f'| {cell(item["topic"])} | {item["strength"]!r} | {item["ts"]} |' for item in explanation["related"])
    else:
        lines.append("关联节点：0；不推断或补关联。")
    lines.extend(["", "关联项目（接口已有 derived 字段）：" + ", ".join(cell(p) for p in explanation["related_projects"] or []) + "。", "",
                  "Agent Swarm / Agent Memory / MCP 的生产证据与信号均为 0；保持原零信号，不借用本主题证据。",
                  "本资产是历史真实库证据链；没有执行 V3/V4、协议锁定或未来七日实验。", ""])
    return "\n".join(lines)


def requests(raw_probe):
    def call(identifier, name, arguments):
        return {"jsonrpc": "2.0", "id": identifier, "method": "tools/call",
                "params": {"name": "astrocyte." + name, "arguments": arguments}}
    result = [{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
        "protocolVersion": "2025-11-25", "capabilities": {}, "clientInfo": {"name": "v2-evidence", "version": "1"}}},
        call(2, "get_attention_state", {})]
    for index, topic in enumerate(TOPICS + ZERO_TOPICS, 3):
        result.append(call(index, "explain_focus", {"topic": topic}))
    for index, name in enumerate(("get_attention_state", "explain_focus", "get_experiences", "report_event", "veto"), 20):
        result.append(call(index, name, {raw_probe: raw_probe}))
    result.extend([call(30, "explain_focus", {"topic": raw_probe}),
                   call(31, "get_experiences", {"pattern": raw_probe, "scope": raw_probe}),
                   call(32, "veto", {"task": raw_probe, "reason": raw_probe}),
                   call(33, "report_event", {"event": {"kind": raw_probe, "ts": 0}}),
                   {"jsonrpc": "2.0", "id": 34, "method": raw_probe}])
    return result


def export(args):
    repo = Path(__file__).resolve().parents[2]
    original = args.database.resolve(strict=True)
    local_root = args.local_root.resolve()
    worktrees = subprocess.check_output(["git", "worktree", "list", "--porcelain"], cwd=repo, text=True)
    roots = [Path(line.removeprefix("worktree ")).resolve() for line in worktrees.splitlines()
             if line.startswith("worktree ")]
    require(all(not local_root.is_relative_to(root) for root in roots), "Local evidence directory must be outside all worktrees")
    local_root.mkdir(parents=True, exist_ok=True)
    run = Path(tempfile.mkdtemp(prefix="run-", dir=local_root))
    started = time.monotonic()
    summary = {"local_run": run.as_posix(), "status": "RED", "started_at": datetime.now(timezone.utc).isoformat()}
    original_before = database_snapshot(original)
    physical_before = physical_snapshot(original)
    try:
        require(physical_before[original.name] == REPLAY_DATABASE_SHA256, "Frozen V1R database identity mismatch")
        source_sha = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip()
        runtime = run / "runtime"
        runtime.mkdir()
        copy = runtime / "state.sqlite"
        with closing(readonly(original)) as source, closing(sqlite3.connect(copy)) as destination:
            source.backup(destination)
        require(database_snapshot(copy) == original_before, "Production copy differs from original")
        with closing(readonly(original)) as connection:
            rows = [dict(row) for row in connection.execute("SELECT * FROM attention_engine_events ORDER BY ts,id")]
            nodes = {row["topic"]: dict(row) for row in connection.execute("SELECT * FROM attention_engine_nodes")}
            cutoff = connection.execute("SELECT max(completed_at) FROM attention_engine_sleep_runs").fetchone()[0]
            zero_signals = {topic: connection.execute("SELECT count(*) FROM attention_engine_signals WHERE topic=?", (topic,)).fetchone()[0]
                            for topic in ZERO_TOPICS}
        lookup = {row["id"]: row for row in rows}
        with (run / "traceback-RAW-LOCAL.jsonl").open("w", encoding="utf-8") as out:
            for row in rows:
                out.write(json.dumps({**row, "raw_bytes": base64.b64encode(row["raw_bytes"]).decode()}, ensure_ascii=False) + "\n")
        binary = run / "astrocyte-mcp.exe"
        with (run / "build-first.log").open("wb") as log:
            built = subprocess.run(["go", "build", "-mod=readonly", "-o", str(binary), "./cmd/mcp"], cwd=repo, stdout=log, stderr=log)
        require(built.returncode == 0, "Production MCP build failed; local log preserved")
        probe = next(value for row in rows for value in sorted(raw_unique(row["raw_bytes"], derived_event(row)))
                     if value not in TOPICS + ZERO_TOPICS)
        commands = requests(probe)
        payload = b"".join(json.dumps(request, ensure_ascii=False).encode() + b"\n" for request in commands)
        (run / "requests-RAW-LOCAL.jsonl").write_bytes(payload)
        invocation = [str(binary), "--data-dir", str(runtime), "--project", "Astrocyte", "--session", "v2-w0-read",
                      "--principal-id", "v2-w0", "--principal-kind", "agent", "--signal-strength", "1",
                      "--decay", "0.1", "--initial-confidence", "0.5"]
        with (run / "stderr-LOCAL.log").open("wb") as log:
            completed = subprocess.run(invocation, input=payload, stdout=subprocess.PIPE, stderr=log, cwd=repo, timeout=180)
        (run / "responses-LOCAL.jsonl").write_bytes(completed.stdout)
        require(completed.returncode == 0, "Production MCP exited unsuccessfully")
        responses = [json.loads(line) for line in completed.stdout.splitlines()]
        require([response["id"] for response in responses] == [request["id"] for request in commands], "MCP response count/order mismatch")
        by_id = {response["id"]: response for response in responses}
        state = tool_result(by_id[2])
        require(set(state) == {"topics"}, "Attention state projection changed")
        require(all(set(row) == {"topic", "attention", "momentum"} for row in state["topics"]), "Focus item projection changed")
        focus = {row["topic"]: row for row in state["topics"]}
        require(len(focus) == len(state["topics"]) and set(focus) == {topic for topic, node in nodes.items()
                                                                   if node["lifecycle"] != "cold"},
                "Attention state incomplete or duplicated")
        staged = run / "markdown"
        staged.mkdir()
        markdowns = {}
        topic_counts = {}
        for index, topic in enumerate(TOPICS + ZERO_TOPICS, 3):
            explanation = tool_result(by_id[index])
            validate_explanation(explanation, topic, nodes, lookup)
            require(all(focus[topic][key] == explanation[key] for key in ("attention", "momentum")), "MCP state/explanation disagreement")
            topic_counts[topic] = {"evidence": len(explanation["evidence"]), "related": len(explanation["related"]),
                                   "kinds": dict(Counter(item["kind"] for item in explanation["evidence"]))}
            if topic in ZERO_TOPICS:
                require(not explanation["evidence"] and zero_signals[topic] == 0, "Zero-signal baseline changed")
            else:
                confidence = nodes[topic]["confidence"] if args.confidence_source == "sqlite-readonly" else None
                markdowns[topic] = render_markdown(explanation, confidence, source_sha, original, cutoff)
                (staged / FILENAMES[topic]).write_text(markdowns[topic], encoding="utf-8")
        error_responses = [response for response in responses if response["id"] >= 20]
        require(all("error" in response or response.get("result", {}).get("isError") is True for response in error_responses),
                "Error probe unexpectedly succeeded")
        audit = audit_raw(rows, responses, list(markdowns.values()))
        (run / "raw-audit-safe.json").write_text(json.dumps(audit, indent=2), encoding="utf-8")
        require(audit["violations"] == 0, "Strict raw audit RED; collision indexes retained locally")
        require(database_snapshot(copy) == original_before, "MCP startup/read mutated production copy")
        summary.update(source_sha=source_sha, replay_source=REPLAY_SOURCE, cutoff_ts=cutoff,
                       original_database_sha256=physical_before[original.name],
                       topic_counts=topic_counts, audit=audit, successful_tool_calls=7,
                       error_responses=len(error_responses), confidence_source=args.confidence_source,
                       production_copy_unchanged=True)
        summary["status"] = "PASS"
    except Exception as error:
        summary["error_type"] = type(error).__name__
        # 外部库异常可能包含原始资料，只有固定 AuditError 文本可以公开。
        summary["error"] = str(error) if isinstance(error, AuditError) else "Failure details suppressed; local artifacts retained"
    finally:
        summary["original_logical_unchanged"] = database_snapshot(original) == original_before
        summary["original_files_unchanged"] = physical_snapshot(original) == physical_before
        if not summary["original_logical_unchanged"] or not summary["original_files_unchanged"]:
            summary["status"] = "RED"
            summary["error"] = "Original database changed"
        summary["ended_at"] = datetime.now(timezone.utc).isoformat()
        summary["elapsed_seconds"] = round(time.monotonic() - started, 3)
        (run / "summary-safe.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    if summary["status"] == "PASS" and args.publish:
        require(args.confidence_source == "sqlite-readonly", "Confidence source unresolved; publish withheld")
        output = repo / "docs" / "evidence"
        output.mkdir(parents=True, exist_ok=True)
        for topic, markdown in markdowns.items():
            (output / FILENAMES[topic]).write_text(markdown, encoding="utf-8")
    print(json.dumps(summary, ensure_ascii=False))
    return 0 if summary["status"] == "PASS" else 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--database", type=Path, required=True)
    parser.add_argument("--local-root", type=Path, required=True)
    parser.add_argument("--confidence-source", choices=("unavailable", "sqlite-readonly"), default="unavailable")
    parser.add_argument("--publish", action="store_true")
    arguments = parser.parse_args()
    raise SystemExit(export(arguments))
