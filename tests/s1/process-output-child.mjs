import { readFile } from 'node:fs/promises';

// 仅输出当前生产源码的明确重复字节，用于核对实际进程的双通道完整性。
const content = await readFile(new URL('../../scripts/process.mjs', import.meta.url));
for (let index = 0; index < 128; index++) {
  process.stdout.write(content);
  process.stderr.write(content);
}
