import { execFileSync } from 'node:child_process'
import { lstat, mkdir, readFile, realpath, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'

const source = 'https://github.com/songconmaisaix31-design/astrocyte-s3-public-stats.git'
const revision = 'da3bdddb085808863090a05f60cc1c64415039cf'
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
  const files = ['README.md', 'go.mod', 'stats.go', 'stats_test.go']
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
  const context = JSON.stringify({ source, revision, originals })
  const roles = [
    '核对四份原文的模块组织、现有 Mean API 和真实测试覆盖，按文件引用已有事实与缺少的证据。',
    '审阅 Median 的设计条件，覆盖奇偶长度、负数、重复值、输入顺序保持与现有 Mean 兼容，不把建议写成已有实现。',
    '审阅空输入、NaN、正负 Infinity、有限大中间值溢出等失败与边界条件，区分现有事实、建议和需要实际运行的检查。',
  ]
  const inputs = []
  for (let index = 0; index < workers; index++) {
    const name = `worker-${index + 1}`, file = `${name}.txt`
    const instruction = `你负责本次公开来源审阅的第${index + 1}项。${roles[index % roles.length]}逐条给出推理与来源位置，检查其他可能解释；收到后续输入时在同一会话继续。禁止调用工具、读写文件或访问其他资料；不要声称运行了测试。部分 Worker 会被计划停止，保留真实停止状态。\n`
    await write(file, instruction + context + '\n')
    inputs.push({ name, input_file: join(inputDirectory, file) })
  }
  const spec = {
    goal: '审阅固定公开 stats 项目，取得多个真实 Worker 的来源核对、设计与失败路径分析；不修改程序。',
    source_repo: source, source_revision: revision, max_workers: workers,
    success_criteria: [
      `${workers}个实际产品 Worker 有独立 Dispatch/session 且真实模型执行区间重叠；不以进程存在或排队代替。`,
      `指定停止一项时其仍真实运行，其余至少${workers - 1}项保留原身份继续；一项在同会话处理后续输入。`,
      '现有事实与设计建议分别表述并引用四份固定原文；未知、未运行测试和被停止项保持原状态。',
      '取得实际完整来源成果，确认全部所属资源退出并保护非所属进程、原账号和数据。',
    ],
    context: { references: [], files },
  }
  await write('task.json', JSON.stringify(spec, null, 2) + '\n')
  await write('followup.txt', '请在原会话继续，复核你的首次回答中各项来源与边界条件，纠正无法由固定原文证明的陈述；结合本会话已有真实消息形成完整说明。禁止调用工具，不读取其他资料，不声称执行了测试。\n' + context + '\n')
  await write('source.json', JSON.stringify({ project_root: project, source, revision, originals }, null, 2) + '\n')
  await write('inputs.json', JSON.stringify(inputs, null, 2) + '\n')
  process.stdout.write(JSON.stringify({ project_root: project, source, revision, workers, input_directory: inputDirectory, spec_file: join(inputDirectory, 'task.json'), inputs_file: join(inputDirectory, 'inputs.json'), source_files_verified: files, model_calls: 0, product_execution: 'NOT_RUN', current_authorization: 'REQUIRED' }, null, 2) + '\n')
} catch {
  process.stderr.write('Public case preparation failed; preserve the explicitly selected case directory and inspect Git access, source identity and actual paths. No project scripts or model calls were executed.\n')
  process.exitCode = 1
}
