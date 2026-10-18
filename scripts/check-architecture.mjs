import { spawnSync } from 'node:child_process';
import { findGo, root } from './process.mjs';

const go = await findGo();
const moduleResult = spawnSync(go, ['list', '-mod=readonly', '-m', '-json'], { cwd: root, encoding: 'utf8', windowsHide: true });
if (moduleResult.error) throw moduleResult.error;
if (moduleResult.status !== 0) throw new Error(moduleResult.stderr);
const moduleName = JSON.parse(moduleResult.stdout).Path;
if (typeof moduleName !== 'string' || !moduleName) throw new Error('Go module identity is unavailable');
const result = spawnSync(go, ['list', '-mod=readonly', '-f', '{{.ImportPath}}|{{join .Imports ","}}', './...'], { cwd: root, encoding: 'utf8', windowsHide: true });
if (result.error) throw result.error;
if (result.status !== 0) throw new Error(result.stderr);
const graph = new Map(result.stdout.trim().split(/\r?\n/).filter(Boolean).map(line => {
  const [name, imports] = line.split('|');
  return [name, imports ? imports.split(',') : []];
}));
if (!graph.size) throw new Error('No Go packages; backend owner work is pending');

// net/url 只解析值对象，其余网络 I/O 不得进入 Domain/app。
const io = /^(database(?:\/|$)|net(?:\/(?!url$)|$)|os(?:\/|$)|io(?:\/|$)|bufio$|syscall$|log(?:\/|$)|modernc\.org\/|golang\.org\/x\/sys)/;
const issues = [];
const checkedLayers = [];
const checkedJudgment = [];
for (const [name, imports] of graph) {
  const relative = name.startsWith(moduleName + '/') ? name.slice(moduleName.length + 1) : name;
  const layer = relative.match(/^internal\/(attention|workspace|swarm|workbench)\/(domain|app)(?:\/|$)/);
  const judgment = /^internal\/judgment(?:\/|$)/.test(relative);
  if (layer) checkedLayers.push(relative);
  if (judgment) checkedJudgment.push(relative);
  for (const dependency of imports) {
    const internal = dependency.startsWith(moduleName + '/') ? dependency.slice(moduleName.length + 1) : '';
    if (layer) {
      const [, context, level] = layer;
      const other = internal.match(/^internal\/(attention|workspace|swarm|workbench)\//);
      if (other && other[1] !== context) issues.push(`${relative} imports another context: ${dependency}`);
      if (io.test(dependency)) issues.push(`${relative} imports I/O: ${dependency}`);
      if (internal.startsWith('internal/adapters/')) issues.push(`${relative} imports an adapter: ${dependency}`);
      if (level === 'domain') {
        if (internal && !internal.startsWith(`internal/${context}/domain`)) issues.push(`${relative} imports outside its Domain: ${dependency}`);
        if (!internal && dependency.includes('.')) issues.push(`${relative} imports an unapproved external dependency: ${dependency}`);
      }
    }
    if (judgment) {
      if (io.test(dependency)) issues.push(`${relative} imports I/O: ${dependency}`);
      // 共享判断契约复用现有 Workspace 值类型和错误类型，禁止依赖应用服务或 adapter。
      if (internal && internal !== 'internal/apierrors' && internal !== 'internal/workspace/domain' && !/^internal\/judgment(?:\/|$)/.test(internal)) {
        issues.push(`${relative} imports outside the shared judgment contract: ${dependency}`);
      }
    }
    if (/^internal\/adapters\/(httpapi|mcp)(?:\/|$)/.test(relative) &&
        (dependency === 'database/sql' || dependency.startsWith('modernc.org/') || /^internal\/adapters\/sqlite(?:\/|$)/.test(internal))) {
      issues.push(`${relative} transport imports persistence: ${dependency}`);
    }
  }
}
if (issues.length) throw new Error(`Architecture violations:\n${issues.join('\n')}`);
console.log(`Go import graph: ${graph.size} packages satisfy Domain/app/context/transport boundaries`);
console.log(`Checked Domain/app packages: ${checkedLayers.join(', ')}`);
console.log(`Checked shared judgment packages: ${checkedJudgment.join(', ')}`);
