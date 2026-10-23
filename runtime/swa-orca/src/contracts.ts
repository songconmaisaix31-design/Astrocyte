import { z } from 'zod'
import { workerLimitSchema } from './task-budget'

export const commandSchema = z.strictObject({
  action: z.enum(['status', 'capabilities', 'run.submit', 'run.show', 'run.export', 'role.list', 'role.show', 'role.create', 'role.update', 'role.archive', 'conversation.list', 'conversation.show', 'conversation.create', 'conversation.append', 'handoff.notify', 'controller.open', 'controller.show', 'controller.close', 'controller.recover', 'task.create', 'task.show', 'task.list', 'task.result', 'task.complete', 'worker.start', 'worker.show', 'worker.list', 'worker.read', 'worker.send', 'worker.stop', 'worker.release', 'request.show']),
  project_id: z.string().optional(), controller: z.string().optional(), task_id: z.string().optional(), run_id: z.string().optional(),
  dispatch_id: z.string().optional(), request_id: z.string().optional(), workdir: z.string().optional(),
  spec_file: z.string().optional(), input_file: z.string().optional(), result_file: z.string().optional(),
  output_dir: z.string().optional(), name: z.string().optional(), agent: z.string().optional(), model: z.string().optional(),
  source: z.string().optional(), cursor: z.string().optional(), limit: z.number().int().optional(),
  parent_task_id: z.string().min(1).optional(), handoff_mode: z.enum(['work', 'consult']).optional(), reply_mode: z.enum(['resume', 'poll']).optional(),
  completion_criteria: z.array(z.string().trim().min(1)).optional(), role_id: z.string().min(1).optional(), conversation_id: z.string().min(1).optional(),
  revision: z.number().int().positive().max(Number.MAX_SAFE_INTEGER).optional(),
  role: z.strictObject({ name: z.string().trim().min(1).max(200), responsibilities: z.string().trim().min(1).max(16000), default_agent: z.string().trim().min(1), default_model: z.string().trim().min(1), skills: z.array(z.string().min(1).max(2048)).max(32).nullable(), knowledge: z.array(z.string().min(1).max(2048)).max(32).nullable() }).optional(),
  conversation: z.strictObject({ title: z.string().trim().min(1).max(200), role_id: z.string().min(1).optional() }).optional(), text: z.string().trim().min(1).max(32000).optional()
})
export const contextSchema = z.looseObject({
  id: z.string(), project_id: z.string(), settings_revision: z.number().int().positive(),
  files: z.array(z.object({ path: z.string(), version: z.string(), text: z.string() })).nullable()
})
export const scopeSchema = z.object({
  project_id: z.string(), controller: z.string(), run_id: z.string(), root_task_id: z.string(),
  authority_domain: z.string(), context: z.unknown()
})
export const capabilityRequirementsSchema = z.strictObject({
  settings_revision: z.number().int().positive(), allowed_actions: z.array(z.string()).nullable(),
  allow_directory: z.boolean(), allowed_subdirs: z.array(z.string()).nullable(), grant_actions: z.array(z.string()).nullable(),
  allowed_tools: z.array(z.string()).nullable(), external_model_cli: z.string(), allow_agent_control: z.boolean()
})
export const settingsSourceSchema = z.strictObject({
  kind: z.enum(['native', 'orca-profile']), path: z.string().min(1).optional(), revision: z.number().int().nonnegative().optional()
})
const nativeConfigurationSourceSchema = z.strictObject({ path: z.string().min(1), version: z.string().min(1) })
export const nativePiLaunchPolicySchema = z.strictObject({
  schema: z.literal('swa.native-launch.v1'),
  account: z.strictObject({
    kind: z.literal('pi-native-api-key'), provider: z.string().min(1), source_home: z.string().min(1),
    auth_source: nativeConfigurationSourceSchema, models_source: nativeConfigurationSourceSchema.nullable(),
    private_home: z.string().min(1),
    model_target: z.strictObject({ model_id: z.string().min(1), api: z.enum(['openai-responses', 'openai-completions', 'azure-openai-responses']), base_url: z.url() })
  }),
  tools: z.strictObject({ mode: z.literal('disabled'), native_contract: z.literal('pi-cli-1.0.1') })
})
export const nativePiPerSessionLaunchPolicySchema = z.strictObject({
  schema: z.literal('swa.native-launch.v1'),
  account: z.strictObject({
    kind: z.literal('pi-native-api-key-per-session'), provider: z.string().min(1), source_home: z.string().min(1),
    auth_source: nativeConfigurationSourceSchema, models_source: nativeConfigurationSourceSchema.nullable(),
    private_root: z.string().min(1),
    model_target: z.strictObject({ model_id: z.string().min(1), api: z.enum(['openai-responses', 'openai-completions', 'azure-openai-responses']), base_url: z.url() })
  }),
  tools: z.strictObject({ mode: z.literal('disabled'), native_contract: z.literal('pi-cli-1.0.1') })
})
export const nativeOpenCodeLaunchPolicySchema = z.strictObject({
  schema: z.literal('swa.native-launch.v1'),
  account: z.strictObject({
    kind: z.literal('opencode-native-api-key'), provider: z.literal('deepseek'), source_home: z.string().min(1),
    config_source: nativeConfigurationSourceSchema, auth_source: nativeConfigurationSourceSchema, database_source: nativeConfigurationSourceSchema,
    credential_id: z.string().min(1), credential_updated_at: z.number().int().nonnegative(), private_root: z.string().min(1),
    model_target: z.strictObject({ model_id: z.literal('deepseek-v4-pro'), api: z.literal('opencode-native-deepseek'), base_url: z.literal('https://api.deepseek.com') })
  }),
  tools: z.strictObject({ mode: z.literal('disabled'), native_contract: z.literal('opencode-cli-2.0.31') })
})
export const nativeCodexLaunchPolicySchema = z.strictObject({
  schema: z.literal('swa.native-launch.v1'),
  account: z.strictObject({
    kind: z.literal('codex-native-api-key'), provider: z.literal('openai'), source_home: z.string().min(1),
    auth_source: nativeConfigurationSourceSchema, catalog_source: nativeConfigurationSourceSchema,
    private_root: z.string().min(1),
    model_target: z.strictObject({ model_id: z.literal('gpt-5.5'), api: z.literal('openai-responses'), base_url: z.literal('https://api.openai.com/v1') })
  }),
  tools: z.strictObject({ mode: z.literal('disabled'), native_contract: z.literal('codex-cli-0.163.0') })
})
export const nativeClaudeLaunchPolicySchema = z.strictObject({
  schema: z.literal('swa.native-launch.v1'),
  account: z.strictObject({
    kind: z.literal('claude-native-api-key'), provider: z.literal('anthropic'), source_home: z.string().min(1),
    auth_source: nativeConfigurationSourceSchema, private_root: z.string().min(1),
    model_target: z.strictObject({ model_id: z.literal('claude-opus-5'), api: z.literal('anthropic-messages'), base_url: z.literal('https://api.anthropic.com') })
  }),
  tools: z.strictObject({ mode: z.literal('disabled'), native_contract: z.literal('claude-cli-2.1.284-sdk-0.3.284') })
})
export const nativeLaunchPolicySchema = z.union([nativePiLaunchPolicySchema, nativeOpenCodeLaunchPolicySchema, nativeCodexLaunchPolicySchema, nativeClaudeLaunchPolicySchema, nativePiPerSessionLaunchPolicySchema])
export type NativeLaunchPolicy = z.infer<typeof nativeLaunchPolicySchema>
export const agentSelectionSchema = z.strictObject({
  agent_id: z.string().min(1), adapter: z.string().min(1), execution_mode: z.literal('structured'),
  executable: z.string().min(1), version: z.string().min(1), model: z.string().min(1), provider: z.string().min(1),
  external_target: z.string().min(1), settings_source: settingsSourceSchema,
  settings_revision: z.number().int().nonnegative(), configuration_version: z.string().min(1),
  native_launch: nativeLaunchPolicySchema
})
export type AgentSelection = z.infer<typeof agentSelectionSchema>
export const structuredInputBindingSchema = z.strictObject({
  parent_request: z.string().min(1), child_request: z.string().min(1), dispatch_id: z.string().min(1),
  agent: z.string().min(1), execution_mode: z.literal('structured'), session_id: z.string().min(1),
  operation_id: z.string().min(1), runtime_fence: z.number().int().nonnegative(),
  consumer_generation: z.number().int().nonnegative(), body: z.unknown()
})
export type CapabilityRequirements = z.infer<typeof capabilityRequirementsSchema>
export const observationSchema = z.strictObject({
  status: z.enum(['available', 'unavailable', 'unknown']), reason: z.string().min(1), source: z.string().min(1), checked_at: z.string().optional(),
  basis: z.enum(['declaration', 'configuration', 'probe', 'session', 'result']).optional()
})
export const declaredAgentFeaturesSchema = z.strictObject({ images: z.boolean(), model_switch: z.boolean(), steering: z.enum(['inject', 'queue']) })
export const modelCatalogObservationSchema = z.strictObject({
  origin: z.enum(['live-session', 'probe', 'configured', 'unknown']), availability: z.enum(['verified', 'unverified']),
  models: z.array(z.strictObject({ id: z.string().min(1), label: z.string().min(1), provider: z.string().min(1).optional() })),
  observed_at: z.iso.datetime().optional()
})
export const executionObservationSchema = z.strictObject({
  selection: z.strictObject({ agent: z.string().min(1), adapter: z.string().min(1), version: z.string().min(1), execution_mode: z.literal('structured'), provider: z.string().min(1), model: z.string().min(1), account_source: z.string().min(1) }),
  launch: observationSchema, control: observationSchema, result: observationSchema, authenticated: observationSchema,
  delivery: z.strictObject({ state: z.enum(['sent', 'not_sent', 'unknown']), phase: z.enum(['journaled', 'queued', 'accepted', 'rejected', 'unknown']), request_id: z.string().min(1).optional(), operation_id: z.string().min(1).optional(), source: z.string().min(1) }),
  resume: z.strictObject({ state: z.enum(['not_requested', 'resumed', 'rebuilt', 'failed', 'unknown']), source: z.string().min(1), session_id: z.string().min(1).optional() }),
  model_catalog: modelCatalogObservationSchema
})
export const registeredAgentSchema = z.strictObject({
  schema: z.literal('swa.registered-agent.v1'), id: z.string().min(1), adapter: z.string().min(1),
  execution_mode: z.enum(['structured', 'terminal']), registered: z.literal(true),
  source: z.strictObject({ kind: z.enum(['structured-registry', 'tui-launcher']) }),
  source_revision: z.string().regex(/^[a-f0-9]{40}$/),
  registry: z.strictObject({ orca_version: z.literal('1.4.230'), artifact_sha256: z.string().regex(/^[a-f0-9]{64}$/) }),
  runtime: z.strictObject({ runtime_id: z.string().min(1), authority_domain: z.string().min(1), data_root: z.string().min(1), install_root: z.string().min(1) })
}).superRefine((value, context) => {
  if ((value.execution_mode === 'structured') !== (value.source.kind === 'structured-registry')) context.addIssue({ code: 'custom', message: 'The registered execution mode and native source disagree' })
})
export const agentCapabilitiesSchema = z.strictObject({
  schema: z.literal('swa.capabilities.v1'), source_revision: z.string().regex(/^[a-f0-9]{40}$/),
  registry: z.strictObject({ orca_version: z.literal('1.4.230'), artifact_sha256: z.string().regex(/^[a-f0-9]{64}$/) }),
  discovery: z.strictObject({ host: z.literal('local'), platform: z.string(), observed_at: z.string(), settings_source: settingsSourceSchema }),
  agents: z.array(z.strictObject({
    id: z.string().min(1), adapter: z.string().min(1), execution_mode: z.literal('structured'),
    registered: observationSchema, enabled: observationSchema,
    installed: observationSchema.extend({ executable: z.string().optional(), version: z.string().optional() }),
    configured: observationSchema.extend({ model: z.string().optional(), provider: z.string().optional(), account_source: z.string().optional() }),
    launchable: observationSchema, authenticated: observationSchema.optional(),
    declared_features: declaredAgentFeaturesSchema.optional(), model_catalog: modelCatalogObservationSchema.optional(),
    control: z.strictObject({ start: observationSchema, send: observationSchema, read: observationSchema, stop: observationSchema, resume: observationSchema, tools: observationSchema, images: observationSchema.optional(), model_switch: observationSchema.optional() }),
    requirements: capabilityRequirementsSchema
  }))
})
export const invocationSchema = z.object({
  command: commandSchema,
  caller: z.strictObject({ id: z.string().min(1), kind: z.enum(['agent', 'human']), project_id: z.string().min(1), settings_revision: z.number().int().positive(), grant_expires_at: z.string().optional() }),
  project: z.looseObject({ id: z.string().min(1), root: z.string().min(1) }),
  scope: scopeSchema,
  handoff: z.strictObject({
    schema: z.literal('swa.handoff-resume.v1'), source_task_id: z.string().min(1).max(512), source_dispatch_id: z.string().min(1).max(512),
    parent_task_id: z.string().min(1).max(512), parent_dispatch_id: z.string().min(1).max(512), mail_id: z.string().min(1).max(512)
  }).optional(),
  submission: z.strictObject({
    schema: z.literal('swa.run-submission.v1'), command: commandSchema,
    caller: z.strictObject({ id: z.string().min(1), kind: z.enum(['agent', 'human']), project_id: z.string().min(1), settings_revision: z.number().int().positive() }),
    spec_input: z.strictObject({ path: z.string().min(1), version: z.string().min(1), text: z.string().min(1) }),
    lead_input: z.strictObject({ path: z.string().min(1), version: z.string().min(1), text: z.string().min(1) }),
    context: contextSchema, selection: agentSelectionSchema,
    task_request_id: z.uuid(), worker_request_id: z.uuid()
  }).optional(),
  selection: agentSelectionSchema.optional(),
  spec: z.looseObject({ goal: z.string().min(1), source_repo: z.string().min(1), source_revision: z.string().min(1), success_criteria: z.array(z.string().min(1)).min(1), max_workers: workerLimitSchema, initial_workers: workerLimitSchema.optional(), role_id: z.string().min(1).optional(), role_revision: workerLimitSchema.optional(), conversation_id: z.string().min(1).optional() }).optional(),
  context: contextSchema.optional(),
  input: z.object({ path: z.string(), version: z.string().min(1), text: z.string().min(1) }).optional()
})
export type Invocation = z.infer<typeof invocationSchema>
export type Scope = z.infer<typeof scopeSchema>
export const helperInput = z.object({ type: z.literal('swa_controller_action'), request_id: z.string().min(1), action: z.string(), invocation: invocationSchema.optional() })
export const helperReply = z.object({ type: z.literal('swa_controller_reply'), request_id: z.string(), controller: z.string(), pid: z.number().int().positive(), result: z.unknown().optional(), error: z.object({ code: z.string(), message: z.string() }).optional() })
export const nativeObject = (input: unknown): Record<string, unknown> => z.record(z.string(), z.unknown()).parse(input)

export function nativeError(error: unknown) {
  const candidate = error instanceof Error ? error : new Error('native_operation_failed')
  const withDetails = z.object({ code: z.string().optional(), data: z.unknown().optional() }).parse(candidate)
  return { code: withDetails.code ?? 'operation_unknown', message: candidate.message, data: withDetails.data, retryable: false, required_action: 'inspect_original_request' }
}
