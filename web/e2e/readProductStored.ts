import { expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const records = process.env.ASTROCYTE_TEST_PRODUCT_RECORDS_DIR;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (!records || !output || resolve(records) === resolve(output)) throw new Error('需要分开的原公开记录目录和本次输出目录');
const paper = JSON.parse(await readFile(resolve(records, 'product-paper-human.json'), 'utf8')) as {
  imports: { status: number; material_id: string; revision: number }[]; distillation_status: number;
};
const paperRead = JSON.parse(await readFile(resolve(records, 'product-paper-readonly.json'), 'utf8')) as {
  manual_distillation_id: string; events: { id: string; type: string; aggregate_id: string }[];
};
const projectRecord = JSON.parse(await readFile(resolve(records, 'product-project-human.json'), 'utf8')) as {
  project_id: string; space_id: string; settings_revision: number; context_files: number; denied_status: number;
};
const proposalRecord = JSON.parse(await readFile(resolve(records, 'product-proposal-human.json'), 'utf8')) as {
  project_id: string; material_id: string; goal_id: string; proposal_id: string; grant_id: string; context_id: string; context_status: number;
};
expect(paper.imports.map(item => item.status)).toEqual([202, 202, 202]);
expect(paper.distillation_status).toBe(200);
expect(projectRecord.context_files).toBe(1);
expect(projectRecord.denied_status).toBe(403);
expect(proposalRecord.context_status).toBe(200);
expect(proposalRecord.project_id).toBe(projectRecord.project_id);
expect(proposalRecord.material_id).toBe(paper.imports[2].material_id);

// 只打开协调者明确交付的产品库，逐项核对原公开记录及其实际事件。
const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
try {
  const read = <T,>(table: string, id: string): T => {
    const row = database.prepare(`SELECT data FROM ${table} WHERE id=?`).get(id) as { data: string } | undefined;
    if (!row) throw new Error(`原产品记录不存在：${table}/${id}`);
    return JSON.parse(row.data) as T;
  };
  const project = read<S['LocalProjectV1']>('local_agent_projects', projectRecord.project_id);
  expect(project.space_id).toBe(projectRecord.space_id);
  expect(project.settings.revision).toBe(projectRecord.settings_revision);
  expect(project.settings).toMatchObject({ allow_directory: true, expand_references: false, allowed_subdirs: ['web/src/pages/workspace/proposals'], allowed_actions: ['read_context'], allowed_tools: [], external_model_cli: '', allow_agent_control: false });
  const materialIDs = [...new Set(paper.imports.map(item => item.material_id))];
  for (const id of materialIDs) {
    const detail = read<S['MaterialDetailV1']>('attention_materials', id);
    const versions = paper.imports.filter(item => item.material_id === id).map(item => item.revision).sort();
    expect(detail.material.current_revision).toBe(Math.max(...versions));
    expect(detail.revisions.map(item => item.revision).sort()).toEqual(versions);
  }
  const distillation = read<S['DistillationV1']>('attention_distillations', paperRead.manual_distillation_id);
  expect(distillation.provenance).toMatchObject({ processor: 'human', mode: 'manual' });
  expect(distillation.input_refs.map(item => ({ material_id: item.material_id, revision: item.revision }))).toEqual([{ material_id: proposalRecord.material_id, revision: 1 }]);
  for (const event of paperRead.events) {
    expect(database.prepare('SELECT id,type,aggregate_id FROM attention_outbox WHERE id=?').get(event.id)).toEqual(event);
  }
  const proposal = read<S['ProposalV1']>('workspace_proposals', proposalRecord.proposal_id);
  const approval = read<S['S2ApprovalGrantV1']>('workspace_approvals', proposalRecord.grant_id);
  const packet = read<S['S2ContextPacketV1']>('workspace_context_packets', proposalRecord.context_id);
  expect(proposal.status).toBe('approved');
  expect(proposal.grant_id).toBe(proposalRecord.grant_id);
  expect(approval.scope.actions).toEqual(['read_context']);
  expect(approval.scope.tools).toEqual([]);
  expect(packet.project_id).toBe(project.id);
  expect(packet.grant_id).toBe(approval.id);
  expect(packet.goal.id).toBe(proposalRecord.goal_id);
  const goalRow = database.prepare('SELECT data FROM workspace_goal_revisions WHERE id=? AND revision=?').get(packet.goal.id, packet.goal.revision) as { data: string };
  expect(JSON.parse(goalRow.data)).toEqual(packet.goal);
  expect(packet.sources.map(item => item.reference)).toEqual([{ material_id: proposalRecord.material_id, revision: 1 }]);
  expect(packet.snapshot.head).toMatch(/^[0-9a-f]{40}$/);
  const workspaceEvents: { event_id: string; event_type: string; aggregate_id: string }[] = [];
  for (const [id, type] of [[proposal.id, 'workspace.proposal_approved'], [packet.id, 'workspace.context_packet_built']]) {
    const rows = database.prepare('SELECT event_id,event_type,aggregate_id FROM foundation_outbox WHERE aggregate_id=? AND event_type=?').all(id, type) as typeof workspaceEvents;
    expect(rows).toHaveLength(1);
    for (const event of rows) {
      expect(database.prepare('SELECT event_type,aggregate_id FROM foundation_outbox WHERE event_id=?').get(event.event_id)).toEqual({ event_type: type, aggregate_id: id });
    }
    workspaceEvents.push(...rows);
  }
  const payload = database.prepare('SELECT payload FROM foundation_outbox WHERE event_id=?').get(workspaceEvents[0].event_id) as { payload: string };
  expect((JSON.parse(payload.payload) as { grant_id: string }).grant_id).toBe(approval.id);
  expect(database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions').get()).toEqual({ count: 0 });
  const result = { material_ids: materialIDs, manual_distillation_id: distillation.id, project_id: project.id,
    settings_revision: project.settings.revision, proposal_id: proposal.id, grant_id: approval.id, context_id: packet.id,
    attention_events: paperRead.events, workspace_events: workspaceEvents, native_sessions: 0,
    read_only: true, http_requests: 0, completed_at: new Date().toISOString() };
  await writeFile(resolve(output, 'product-stored-readonly.json'), JSON.stringify(result, null, 2), { flag: 'wx' });
  console.log(JSON.stringify(result));
} finally { database.close(); }
