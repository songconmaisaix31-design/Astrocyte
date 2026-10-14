# 只下载固定版本的公开模型和运行环境。
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
from pathlib import Path
import shutil
import time
import urllib.request
import zipfile

from local_store import PRIVATE_ROOT, write_new

REVISION = '938a7432affaec9157f883a87164e2646ae17555'
MODEL = 'Qwen3-32B-Q4_K_M.gguf'
MODEL_BYTES = 19762149024
MODEL_SHA256 = 'efd971561896866f0e910cce52761ca77b1b138090c7f15fe284676d57d1f689'
TAG = 'b11574'
ASSET = 'llama-b11574-bin-win-vulkan-x64.zip'


def download(url, path):
    start = time.monotonic()
    count, digest = 0, hashlib.sha256()
    with urllib.request.urlopen(url, timeout=60) as response, path.open('xb') as target:
        while block := response.read(4 * 1024 * 1024):
            target.write(block)
            digest.update(block)
            count += len(block)
    return {'name': path.name, 'bytes': count, 'sha256': digest.hexdigest(),
            'seconds': round(time.monotonic() - start, 3)}


def main():
    target = PRIVATE_ROOT / 'candidate-qwen32b'
    target.mkdir(parents=True, exist_ok=False)
    if shutil.disk_usage(target).free < MODEL_BYTES + 2 * 1024**3:
        raise ValueError('insufficient candidate download disk')
    release = json.load(urllib.request.urlopen(f'https://api.github.com/repos/ggml-org/llama.cpp/releases/tags/{TAG}', timeout=30))
    asset = next(a for a in release['assets'] if a['name'] == ASSET)
    write_new(target / 'public-download-source.json', {'model': 'Qwen/Qwen3-32B-GGUF', 'revision': REVISION,
              'quantization': 'Q4_K_M', 'model_bytes': MODEL_BYTES, 'model_sha256': MODEL_SHA256,
              'runtime_tag': TAG, 'runtime_asset': asset['name'], 'runtime_asset_digest': asset.get('digest'),
              'private_model_calls': 0, 'status': 'CANDIDATE_ONLY'}, PRIVATE_ROOT)
    try:
        with ThreadPoolExecutor(max_workers=2) as pool:
            model = pool.submit(download, f'https://huggingface.co/Qwen/Qwen3-32B-GGUF/resolve/{REVISION}/{MODEL}', target / MODEL)
            runtime = pool.submit(download, asset['browser_download_url'], target / ASSET)
            runtime_result = runtime.result()
            if asset.get('digest') and asset['digest'] != 'sha256:' + runtime_result['sha256']:
                raise ValueError('official runtime digest mismatch')
            with zipfile.ZipFile(target / ASSET) as archive:
                for member in archive.infolist():
                    resolved = (target / 'runtime' / member.filename).absolute()
                    if not resolved.is_relative_to((target / 'runtime').absolute()):
                        raise ValueError('runtime ZIP path rejected')
                archive.extractall(target / 'runtime')
            print(json.dumps({'runtime_download': runtime_result}), flush=True)
            model_result = model.result()
            if model_result['bytes'] != MODEL_BYTES or model_result['sha256'] != MODEL_SHA256:
                raise ValueError('official model digest/size mismatch')
        write_new(target / 'download-result.json', {'status': 'DOWNLOADED_ONLY', 'model': model_result,
                                                  'runtime': runtime_result}, PRIVATE_ROOT)
        print(json.dumps({'model_download': model_result}), flush=True)
    except Exception as error:
        write_new(target / 'download-first-error.json', {'status': 'FAILED', 'error_type': type(error).__name__,
                                                        'error_local': str(error)}, PRIVATE_ROOT)
        raise


if __name__ == '__main__':
    main()
