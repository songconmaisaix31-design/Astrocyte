import { createRequire } from 'node:module';
import { DatabaseSync } from 'node:sqlite';
import fs from 'node:fs';
import path from 'node:path';

// stdout只承载经筛选的引用和数量，第三方诊断留给本机stderr。
const input = JSON.parse(fs.readFileSync(0, 'utf8'));
process.env.CODEGRAPH_TELEMETRY = '0';
const require = createRequire(import.meta.url);
const pkg = require(path.join(input.sdk, 'package.json'));
if (pkg.version !== '1.0.1') throw new Error('unsupported_codegraph_version');
const { CodeGraph, QueryBuilder, setLogger, silentLogger } = require(input.sdk);
setLogger(silentLogger);

const span = node => ({ path: node.filePath, start_line: node.startLine, end_line: node.endLine });
const symbol = (node, score = 0) => ({ id: node.id, name: node.name, kind: node.kind, source_span: span(node), score });
const within = file => input.roots.some(root => root === '.' || file === root || file.startsWith(`${root}/`));

async function run() {
  if (['index', 'sync', 'rebuild'].includes(input.action)) {
    const graph = input.action === 'sync'
      ? await CodeGraph.open(input.root, { sync: false })
      : await CodeGraph.init(input.root, { index: false });
    try {
      const result = input.action === 'sync' ? await graph.sync() : await graph.indexAll();
      if (result.success === false || result.filesErrored > 0 || result.errors?.length) throw new Error('codegraph_index_failed');
      const changes = graph.getChangedFiles();
      if (changes.added.length || changes.modified.length || changes.removed.length) throw new Error('codegraph_index_incomplete');
      const stats = graph.getStats();
      if (stats.nodeCount < 1) throw new Error('codegraph_index_empty');
      return { version: pkg.version, indexed: true, files_added: result.filesAdded ?? 0, files_modified: result.filesModified ?? 0, files_removed: result.filesRemoved ?? 0 };
    } finally { graph.close(); }
  }

  // CodeGraph.open的readOnly参数在1.0.1未生效；官方QueryBuilder使用真正只读的SQLite连接。
  const db = new DatabaseSync(path.join(input.root, '.codegraph', 'codegraph.db'), { readOnly: true });
  try {
    const queries = new QueryBuilder(db);
    const symbols = [];
    const relations = [];
    if (input.action === 'query') {
      const results = queries.searchNodes(input.text, { limit: input.limit + 1 });
      if (results.length > input.limit) throw new Error('repository_result_budget_exceeded');
      for (const row of results) {
        if (!within(row.node.filePath)) throw new Error('repository_scope_mismatch');
        symbols.push(symbol(row.node, row.score));
      }
    } else if (input.action === 'impact') {
      const initial = queries.getNodeById(input.symbol);
      if (!initial || !within(initial.filePath)) throw new Error('repository_symbol_not_found');
      const nodes = new Map([[initial.id, initial]]);
      const queue = [[initial.id, 0]];
      const seenEdges = new Set();
      for (let cursor = 0; cursor < queue.length; cursor += 1) {
        const [id, depth] = queue[cursor];
        if (depth >= input.depth) continue;
        for (const edge of queries.getIncomingEdges(id, ['calls', 'imports', 'implements', 'extends', 'references', 'type_of', 'instantiates', 'overrides'])) {
          const node = queries.getNodeById(edge.source);
          if (!node || !within(node.filePath)) throw new Error('repository_scope_mismatch');
          const key = JSON.stringify([edge.source, edge.target, edge.kind]);
          if (seenEdges.has(key)) continue;
          seenEdges.add(key);
          if (relations.length >= input.limit) throw new Error('repository_result_budget_exceeded');
          relations.push({ source: edge.source, target: edge.target, kind: edge.kind, evidence_kind: 'codegraph_static_heuristic', source_span: span(node), extractor_version: pkg.version });
          if (!nodes.has(node.id)) {
            if (nodes.size >= input.limit) throw new Error('repository_result_budget_exceeded');
            nodes.set(node.id, node); queue.push([node.id, depth + 1]);
          }
        }
      }
      for (const node of nodes.values()) symbols.push(symbol(node));
    } else { throw new Error('unsupported_repository_action'); }
    return { symbols, relations, version: pkg.version };
  } finally { db.close(); }
}

try {
  process.stdout.write(JSON.stringify(await run()));
} catch (error) {
  if (error?.message === 'repository_result_budget_exceeded') {
    process.stdout.write(JSON.stringify({ version: pkg.version, error: 'budget_exhausted' }));
  } else {
    throw error;
  }
}
