/// <reference types="node" />
import { isAbsolute, relative, resolve, sep } from 'node:path';

export function relayEvidenceDirectory() {
  const input = process.env.ASTROCYTE_TEST_RELAY_EVIDENCE_DIR;
  if (!input) throw new Error('需要本次窗口独立的 ASTROCYTE_TEST_RELAY_EVIDENCE_DIR');
  const root = resolve(import.meta.dirname, '../../.verification');
  const directory = resolve(input);
  const path = relative(root, directory);
  if (!path || isAbsolute(path) || path === '..' || path.startsWith(`..${sep}`)) {
    throw new Error('本次窗口证据目录必须位于当前工作树 .verification 内');
  }
  return directory;
}
