"""Private JSONL evidence and read-only event projection. No engine implementation."""
from contextlib import contextmanager
from datetime import datetime, timedelta, timezone
import json
import os
from pathlib import Path
import sqlite3

ZONE = timezone(timedelta(hours=8))
PRIVATE_ROOT = Path('C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v3-w2')
EVIDENCE_ROOT = PRIVATE_ROOT.parent
BASELINE = Path('C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1/replay-002/state.sqlite')
REPO = Path(__file__).resolve().parents[2]


def unredirected(path):
    return all(not p.is_symlink() and not os.path.isjunction(p) for p in [path, *path.parents])


def private_path(path, root=PRIVATE_ROOT):
    original, original_root = Path(path).absolute(), Path(root).absolute()
    p, r = original.resolve(), original_root.resolve()
    if not unredirected(original) or not unredirected(original_root) or not p.is_relative_to(r) or p.is_relative_to(REPO):
        raise ValueError('local evidence path rejected')
    return p


def local_input(path):
    """Explicit W1/W2 evidence file reads allowed; writes remain W2-exclusive."""
    original = Path(path).absolute()
    p = original.resolve()
    if not p.is_relative_to(EVIDENCE_ROOT.resolve()) or not unredirected(original) or not p.is_file():
        raise ValueError('external local evidence input rejected')
    return p


def write_new(path, value, root=PRIVATE_ROOT):
    p = private_path(path, root)
    p.parent.mkdir(parents=True, exist_ok=True)
    with p.open('x', encoding='utf-8', newline='\n') as f:
        json.dump(value, f, ensure_ascii=False, indent=2)
        f.write('\n')


def append_row(path, row, root=PRIVATE_ROOT):
    p = private_path(path, root)
    p.parent.mkdir(parents=True, exist_ok=True)
    # Single W2 owner; fsync before reporting success. No overwrite/retry orchestration.
    with p.open('a', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(row, ensure_ascii=False, separators=(',', ':')) + '\n')
        f.flush()
        os.fsync(f.fileno())


def rows(path):
    with Path(path).open(encoding='utf-8') as f:
        return [json.loads(line) for line in f if line.strip()]


def bounds(day):
    start = datetime.fromisoformat(day).replace(tzinfo=ZONE)
    return int(start.timestamp() * 1000), int((start + timedelta(days=1)).timestamp() * 1000)


@contextmanager
def readonly_db(path):
    p = Path(path).absolute()
    if not p.is_file() or not unredirected(p):
        raise ValueError('event database path rejected')
    conn = sqlite3.connect(p.as_uri() + '?mode=ro', uri=True)
    try:
        conn.execute('PRAGMA query_only=ON')
        conn.execute('BEGIN')  # coherent snapshot; raw/data/observation never selected
        yield conn
    finally:
        conn.rollback()
        conn.close()


def events(path, start, end):
    with readonly_db(path) as conn:
        cursor = conn.execute('SELECT e.id,e.source,e.ts,e.kind,e.topic FROM attention_engine_events e '
                              'WHERE e.mapping_status=? AND e.ts>=? AND e.ts<? '
                              'AND EXISTS(SELECT 1 FROM attention_engine_signals s WHERE s.event_id=e.id) ORDER BY e.ts,e.id',
                              ('mapped', start, end))
        return [dict(zip(('event_id', 'source', 'ts', 'kind', 'topic'), r)) for r in cursor]


def complete_day(row, day):
    start, end = bounds(day)
    return (row.get('day') == day and row.get('window_start') == start and row.get('window_end') == end
            and row.get('capture_complete') is True and row.get('selection_complete') is True
            and row.get('sleep_status') in ('completed', 'already_completed')
            and bool(row.get('sleep_receipt_id')) and not row.get('late') and not row.get('source_errors'))


def distribution(path):
    with readonly_db(path) as conn:
        return {'events': conn.execute('SELECT COUNT(*) FROM attention_engine_events').fetchone()[0],
                'topics': dict(conn.execute('SELECT topic,COUNT(*) FROM attention_engine_events GROUP BY topic')),
                'kinds': dict(conn.execute('SELECT kind,COUNT(*) FROM attention_engine_events GROUP BY kind')),
                'sleep_runs': conn.execute('SELECT COUNT(*) FROM attention_engine_sleep_runs').fetchone()[0]}
