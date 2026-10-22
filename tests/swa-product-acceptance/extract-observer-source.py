"""只物化已核原 Git blob 的普通归档条目，不执行项目代码。"""

import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import subprocess
import sys
import zipfile


def main():
    arguments = sys.argv[1:]
    profile = None
    if len(arguments) == 6 and arguments[-2:] == ["--profile", "stepfun-pi-n3"]:
        profile = arguments[-1]
        arguments = arguments[:4]
    archive_name, repository_name, revision, target_name = arguments
    if not re.fullmatch(r"[0-9a-f]{40}", revision):
        raise ValueError("source_revision_invalid")
    repository = Path(repository_name)
    target = Path(target_name)
    if not repository.is_absolute() or not target.is_absolute():
        raise ValueError("source_absolute_paths_required")
    if target.exists() or target.is_symlink():
        raise ValueError("source_target_already_exists")
    allowed = [
        "third_party/orca-v1.4.230/src/shared",
        "third_party/orca-v1.4.230/src/main/daemon/swa-linux-process-identity.ts",
        "third_party/orca-v1.4.230/src/main/agent-hooks/managed-hook-owner-identity.ts",
        "third_party/orca-v1.4.230/src/main/pty-process-table-parser.ts",
        "third_party/orca-v1.4.230/package.json",
        "third_party/orca-v1.4.230/tsconfig.json",
        "third_party/orca-v1.4.230/config/tsconfig.node.json",
    ]
    if profile == "stepfun-pi-n3":
        allowed.extend([
            "third_party/orca-v1.4.230/src/main/pi/rpc-protocol.ts",
            "scripts/build-swa-runtime.mjs",
            "third_party/orca-v1.4.230/src/main/native-chat/agent-session-timeline/provider-timeline-identity.ts",
            "third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-payload-bounds.ts",
        ])
    git_environment = dict(os.environ, GIT_OPTIONAL_LOCKS="0", GIT_PAGER="cat")
    result = subprocess.run(
        ["git", "--no-pager", "-C", str(repository), "ls-tree", "-rz", "--full-tree", revision, "--", *allowed],
        env=git_environment, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
        timeout=30,
    )
    expected = {}
    for row in result.stdout.split(b"\0"):
        if not row:
            continue
        metadata, encoded_path = row.split(b"\t", 1)
        mode, kind, blob = metadata.decode("ascii").split(" ")
        path = encoded_path.decode("utf-8", errors="strict")
        if mode not in ("100644", "100755") or kind != "blob" or path in expected:
            raise ValueError("source_git_entry_not_ordinary")
        expected[path] = blob
    if not expected or len(expected) > 4096:
        raise ValueError("source_inventory_out_of_bounds")
    if Path(archive_name).stat().st_size > 64 * 1024 * 1024:
        raise ValueError("source_archive_exceeds_bound")
    verified = {}
    directories = set()
    total = 0
    with zipfile.ZipFile(archive_name) as archive:
        if archive.comment != revision.encode("ascii"):
            raise ValueError("source_archive_revision_comment_differs")
        if len(archive.infolist()) > 8192:
            raise ValueError("source_archive_inventory_exceeds_bound")
        for item in archive.infolist():
            name = item.filename
            normalized = PurePosixPath(name)
            if "\\" in name or normalized.is_absolute() or any(part in ("", ".", "..") for part in name.rstrip("/").split("/")):
                raise ValueError("source_archive_path_invalid")
            mode = item.external_attr >> 16
            kind = stat.S_IFMT(mode)
            if item.is_dir():
                if kind not in (0, stat.S_IFDIR) or name in directories:
                    raise ValueError("source_archive_directory_invalid")
                if not any(path.startswith(name) for path in expected):
                    raise ValueError("source_archive_directory_outside_inventory")
                directories.add(name)
                continue
            if kind not in (0, stat.S_IFREG) or name not in expected or name in verified:
                raise ValueError("source_archive_file_outside_inventory")
            if item.file_size > 8 * 1024 * 1024:
                raise ValueError("source_file_exceeds_bound")
            total += item.file_size
            if total > 128 * 1024 * 1024:
                raise ValueError("source_total_exceeds_bound")
            data = archive.read(item)
            blob = hashlib.sha1(b"blob " + str(len(data)).encode("ascii") + b"\0" + data).hexdigest()
            if blob != expected[name]:
                raise ValueError("source_archive_differs_from_git_blob")
            verified[name] = data
    if verified.keys() != expected.keys():
        raise ValueError("source_archive_inventory_incomplete")
    target.mkdir(mode=0o700)
    summaries = []
    for name, data in verified.items():
        destination = target / name
        destination.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
        with destination.open("xb") as stream:
            stream.write(data)
        destination.chmod(0o600)
        summaries.append({"path": name, "blob": expected[name], "sha256": hashlib.sha256(data).hexdigest(), "bytes": len(data)})
    print(json.dumps({"revision": revision, "files": len(summaries), "bytes": total, "entries": summaries}))


try:
    main()
except Exception:
    # Git 环境、未知路径和原始命令错误不进入诊断输出。
    print("source_archive_verification_failed", file=sys.stderr)
    sys.exit(1)
