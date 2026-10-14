# 只读取两个已授权仓库的指定 HEAD，原始 Git subject 留在本地。
import argparse
import base64
from collections import Counter
from datetime import datetime
import json
import os
from pathlib import Path
import subprocess
import sys

from diagnose_native import END, START, unredirected

ROOTS = (('Morphogenesis', Path('C:/Users/DW/orca/Morphogenesis')), ('Astrocyte', Path('C:/Users/DW/orca/Astrocyte')))


def git(root, *args):
    result = subprocess.run(['git', '-C', str(root), *args], capture_output=True, check=True)
    return result.stdout


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--out', required=True)
    args = parser.parse_args()
    out = Path(args.out).absolute()
    allowed = Path(os.environ['LOCALAPPDATA']) / 'astrocyte-validation/v1-20261012'
    if not out.is_relative_to(allowed) or not unredirected(out):
        raise ValueError('Unauthorized evidence root')
    out.mkdir(parents=True, exist_ok=True)
    all_ids, inventory, records = set(), {}, []
    zone = datetime.fromisoformat('2026-10-12T00:00:00+08:00').tzinfo
    for label, root in ROOTS:
        if not root.is_dir() or not unredirected(root):
            raise ValueError('Authorized repository unavailable')
        branch = git(root, 'symbolic-ref', '--short', 'HEAD').decode().strip()
        head = git(root, 'rev-parse', 'HEAD').decode().strip()
        log = git(root, 'log', head, '--since-as-filter=2026-09-12T00:00:00+08:00', '--until=2026-10-12T00:00:00+08:00', '--format=%H%x00%ct%x00%s')
        ids, days = set(), Counter()
        for line in log.splitlines():
            commit, seconds, subject = line.split(b'\0', 2)
            commit = commit.decode('ascii')
            ts = int(seconds) * 1000
            if not START <= ts < END or commit in ids:
                continue
            ids.add(commit)
            days[datetime.fromtimestamp(ts / 1000, zone).date().isoformat()] += 1
            if commit not in all_ids:
                all_ids.add(commit)
                records.append({'project': str(root), 'git_line_base64': base64.b64encode(line).decode('ascii'), 'record': {'commit': commit, 'ts': ts, 'subject': subject.decode('utf-8')}})
        inventory[label] = {'branch': branch, 'head': head, 'unique_commits': len(ids), 'days': dict(sorted(days.items()))}
    with (out / 'git-log-local.jsonl').open('x', encoding='utf-8') as handle:
        for row in sorted(records, key=lambda r: (r['record']['ts'], r['record']['commit'])):
            handle.write(json.dumps(row, ensure_ascii=False) + '\n')
    inventory['combined_unique_commits'] = len(all_ids)
    with (out / 'git-inventory.json').open('x', encoding='utf-8') as handle:
        json.dump(inventory, handle, indent=2)
    print(json.dumps(inventory, sort_keys=True))


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print('Git inventory failed: ' + type(error).__name__, file=sys.stderr)
        sys.exit(1)
