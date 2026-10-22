import { useEffect, useRef, useState } from 'react';
import type { components } from '../../api/schema';
import { localProjectsApi } from '../../api/s1';
import { swaApi, SwaCommandError, SwaRuntimeUnavailableError, type SwaCommand, type SwaEnvelope } from '../../api/swa';
import { useReadApi, formatError } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { SectionCard } from '../../components/SectionCard';
import { readSwaFleet, readSwaAgents, readSwaTranscript, recoverSwaRequest, observationJSON, type PendingSwaRequest } from './swaObservation';
import styles from './SwaWorkspace.module.css';

type Project = components['schemas']['LocalProjectV1'];

export function SwaWorkspace() {
  const projects = useReadApi(signal => localProjectsApi.listProjects({ signal }));
  const [projectID, setProjectID] = useState('');
  const project = projects.data?.items.find(item => item.id === projectID);
  return <SectionCard title="SWA 执行工作台" padded>
    <div className={styles.workspace}>
      <p className={styles.note}>连接 CLI 使用的同一项目与任务。当前人工会话和项目批准由服务逐次核查；关闭页面保留后台任务。</p>
      <QueryState state={projects} empty={data => !data.items.length} emptyTitle="尚无已登记项目，请先在共同工作区登记并审阅权限">{data => <label className={styles.form}>选择已登记项目<select aria-label="SWA 项目" value={projectID} disabled={projects.loading || projects.stale} onChange={event => setProjectID(event.target.value)}><option value="">请选择项目</option>{data.items.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>}</QueryState>
      <a href="/workspace?tab=overview">管理项目与批准范围</a>
      {project && !projects.stale && <ProjectRuntime key={`${project.id}:${project.settings.revision}`} project={project} />}
    </div>
  </SectionCard>;
}

function ProjectRuntime({ project }: { project: Project }) {
  const status = useReadApi(signal => swaApi.command(project.id, { action: 'status' }, signal), { key: project.id });
  const capabilities = useReadApi(signal => swaApi.command(project.id, { action: 'capabilities' }, signal), { key: project.id, enabled: !!status.data && !status.error });
  const [controller, setController] = useState(''), [taskID, setTaskID] = useState('');
  const [attached, setAttached] = useState<{ controller: string; taskID: string } | null>(null);
  const [workdir, setWorkdir] = useState(project.root);
  const open = useSwaCommand(project.id, 'controller.open');
  const available = !!status.data && !status.loading && !status.stale && !!capabilities.data && !capabilities.stale;
  const canStart = available && project.settings.allow_agent_control && project.settings.allowed_actions?.includes('start');
  return <div className={styles.workspace}>
    <p className={styles.identity}>项目 · {project.name} · {project.id} · 权限版本 {project.settings.revision}</p>
    <QueryState state={status}>{reply => <details><summary>执行服务 · 已取得状态快照</summary><p className={styles.note}>查询完成仅证明本次观察。任务状态和资源退出仍以各自原生记录为准。</p><pre className={styles.output}>{observationJSON(reply)}</pre></details>}</QueryState>
    {!!status.data && <QueryState state={capabilities}>{reply => <details><summary>实际 Agent 与逐项能力</summary><p className={styles.note}>注册、安装、配置、可启动和控制能力分别核实。此清单不创建 Worker。</p><pre className={styles.output}>{observationJSON(reply.result)}</pre></details>}</QueryState>}
    <details><summary>新建 Controller 与固定任务</summary><p className={styles.note}>以下路径指向当前服务机器。页面不上传文件或修改项目批准；任务规格文件中的 max_workers 限制总创建和预留名额，释放 Worker 不归还名额。</p><form className={styles.form} onSubmit={event => { event.preventDefault(); void open.run({ action: 'controller.open', project_id: project.id, workdir: workdir.trim() }, reply => { if (!reply.controller) throw new Error('回执缺少真实 Controller，请查询原请求'); setController(reply.controller); }); }}><label>已批准的服务端工作目录<input required value={workdir} onChange={event => setWorkdir(event.target.value)} /></label><button className="ac-button secondary compact" type="submit" disabled={!canStart || open.pending || !!open.unresolved}>创建 Controller</button></form><OperationState command={open} onRecovered={reply => { if (reply.controller) setController(reply.controller); }} />
      {controller.trim() && <TaskCreateForm key={controller.trim()} projectID={project.id} controller={controller.trim()} disabled={!canStart} onCreated={id => { setTaskID(id); setAttached({ controller: controller.trim(), taskID: id }); }} />}
    </details>
    <form className={styles.form} onSubmit={event => { event.preventDefault(); setAttached({ controller: controller.trim(), taskID: taskID.trim() }); }}>
      <label>CLI 回执中的 Controller<input required value={controller} onChange={event => setController(event.target.value)} /></label>
      <label>同一任务的根 Task ID<input required value={taskID} onChange={event => setTaskID(event.target.value)} /></label>
      <button className="ac-button compact" disabled={!status.data || status.loading || status.stale || !controller.trim() || !taskID.trim()} type="submit">连接已有任务</button>
    </form>
    <p className={styles.note}>可连接 CLI 的真实任务，也可使用上方已批准文件建立任务。所有操作使用同一执行服务与原生身份。</p>
    {attached && capabilities.data && <TaskWorkspace key={`${attached.controller}:${attached.taskID}`} project={project} capabilities={capabilities.data} {...attached} />}
  </div>;
}

function TaskCreateForm({ projectID, controller, disabled, onCreated }: { projectID: string; controller: string; disabled: boolean; onCreated: (id: string) => void }) {
  const [specFile, setSpecFile] = useState('');
  const create = useSwaCommand(projectID, `task.create:${controller}`);
  return <><form className={styles.form} onSubmit={event => { event.preventDefault(); void create.run({ action: 'task.create', controller, spec_file: specFile.trim() }, reply => { if (!reply.task_id) throw new Error('回执缺少真实根 Task ID，请查询原请求'); onCreated(reply.task_id); }); }}><label>已批准的服务端任务规格文件<input required value={specFile} onChange={event => setSpecFile(event.target.value)} /></label><button className="ac-button secondary compact" type="submit" disabled={disabled || !specFile.trim() || create.pending || !!create.unresolved}>在当前 Controller 建立任务</button></form><OperationState command={create} controller={controller} onRecovered={reply => { if (reply.task_id) onCreated(reply.task_id); }} /></>;
}

function TaskWorkspace({ project, controller, taskID, capabilities }: { project: Project; controller: string; taskID: string; capabilities: SwaEnvelope }) {
  const fleet = useReadApi(async signal => readSwaFleet(await swaApi.command(project.id, { action: 'worker.list', controller, task_id: taskID }, signal), project.id, controller, taskID), { key: `${project.id}:${controller}:${taskID}` });
  const [dispatchID, setDispatchID] = useState('');
  const selected = fleet.data?.workers.find(worker => worker.dispatchID === dispatchID);
  const [outputDir, setOutputDir] = useState('');
  const exportCommand = useSwaCommand(project.id, `task.result:${controller}:${taskID}`);
  return <div className={styles.workspace}>
    <div><button className="ac-button secondary compact" type="button" disabled={fleet.loading} onClick={fleet.retry}>刷新任务快照</button></div>
    <QueryState state={fleet}>{data => <>
      <p className={styles.identity}>Run · {data.runID} · 根任务 {taskID}</p>
      {!fleet.stale && <WorkerStartForm key={`${project.id}:${controller}:${taskID}`} project={project} controller={controller} taskID={taskID} capabilities={capabilities} disabled={fleet.loading} onStarted={fleet.retry} />}
      <p className={styles.note}>本页 {data.workers.length} / 服务返回总数 {data.total} 个 Worker。{data.hasMore ? '清单不完整：还有未加载的 Worker，请使用 CLI 核查完整范围。' : '服务返回的本次清单已完整。'} 未提供协作依赖，不生成连接线。</p>
      <div className={styles.columns}>
        <ul className={styles.workers} aria-label="实际 Worker">{data.workers.map(worker => <li key={worker.dispatchID}><button className={styles.worker} aria-pressed={worker.dispatchID === selected?.dispatchID} type="button" disabled={fleet.stale || fleet.loading} onClick={() => setDispatchID(worker.dispatchID)}><strong>{worker.dispatchID}</strong><small>Task · {worker.taskID}</small><span>业务 · {worker.dispatchStatus}</span><span>Worker · {worker.workerState}</span><span>资源 · {worker.terminalState ?? '未观察，保持未知'}</span></button></li>)}</ul>
        {selected && !fleet.stale ? <WorkerWorkspace key={selected.dispatchID} project={project} controller={controller} dispatchID={selected.dispatchID} onChanged={fleet.retry} /> : <p className={styles.note}>{data.workers.length ? '选择 Worker 查看原生输出与可用操作。' : '服务本次没有返回 Worker。'}</p>}
      </div>
    </>}</QueryState>
    <details><summary>完整成果导出</summary><p className={styles.note}>成果导出与消息预览分别处理。服务核查全部实际来源与大小边界后，将成果写入当前项目已批准父目录下的全新目录；浏览器不会把预览当作完整成果下载。</p><form className={styles.form} onSubmit={event => { event.preventDefault(); void exportCommand.run({ action: 'task.result', controller, task_id: taskID, output_dir: outputDir.trim() }); }}><label>服务端新的绝对输出目录<input required value={outputDir} onChange={event => setOutputDir(event.target.value)} /></label><button className="ac-button secondary compact" type="submit" disabled={fleet.loading || fleet.stale || !fleet.data || exportCommand.pending || !outputDir.trim()}>导出实际成果</button></form><OperationState command={exportCommand} controller={controller} /></details>
  </div>;
}

function WorkerWorkspace({ project, controller, dispatchID, onChanged }: { project: Project; controller: string; dispatchID: string; onChanged: () => void }) {
  const detail = useReadApi(signal => swaApi.command(project.id, { action: 'worker.show', controller, dispatch_id: dispatchID }, signal), { key: dispatchID });
  const [cursor, setCursor] = useState(''), [readCursor, setReadCursor] = useState<string | null>(null);
  const transcript = useReadApi(signal => swaApi.command(project.id, { action: 'worker.read', controller, dispatch_id: dispatchID, source: 'transcript', limit: 100, ...(readCursor ? { cursor: readCursor } : {}) }, signal), { key: `${dispatchID}:${readCursor}`, enabled: readCursor !== null });
  const [inputFile, setInputFile] = useState(''), [confirmStop, setConfirmStop] = useState(false);
  const command = useSwaCommand(project.id, `worker.send:${controller}:${dispatchID}`);
  const stop = useSwaCommand(project.id, `worker.stop:${controller}:${dispatchID}`);
  const blocked = command.pending || !!command.unresolved || stop.pending || !detail.data || detail.stale || detail.loading;
  const stopBlocked = stop.pending || !!stop.unresolved || !detail.data || detail.stale || detail.loading;
  return <section className={styles.detail} aria-label={`Worker ${dispatchID}`}>
    <h3>当前 Worker</h3><code className={styles.identity}>{dispatchID}</code>
    <QueryState state={detail}>{reply => <details><summary>原生身份、业务状态和资源观察</summary><pre className={styles.output}>{observationJSON(reply.result)}</pre></details>}</QueryState>
    <form className={styles.form} onSubmit={event => { event.preventDefault(); setReadCursor(cursor.trim()); transcript.retry(); }}><label>继续读取游标（空值从头）<input value={cursor} onChange={event => setCursor(event.target.value)} /></label><button className="ac-button secondary compact" type="submit" disabled={detail.loading || detail.stale || !detail.data || transcript.loading}>读取实际会话</button></form>
    {readCursor !== null && <QueryState state={transcript}>{reply => <Transcript reply={reply} />}</QueryState>}
    <form className={styles.form} onSubmit={event => { event.preventDefault(); void command.run({ action: 'worker.send', controller, dispatch_id: dispatchID, input_file: inputFile.trim() }, () => { detail.retry(); onChanged(); }); }}><label>追加输入的已批准服务端 UTF-8 文件<input required value={inputFile} onChange={event => setInputFile(event.target.value)} /></label><button className="ac-button compact" type="submit" disabled={blocked || !inputFile.trim() || !project.settings.allowed_actions?.includes('send')}>提交此 Worker 的输入</button></form>
    <p className={styles.note}>当前接口读取项目中的明确文件；浏览器文本上传尚无共享合同。模型输入仍受当前项目批准、原生工具范围和账号限制。CLI Agent 创建的任务绑定原调用身份；人工会话可观察与安全停止，跨身份启动或追加输入会由服务拒绝。</p>
    <button className="ac-button secondary compact" type="button" disabled={stopBlocked} onClick={() => setConfirmStop(true)}>停止此 Worker…</button>
    {confirmStop && <div className={styles.confirm}><p>确认停止 Dispatch {dispatchID}？服务将核对其所属资源；其他 Worker 保留。操作提交不代表资源已经退出。</p><button className="ac-button compact" type="button" disabled={stopBlocked} onClick={() => { setConfirmStop(false); void stop.run({ action: 'worker.stop', controller, dispatch_id: dispatchID }, () => { detail.retry(); onChanged(); }); }}>确认停止 {dispatchID}</button><button className="ac-button secondary compact" type="button" onClick={() => setConfirmStop(false)}>取消</button></div>}
    <OperationState command={command} controller={controller} onRecovered={() => { detail.retry(); onChanged(); }} />
    <OperationState command={stop} controller={controller} onRecovered={() => { detail.retry(); onChanged(); }} />
  </section>;
}

function WorkerStartForm({ project, controller, taskID, capabilities, disabled, onStarted }: { project: Project; controller: string; taskID: string; capabilities: SwaEnvelope; disabled: boolean; onStarted: () => void }) {
  let agents: ReturnType<typeof readSwaAgents> = [], capabilityError: string | null = null;
  try { agents = readSwaAgents(capabilities).filter(agent => agent.id === project.settings.external_model_cli); } catch (error) { capabilityError = formatError(error); }
  const [agentID, setAgentID] = useState(''), [name, setName] = useState(''), [inputFile, setInputFile] = useState('');
  const agent = agents.find(item => item.id === agentID);
  const command = useSwaCommand(project.id, `worker.start:${controller}:${taskID}`);
  return <details><summary>向同一任务添加 Worker</summary><p className={styles.note}>数量由服务的任务规格与资源准入核查。每次启动建立独立 Dispatch，页面不预造身份；已安装或可启动不等于模型执行已验证。当前人工身份只能沿属于自己的任务创建执行，不能继承 CLI Agent 的授权。</p><form className={styles.form} onSubmit={event => { event.preventDefault(); if (!agent) return; void command.run({ action: 'worker.start', controller, task_id: taskID, name: name.trim(), agent: agent.id, model: agent.model, input_file: inputFile.trim() }, reply => { if (!reply.dispatch_id) throw new Error('回执缺少真实 Dispatch，请查询原请求'); onStarted(); }); }}>
    <label>项目已批准且实际发现的 Agent<select required value={agentID} onChange={event => setAgentID(event.target.value)}><option value="">选择 Agent</option>{agents.map(item => <option key={item.id} value={item.id} disabled={!item.launchable || !item.model}>{item.id} · {item.model || '模型未提供'}{item.launchable ? '' : ' · 尚不可启动'}</option>)}</select></label>
    <label>此 Worker 名称<input required value={name} onChange={event => setName(event.target.value)} /></label><label>已批准的服务端输入文件<input required value={inputFile} onChange={event => setInputFile(event.target.value)} /></label><button className="ac-button compact" type="submit" disabled={disabled || command.pending || !!command.unresolved || !agent?.launchable || !agent.model || !name.trim() || !inputFile.trim() || !project.settings.allow_agent_control || !project.settings.allowed_actions?.includes('start')}>启动此 Worker</button>
  </form>{capabilityError && <p role="alert">{capabilityError}</p>}{!agents.length && <p role="status">当前能力清单未返回项目已批准的 Agent，启动不可用。</p>}<OperationState command={command} controller={controller} onRecovered={onStarted} /></details>;
}

function Transcript({ reply }: { reply: SwaEnvelope }) {
  let page: ReturnType<typeof readSwaTranscript>;
  try { page = readSwaTranscript(reply); } catch (error) { return <div role="alert"><p>{formatError(error)}</p><pre className={styles.output}>{observationJSON(reply.result)}</pre></div>; }
  return <section aria-label="原生聊天记录"><p className={styles.note}>来源 · {page.sourceIdentity}。{page.limited ? '本页未包含全部消息。' : '本页已读取至当前边界。'}{page.complete ? '服务标记内容完整。' : '完整成果尚未确认。'} 查询成功不代表任务完成。</p>{page.messages.map(message => <article className={styles.message} key={message.id}><strong>{message.role}</strong>{message.blocks.map((block, index) => <pre key={index} className={styles.output}>{block.type === 'text' ? block.text : `${block.type}\n${block.text}`}</pre>)}</article>)}<p className={styles.identity}>继续读取游标 · {page.nextCursor}</p></section>;
}

function useSwaCommand(projectID: string, scope: string) {
  const [pending, setPending] = useState(false), [error, setError] = useState<string | null>(null);
  const storageKey = `swa-pending:${projectID}:${scope}`;
  const [reply, setReply] = useState<SwaEnvelope | null>(null), [unresolved, setUnresolved] = useState<PendingSwaRequest | null>(() => {
    const saved = typeof window === 'undefined' ? null : sessionStorage.getItem(storageKey);
    if (!saved) return null;
    const parsed: unknown = JSON.parse(saved);
    if (!parsed || typeof parsed !== 'object' || !('requestID' in parsed) || typeof parsed.requestID !== 'string' || !('action' in parsed) || typeof parsed.action !== 'string') throw new Error('保存的请求身份不可解析，请核查本标签页记录');
    const record = parsed as Record<string, unknown>;
    for (const key of ['controller', 'taskID', 'dispatchID'] as const) if (key in record && typeof record[key] !== 'string') throw new Error('保存的请求选择范围不可解析');
    return parsed as PendingSwaRequest;
  });
  const active = useRef(false), busy = useRef(false);
  useEffect(() => { active.current = true; return () => { active.current = false; }; }, []);
  async function run(body: SwaCommand, onSuccess?: (reply: SwaEnvelope) => void) {
    if (busy.current || unresolved) return;
    busy.current = true;
    const mutation = ['controller.open', 'task.create', 'worker.start', 'worker.send', 'worker.stop'].includes(body.action);
    const requestID = mutation ? crypto.randomUUID() : undefined;
    setPending(true); setError(null); setReply(null);
    try {
      if (requestID) { const original: PendingSwaRequest = { requestID, action: body.action, ...(body.controller ? { controller: body.controller } : {}), ...(body.task_id ? { taskID: body.task_id } : {}), ...(body.dispatch_id ? { dispatchID: body.dispatch_id } : {}) }; sessionStorage.setItem(storageKey, JSON.stringify(original)); setUnresolved(original); }
      const result = await swaApi.command(projectID, { ...body, ...(requestID ? { request_id: requestID } : {}) });
      if (active.current) { setReply(result); onSuccess?.(result); sessionStorage.removeItem(storageKey); setUnresolved(null); }
    } catch (failure) {
      if (failure instanceof SwaRuntimeUnavailableError) sessionStorage.removeItem(storageKey);
      if (active.current) {
        setError(formatError(failure));
        if (failure instanceof SwaCommandError) setReply(failure.reply);
        if (failure instanceof SwaRuntimeUnavailableError) setUnresolved(null);
        // 其余不确定错误保留原请求，不推断已执行效果。
      }
    } finally { busy.current = false; if (active.current) setPending(false); }
  }
  async function inspect(onRecovered?: (reply: SwaEnvelope) => void) {
    if (!unresolved || busy.current) return;
    busy.current = true; setPending(true);
    try {
      const result = await swaApi.command(projectID, { action: 'request.show', ...(unresolved.controller ? { controller: unresolved.controller } : { project_id: projectID }), request_id: unresolved.requestID });
      if (active.current) {
        setReply(result);
        const recovered = recoverSwaRequest(result, projectID, unresolved);
        if (recovered) { onRecovered?.(recovered); sessionStorage.removeItem(storageKey); setUnresolved(null); setError(null); }
      }
    } catch (failure) { if (active.current) setError(formatError(failure)); }
    finally { busy.current = false; if (active.current) setPending(false); }
  }
  return { pending, error, reply, unresolved, run, inspect };
}

function OperationState({ command, onRecovered }: { command: ReturnType<typeof useSwaCommand>; controller?: string; onRecovered?: (reply: SwaEnvelope) => void }) {
  return <div aria-live="polite">{command.pending && <p role="status">正在提交，请保留当前操作…</p>}{command.error && <p role="alert">{command.error}</p>}{command.unresolved && <div role="alert"><p>本次操作尚未确认 · {command.unresolved.action} · 原请求 {command.unresolved.requestID}。请保留此标识；不会自动重复发送。</p><button className="ac-button secondary compact" type="button" disabled={command.pending} onClick={() => { void command.inspect(onRecovered); }}>查询原请求</button></div>}{command.reply && <details open><summary>服务实际回执 · {command.reply.ok ? '本次命令已返回' : '未确认成功'}</summary><p className={styles.note}>请分别核对原生业务状态、输出完整性与资源退出；成功返回不表示全部任务成功。</p><pre className={styles.output}>{observationJSON(command.reply)}</pre></details>}</div>;
}
