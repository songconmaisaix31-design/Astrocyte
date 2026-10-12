"""Manual V4 startup checks; historical material never becomes day0 answers."""
import hashlib
import json

from a_material import verified_material
from inputs import load_native, messages, questions, prompt_bytes
from local_store import BASELINE, PRIVATE_ROOT, append_row, bounds, events, local_input, private_path, write_new
from native_model import MODEL, authorized, input_bytes, process
from protocol import rules
from runner import mcp_command, snapshot
from transport import a_memory

SCHEMA = {'type': 'object', 'properties': {'answer_json': {'type': 'string'}},
          'required': ['answer_json'], 'additionalProperties': False}


def prepare_b(day, selection, db, out, protocol='docs/EXPERIMENT.md'):
    out = private_path(out)
    out.mkdir(parents=True, exist_ok=False)
    start, end = bounds(day)
    native, issues = load_native(local_input(selection), end - 7 * 86400000, end)
    write_new(out / 'source-issues-local.json', issues)
    if issues:
        raise ValueError('full B native population incomplete')
    qs = questions(day, events(local_input(db), start, end), rules(protocol)['baseline_topics'])
    native.sort(key=lambda r: (r['ts'], r['source'], r['file'].casefold(), r['line']))
    prompts = [messages(q, native) for q in qs]
    write_new(out / 'b-full-population-local.json', {'native': native, 'questions': qs, 'messages_by_question': prompts})
    safe = {'status': 'LOCAL_FULL_POPULATION_NOT_INFERENCE', 'evidence': 'historical_NOT_DAY0',
            'available': len(native), 'loaded': len(native), 'unloaded': 0,
            'source_issues': 0, 'prompt_utf8_bytes': [prompt_bytes(p) for p in prompts],
            'native_stdin_utf8_bytes': [input_bytes(p) for p in prompts],
            'token_budget_accepted': False, 'byte_budget_accepted': False}
    write_new(out / 'result-safe.json', safe)
    return safe


def a_smoke(binary, out):
    out = private_path(out)
    out.mkdir(parents=True, exist_ok=False)
    before = hashlib.sha256(BASELINE.read_bytes()).hexdigest()
    db = snapshot(BASELINE, out / 'mcp-view/state.sqlite')
    results = []
    for topic in ('Morphogenesis', 'Astrocyte/Attention Engine', 'McPet'):
        memory, receipt = a_memory(mcp_command(binary, db.parent, topic), topic)
        append_row(out / 'mcp-receipts-local.jsonl', receipt)
        if memory is None:
            raise ValueError('actual A MCP calls failed')
        material, audit = verified_material(memory, db)
        write_new(out / (topic.split('/')[0] + '-material-local.json'), {'material': material, 'memory': memory, 'receipt': receipt, 'audit': audit})
        results.append(audit)
    safe = {'status': 'ACTUAL_LAYERED_A_VERIFIED_NOT_INFERENCE', 'audits': results,
            'baseline_unchanged': before == hashlib.sha256(BASELINE.read_bytes()).hexdigest(),
            'tools_called': 9, 'model_calls': 0, 'evidence': 'historical_NOT_DAY0'}
    write_new(out / 'result-safe.json', safe)
    if not safe['baseline_unchanged']:
        raise ValueError('frozen baseline changed')
    return safe


def public_probe(out, context_tokens=872000):
    prompt = [{'role': 'user', 'content': 'This is a public route probe with no private data. Return {"ready":true}.'}]
    answer, safe = process(prompt, out, context_tokens, SCHEMA, private=False)
    if json.loads(answer['answer_json']) != {'ready': True}:
        raise ValueError('public probe output differs')
    return safe


def b_smoke(candidate, out, authorization=None, context_tokens=872000, output_reserve=4096):
    """Exactly one whole-population call, or a persisted NOT_SENT privacy block."""
    candidate = json.loads(local_input(candidate).read_text(encoding='utf-8'))
    native = candidate['native']
    chronological = sorted(native, key=lambda r: (r['ts'], r['source'], r['file'].casefold(), r['line']))
    qs = candidate['questions']
    if len(qs) != 4 or [q['dimension'] for q in qs] != ['fact', 'fact', 'reason', 'prediction']:
        raise ValueError('paired four-question candidate required')
    expected = [messages(q, chronological) for q in qs]
    if candidate['messages_by_question'] != expected:
        raise ValueError('B complete original population/order differs')
    prompt = expected[2]
    out = private_path(out)
    out.mkdir(parents=True, exist_ok=False)
    base = {'requested_model': MODEL, 'evidence': 'startup_full_B_NOT_DAY0',
            'available': len(native), 'loaded': len(native), 'unloaded': 0,
            'prompt_utf8_bytes': prompt_bytes(prompt), 'native_stdin_utf8_bytes': input_bytes(prompt),
            'context_tokens_requested': context_tokens, 'output_reserve': output_reserve,
            'timeout_seconds': 1800, 'strong_b_accepted': False}
    write_new(out / 'input-local.json', {'question': qs[2], 'messages': prompt, 'loading': base})
    if not authorized(authorization):
        safe = {**base, 'status': 'PRIVATE_EGRESS_BLOCKED', 'delivery': 'not_sent', 'model_calls': 0}
        append_row(out / 'receipts-local.jsonl', safe)
        write_new(out / 'result-safe.json', safe)
        raise ValueError('PRIVATE_EGRESS_BLOCKED: scoped user decision pending; original NOT_SENT retained')
    append_row(out / 'receipts-local.jsonl', {**base, 'status': 'STARTED', 'delivery': 'unknown_no_auto_retry'})
    try:
        answer, response = process(prompt, out / 'native', context_tokens, SCHEMA, private=True, authorization=authorization)
        decoded = json.loads(answer['answer_json'])
        if not isinstance(decoded, dict) or not isinstance(decoded.get('explanation'), str) or not all(isinstance(decoded.get(k), list) for k in ('evidence_event_ids', 'unknowns')):
            raise ValueError('full B reason answer schema differs')
        usage = response['usage']
        if usage['input_tokens'] + output_reserve > context_tokens or usage['output_tokens'] > output_reserve:
            raise ValueError('full B actual usage violates shared context/output reserve')
        append_row(out / 'answers-local.jsonl', {'question_id': qs[2]['question_id'], 'arm': 'B', 'answer': decoded, 'response_local': response})
        safe = {**base, 'status': 'ACTUAL_FULL_B_COMPLETE_PENDING_ROOT_ROUTE_ACCEPTANCE',
                'model_calls': 1, 'usage': usage, 'total_seconds': response['total_seconds'],
                'stdin_input_complete': response['stdin_input_complete'],
                'first_token_time_verified': False, 'identity_and_tokenizer_acceptance': False}
        append_row(out / 'receipts-local.jsonl', safe)
        write_new(out / 'result-safe.json', safe)
        return safe
    except Exception as error:
        append_row(out / 'receipts-local.jsonl', {**base, 'status': 'FAILED', 'error_type': type(error).__name__,
                   'error_local': str(error), 'delivery': 'unknown_no_auto_retry'})
        raise
