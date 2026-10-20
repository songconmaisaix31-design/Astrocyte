import { open, realpath } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { join } from 'node:path'
import { Readable } from 'node:stream'
import { pathToFileURL } from 'node:url'

export async function readJSONL(candidate, file, maximum = 64 * 1024 * 1024) {
  const requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
  const { default: parser } = await import(pathToFileURL(requireRuntime.resolve('stream-json/jsonl/parser.js')).href)
  const handle = await open(file, 'r')
  try {
    const before = await handle.stat({ bigint: true })
    if (!before.isFile() || before.size > BigInt(maximum)) throw new Error('Actual JSONL source exceeds its explicit reader bound')
    const bytes = await handle.readFile()
    new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    if (bytes.length && bytes.at(-1) !== 10) throw new Error('Actual JSONL source is not at a complete record boundary')
    const rows = []
    for await (const row of Readable.from([bytes]).pipe(parser.asStream())) rows.push(row.value)
    const after = await handle.stat({ bigint: true })
    if (before.dev !== after.dev || before.ino !== after.ino || before.birthtimeNs !== after.birthtimeNs || before.size !== after.size || before.mtimeNs !== after.mtimeNs || (await realpath(file)).toLowerCase() !== file.toLowerCase()) throw new Error('Actual JSONL source changed during observation')
    return rows
  } finally { await handle.close() }
}
