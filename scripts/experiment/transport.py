"""Local inference only; actual MCP subprocess, no fixture fallback."""
import ipaddress
import json
from pathlib import Path
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise ValueError('local inference redirect rejected')


class LocalModel:
    def __init__(self, endpoint, model, context_tokens, timeout=1800):
        parsed = urllib.parse.urlsplit(endpoint)
        try:
            local = ipaddress.ip_address(parsed.hostname or '').is_loopback
        except ValueError:
            local = False
        if not local or parsed.scheme != 'http' or parsed.username or parsed.password or parsed.path not in ('', '/') or parsed.query or parsed.fragment:
            raise ValueError('numeric loopback inference URL required')
        self.endpoint, self.model, self.context_tokens = endpoint.rstrip('/'), model, context_tokens
        self.timeout = timeout
        self.last_response = None
        self.last_error = None
        self.expected_identity = None
        self.opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())

    def request(self, path, data=None):
        body = None if data is None else json.dumps(data, ensure_ascii=False).encode('utf-8')
        request = urllib.request.Request(self.endpoint + path, data=body, headers={'Content-Type': 'application/json'})
        try:
            with self.opener.open(request, timeout=self.timeout) as response:
                self.last_response = json.load(response)
                return self.last_response
        except urllib.error.HTTPError as error:
            self.last_error = {'http_status': error.code, 'body_local': error.read().decode('utf-8', errors='replace')}
            raise

    def identity(self):
        models, props = self.request('/v1/models'), self.request('/props')
        if self.model not in [m['id'] for m in models['data']]:
            raise ValueError('requested model not served')
        actual = props.get('default_generation_settings', {}).get('n_ctx')
        if actual != self.context_tokens or not props.get('chat_template') or not props.get('model_path'):
            raise ValueError('actual context/template/model_path not verified')
        identity = {'model': self.model, 'context_tokens': actual, 'model_path': props['model_path'],
                    'chat_template': props['chat_template'], 'endpoint': self.endpoint}
        if self.expected_identity is not None and any(identity[k] != self.expected_identity[k] for k in identity):
            raise ValueError('actual model/template identity changed')
        self.expected_identity = identity
        return identity

    def count(self, messages):
        prompt = self.request('/apply-template', {'messages': messages, 'add_generation_prompt': True,
                                                 'chat_template_kwargs': {'enable_thinking': False}})['prompt']
        return len(self.request('/tokenize', {'content': prompt, 'add_special': True})['tokens'])

    def answer(self, messages, output_tokens, seed):
        self.identity()
        expected = self.count(messages)
        if expected + output_tokens > self.context_tokens:
            raise ValueError('prompt plus output exceed context')
        result = self.completion_stream({'model': self.model, 'messages': messages,
                              'max_tokens': output_tokens, 'temperature': 0, 'seed': seed,
                              'stream': True, 'stream_options': {'include_usage': True}, 'cache_prompt': False,
                              'chat_template_kwargs': {'enable_thinking': False},
                              'response_format': {'type': 'json_object'}})
        if result.get('model') != self.model:
            raise ValueError('response model identity changed')
        if result.get('usage', {}).get('prompt_tokens') != expected:
            raise ValueError('token usage mismatch or silent truncation')
        choice = result['choices'][0]
        if choice.get('finish_reason') != 'stop':
            raise ValueError('generation incomplete')
        answer = json.loads(choice['message']['content'])
        if not isinstance(answer, dict):
            raise ValueError('answer object required')
        return answer, result

    def completion_stream(self, payload):
        """Actual provider bytes/deltas/termination, bounded idle and total elapsed."""
        start = time.monotonic()
        self.stream_receipt = {'first_provider_byte_seconds': None, 'first_model_delta_seconds': None,
                               'timeout_seconds': self.timeout, 'complete': False}
        self.last_response = None
        request = urllib.request.Request(self.endpoint + '/v1/chat/completions',
                   data=json.dumps(payload, ensure_ascii=False).encode('utf-8'),
                   headers={'Content-Type': 'application/json', 'Accept': 'text/event-stream'})
        text, usage, finish, done = [], None, None, False
        buffer = b''
        try:
            with self.opener.open(request, timeout=self.timeout) as response:
                if 'text/event-stream' not in response.headers.get('Content-Type', ''):
                    raise ValueError('provider did not return actual SSE stream')
                while True:
                    if response.fp is None:
                        break
                    remaining = self.timeout - (time.monotonic() - start)
                    if remaining <= 0:
                        raise TimeoutError('total inference timeout')
                    response.fp.raw._sock.settimeout(remaining)
                    chunk = response.read1(4096)
                    elapsed = time.monotonic() - start
                    if elapsed > self.timeout:
                        raise TimeoutError('total inference timeout')
                    if not chunk:
                        break
                    if self.stream_receipt['first_provider_byte_seconds'] is None:
                        self.stream_receipt['first_provider_byte_seconds'] = elapsed
                    buffer += chunk
                    while b'\n' in buffer:
                        line, buffer = buffer.split(b'\n', 1)
                        line = line.rstrip(b'\r')
                        if not line.startswith(b'data:'):
                            continue
                        value = line[5:].strip()
                        if value == b'[DONE]':
                            done = True
                            continue
                        frame = json.loads(value)
                        self.last_response = frame
                        if frame.get('error') or frame.get('model') != self.model:
                            raise ValueError('provider streaming model/error changed')
                        if frame.get('usage') is not None:
                            usage = frame['usage']
                        for choice in frame.get('choices', []):
                            content = choice.get('delta', {}).get('content')
                            if content:
                                if self.stream_receipt['first_model_delta_seconds'] is None:
                                    self.stream_receipt['first_model_delta_seconds'] = elapsed
                                text.append(content)
                            if choice.get('finish_reason') is not None:
                                finish = choice['finish_reason']
                if buffer.strip() or not done or finish != 'stop' or usage is None:
                    raise ValueError('provider stream incomplete or missing usage')
            self.stream_receipt.update(total_seconds=time.monotonic() - start, complete=True)
            return {'model': self.model, 'usage': usage,
                    'choices': [{'finish_reason': finish, 'message': {'content': ''.join(text)}}],
                    'stream_receipt': self.stream_receipt}
        except urllib.error.HTTPError as error:
            self.last_error = {'http_status': error.code, 'body_local': error.read().decode('utf-8', errors='replace')}
            raise
        finally:
            self.stream_receipt['total_seconds'] = time.monotonic() - start


def probe(endpoints):
    report = []
    for endpoint in endpoints:
        start = time.monotonic()
        try:
            client = LocalModel(endpoint, '', 0, timeout=2)
            response = client.request('/v1/models')
            report.append({'endpoint': endpoint, 'status': 'MODEL_LIST_ONLY',
                           'models': [m['id'] for m in response.get('data', [])],
                           'elapsed_seconds': round(time.monotonic() - start, 3)})
        except Exception as error:
            report.append({'endpoint': endpoint, 'status': 'UNAVAILABLE', 'error_type': type(error).__name__,
                           'elapsed_seconds': round(time.monotonic() - start, 3)})
    return report


def mcp_call(command, calls, timeout=120):
    requests = [{'jsonrpc': '2.0', 'id': 1, 'method': 'initialize', 'params': {
        'protocolVersion': '2025-11-25', 'capabilities': {}, 'clientInfo': {'name': 'v3-w2', 'version': '1'}}}]
    requests += [{'jsonrpc': '2.0', 'id': i+2, 'method': 'tools/call',
                  'params': {'name': name, 'arguments': args}} for i, (name, args) in enumerate(calls)]
    try:
        run = subprocess.run(command, input='\n'.join(json.dumps(r) for r in requests) + '\n',
                             capture_output=True, text=True, encoding='utf-8', timeout=timeout, shell=False)
    except (subprocess.TimeoutExpired, OSError) as error:
        def local_text(value):
            return value.decode('utf-8', errors='replace') if isinstance(value, bytes) else value
        return None, {'command': command, 'status': 'FAILED', 'error_type': type(error).__name__,
                      'stdout_local': local_text(getattr(error, 'stdout', None)),
                      'stderr_local': local_text(getattr(error, 'stderr', None)),
                      'delivery': 'unknown_no_auto_retry'}
    receipt = {'command': command, 'returncode': run.returncode, 'stdout_local': run.stdout,
               'stderr_local': run.stderr}
    try:
        responses = [json.loads(line) for line in run.stdout.splitlines() if line.strip()]
        by_id = {r['id']: r for r in responses}
        if run.returncode or len(by_id) != len(requests) or 'error' in by_id[1]:
            raise ValueError('MCP startup/response failed')
        values = []
        for i in range(len(calls)):
            response = by_id[i+2]
            if 'error' in response or response['result'].get('isError'):
                raise ValueError('MCP tool failed')
            result = response['result']
            value = result.get('structuredContent')
            if value is None:
                value = json.loads(''.join(c['text'] for c in result['content'] if c['type'] == 'text'))
            values.append(value)
        receipt['status'] = 'OK'
        return values, receipt
    except Exception as error:
        receipt.update(status='FAILED', error_type=type(error).__name__)
        return None, receipt


def a_memory(command, topic):
    names = ['astrocyte.get_attention_state', 'astrocyte.explain_focus', 'astrocyte.get_experiences']
    values, receipt = mcp_call(command, [(names[0], {}), (names[1], {'topic': topic}),
                                       (names[2], {'pattern': topic})])
    return (dict(zip(names, values)) if values is not None else None), receipt
