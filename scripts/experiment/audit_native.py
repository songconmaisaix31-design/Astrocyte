"""Kind-blind native message inventory; only aggregate metadata on stdout."""
from collections import Counter
from datetime import datetime
import json
from pathlib import Path

from inputs import native_message, select_native_files, timestamp
from local_store import PRIVATE_ROOT, ZONE, bounds, write_new


def audit(day):
    _, end = bounds(day)
    start = end - 7 * 86400000
    selection = select_native_files({'claude_code': 'C:/Users/DW/.claude/projects', 'codex': 'C:/Users/DW/.codex/sessions'})
    write_new(PRIVATE_ROOT / 'native-selection-first.json', selection)
    total, files_local = Counter(), []
    for entry in selection:
        for name in entry['files']:
            counts, canonical, mirrors = Counter(), [], []
            with (Path(entry['root']) / name).open(encoding='utf-8') as f:
                for line_number, line in enumerate(f, 1):
                    try:
                        record = json.loads(line)
                        ts = timestamp(record)
                        message = native_message(entry['source'], record)
                        payload = record.get('payload', {})
                        event_message = entry['source'] == 'codex' and record.get('type') == 'event_msg' and payload.get('type') in ('user_message', 'agent_message')
                        if message is None and not event_message:
                            continue
                        if ts is None:
                            counts['message_timestamp_unknown'] += 1
                            continue
                        if not start <= ts < end:
                            continue
                        if message:
                            role, content, _ = message
                            counts['canonical_' + role] += 1
                            counts['canonical_json_chars'] += len(json.dumps(content, ensure_ascii=False))
                            if isinstance(content, str):
                                text = content
                            elif isinstance(content, list):
                                text = ''.join(c.get('text', '') for c in content if isinstance(c, dict) and c.get('type') in ('text', 'input_text', 'output_text'))
                            else:
                                text = ''
                            canonical.append((line_number, ts, role, text))
                        if event_message:
                            role = 'user' if payload['type'] == 'user_message' else 'assistant'
                            text = payload.get('message')
                            counts['event_msg_' + role] += 1
                            counts['event_msg_chars'] += len(text) if isinstance(text, str) else 0
                            mirrors.append((line_number, ts, role, text))
                    except Exception:
                        counts['line_parse_errors'] += 1
            unmatched, consumed = [], set()
            for mirror in mirrors:
                # Inventory evidence only, not a parser dedup rule: matching content is never
                # enough to remove a native message. Record how many have nearby counterparts.
                nearby = [i for i, r in enumerate(canonical) if i not in consumed and r[2:] == mirror[2:] and abs(r[1] - mirror[1]) <= 2000 and abs(r[0] - mirror[0]) <= 8]
                if nearby:
                    consumed.add(min(nearby, key=lambda i: abs(canonical[i][0] - mirror[0])))
                    counts['event_near_canonical_same_text'] += 1
                else:
                    unmatched.append({'line': mirror[0], 'ts': mirror[1], 'role': mirror[2], 'chars': len(mirror[3]) if isinstance(mirror[3], str) else 0})
                    counts['event_without_near_canonical'] += 1
            files_local.append({'source': entry['source'], 'file': name, 'counts': dict(counts), 'unmatched_event_metadata': unmatched})
            for k, v in counts.items():
                total[entry['source'] + ':' + k] += v
    report = {'evidence': 'real_native_inventory_NOT_MODEL_NOT_DAY0', 'day_end': day,
              'window_start': start, 'window_end': end, 'files': {e['source']: len(e['files']) for e in selection},
              'counts': dict(total), 'parser_completeness': 'REVIEW_REQUIRED', 'raw_printed': False}
    write_new(PRIVATE_ROOT / 'native-inventory-first-local.json', {'report': report, 'files': files_local})
    write_new(PRIVATE_ROOT / 'native-inventory-first-safe.json', report)
    return report


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--day', required=True)
    a = parser.parse_args()
    print(json.dumps(audit(a.day), ensure_ascii=False))
