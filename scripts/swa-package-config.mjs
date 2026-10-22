import { execFileSync } from 'node:child_process'
import { copyFile, lstat, mkdir, readFile, readdir, realpath } from 'node:fs/promises'
import { isAbsolute, join, relative } from 'node:path'
import { parseArgs } from 'node:util'

export function packageOptions(args = process.argv.slice(2), environment = process.env) {
  const { values } = parseArgs({ args, options: {
    output: { type: 'string' }, edition: { type: 'string', default: 'cli' }, python: { type: 'string' },
    'node-inputs': { type: 'string' }, 'native-source': { type: 'string' },
    'native-work': { type: 'string' }, 'native-evidence': { type: 'string' }
  }, allowPositionals: false })
  if (!['cli', 'combined'].includes(values.edition)) throw new Error('The edition must be cli or combined')
  const python = values.python ?? environment.SWA_BUILD_PYTHON
  if (!python || !isAbsolute(python)) throw new Error('Provide an absolute --python or SWA_BUILD_PYTHON executable')
  if (!values.output || !isAbsolute(values.output)) throw new Error('An explicit absolute --output inside dist is required')
  const reuse = ['native-source', 'native-work', 'native-evidence'].filter(key => values[key])
  if (reuse.length !== 0 && reuse.length !== 3) throw new Error('Native reuse requires the explicit candidate, original work and successful evidence together')
  return { ...values, python }
}

export async function buildPython(path) {
  const executable = await realpath(path)
  if (!(await lstat(executable)).isFile()) throw new Error('The selected Python executable must be an actual file')
  const info = JSON.parse(execFileSync(executable, ['-c', 'import json,sys; print(json.dumps({"executable":sys.executable,"version":list(sys.version_info[:3])}))'], {
    encoding: 'utf8', windowsHide: true, timeout: 60_000
  }))
  if (!info || typeof info.executable !== 'string' || !isAbsolute(info.executable) || await realpath(info.executable) !== executable) throw new Error('The actual Python executable differs from the selected native build input')
  const version = info.version
  if (!Array.isArray(version) || version.length !== 3 || !version.every(Number.isSafeInteger) || version.some(value => value < 0) || version[0] !== 3 || version[1] < 6) throw new Error('The selected interpreter must be Python 3.6 or newer within Python 3')
  return { executable, version: version.join('.') }
}

export async function checkWebAssets(directory) {
  const root = await realpath(directory)
  async function file(path) {
    if (typeof path !== 'string' || !path || isAbsolute(path) || path.includes('\\')) throw new Error('Invalid web asset path')
    const actual = await realpath(join(root, path)), within = relative(root, actual)
    if (!within || within === '..' || within.startsWith('../') || within.startsWith('..\\') || isAbsolute(within) || !(await lstat(actual)).isFile()) throw new Error('Web asset escaped its actual root or is not a file: ' + path)
    return actual
  }
  await file('index.html')
  const manifest = JSON.parse(await readFile(await file('.vite/manifest.json'), 'utf8'))
  if (!manifest || typeof manifest !== 'object' || Array.isArray(manifest) || manifest['index.html']?.isEntry !== true) throw new Error('The Vite manifest must identify index.html as its entry')
  let referencedFiles = 0
  for (const [key, entry] of Object.entries(manifest)) {
    if (!key || !entry || typeof entry !== 'object' || Array.isArray(entry)) throw new Error('Invalid Vite manifest entry')
    await file(entry.file)
    referencedFiles++
    for (const field of ['css', 'assets']) {
      if (entry[field] === undefined) continue
      if (!Array.isArray(entry[field])) throw new Error('Invalid Vite manifest asset list')
      for (const path of entry[field]) { await file(path); referencedFiles++ }
    }
    for (const field of ['imports', 'dynamicImports']) {
      if (entry[field] === undefined) continue
      if (!Array.isArray(entry[field]) || entry[field].some(key => typeof key !== 'string' || !Object.hasOwn(manifest, key))) throw new Error('Missing Vite manifest import')
    }
  }
  return { entries: Object.keys(manifest).length, referenced_files: referencedFiles }
}

export async function copyWebLicenses(webRoot, candidate, pnpm) {
  if (!isAbsolute(pnpm)) throw new Error('The pinned pnpm JavaScript entry must be absolute')
  const pkg = JSON.parse(await readFile(join(webRoot, 'package.json'), 'utf8'))
  const tree = JSON.parse(execFileSync(process.execPath, [pnpm, '--filter', pkg.name, 'list', '--prod', '--depth', 'Infinity', '--json'], { cwd: webRoot, encoding: 'utf8', windowsHide: true, timeout: 60_000 }))
  if (!Array.isArray(tree) || tree.length !== 1 || tree[0].name !== pkg.name || await realpath(tree[0].path) !== await realpath(webRoot)) throw new Error('The production dependency tree does not identify the selected web package')
  for (const [name, expected] of Object.entries(pkg.dependencies)) if (tree[0].dependencies?.[name]?.version !== expected) throw new Error('The web dependency tree differs from the pinned dependency: ' + name)
  const seen = new Set()
  async function copy(name, dependency) {
    const source = await realpath(dependency.path), expected = dependency.version
    const installed = JSON.parse(await readFile(join(source, 'package.json'), 'utf8'))
    if (installed.name !== name || installed.version !== expected) throw new Error('Web license source differs from the pinned dependency: ' + name)
    const identity = name + '@' + expected
    if (!seen.has(identity)) {
      seen.add(identity)
      const files = (await readdir(source)).filter(file => /^(?:licen[cs]e|copying|notice)(?:[.-].*)?$/i.test(file))
      if (!files.some(file => /^(?:licen[cs]e|copying)(?:[.-].*)?$/i.test(file))) throw new Error('Missing web dependency license: ' + name)
      const destination = join(candidate, 'third-party-licenses', 'web', name.replaceAll('/', '__') + '@' + expected)
      await mkdir(destination, { recursive: true })
      for (const file of files) {
        if (!(await lstat(join(source, file))).isFile()) throw new Error('Web license must be an ordinary file: ' + name)
        await copyFile(join(source, file), join(destination, file))
      }
    }
    for (const [name, child] of Object.entries({ ...dependency.dependencies, ...dependency.optionalDependencies })) await copy(name, child)
  }
  for (const [name, dependency] of Object.entries({ ...tree[0].dependencies, ...tree[0].optionalDependencies })) await copy(name, dependency)
  return { packages: seen.size }
}
