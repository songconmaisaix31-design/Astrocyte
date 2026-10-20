import { z } from 'zod'

export const commandSchema = z.strictObject({
  action: z.enum(['status', 'capabilities', 'controller.open', 'controller.show', 'controller.close', 'task.create', 'task.show', 'task.list', 'task.result', 'task.complete', 'worker.start', 'worker.show', 'worker.list', 'worker.read', 'worker.send', 'worker.stop', 'worker.release', 'request.show']),
  project_id: z.string().optional(), controller: z.string().optional(), task_id: z.string().optional(),
  dispatch_id: z.string().optional(), request_id: z.string().optional(), workdir: z.string().optional(),
  spec_file: z.string().optional(), input_file: z.string().optional(), result_file: z.string().optional(),
  output_dir: z.string().optional(), name: z.string().optional(), agent: z.string().optional(), model: z.string().optional(),
  source: z.string().optional(), cursor: z.string().optional(), limit: z.number().int().optional()
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
export const agentSelectionSchema = z.strictObject({
  agent_id: z.string().min(1), adapter: z.string().min(1), execution_mode: z.literal('structured'),
  executable: z.string().min(1), version: z.string().min(1), model: z.string().min(1), provider: z.string().min(1),
  external_target: z.string().min(1), settings_source: settingsSourceSchema,
  settings_revision: z.number().int().nonnegative(), configuration_version: z.string().min(1)
})
export type AgentSelection = z.infer<typeof agentSelectionSchema>
export type CapabilityRequirements = z.infer<typeof capabilityRequirementsSchema>
export const observationSchema = z.strictObject({
  status: z.enum(['available', 'unavailable', 'unknown']), reason: z.string().min(1), source: z.string().min(1), checked_at: z.string().optional()
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
    launchable: observationSchema,
    control: z.strictObject({ start: observationSchema, send: observationSchema, read: observationSchema, stop: observationSchema, resume: observationSchema, tools: observationSchema }),
    requirements: capabilityRequirementsSchema
  }))
})
export const invocationSchema = z.object({
  command: commandSchema,
  caller: z.object({ id: z.string().min(1), kind: z.enum(['agent', 'human']), project_id: z.string().min(1), settings_revision: z.number().int().positive() }),
  project: z.looseObject({ id: z.string().min(1), root: z.string().min(1) }),
  scope: scopeSchema,
  selection: agentSelectionSchema.optional(),
  spec: z.looseObject({ goal: z.string().min(1), source_repo: z.string().min(1), source_revision: z.string().min(1), success_criteria: z.array(z.string().min(1)).min(1), max_workers: z.literal(2) }).optional(),
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
