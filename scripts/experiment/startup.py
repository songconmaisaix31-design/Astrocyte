# 手动检查 V4 启动条件，历史资料保持历史状态。
import hashlib
import json

from a_material import verified_material
from inputs import load_native, messages, questions, prompt_bytes
from local_store import BASELINE, PRIVATE_ROOT, append_row, bounds, events, local_input, private_path, write_new
from native_model import MODEL, ANSWER_SCHEMA, authorized, input_bytes, context_bytes, context_envelope, process
from protocol import rules
from runner import mcp_command, snapshot
from transport import a_memory

SCHEMA = ANSWER_SCHEMA


def validate_reason_answer(decoded):
    # 已登记的 explanation 允许文本或结构化说明，返回对象保持原样。
    if not isinstance(decoded, dict) or not isinstance(decoded.get('explanation'), (str, dict)) or not decoded['explanation'] or not all(isinstance(decoded.get(k), list) for k in ('evidence_event_ids', 'unknowns')):
        raise ValueError('full B reason answer schema differs')


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


def public_context_probe(out, repetitions=160000, pattern='public', input_transport='model_instructions_file'):
    if not 1 <= repetitions <= 160000:
        raise ValueError('public synthetic repetition bounds exceeded')
    q = {'prompt': 'This is public synthetic context. Return {"ready":true,"first_marker":"the first marker in the memory data","last_marker":"the last marker in the memory data"}.'}
    if pattern not in ('public', 'spaces'):
        raise ValueError('unregistered public synthetic pattern')
    memory = {'first_marker': 'PUBLIC_BEGIN_20261013', 'public_filler': ('public ' if pattern == 'public' else '       ') * repetitions, 'last_marker': 'PUBLIC_END_20261013'}
    prompt = messages(q, memory)
    answer, safe = process(prompt, out, 872000, SCHEMA, private=False, input_transport=input_transport)
    if json.loads(answer['answer_json']) != {'ready': True, 'first_marker': memory['first_marker'], 'last_marker': memory['last_marker']}:
        raise ValueError('public long-context begin/end markers differ')
    return safe


def b_smoke(candidate, out, authorization=None, context_tokens=872000, output_reserve=4096, input_transport='stdin'):
    """Exactly one whole-population call, or a persisted NOT_SENT privacy block."""
    candidate = json.loads(local_input(candidate).read_text(encoding='utf-8'))
    native = candidate['native']
    chronological = sorted(native, key=lambda r: (r['ts'], r['source'], r['file'].casefold(), r['line']))
    qs = candidate['questions']
    if len(qs) != 4 or [q['dimension'] for q in qs] != ['fact', 'fact', 'reason', 'prediction']:
        raise ValueError('paired four-question candidate required')
    if len({q['day'] for q in qs}) != 1:
        raise ValueError('B candidate must match one event day')
    _, end = bounds(qs[0]['day'])
    if any(not end - 7 * 86400000 <= r['ts'] < end for r in native):
        raise ValueError('B record outside authorized seven-day event window')
    expected = [messages(q, chronological) for q in qs]
    if candidate['messages_by_question'] != expected:
        raise ValueError('B complete original population/order differs')
    prompt = expected[2]
    out = private_path(out)
    out.mkdir(parents=True, exist_ok=False)
    base = {'requested_model': MODEL, 'evidence': 'startup_full_B_NOT_DAY0',
            'available': len(native), 'loaded': len(native), 'unloaded': 0,
            'prompt_utf8_bytes': prompt_bytes(prompt), 'native_stdin_utf8_bytes': len(context_envelope(prompt, input_transport)[1].encode('utf-8')) if input_transport != 'stdin' else input_bytes(prompt),
            'input_transport': input_transport, 'all_input_utf8_bytes': context_bytes(prompt, input_transport) if input_transport != 'stdin' else input_bytes(prompt),
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
        answer, response = process(prompt, out / 'native', context_tokens, SCHEMA, private=True, authorization=authorization, input_transport=input_transport)
        decoded = json.loads(answer['answer_json'])
        validate_reason_answer(decoded)
        usage = response['usage']
        if response['total_seconds'] > 1800 or usage['input_tokens'] + output_reserve > context_tokens or usage['output_tokens'] > output_reserve:
            raise ValueError('full B actual usage violates shared context/output reserve')
        append_row(out / 'answers-local.jsonl', {'question_id': qs[2]['question_id'], 'arm': 'B', 'answer': decoded, 'response_local': response})
        safe = {**base, 'status': 'ACTUAL_FULL_B_COMPLETE_PENDING_ROOT_ROUTE_ACCEPTANCE',
                'model_calls': 1, 'usage': usage, 'total_seconds': response['total_seconds'],
                'stdin_input_complete': response['stdin_input_complete'],
                'stdin_utf8_bytes': response['stdin_utf8_bytes'],
                'initial_context_utf8_bytes': response['initial_context_utf8_bytes'],
                'initial_context_file_full_input_verified': response['initial_context_file_full_input_verified'],
                'full_packet_reassembly_verified': response.get('full_packet_reassembly_verified', False),
                'actual_response_complete': response['actual_response_complete'],
                'streamed_output_matches': response['streamed_output_matches'],
                'no_tools': response['no_tools'], 'no_compact': response['no_compact'],
                'first_process_byte_seconds': response['first_process_byte_seconds'],
                'first_agent_message_event_seconds': response['first_agent_message_event_seconds'],
                'first_token_time_verified': False, 'identity_and_tokenizer_acceptance': False}
        append_row(out / 'receipts-local.jsonl', safe)
        write_new(out / 'result-safe.json', safe)
        return safe
    except Exception as error:
        append_row(out / 'receipts-local.jsonl', {**base, 'status': 'FAILED', 'error_type': type(error).__name__,
                   'error_local': str(error), 'delivery': 'unknown_no_auto_retry'})
        raise
