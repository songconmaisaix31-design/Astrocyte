import { execFileSync } from 'node:child_process'
import { lstat, mkdir, readFile, realpath, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { publicCase, publicFiles as files, publicRevision as revision, publicSource as source } from './public-case-definition.mjs'

const repository = await realpath(fileURLToPath(new URL('../../', import.meta.url)))
let values, workers
try {
  values = parseArgs({ options: { root: { type: 'string' }, workers: { type: 'string', default: '3' } }, allowPositionals: false }).values
  if (!values.root || !isAbsolute(values.root)) throw new Error('Select an explicit absolute --root that does not exist')
  if (!/^[1-9]\d*$/.test(values.workers) || !Number.isSafeInteger(Number(values.workers)) || Number(values.workers) < 3 || Number(values.workers) > 100) throw new Error('--workers must be within the current complete-inventory observation bound')
  workers = Number(values.workers)
} catch {
  process.stderr.write('Invalid arguments: use --root ABSOLUTE_NEW_DIRECTORY and optional --workers INTEGER_3_TO_100; this helper follows the current complete-inventory observation bound, not a product quota limit\n')
  process.exit(2)
}

const contained = (root, path) => {
  const within = relative(root, path)
  return !within || within !== '..' && !within.startsWith('../') && !within.startsWith('..\\') && !isAbsolute(within)
}

async function ordinaryDirectory(path) {
  for (let current = path; ; current = dirname(current)) {
    const info = await lstat(current)
    if (!info.isDirectory() || info.isSymbolicLink()) throw new Error('The case parent must contain only actual directories')
    if (dirname(current) === current) break
  }
  const actual = await realpath(path)
  if ((process.platform === 'win32' ? actual.toLowerCase() !== path.toLowerCase() : actual !== path)) throw new Error('The case parent has a redirected canonical location')
  return actual
}

try {
  const root = resolve(values.root)
  await ordinaryDirectory(dirname(root))
  if (contained(repository, root) || contained(root, repository)) throw new Error('The public case must stay separate from the development checkout')
  await mkdir(root, { mode: 0o700 })
  const hooks = join(root, 'empty-hooks')
  await mkdir(hooks, { mode: 0o700 })
  const project = join(root, 'public-project')
  // 保留平台 HTTPS Git 代理；仅关闭本项 checkout 的 hook、自动换行和子模块执行。
  const git = args => execFileSync('git', ['-c', `core.hooksPath=${hooks}`, '-c', 'core.autocrlf=false', '-c', 'core.fsmonitor=false', '-c', 'submodule.recurse=false', ...args], {
    encoding: 'utf8', env: { ...process.env, GIT_TERMINAL_PROMPT: '0' }, timeout: 120_000, maxBuffer: 4 * 1024 * 1024,
  })
  git(['clone', '--no-checkout', '--no-recurse-submodules', source, project])
  git(['-C', project, 'checkout', '--detach', revision])
  if (git(['-C', project, 'rev-parse', 'HEAD']).trim() !== revision || git(['-C', project, 'remote', 'get-url', 'origin']).trim() !== source) throw new Error('The actual public source identity differs from the fixed input')
  const originals = []
  for (const path of files) {
    const actual = join(project, path)
    const info = await lstat(actual)
    if (!info.isFile() || info.isSymbolicLink() || !contained(project, await realpath(actual))) throw new Error('The selected public original is not an ordinary source file')
    const text = git(['-C', project, 'show', `${revision}:${path}`])
    if (await readFile(actual, 'utf8') !== text) throw new Error('The checked-out original differs from its fixed Git revision')
    originals.push({ path, source_revision: revision, text })
  }
  const inputDirectory = join(project, '.verification', 'swa-product-input')
  await mkdir(inputDirectory, { recursive: true, mode: 0o700 })
  const write = (name, value) => writeFile(join(inputDirectory, name), value, { flag: 'wx', mode: 0o600 })
  const prepared = publicCase(workers, originals)
  const inputs = []
  for (const { name, file, body } of prepared.inputs) {
    await write(file, body)
    inputs.push({ name, input_file: join(inputDirectory, file) })
  }
  await write('task.json', JSON.stringify(prepared.spec, null, 2) + '\n')
  await write('followup.txt', prepared.followup)
  await write('source.json', JSON.stringify({ project_root: project, source, revision, originals }, null, 2) + '\n')
  await write('inputs.json', JSON.stringify(inputs, null, 2) + '\n')
  process.stdout.write(JSON.stringify({ project_root: project, source, revision, workers, input_directory: inputDirectory, spec_file: join(inputDirectory, 'task.json'), inputs_file: join(inputDirectory, 'inputs.json'), source_files_verified: files, model_calls: 0, product_execution: 'NOT_RUN', current_authorization: 'REQUIRED' }, null, 2) + '\n')
} catch {
  process.stderr.write('Public case preparation failed; preserve the explicitly selected case directory and inspect Git access, source identity and actual paths. No project scripts or model calls were executed.\n')
  process.exitCode = 1
}
