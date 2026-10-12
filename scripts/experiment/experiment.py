"""Run from repository root. Stdout contains status/counts only; raw stays local."""
import argparse
from datetime import datetime
import hashlib
import json
from pathlib import Path
import sys

from evaluation import prediction, summarize
from inputs import load_native, pack_b, questions, select_native_files
from local_store import BASELINE, PRIVATE_ROOT, ZONE, append_row, bounds, distribution, events, local_input, private_path, rows, write_new
from protocol import check_lock, lock_protocol, rules
from runner import import_user, mcp_command, run_day, run_judge, snapshot
from transport import LocalModel, a_memory, probe


def read_private(path):
    return json.loads(private_path(path).read_text(encoding='utf-8'))


def args():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='command', required=True)
    p = sub.add_parser('baseline')
    p.add_argument('--out', required=True)
    p = sub.add_parser('probe')
    p.add_argument('--out', required=True)
    p = sub.add_parser('select-native')
    p.add_argument('--out', required=True)
    p = sub.add_parser('preview')
    p.add_argument('--day', required=True)
    p.add_argument('--db', default=str(BASELINE))
    p.add_argument('--out', required=True)
    p = sub.add_parser('mcp-smoke')
    p.add_argument('--binary', required=True)
    p.add_argument('--out', required=True)
    p = sub.add_parser('model-smoke')
    p.add_argument('--endpoint', required=True)
    p.add_argument('--model', required=True)
    p.add_argument('--context-tokens', type=int, required=True)
    p.add_argument('--out', required=True)
    p.add_argument('--selection')
    p.add_argument('--day')
    p = sub.add_parser('lock')
    p.add_argument('--protocol', default='docs/EXPERIMENT.md')
    p.add_argument('--route', required=True)
    p.add_argument('--out', required=True)
    p.add_argument('--root-reviewed', action='store_true')
    p = sub.add_parser('run-day')
    p.add_argument('--day', required=True)
    p.add_argument('--db', required=True)
    p.add_argument('--selection', required=True)
    p.add_argument('--coverage', required=True)
    p.add_argument('--protocol', default='docs/EXPERIMENT.md')
    p.add_argument('--lock', required=True)
    p.add_argument('--binary', required=True)
    p.add_argument('--out', required=True)
    p = sub.add_parser('import-user')
    p.add_argument('--source', required=True)
    p.add_argument('--day-dir', required=True)
    p = sub.add_parser('judge')
    p.add_argument('--day-dir', required=True)
    p.add_argument('--protocol', default='docs/EXPERIMENT.md')
    p.add_argument('--lock', required=True)
    p = sub.add_parser('score-predictions')
    p.add_argument('--day-dir', required=True)
    p.add_argument('--db', required=True)
    p.add_argument('--coverage', required=True)
    p.add_argument('--protocol', default='docs/EXPERIMENT.md')
    p.add_argument('--lock', required=True)
    p = sub.add_parser('summary')
    p.add_argument('--day-dirs', nargs='+', required=True)
    p.add_argument('--protocol', default='docs/EXPERIMENT.md')
    p.add_argument('--lock', required=True)
    p.add_argument('--out', required=True)
    return parser.parse_args()


def execute(a):
    if a.command == 'baseline':
        report = {'evidence': 'real_readonly_V1R', **distribution(BASELINE)}
        write_new(a.out, report)
        return report
    if a.command == 'probe':
        report = probe([f'http://127.0.0.1:{port}' for port in (11434, 1234, 8080, 8000)])
        write_new(a.out, report)
        return {'status': 'PROBED', 'available': sum(r['status'] != 'UNAVAILABLE' for r in report)}
    if a.command == 'select-native':
        selection = select_native_files({'claude_code': 'C:/Users/DW/.claude/projects', 'codex': 'C:/Users/DW/.codex/sessions'})
        write_new(a.out, selection)
        return {'status': 'LOCAL_SELECTION', 'files': {r['source']: len(r['files']) for r in selection}}
    if a.command == 'preview':
        start, end = bounds(a.day)
        actual = events(a.db, start, end)
        config = rules('docs/EXPERIMENT.md')
        qs = questions(a.day, actual, config['baseline_topics'])
        write_new(a.out, {'evidence': 'historical_preview_NOT_DAY0', 'questions': qs, 'event_count': len(actual)})
        return {'status': 'PREVIEW_ONLY', 'questions': len(qs), 'events': len(actual)}
    if a.command == 'mcp-smoke':
        out = private_path(a.out)
        out.mkdir(parents=True, exist_ok=False)
        before = hashlib.sha256(BASELINE.read_bytes()).hexdigest()
        copy = snapshot(BASELINE, out / 'mcp-view/state.sqlite')
        memory, receipt = a_memory(mcp_command(a.binary, copy.parent, 'Astrocyte/Attention Engine'), 'Astrocyte/Attention Engine')
        write_new(out / 'mcp-local.json', {'memory': memory, 'receipt': receipt})
        after = hashlib.sha256(BASELINE.read_bytes()).hexdigest()
        result = {'evidence': 'real_MCP_historical_snapshot_NOT_MODEL_NOT_DAY0', 'status': receipt['status'],
                  'baseline_unchanged': before == after, 'baseline_sha256': before,
                  'tools_called': 3, 'model_calls': 0}
        write_new(out / 'result-safe.json', result)
        if memory is None or before != after:
            raise ValueError('real MCP smoke failed; inspect private receipt')
        return result
    if a.command == 'model-smoke':
        out = private_path(a.out)
        out.mkdir(parents=True, exist_ok=False)
        client = LocalModel(a.endpoint, a.model, a.context_tokens)
        append_row(out / 'model-local.jsonl', {'status': 'STARTED', 'evidence': 'candidate_smoke_NOT_DAY0'})
        try:
            identity = client.identity()
            smoke = [{'role': 'system', 'content': '只输出JSON。'}, {'role': 'user', 'content': '输出 {"ready":true}。'}]
            answer, response = client.answer(smoke, 256, 2026101303)
            append_row(out / 'model-local.jsonl', {'identity': identity, 'answer': answer, 'response_local': response, 'status': 'PUBLIC_SMOKE_OK'})
            result = {'status': 'PUBLIC_SMOKE_OK', 'model_calls': 1, 'context_declared': a.context_tokens,
                      'actual_full_budget_inference': False, 'strong_b_accepted': False}
            if a.selection:
                if not a.day:
                    raise ValueError('B smoke day required')
                _, end = bounds(a.day)
                native, issues = load_native(local_input(a.selection), end - 7 * 86400000, end)
                write_new(out / 'native-issues-local.json', issues)
                if issues:
                    raise ValueError('B native source incomplete')
                q = questions(a.day, [], rules('docs/EXPERIMENT.md')['baseline_topics'])[2]
                budget = {'context_tokens': a.context_tokens, 'output_reserve': 4096,
                          'system_reserve': 2048, 'question_reserve': 2048, 'template_reserve': 1024}
                prompt, loading = pack_b(q, native, client.count, budget)
                write_new(out / 'b-input-local.json', {'messages': prompt, 'loading': loading})
                append_row(out / 'model-local.jsonl', {'status': 'B_STARTED', 'at': datetime.now(ZONE).isoformat(), 'loading': loading})
                answer, response = client.answer(prompt, 4096, q['seed'])
                append_row(out / 'model-local.jsonl', {'status': 'B_SMOKE_OK', 'answer': answer, 'response_local': response, 'loading': loading})
                result.update(status='B_SMOKE_OK', model_calls=2, actual_full_budget_inference=True,
                              b_available=loading['available'], b_loaded=loading['loaded'], b_unloaded=loading['unloaded'],
                              prompt_tokens=loading['prompt_tokens'])
            write_new(out / 'result-safe.json', result)
            return result
        except Exception as error:
            append_row(out / 'model-local.jsonl', {'status': 'FAILED', 'error_type': type(error).__name__, 'error_local': str(error),
                                                'http_error_local': client.last_error, 'last_response_local': client.last_response,
                                                'delivery': 'unknown_no_auto_retry', 'at': datetime.now(ZONE).isoformat()})
            raise
    if a.command == 'lock':
        record = lock_protocol(a.protocol, a.out, read_private(a.route), a.root_reviewed, PRIVATE_ROOT)
        return {'status': 'LOCKED_BY_ROOT', 'day0': record['day0'], 'protocol_sha256': record['protocol_sha256']}
    if a.command == 'run-day':
        run_day(a.day, local_input(a.db), local_input(a.selection), rows(local_input(a.coverage)), a.protocol, private_path(a.lock), a.binary, a.out)
        return {'status': 'DAY_COLLECTED_CHECK_ARM_STATUSES'}
    if a.command == 'import-user':
        import_user(a.source, a.day_dir)
        return {'status': 'USER_RATING_IMPORTED'}
    if a.command == 'judge':
        budget = rules(a.protocol)['model']
        client = LocalModel(budget['endpoint'], budget['model'], budget['context_tokens'])
        run_judge(a.day_dir, client, a.protocol, private_path(a.lock))
        return {'status': 'INDEPENDENT_JUDGE_RECORDED'}
    if a.command == 'score-predictions':
        locked = check_lock(a.protocol, private_path(a.lock))
        p = private_path(a.day_dir)
        question = next(q for q in rows(p / 'questions.jsonl') if q['dimension'] == 'prediction')
        latest = {(r['question_id'], r['arm']): r for r in rows(p / 'answers.jsonl')}
        for arm in 'ABC':
            answer = latest.get((question['question_id'], arm), {})
            score = prediction(question, answer.get('answer', {}), local_input(a.db), rows(local_input(a.coverage)), locked=locked)
            append_row(p / 'prediction-scores.jsonl', {'question_id': question['question_id'], 'arm': arm, **score})
        return {'status': 'MATURITY_CHECK_RECORDED'}
    if a.command == 'summary':
        locked = check_lock(a.protocol, private_path(a.lock))
        collected = {name: [] for name in ('questions', 'answers', 'fact-scores', 'user-scores', 'prediction-scores', 'judge-scores')}
        mappings = {}
        for folder in a.day_dirs:
            p = private_path(folder)
            receipt = rows(p / 'receipts.jsonl')[0]
            if receipt.get('evidence') != 'prospective' or receipt.get('protocol_sha256') != locked['protocol_sha256'] or receipt.get('day0') != locked['day0']:
                raise ValueError('summary cannot consume prelock or another protocol day')
            for name in collected:
                file = p / (name + '.jsonl')
                if file.exists():
                    collected[name].extend(rows(file))
            file = p / 'blind-mapping.json'
            if file.exists():
                mappings.update(read_private(file))
        result = summarize(collected['questions'], collected['answers'], collected['fact-scores'],
                           collected['user-scores'], collected['prediction-scores'], mappings, locked['day0'],
                           judge_scores=collected['judge-scores'])
        write_new(a.out, result)
        return result


def main():
    a = args()
    try:
        result = execute(a)
        print(json.dumps(result, ensure_ascii=False))
    except Exception as error:
        # Never print private HTTP body, path, prompt or exception string.
        status = 'INVALID' if str(error).startswith('INVALID:') else 'FAILED'
        append_row(PRIVATE_ROOT / 'cli-failures-local.jsonl', {'status': status, 'command': a.command,
                   'error_type': type(error).__name__, 'error_local': str(error), 'at': datetime.now(ZONE).isoformat()})
        print(json.dumps({'status': status, 'error_type': type(error).__name__, 'command': a.command}), file=sys.stderr)
        return 1
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
