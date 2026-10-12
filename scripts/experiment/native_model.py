"""Released isolated native Codex text processor, explicit V4 route boundary.

No credential reads, tool execution, child agents, automatic replay or private
egress without a scoped user authorization record. Public probe is separate.
"""
import json
import os
from pathlib import Path
import queue
import subprocess
import threading
import time

from local_store import PRIVATE_ROOT, append_row, local_input, private_path, write_new

MODEL = 'gpt-6.1-sol'
EXECUTABLE = Path('C:/Users/DW/AppData/Roaming/npm/node_modules/@openai/codex/node_modules/@openai/codex-win32-x64/vendor/x86_64-pc-windows-msvc/bin/codex.exe')
DISABLED = ['shell_tool', 'unified_exec', 'plugins', 'apps', 'browser_use', 'browser_use_external',
            'browser_use_full_cdp_access', 'computer_use', 'view_image', 'image_generation',
            'multi_agent', 'multi_agent_v2', 'memories', 'hooks', 'code_mode', 'code_mode_host',
            'skill_search', 'tool_suggest', 'workspace_dependencies', 'realtime_conversation',
            'remote_plugin', 'remote_models', 'in_app_browser', 'in_app_local_automation', 'sleep_tool']
ENDPOINT_BOUNDARY = 'native OpenAI Responses provider; external provider inference, CLI-owned existing authentication'
ANSWER_PREFIX = 'Return the requested answer object serialized into the answer_json string.\n'


def input_bytes(messages):
    return len((ANSWER_PREFIX + json.dumps(messages, ensure_ascii=False, separators=(',', ':'))).encode('utf-8'))


def environment():
    return {k: os.environ[k] for k in ['SystemRoot', 'WINDIR', 'COMSPEC', 'PATH', 'APPDATA',
            'LOCALAPPDATA', 'USERPROFILE', 'HOME', 'CODEX_HOME', 'TEMP', 'TMP'] if k in os.environ}


def arguments(out, context_tokens):
    args = [str(EXECUTABLE), 'exec', '--ignore-user-config', '--ignore-rules', '--strict-config',
            '--ephemeral', '--skip-git-repo-check', '--sandbox', 'read-only', '-C', str(PRIVATE_ROOT / 'native-empty-workdir'), '-m', MODEL,
            '-c', 'model_reasoning_effort="high"', '-c', 'service_tier="priority"',
            '-c', f'model_context_window={context_tokens}', '-c', 'model_auto_compact_token_limit=9223372036854775807',
            '-c', 'project_doc_max_bytes=0', '-c', 'web_search="disabled"', '--enable', 'skip_host_skill_discovery',
            '-c', 'model_provider="v4-isolated"',
            '-c', 'model_providers.v4-isolated={name="OpenAI",wire_api="responses",requires_openai_auth=true,request_max_retries=0,stream_max_retries=0,stream_idle_timeout_ms=1800000}']
    for feature in DISABLED:
        args.extend(['--disable', feature])
    return args + ['-c', 'agents.enabled=false', '-c', 'skills.include_instructions=false',
                   '-c', 'include_apps_instructions=false', '-c', 'mcp_servers={}', '--json',
                   '--output-schema', str(out / 'schema.json'), '--output-last-message', str(out / 'output.json'), '-']


def authorized(authorization):
    if authorization is None:
        return False
    row = json.loads(local_input(authorization).read_text(encoding='utf-8'))
    return (row.get('authority') == 'user' and row.get('status') == 'APPROVED'
            and row.get('model') == MODEL and row.get('endpoint_boundary') == ENDPOINT_BOUNDARY
            and row.get('scope') == 'V4 ABC selected A derived and full seven-day Claude/Codex B'
            and bool(row.get('decision_reference')))


class NativeCodexModel:
    """Native full-population route; actual usage is checked after completion.

    The current CLI has no exact preflight tokenizer. It cannot select messages
    near a context boundary. Whole-population requests preserve every input and
    must fail if actual usage exceeds the shared budget, never truncate.
    """
    def __init__(self, budget):
        self.budget = budget
        self.expected_identity = None
        self.last_response = self.last_error = None
        self.index = 0

    def identity(self):
        b = self.budget
        if b.get('model') != MODEL or not b.get('actual_identity_verified'):
            raise ValueError('actual gpt identity/context/configuration gate pending')
        return {k: b[k] for k in ('model', 'context_tokens', 'model_path', 'chat_template', 'endpoint')}

    def count(self, messages):
        raise ValueError('native exact preflight tokenizer not exposed; boundary selection unavailable')

    def answer(self, messages, output_tokens, seed):
        self.identity()
        self.index += 1
        schema = {'type': 'object', 'properties': {'answer_json': {'type': 'string'}},
                  'required': ['answer_json'], 'additionalProperties': False}
        out = PRIVATE_ROOT / ('native-completion-' + str(time.time_ns()))
        answer, response = process(messages, out, self.budget['context_tokens'], schema,
                                   private=True, authorization=self.budget.get('egress_authorization'), timeout=1800)
        self.last_response = response
        if response['usage']['input_tokens'] + output_tokens > self.budget['context_tokens'] or response['usage']['output_tokens'] > output_tokens:
            raise ValueError('native full-input actual usage violates shared context/output reserve')
        decoded = json.loads(answer['answer_json'])
        if not isinstance(decoded, dict):
            raise ValueError('native structured answer object required')
        return decoded, response


def process(messages, out, context_tokens, schema, private=True, authorization=None, timeout=1800):
    if private and not authorized(authorization):
        raise ValueError('PRIVATE_EGRESS_BLOCKED: local-only raw order requires explicit scoped user decision')
    out = private_path(out)
    out.mkdir(parents=True, exist_ok=False)
    write_new(out / 'schema.json', schema)
    working = private_path(PRIVATE_ROOT / 'native-empty-workdir')
    working.mkdir(parents=True, exist_ok=True)
    if any(working.iterdir()):
        raise ValueError('fixed native working root must remain empty')
    prompt = json.dumps(messages, ensure_ascii=False, separators=(',', ':'))
    if 'answer_json' in schema.get('properties', {}):
        prompt = ANSWER_PREFIX + prompt
    write_new(out / 'input-local.json', {'messages': messages, 'utf8_bytes': len(prompt.encode('utf-8'))})
    args = arguments(out, context_tokens)
    write_new(out / 'configuration-local.json', {'model': MODEL, 'context_tokens_requested': context_tokens,
              'argv': args, 'timeout_seconds': timeout, 'endpoint_boundary': ENDPOINT_BOUNDARY,
              'private': private, 'sampling': 'provider default, equal ABC', 'reasoning': 'high', 'service_tier': 'priority'})
    append_row(out / 'receipts-local.jsonl', {'status': 'STARTED', 'delivery': 'unknown_no_auto_retry'})
    start = time.monotonic()
    flags = subprocess.CREATE_NO_WINDOW if os.name == 'nt' else 0
    try:
        child = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                 cwd=working, env=environment(), creationflags=flags)
    except OSError as error:
        append_row(out / 'receipts-local.jsonl', {'status': 'FAILED', 'error_type': type(error).__name__,
                   'error_local': str(error), 'delivery': 'not_started', 'total_seconds': time.monotonic() - start})
        raise
    records = queue.Queue()
    def read(channel, stream):
        # Real pipe bytes are consumed as they arrive, not buffered until complete lines.
        while True:
            chunk = os.read(stream.fileno(), 4096)
            if not chunk:
                records.put((channel, None, time.monotonic() - start))
                return
            records.put((channel, chunk, time.monotonic() - start))
    for channel, stream in [('stdout', child.stdout), ('stderr', child.stderr)]:
        threading.Thread(target=read, args=(channel, stream), daemon=True).start()
    fed = threading.Event()
    input_receipt = {'expected_bytes': len(prompt.encode('utf-8')), 'written_bytes': 0, 'closed': False}
    def feed():
        try:
            input_receipt['written_bytes'] = child.stdin.write(prompt.encode('utf-8'))
            child.stdin.close()
            input_receipt['closed'] = True
        except (OSError, BrokenPipeError) as error:
            input_receipt['error_type'] = type(error).__name__
        finally:
            fed.set()
    threading.Thread(target=feed, daemon=True).start()
    closed, first, first_model, usage, completed = set(), None, None, None, False
    buffers = {'stdout': b'', 'stderr': b''}
    events = []
    try:
        with (out / 'stdout-local.jsonl').open('xb') as stdout, (out / 'stderr-local.log').open('xb') as stderr:
            files = {'stdout': stdout, 'stderr': stderr}
            while len(closed) < 2:
                remaining = timeout - (time.monotonic() - start)
                if remaining <= 0:
                    raise TimeoutError('total elapsed 1800-second boundary reached')
                try:
                    channel, chunk, elapsed = records.get(timeout=min(remaining, 1))
                except queue.Empty:
                    continue
                if chunk is None:
                    closed.add(channel)
                    continue
                files[channel].write(chunk)
                files[channel].flush()
                if first is None:
                    first = elapsed  # transport/process bytes, not model tokens
                if channel == 'stdout':
                    buffers[channel] += chunk
                    while b'\n' in buffers[channel]:
                        line, buffers[channel] = buffers[channel].split(b'\n', 1)
                        if not line.strip():
                            continue
                        event = json.loads(line)
                        events.append(event)
                        append_row(out / 'stream-times-local.jsonl', {'type': event.get('type'), 'elapsed_seconds': elapsed})
                        item = event.get('item', {})
                        if item.get('type') == 'agent_message' and first_model is None:
                            first_model = elapsed
                        if event.get('type') == 'turn.completed':
                            completed, usage = True, event.get('usage')
            code = child.wait(timeout=max(.01, timeout - (time.monotonic() - start)))
        total = time.monotonic() - start
        if not fed.wait(timeout=max(.01, timeout - total)) or not input_receipt['closed'] or input_receipt['written_bytes'] != input_receipt['expected_bytes']:
            raise ValueError('native full stdin input delivery incomplete')
        if code or not completed or not isinstance(usage, dict) or not all(type(usage.get(k)) is int for k in ('input_tokens', 'output_tokens')):
            raise ValueError('native complete termination or actual usage missing')
        if any(e.get('type') == 'context.compacted' or e.get('item', {}).get('type') in ('command_execution', 'mcp_tool_call', 'web_search', 'collab_tool_call') for e in events):
            raise ValueError('isolated native processor attempted forbidden tools')
        answer = json.loads((out / 'output.json').read_text(encoding='utf-8'))
        streamed = [e['item']['text'] for e in events if e.get('type') == 'item.completed' and e.get('item', {}).get('type') == 'agent_message']
        if not streamed or json.loads(streamed[-1]) != answer or not isinstance(answer, dict):
            raise ValueError('native stream and complete output disagree')
        report = {'status': 'PUBLIC_NATIVE_COMPLETE' if not private else 'PRIVATE_NATIVE_COMPLETE',
                  'requested_model': MODEL, 'response_model_echo': None, 'usage': usage,
                  'first_process_byte_seconds': first, 'first_agent_message_event_seconds': first_model,
                  'first_token_time_verified': False, 'total_seconds': total,
                  'actual_response_complete': True, 'streamed_output_matches': True,
                  'stdin_input_complete': True, 'stdin_utf8_bytes': input_receipt['written_bytes'],
                  'provider_input_wire_verified': False,
                  'identity_and_tokenizer_acceptance': False, 'private_egress_authorized': private}
        append_row(out / 'receipts-local.jsonl', report)
        write_new(out / 'result-safe.json', report)
        return answer, report
    except Exception as error:
        if child.poll() is None:
            child.kill()
        child.wait()
        append_row(out / 'receipts-local.jsonl', {'status': 'FAILED', 'error_type': type(error).__name__,
                   'error_local': str(error), 'delivery': 'unknown_no_auto_retry',
                   'first_process_byte_seconds': first, 'first_agent_message_event_seconds': first_model,
                   'total_seconds': time.monotonic() - start, 'stdin_receipt': input_receipt})
        raise
    finally:
        for stream in (child.stdin, child.stdout, child.stderr):
            stream.close()
