"""Only the requested full-protocol SHA lock; no experiment scheduler or manifests."""
from datetime import datetime
import hashlib
import json
from pathlib import Path
import re

from inputs import SEED, SYSTEM, TEMPLATES
from evaluation import JUDGE_SYSTEM, PROJECT_ALIASES, KIND_ALIASES
from local_store import EVIDENCE_ROOT, ZONE, unredirected, write_new
from native_model import authorized


def rules(path):
    text = Path(path).read_text(encoding='utf-8')
    match = re.search(r'<!-- experiment-rules -->\s*```json\s*(.*?)\s*```', text, re.S)
    if not match:
        raise ValueError('normative experiment-rules missing')
    config = json.loads(match.group(1))
    if config['seed'] != SEED or config['system_prompt'] != SYSTEM or config['question_templates'] != [list(t) for t in TEMPLATES] or config['judge_system_prompt'] != JUDGE_SYSTEM:
        raise ValueError('protocol/template implementation differs')
    if config.get('project_aliases') != PROJECT_ALIASES or config.get('kind_aliases') != KIND_ALIASES or config.get('prediction_k') != 3 or config.get('fact_metric') != 'set_precision_recall_f1' or config.get('prediction_metric') != 'per_topic_binary_precision_recall_exact_order':
        raise ValueError('protocol/scoring implementation differs')
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
    if budget.get('backend') == 'native_codex':
        if budget.get('model') != 'gpt-6.1-sol' or config.get('a_representation') != 'layered-A-v4-full-state-header-kind-day-all-event-refs-approved':
            raise ValueError('V4 fixed model/layered A differs')
        if not budget.get('actual_identity_verified') or not route.get('full_population_usage_verified') or not route.get('native_request_input_complete_verified') or type(budget.get('input_utf8_bytes_limit')) is not int:
            raise ValueError('UNLOCKED: actual V4 identity/usage/output/UTF8 budget acceptance pending')
    # W2 prepares this tool. Only root invokes after integration and actual route acceptance.
    if not reviewed or route.get('status') != 'ACCEPTED' or not budget.get('identity'):
        raise ValueError('protocol remains UNLOCKED: route/root review pending')
    required = ('model', 'context_tokens', 'model_path', 'chat_template', 'endpoint')
    if any(route.get(k) != budget.get(k) for k in required):
        raise ValueError('protocol route differs from accepted actual identity')
    runtime = route.get('local_runtime_verified')
    if budget.get('backend') == 'native_codex':
        runtime = route.get('external_runtime_verified') and authorized(budget.get('egress_authorization'))
    if not runtime or not route.get('strong_b_accepted') or not route.get('budget_smoke_passed') or not route.get('production_live_accepted'):
        raise ValueError('actual runtime/scoped privacy/B budget/live acceptance pending')
    original = Path(route.get('formal_live_root', '')).absolute()
    formal = original.resolve()
    if not route.get('formal_live_root') or not formal.is_relative_to((EVIDENCE_ROOT / 'v4-w1').resolve()) or not unredirected(original) or 'prelock' in str(formal).lower():
        raise ValueError('root-owned separate formal live directory required')
    day0 = datetime.now(ZONE).date().isoformat()
    record = {'authority': 'root', 'reviewed': True, 'protocol_sha256': digest(protocol),
              'day0': day0, 'locked_at': datetime.now(ZONE).isoformat(), 'route': route}
    write_new(output, record, root)
    return record
