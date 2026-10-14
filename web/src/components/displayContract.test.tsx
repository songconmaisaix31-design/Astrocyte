import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { MaterialList } from '../pages/attention/AttentionPage';
import { MaterialSummary } from '../pages/attention/MaterialWorkspace';
import { MissionList, MissionDetail, WorkItemList, WorkItemDetail } from '../pages/swarm/SwarmPage';
import { ProjectDetail, ProposalList, ProposalDetail, SessionList, SessionDetail } from '../pages/workspace/WorkspacePage';
import { DesignGraph, DesignTimeline, ResearchRoutes } from './DesignExamples';
import type { components } from '../api/schema';

type S = components['schemas'];
// 明确的函数参数输入只验证显示契约，不表示存在相应来源、会话或执行结果。
const material: S['MaterialV1'] = { id: 'input-material', kind: 'paper', source_locator: '输入定位', lifecycle: 'active', current_revision: 1, title: '输入标题' };
const workItem: S['WorkItemV1'] = { id: 'input-workitem', version: 2, status: 'open', holder_id: null, context_packet_id: null, artifact_ids: ['input-artifact'] };
const mission: S['MissionV1'] = { id: 'input-mission', version: 3, status: 'blocked', grant_id: 'input-grant', goal: '输入目标', work_items: [workItem], artifact_ids: ['input-artifact'], blockers: ['输入阻塞原因'] };
const project: S['ProjectV1'] = { id: 'input-project', name: '输入项目', root_path: '输入目录', environment_id: 'input-environment', repo_id: null, worktree_id: null };
const proposal: S['ProposalV1'] = { id: 'input-proposal', revision: 1, status: 'declined', title: '输入提案', goal: '输入目标', scope: { project_ids: [], allowed_roots: [], allowed_actions: [], data_egress_rules: [] }, deliverables: ['输入交付物'], budget: { calls: null, tokens: null, money: null, currency: null }, stop_conditions: ['输入停止条件'] };
const session: S['SessionV1'] = { id: 'input-session', adapter: 'input-adapter', native_session_id: null, project_id: 'input-project', binding_status: 'unavailable', capabilities: { native_resume: null, handoff: null }, context_state: 'unknown' };
const select = () => { throw new Error('静态显示测试不得触发操作'); };

describe('资料显示契约', () => {
  it('显示全部资料类型输入', () => {
    const html = renderToStaticMarkup(<MaterialList items={(['paper', 'video', 'text', 'file'] as const).map(kind => ({ ...material, id: kind, kind }))} onSelect={select} />);
    for (const label of ['论文', '视频', '文本', '文件']) expect(html).toContain(label);
  });
  it('长中文标题保留完整内容和两行显示样式', () => {
    const title = '中文标题'.repeat(50);
    const html = renderToStaticMarkup(<MaterialList items={[{ ...material, title }]} onSelect={select} />);
    expect(html).toContain(title);
    expect(html).toContain('line-clamp-2');
  });
  it('完整显示失败输入状态', () => {
    expect(renderToStaticMarkup(<MaterialList items={[{ ...material, import_status: 'failed' }]} onSelect={select} />)).toContain('失败');
  });
  it('完整显示归档输入状态', () => {
    expect(renderToStaticMarkup(<MaterialList items={[{ ...material, lifecycle: 'archived' }]} onSelect={select} />)).toContain('已归档');
  });
  it('缺失字段保持未提供或未知', () => {
    const html = renderToStaticMarkup(<MaterialSummary item={material} />);
    expect(html).toContain('收藏理由');
    expect(html).toContain('未提供');
    expect(html).toContain('人工使用');
  });
});

describe('项目、提案和会话显示契约', () => {
  it('显示项目原始目录和标识', () => {
    const html = renderToStaticMarkup(<ProjectDetail item={project} />);
    expect(html).toContain(project.root_path);
    expect(html).toContain(project.id);
  });
  it('显示提案标题、状态和交付数量', () => {
    const html = renderToStaticMarkup(<ProposalList items={[proposal]} onSelect={select} />);
    expect(html).toContain(proposal.title);
    expect(html).toContain('已拒绝');
    expect(html).toContain('1');
  });
  it('提案详情保留未知预算和停止条件', () => {
    const html = renderToStaticMarkup(<ProposalDetail item={proposal} />);
    expect(html).toContain('未知');
    expect(html).toContain('输入停止条件');
    expect(html).toContain('输入交付物');
  });
  it('会话列表保留 unknown 能力', () => {
    const html = renderToStaticMarkup(<SessionList items={[session]} onSelect={select} />);
    expect(html).toContain(session.adapter);
    expect(html).toContain('未知');
  });
  it('会话能力保留三值语义', () => {
    for (const [value, label] of [[null, '未知'], [false, '不可用'], [true, '可用']] as const) {
      const html = renderToStaticMarkup(<SessionDetail item={{ ...session, capabilities: { native_resume: value, handoff: value } }} />);
      expect(html).toContain(label);
      expect(html).toContain('能力');
      expect(html).toContain('上下文状态');
    }
  });
});

describe('Mission 显示契约，实际生命周期验收 NOT_RUN', () => {
  it('显示运行、阻塞和失败输入状态', () => {
    const html = renderToStaticMarkup(<MissionList items={(['running', 'blocked', 'failed'] as const).map(status => ({ ...mission, id: status, status }))} onSelect={select} />);
    for (const label of ['执行中', '已阻塞', '失败']) expect(html).toContain(label);
  });
  it('详情显示工作项、产物引用和阻塞原因', () => {
    const html = renderToStaticMarkup(<MissionDetail item={mission} />);
    for (const value of ['工作项', workItem.id, '产物引用', 'input-artifact', '输入阻塞原因']) expect(html).toContain(value);
  });
  it('工作项列表保持数量和输入状态', () => {
    const html = renderToStaticMarkup(<WorkItemList items={[workItem]} missions={[mission]} onSelect={select} />);
    expect(html.match(/role="button"/g)).toHaveLength(1);
    expect(html).toContain(workItem.id);
    expect(html).toContain('未认领');
  });
  it('工作项详情保留所属任务和产物引用', () => {
    const html = renderToStaticMarkup(<WorkItemDetail item={workItem} missionId={mission.id} />);
    for (const value of [mission.id, 'input-artifact', '未认领', '上下文包']) expect(html).toContain(value);
  });
});

describe('缺少事实来源时的协作显示', () => {
  it('拓扑按钮禁用且没有构造任务', () => {
    const html = renderToStaticMarkup(<DesignGraph />);
    expect(html).toContain('disabled');
    expect(html).toContain('拓扑尚未接入');
    expect(html).not.toContain('T-11');
  });
  it('时间线显示缺少接口，保留未知观察状态', () => {
    const html = renderToStaticMarkup(<DesignTimeline />);
    expect(html).toContain('事件时间线尚未接入');
    expect(html).not.toContain('14:32');
  });
  it('研究路线缺少关系来源时明确提示', () => {
    expect(renderToStaticMarkup(<ResearchRoutes />)).toContain('研究路线尚未接入');
  });
});
