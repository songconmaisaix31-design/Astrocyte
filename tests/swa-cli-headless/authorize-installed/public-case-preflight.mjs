import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { lstat, readFile, realpath } from 'node:fs/promises'
import { dirname, isAbsolute, join, resolve } from 'node:path'
import { pathToFileURL } from 'node:url'

const [definition, project, requestedWorkers = '3'] = process.argv.slice(2)
assert.ok(process.argv.length >= 4 && process.argv.length <= 5)
assert.ok(['3', '4', '6'].includes(requestedWorkers))
const workers = Number(requestedWorkers)
assert.ok(isAbsolute(definition) && isAbsolute(project))
async function ordinary(path, directory = false) {
  assert.equal(await realpath(path), resolve(path))
  for (let current = path; ; current = dirname(current)) {
    const info = await lstat(current)
    assert.equal(info.isSymbolicLink(), false)
    assert.equal(current === path && !directory ? info.isFile() : info.isDirectory(), true)
    if (dirname(current) === current) return path
  }
}
await ordinary(project, true)
await ordinary(definition)
const { publicCase, publicFiles, publicRevision, publicSource } = await import(pathToFileURL(definition).href)
assert.equal(publicSource, 'https://github.com/songconmaisaix31-design/astrocyte-s3-public-stats.git')
assert.equal(publicRevision, 'da3bdddb085808863090a05f60cc1c64415039cf')
assert.deepEqual(publicFiles, ['README.md', 'go.mod', 'stats.go', 'stats_test.go'])
const git = args => execFileSync('git', ['-c', 'core.fsmonitor=false', '-c', `core.hooksPath=${join(dirname(project), 'empty-hooks')}`, '-C', project, ...args], { encoding: 'utf8', timeout: 10_000, maxBuffer: 1024 * 1024 })
assert.equal(git(['rev-parse', 'HEAD']).trim(), publicRevision)
assert.equal(git(['remote', 'get-url', 'origin']).trim(), publicSource)
const originals = []
const hashes = {}
const read = async path => {
  const bytes = await readFile(await ordinary(path))
  hashes[path.slice(project.length + 1)] = createHash('sha256').update(bytes).digest('hex')
  return new TextDecoder('utf-8', { fatal: true }).decode(bytes)
}
for (const path of publicFiles) {
  const text = await read(join(project, path))
  assert.equal(text, git(['show', `${publicRevision}:${path}`]))
  originals.push({ path, source_revision: publicRevision, text })
}
const directory = join(project, '.verification/swa-product-input')
const prepared = publicCase(workers, originals)
assert.deepEqual(JSON.parse(await read(join(directory, 'task.json'))), prepared.spec)
assert.deepEqual(JSON.parse(await read(join(directory, 'source.json'))), { project_root: project, source: publicSource, revision: publicRevision, originals })
assert.deepEqual(JSON.parse(await read(join(directory, 'inputs.json'))), prepared.inputs.map(({ name, file }) => ({ name, input_file: join(directory, file) })))
for (const input of prepared.inputs) assert.equal(await read(join(directory, input.file)), input.body)
assert.equal(await read(join(directory, 'followup.txt')), prepared.followup)
process.stdout.write(JSON.stringify({ status: 'VERIFIED_FIXED_PUBLIC_INPUT', source: publicSource, revision: publicRevision, workers, files: hashes, http_calls: 0, model_calls: 0, project_registration: 'NOT_RUN' }) + '\n')
