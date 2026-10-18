import { parseArgs } from 'node:util';
import { statSync } from 'node:fs';
import { isAbsolute, join } from 'node:path';
import { spawnSync } from 'node:child_process';

const { values } = parseArgs({ options: Object.fromEntries([
  'run-id', 'task-id', 'dispatch-id', 'from-handle', 'candidate-sha', 'decision', 'report-path', 'reason', 'orca-path',
].map(name => [name, { type: 'string' }])) });

for (const name of ['run-id', 'task-id', 'dispatch-id', 'from-handle', 'candidate-sha', 'decision', 'report-path', 'reason']) {
  if (!values[name]?.trim()) throw new Error(`缺少必填参数 --${name}`);
}
for (const [name, pattern] of [
  ['run-id', /^run_[a-z0-9]+$/], ['task-id', /^task_[a-z0-9]+$/],
  ['dispatch-id', /^ctx_[a-z0-9]+$/], ['from-handle', /^term_[a-f0-9-]+$/],
  ['candidate-sha', /^[a-f0-9]{40}$/],
]) {
  if (!pattern.test(values[name])) throw new Error(`参数 --${name} 无效`);
}
if (!['accept', 'revise', 'insufficient_evidence'].includes(values.decision)) throw new Error('审查决策无效');
if (!isAbsolute(values['report-path'])) throw new Error('审查报告必须使用实际绝对路径');
const report = statSync(values['report-path']);
if (!report.isFile() || report.size === 0) throw new Error('审查报告必须是已有的非空文件');

let executable = values['orca-path'];
if (!executable) {
  if (process.platform !== 'win32' || !process.env.LOCALAPPDATA) throw new Error('必须明确提供当前会话已选 Orca CLI 的绝对路径');
  executable = join(process.env.LOCALAPPDATA, 'Programs', 'orca', 'resources', 'bin', 'orca.exe');
}
if (!isAbsolute(executable) || !statSync(executable).isFile()) throw new Error('Orca CLI 必须是已安装的实际可执行文件');

const body = [
  `Decision: ${values.decision}`, `Candidate SHA: ${values['candidate-sha']}`, `Run: ${values['run-id']}`,
  `Reviewer Task: ${values['task-id']}`, `Reviewer Dispatch: ${values['dispatch-id']}`,
  `Reviewer terminal: ${values['from-handle']}`, `Report: ${values['report-path']}`, `Reason: ${values.reason}`,
].join('\n');
const result = spawnSync(executable, [
  'orchestration', 'send', '--from', values['from-handle'], '--to', `run:${values['run-id']}`,
  '--type', 'handoff', '--subject', 'ReviewDecision', '--body', body,
  '--task-id', values['task-id'], '--dispatch-id', values['dispatch-id'], '--report-path', values['report-path'], '--json',
], { shell: false, windowsHide: true, encoding: 'utf8', maxBuffer: 1024 * 1024, timeout: 60_000 });
if (result.stdout) process.stdout.write(result.stdout);
if (result.stderr) process.stderr.write(result.stderr);
if (result.error) throw result.error;
if (result.status !== 0) throw new Error(`Orca 审查通知失败（${result.status ?? result.signal}）；请保留回执核查，禁止自动重发`);
const receipt = JSON.parse(result.stdout);
if (receipt.ok !== true) throw new Error('Orca 未确认审查通知入队；请保留回执核查，禁止自动重发');
process.stdout.write('审查通知已持久入队；主控收到并回复后才能确认唤起。\n');
