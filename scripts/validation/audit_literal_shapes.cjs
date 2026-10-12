// Local AST inspection only. Never evaluates historical JS or prints literals.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.argv[3] || path.resolve(__dirname, '../../web/node_modules/typescript'));
const root = process.argv[2];
const audit = JSON.parse(fs.readFileSync(path.join(root, 'semantic-audit-safe.json'), 'utf8'));
const frozen = path.resolve(root, '../../v21-w1/tool-records-local.jsonl');
const source = new Map();
for (const line of fs.readFileSync(frozen, 'utf8').trim().split(/\r?\n/)) {
  const e = JSON.parse(line);
  for (const block of e.blocks) source.set(`native:${e.source}:${e.file_index}:${e.line}:block:${block}`, { e, block });
}
const counts = {};
const inc = key => { counts[key] = (counts[key] || 0) + 1; };
const closedProperties = new Set(['cmd', 'command', 'workdir', 'yield_time_ms', 'max_output_tokens', 'session_id', 'chars', 'path', 'task_id', 'worker_id']);
function shape(node, declarations) {
  if (!node) return 'absent';
  if (ts.isStringLiteralLike(node)) return ts.SyntaxKind[node.kind];
  if (ts.isIdentifier(node)) return 'Identifier:' + (declarations.has(node.text) ? shape(declarations.get(node.text), new Map()) : 'dynamic');
  if (ts.isTaggedTemplateExpression(node)) return 'TaggedTemplate:' + (node.tag.getText() === 'String.raw' ? 'String.raw' : 'other');
  if (ts.isObjectLiteralExpression(node)) return 'Object{' + node.properties.map(p => {
    if (!ts.isPropertyAssignment(p)) return ts.SyntaxKind[p.kind];
    const key = closedProperties.has(p.name.getText()) ? p.name.getText() : 'other';
    return key + ':' + shape(p.initializer, declarations);
  }).join(',') + '}';
  if (ts.isArrayLiteralExpression(node)) return 'Array[' + [...new Set(node.elements.map(n => shape(n, declarations)))].join(',') + ']';
  if (ts.isTemplateExpression(node)) return 'TemplateExpression[' + [...new Set(node.templateSpans.map(n => shape(n.expression, declarations)))].join(',') + ']';
  if (ts.isCallExpression(node)) {const text=node.expression.getText();const label=['JSON.stringify','String','Number','Object.keys'].includes(text)?text:(ts.isPropertyAccessExpression(node.expression)&&['replace','replaceAll','join'].includes(node.expression.name.text)?node.expression.name.text:'helper');return 'Call:'+label+'('+node.arguments.map(n=>shape(n,declarations)).join(',')+')';}
  if (ts.isBinaryExpression(node)) return 'Binary:'+ts.SyntaxKind[node.operatorToken.kind]+'('+shape(node.left,declarations)+','+shape(node.right,declarations)+')';
  return ts.SyntaxKind[node.kind];
}
for (const row of audit.filter(r => r.reason === 'dynamic_arguments')) {
  const entry = source.get(row.parent_id);
  if (!entry) { inc('missing_parent'); continue; }
  const raw = JSON.parse(Buffer.from(entry.e.raw_base64, 'base64'));
  let input = raw.payload?.arguments ?? raw.payload?.input;
  if (typeof input !== 'string') { inc('non_string_wrapper'); continue; }
  const tree = ts.createSourceFile('local.js', input, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
  const calls = [], declarations = new Map(), functions=[];
  function visit(node) {
    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.initializer) declarations.set(node.name.text, node.initializer);
    if (ts.isFunctionDeclaration(node) || ts.isArrowFunction(node)) functions.push(node);
    if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && ts.isIdentifier(node.expression.expression) && node.expression.expression.text === 'tools') calls.push(node);
    ts.forEachChild(node, visit);
  }
  visit(tree); calls.sort((a, b) => a.pos - b.pos);
  for(const f of functions){if(!f.body)continue;let body=f.body;if(ts.isBlock(body)){const statements=body.statements;if(statements.length===1&&ts.isReturnStatement(statements[0]))body=statements[0].expression;else continue;}inc('pure_helper_body:'+shape(body,declarations));}
  const ordinal = Number(row.event_id.split('/semantic/')[1]?.split('.')[1]);
  const call = calls[ordinal];
  inc(row.inner_tool + ':' + shape(call?.arguments[0], declarations));
}
console.log(JSON.stringify(counts, null, 2));
