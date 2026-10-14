# 盘点并冻结不按 kind 筛选的样本，输出只包含安全统计。
import argparse
import base64
from collections import Counter
from datetime import datetime
import json
import os
from pathlib import Path
import sys

START = datetime.fromisoformat("2026-09-12T00:00:00+08:00").timestamp() * 1000
END = datetime.fromisoformat("2026-10-12T00:00:00+08:00").timestamp() * 1000
POLICY = {
    "window": "[2026-09-12T00:00:00+08:00,2026-10-12T00:00:00+08:00)",
    "per_client": 50,
    "selection": "Sort timestamped JSON objects by (ts, relative-file-order, line); partition into 50 equal-count chronological strata; in each choose least-used session, then closest to stratum midpoint, then file and line. No kind or payload filtering; headers participate.",
    "missing_timestamp": "Count in inventory, excluded only because calendar membership is unknown.",
    "raw": "Local frozen sample-first.jsonl only; never commit or print.",
}


def timestamp(record):
    value = record.get("timestamp")
    if value is None and record.get("type") == "session_meta":
        value = record.get("payload", {}).get("timestamp")
    if isinstance(value, (int, float)):
        return int(value)
    if isinstance(value, str):
        try:
            return int(datetime.fromisoformat(value.replace("Z", "+00:00")).timestamp() * 1000)
        except ValueError:
            raise ValueError("Native timestamp is invalid") from None
    return None


def unredirected(path):
    for part in [path, *path.parents]:
        if part.is_symlink() or os.path.isjunction(part):
            return False
    return True


def select(rows, count=50):
    rows = sorted(rows, key=lambda r: (r[0], r[1], r[2]))
    used = Counter()
    chosen = []
    for i in range(min(count, len(rows))):
        lo, hi = i * len(rows) // count, (i + 1) * len(rows) // count
        if len(rows) < count:
            lo, hi = i, i + 1
        middle = (lo + hi - 1) / 2
        index = min(range(lo, hi), key=lambda j: (used[rows[j][3]], abs(j - middle), rows[j][1], rows[j][2]))
        row = rows[index]
        used[row[3]] += 1
        chosen.append(row)
    return chosen


def inventory(root, source):
    counts = Counter()
    layouts = Counter()
    kinds = Counter()
    days = Counter()
    files = []
    rows = []
    sessions = {}
    if not root.is_dir() or not unredirected(root):
        counts["root_authorization_rejected"] += 1
        raise ValueError("Authorized native root unavailable or redirected")
    counts["authorized_roots"] += 1
    # os.walk 禁止跟随目录链接，junction 也必须拒绝。
    def onerror(error):
        raise OSError("Native directory read failed") from None
    for parent, dirs, names in os.walk(root, followlinks=False, onerror=onerror):
        safe = []
        for name in sorted(dirs):
            child = Path(parent) / name
            if child.is_symlink() or os.path.isjunction(child):
                counts["redirected_directories_rejected"] += 1
                raise ValueError("Native directory is redirected")
            else:
                safe.append(name)
        dirs[:] = safe
        for name in sorted(names):
            if name.endswith(".jsonl"):
                files.append(Path(parent) / name)
    files.sort(key=lambda p: str(p.relative_to(root)).casefold())
    for file_index, path in enumerate(files):
        relative = path.relative_to(root)
        if source == "codex":
            layout = "dated" if len(relative.parts) >= 4 and all(x.isdigit() for x in relative.parts[:3]) else "legacy_other"
        else:
            layout = "subagent" if "subagents" in relative.parts else "main"
        layouts[layout + "_files"] += 1
        if not unredirected(path) or not path.is_file():
            counts["file_authorization_rejected"] += 1
            raise ValueError("Native file unavailable or redirected")
        counts["readable_files"] += 1
        counts["available_bytes"] += path.stat().st_size
        session, project = "", ""
        seen_header = False
        window_file = False
        try:
            with path.open("rb") as handle:
                for line, raw in enumerate(handle, 1):
                    counts["lines_all"] += 1
                    try:
                        record = json.loads(raw)
                    except (ValueError, UnicodeDecodeError):
                        counts["parse_error_all"] += 1
                        raise ValueError("Native source contains invalid JSON") from None
                    if not isinstance(record, dict):
                        counts["non_object_all"] += 1
                        raise ValueError("Native source record is not an object")
                    counts["parsed_objects_all"] += 1
                    payload = record.get("payload")
                    payload = payload if isinstance(payload, dict) else {}
                    if source == "codex" and record.get("type") == "session_meta":
                        session, project = payload.get("id", ""), payload.get("cwd", "")
                        counts["session_header_records_all"] += 1
                        seen_header = bool(session and project)
                    elif source == "claude_code":
                        session = record.get("sessionId") or session
                        project = record.get("cwd") or project
                        if not seen_header and session and project:
                            counts["session_header_records_all"] += 1
                            seen_header = True
                    ts = timestamp(record)
                    if ts is None:
                        counts["timestamp_unknown_all"] += 1
                        continue
                    if not START <= ts < END:
                        counts["outside_window_all"] += 1
                        continue
                    window_file = True
                    counts["window_records"] += 1
                    layouts[layout + "_window_records"] += 1
                    if not session or not project:
                        counts["header_missing_window"] += 1
                    key = (session or "missing", file_index if not session else -1)
                    if key not in sessions:
                        sessions[key] = len(sessions)
                    rows.append((ts, file_index, line, sessions[key], project, bool(session and project)))
                    # 聚合输出只允许固定集合中的原生类型标记。
                    kind = record.get("type")
                    if kind not in {"session_meta", "turn_context", "response_item", "event_msg", "assistant", "user", "system", "progress", "file-history-snapshot", "queue-operation", "summary"}:
                        kind = "other"
                    kinds[kind] += 1
                    day = datetime.fromtimestamp(ts / 1000, datetime.fromisoformat("2026-10-12T00:00:00+08:00").tzinfo).date().isoformat()
                    days[day] += 1
        except OSError:
            counts["read_error_files"] += 1
            raise OSError("Native file read failed") from None
        counts["header_complete_files" if seen_header else "header_missing_files"] += 1
        counts["window_files"] += int(window_file)
    counts["jsonl_files"] = len(files)
    counts["window_sessions"] = len(sessions)
    return counts, layouts, kinds, days, files, rows


def write_json(path, value):
    with path.open("x", encoding="utf-8") as handle:
        json.dump(value, handle, indent=2, sort_keys=True)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", required=True)
    args = parser.parse_args()
    out = Path(args.out).absolute()
    allowed = Path(os.environ["LOCALAPPDATA"]) / "astrocyte-validation" / "v1-20261012"
    if not out.is_relative_to(allowed) or not unredirected(out):
        raise SystemExit("Evidence output outside authorized local evidence root")
    out.mkdir(parents=True, exist_ok=True)
    write_json(out / "sampling-policy.json", POLICY)
    aggregate, manifest = {}, []
    with (out / "sample-first.jsonl").open("xb") as snapshot:
        for source, relative in [("claude_code", ".claude/projects"), ("codex", ".codex/sessions")]:
            root = Path.home() / relative
            counts, layouts, kinds, days, files, rows = inventory(root, source)
            chosen = select(rows)
            by_file = {}
            for ordinal, row in enumerate(chosen):
                by_file.setdefault(row[1], {})[row[2]] = (ordinal, row)
            frozen = {}
            for file_index, wanted in by_file.items():
                with files[file_index].open("rb") as handle:
                    for line, raw in enumerate(handle, 1):
                        if line in wanted:
                            ordinal, row = wanted[line]
                            current = json.loads(raw)
                            if timestamp(current) != row[0]:
                                raise SystemExit("Source changed during sampling; first evidence is incomplete")
                            frozen[ordinal] = (row, raw.rstrip(b"\r\n"))
                        if line >= max(wanted):
                            break
            for ordinal in sorted(frozen):
                row, raw = frozen[ordinal]
                item = {"source": source, "ordinal": ordinal, "file_index": row[1], "line": row[2], "session_index": row[3], "ts": row[0], "header_complete": row[5]}
                manifest.append(item)
                local = dict(item, project=row[4], raw_base64=base64.b64encode(raw).decode("ascii"))
                snapshot.write(json.dumps(local, separators=(",", ":")).encode() + b"\n")
            aggregate[source] = {"counts": dict(counts), "layouts": dict(layouts), "native_types_window": dict(kinds), "days": dict(days), "sampled": len(frozen), "sample_sessions": len({r[3] for r in chosen}), "first_ts": min((r[0] for r in rows), default=None), "last_ts": max((r[0] for r in rows), default=None)}
    write_json(out / "inventory.json", aggregate)
    write_json(out / "sample-manifest.json", manifest)
    print(json.dumps(aggregate, sort_keys=True))


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        # 异常文本不能泄露原生资料及路径。
        print("Diagnostic failed: " + type(error).__name__, file=sys.stderr)
        sys.exit(1)
