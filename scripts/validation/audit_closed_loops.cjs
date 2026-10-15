// Read-only local TypeScript AST/source join. Output is syntax shape only.
// Never evaluates JS, prints literals, or emits source arguments/metadata.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.argv[4]);
const before = JSON.parse(fs.readFileSync(path.join(process.argv[2], 'semantic-audit-safe.json'), 'utf8'));
const after = new Map(JSON.parse(fs.readFileSync(path.join(process.argv[3], 'semantic-audit-safe.json'), 'utf8')).map(x => [x.event_id, x]));
const frozen = path.resolve(process.argv[2], '../../v21-w1/tool-records-local.jsonl');
const parents = new Map();
for (const line of fs.readFileSync(frozen, 'utf8').trim().split(/\r?\n/)) {
  const e = JSON.parse(line);
  for (const block of e.blocks) parents.set(`native:${e.source}:${e.file_index}:${e.line}:block:${block}`, { e, block });
}
const props = new Set(['cmd', 'command', 'workdir', 'yield_time_ms', 'max_output_tokens', 'session_id', 'chars', 'path']);
function shape(n) {
  if (!n) return 'absent';
  if (ts.isStringLiteralLike(n)) return 'literal_string';
  if (ts.isIdentifier(n)) return 'binding';
  if (ts.isPropertyAccessExpression(n)) return 'property_access';
  if (ts.isElementAccessExpression(n)) return 'index_access';
  if (ts.isArrayLiteralExpression(n)) return `array(${[...new Set(n.elements.map(shape))].join(',')})`;
  if (ts.isObjectLiteralExpression(n)) return 'object(' + n.properties.map(p => ts.isPropertyAssignment(p) ? (props.has(p.name.getText()) ? p.name.getText() : 'other') + ':' + shape(p.initializer) : ts.SyntaxKind[p.kind]).join(',') + ')';
  if (ts.isTemplateExpression(n)) return 'template(' + [...new Set(n.templateSpans.map(s => shape(s.expression)))].join(',') + ')';
  if (ts.isCallExpression(n)) return 'call(' + n.arguments.map(shape).join(',') + ')';
  if (ts.isBinaryExpression(n)) return ts.SyntaxKind[n.operatorToken.kind] + '(' + shape(n.left) + ',' + shape(n.right) + ')';
  return ts.SyntaxKind[n.kind];
}
const rows = [], counts = {};
for (const row of before.filter(x => x.reason === 'dynamic_arguments')) {
  const entry = parents.get(row.parent_id);
  if (!entry) throw new Error('missing frozen parent');
  const raw = JSON.parse(Buffer.from(entry.e.raw_base64, 'base64'));
  const input = raw.payload?.arguments ?? raw.payload?.input;
  if (typeof input !== 'string') continue;
  const tree = ts.createSourceFile('local.js', input, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
  const calls = [], declarations = new Map();
  function visit(n) {
    if (ts.isVariableDeclaration(n) && ts.isIdentifier(n.name) && n.initializer && (n.parent.flags & ts.NodeFlags.Const)) declarations.set(n.name.text, n.initializer);
    if (ts.isCallExpression(n) && ts.isPropertyAccessExpression(n.expression) && ts.isIdentifier(n.expression.expression) && n.expression.expression.text === 'tools') calls.push(n);
    ts.forEachChild(n, visit);
  }
  visit(tree); calls.sort((a, b) => a.pos - b.pos);
  const call = calls[Number(row.event_id.split('/semantic/')[1]?.split('.')[1])];
  let loop = call?.parent;
  while (loop && !ts.isForOfStatement(loop)) loop = loop.parent;
  if (!loop) continue;
  const array = ts.isIdentifier(loop.expression) ? declarations.get(loop.expression.text) : loop.expression;
  if (!array || !ts.isArrayLiteralExpression(array)) continue;
  const outcome = after.get(row.event_id);
  if (!outcome) throw new Error('changed lexical child identity');
  const bodyFlags = new Set();
  function bodyVisit(n) {
    if (ts.isIfStatement(n) || ts.isForOfStatement(n) || ts.isForStatement(n) || ts.isWhileStatement(n) || ts.isFunctionDeclaration(n) || ts.isArrowFunction(n) || ts.isReturnStatement(n) || ts.isBreakStatement(n) || ts.isContinueStatement(n) || ts.isConditionalExpression(n)) bodyFlags.add(ts.SyntaxKind[n.kind]);
    ts.forEachChild(n, bodyVisit);
  }
  bodyVisit(loop.statement);
  const structure = { event_id: row.event_id, binding: shape(loop.initializer.declarations?.[0]?.name), declaration: (loop.initializer.flags & ts.NodeFlags.Const) ? 'const' : (loop.initializer.flags & ts.NodeFlags.Let) ? 'let' : 'other', source: shape(array), argument: shape(call.arguments[0]), body_flags: [...bodyFlags].sort(), after_bucket: outcome.bucket, after_reason: outcome.reason };
  rows.push(structure);
  const key = outcome.bucket + '/' + outcome.reason + '/' + structure.declaration + '/' + structure.body_flags.join(',') + '/' + structure.argument;
  counts[key] = (counts[key] || 0) + 1;
}
console.log(JSON.stringify({ count: rows.length, counts, rows }, null, 2));
