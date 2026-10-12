"""Only the requested full-protocol SHA lock; no experiment scheduler or manifests."""
from datetime import datetime
import hashlib
import json
from pathlib import Path
import re

from inputs import SEED, SYSTEM, TEMPLATES
from evaluation import JUDGE_SYSTEM
from local_store import EVIDENCE_ROOT, ZONE, unredirected, write_new


def rules(path):
    text = Path(path).read_text(encoding='utf-8')
    match = re.search(r'<!-- experiment-rules -->\s*```json\s*(.*?)\s*```', text, re.S)
    if not match:
        raise ValueError('normative experiment-rules missing')
    config = json.loads(match.group(1))
    if config['seed'] != SEED or config['system_prompt'] != SYSTEM or config['question_templates'] != [list(t) for t in TEMPLATES] or config['judge_system_prompt'] != JUDGE_SYSTEM:
        raise ValueError('protocol/template implementation differs')
    return config


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def check_lock(protocol, lock):
    record = json.loads(Path(lock).read_text(encoding='utf-8'))
    if record['protocol_sha256'] != digest(protocol):
        raise ValueError('INVALID: protocol changed after lock')
    if record.get('authority') != 'root' or not record.get('reviewed'):
        raise ValueError('INVALID: authoritative root review missing')
    return record


def lock_protocol(protocol, output, route, reviewed, root):
    config = rules(protocol)
    budget = config['model']
    # W2 prepares this tool. Only root invokes after integration and actual route acceptance.
    if not reviewed or route.get('status') != 'ACCEPTED' or not budget.get('identity'):
        raise ValueError('protocol remains UNLOCKED: route/root review pending')
    required = ('model', 'context_tokens', 'model_path', 'chat_template', 'endpoint')
    if any(route.get(k) != budget.get(k) for k in required):
        raise ValueError('protocol route differs from accepted actual identity')
    if not route.get('local_runtime_verified') or not route.get('strong_b_accepted') or not route.get('budget_smoke_passed') or not route.get('production_live_accepted'):
        raise ValueError('actual local route/B budget/live acceptance pending')
    formal = Path(route.get('formal_live_root', '')).absolute()
    if not route.get('formal_live_root') or not formal.is_relative_to((EVIDENCE_ROOT / 'v4-w1').absolute()) or not unredirected(formal) or 'prelock' in str(formal).lower():
        raise ValueError('root-owned separate formal live directory required')
    day0 = datetime.now(ZONE).date().isoformat()
    record = {'authority': 'root', 'reviewed': True, 'protocol_sha256': digest(protocol),
              'day0': day0, 'locked_at': datetime.now(ZONE).isoformat(), 'route': route}
    write_new(output, record, root)
    return record
