import { spawnSync } from 'node:child_process';
import { readFile } from 'node:fs/promises';
import { findGo, root } from './process.mjs';

const moduleName = (await readFile(new URL('../go.mod', import.meta.url), 'utf8')).match(/^module\s+(\S+)/m)[1];
const result = spawnSync(await findGo(), ['list', '-mod=readonly', '-f', '{{.ImportPath}}|{{join .Imports ","}}', './...'], { cwd: root, encoding: 'utf8', windowsHide: true });
if (result.error) throw result.error;
if (result.status !== 0) throw new Error(result.stderr);
const graph = new Map(result.stdout.trim().split(/\r?\n/).filter(Boolean).map(line => {
  const [name, imports] = line.split('|');
  return [name, imports ? imports.split(',') : []];
}));
if (!graph.size) throw new Error('No Go packages; backend owner work is pending');

// net/url parses value objects and performs no network I/O. Keep net itself and
// every other net/* package (including net/http) outside Domain/app.
const io = /^(database(?:\/|$)|net(?:\/(?!url$)|$)|os(?:\/|$)|io(?:\/|$)|bufio$|syscall$|log(?:\/|$)|modernc\.org\/|golang\.org\/x\/sys)/;
const issues = [];
for (const [name, imports] of graph) {
  const relative = name.startsWith(moduleName + '/') ? name.slice(moduleName.length + 1) : name;
  const layer = relative.match(/^internal\/(attention|workspace|swarm)\/(domain|app)(?:\/|$)/);
  for (const dependency of imports) {
    const internal = dependency.startsWith(moduleName + '/') ? dependency.slice(moduleName.length + 1) : '';
    if (layer) {
      const [, context, level] = layer;
      const other = internal.match(/^internal\/(attention|workspace|swarm)\//);
      if (other && other[1] !== context) issues.push(`${relative} imports another context: ${dependency}`);
      if (io.test(dependency)) issues.push(`${relative} imports I/O: ${dependency}`);
      if (internal.startsWith('internal/adapters/')) issues.push(`${relative} imports an adapter: ${dependency}`);
      if (level === 'domain') {
        if (internal && !internal.startsWith(`internal/${context}/domain`)) issues.push(`${relative} imports outside its Domain: ${dependency}`);
        if (!internal && dependency.includes('.')) issues.push(`${relative} imports an unapproved external dependency: ${dependency}`);
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
