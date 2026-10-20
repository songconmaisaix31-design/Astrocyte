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
export const invocationSchema = z.object({
  command: commandSchema,
  caller: z.object({ id: z.string().min(1), kind: z.enum(['agent', 'human']), project_id: z.string().min(1), settings_revision: z.number().int().positive() }),
  project: z.looseObject({ id: z.string().min(1), root: z.string().min(1) }),
  scope: scopeSchema,
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
