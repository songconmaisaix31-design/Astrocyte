# 使用明确的公开运行参数启动已下载的本地候选模型。
import argparse
import json
from pathlib import Path
import socket
import subprocess

from download_candidate import MODEL
from local_store import PRIVATE_ROOT, private_path, write_new


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--out', required=True)
    parser.add_argument('--context-tokens', type=int, default=131072)
    parser.add_argument('--gpu-layers', type=int, default=14)
    parser.add_argument('--port', type=int, default=18080)
    a = parser.parse_args()
    base = PRIVATE_ROOT / 'candidate-qwen32b'
    result = json.loads((base / 'download-result.json').read_text(encoding='utf-8'))
    if result['status'] != 'DOWNLOADED_ONLY':
        raise ValueError('pinned verified public download required')
    with socket.socket() as check:
        check.bind(('127.0.0.1', a.port))
    out = private_path(a.out)
    out.mkdir(parents=True, exist_ok=False)
    argv = [str(base / 'runtime/llama-server.exe'), '-m', str(base / MODEL),
            '--alias', 'Qwen3-32B-Q4_K_M', '--host', '127.0.0.1', '--port', str(a.port),
            '--ctx-size', str(a.context_tokens), '--parallel', '1', '--device', 'Vulkan1',
            '--gpu-layers', str(a.gpu_layers), '--fit', 'off', '--flash-attn', 'on',
            '--cache-type-k', 'q4_0', '--cache-type-v', 'q4_0', '--batch-size', '512', '--ubatch-size', '128',
            '--threads', '24', '--threads-batch', '24', '--no-context-shift', '--no-webui',
            '--chat-template-kwargs', '{"enable_thinking":false}']
    if a.context_tokens > 32768:
        argv += ['--rope-scaling', 'yarn', '--rope-scale', str(a.context_tokens / 32768), '--yarn-orig-ctx', '32768']
    with (out / 'server-stdout-local.log').open('xb') as stdout, (out / 'server-stderr-local.log').open('xb') as stderr:
        process = subprocess.Popen(argv, cwd=base / 'runtime', stdout=stdout, stderr=stderr,
                                   creationflags=getattr(subprocess, 'CREATE_NO_WINDOW', 0))
    record = {'pid': process.pid, 'command': argv, 'endpoint': f'http://127.0.0.1:{a.port}',
              'model': 'Qwen3-32B-Q4_K_M', 'context_tokens_configured': a.context_tokens,
              'evidence': 'candidate_process_ONLY_NOT_ACCEPTED_NOT_DAY0'}
    write_new(out / 'process-local.json', record)
    print(json.dumps({'status': 'CANDIDATE_LAUNCHED_NOT_READY', 'pid': process.pid,
                      'context_tokens_configured': a.context_tokens}))


if __name__ == '__main__':
    main()
