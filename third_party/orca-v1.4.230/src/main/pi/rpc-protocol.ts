import { z } from 'zod'
import { parseTree, type Node as JsonNode, type ParseError } from 'jsonc-parser'

export function parseStrictPiJson(line: string): unknown {
  const errors: ParseError[] = []
  const root = parseTree(line, errors, { allowTrailingComma: false, disallowComments: true, allowEmptyContent: false })
  if (!root || errors.length) throw new Error('native_pi_invalid_json')
  function unique(node: JsonNode) {
    if (node.type === 'object') {
      const keys = new Set<string>()
      for (const property of node.children ?? []) { const key = property.children?.[0]?.value; if (typeof key !== 'string' || keys.has(key)) throw new Error('native_pi_duplicate_json_key'); keys.add(key) }
    }
    for (const child of node.children ?? []) unique(child)
  }
  unique(root)
  // 库的getNodeValue返回null-prototype对象；原生JSON与现有语义比较使用普通JSON对象。
  try { return JSON.parse(line) as unknown } catch { throw new Error('native_pi_invalid_json') }
}

export const piRpcModelSchema = z.looseObject({
  id: z.string(),
  provider: z.string(),
  name: z.string().optional(),
  contextWindow: z.number().nonnegative().optional(),
  reasoning: z.boolean().optional(),
  input: z.array(z.string()).optional()
})
export type PiRpcModel = z.infer<typeof piRpcModelSchema>

export const piRpcStateSchema = z.looseObject({
  model: piRpcModelSchema.nullish(),
  thinkingLevel: z.string().optional(),
  sessionFile: z.string().min(1),
  sessionId: z.string().optional(),
  isStreaming: z.boolean(),
  isCompacting: z.boolean(),
  autoCompactionEnabled: z.boolean().optional(),
  pendingMessageCount: z.number().int().nonnegative().optional()
})
export type PiRpcState = z.infer<typeof piRpcStateSchema>

export function piRpcIdle(state: PiRpcState): boolean {
  return !state.isStreaming && !state.isCompacting && (state.pendingMessageCount ?? 0) === 0
}

export const piRpcContentSchema = z.looseObject({
  type: z.string(),
  text: z.string().optional(),
  thinking: z.string().optional()
})
export const piRpcUsageSchema = z.object({
  input: z.number().nonnegative(),
  output: z.number().nonnegative(),
  cacheRead: z.number().nonnegative(),
  cacheWrite: z.number().nonnegative()
})
export const piRpcMessageSchema = z.looseObject({
  role: z.string(),
  content: z.union([z.string(), z.array(piRpcContentSchema)]),
  stopReason: z.string().optional(),
  errorMessage: z.string().optional(),
  usage: piRpcUsageSchema.optional(),
  timestamp: z.number().optional()
})
export const piRpcMessageEventSchema = z.object({
  type: z.enum(['message_start', 'message_update', 'message_end']),
  message: piRpcMessageSchema.optional(),
  usage: piRpcUsageSchema.optional(),
  assistantMessageEvent: z
    .looseObject({
      type: z.string(),
      delta: z.string().optional(),
      contentIndex: z.number().int().nonnegative().optional(),
      partial: piRpcMessageSchema.optional()
    })
    .optional()
})
export type PiRpcMessageEvent = z.infer<typeof piRpcMessageEventSchema>

export const piRpcToolEventSchema = z.object({
  type: z.enum(['tool_execution_start', 'tool_execution_update', 'tool_execution_end']),
  toolCallId: z.string(),
  toolName: z.string(),
  args: z.unknown().optional(),
  partialResult: z.unknown().optional(),
  result: z.unknown().optional(),
  isError: z.boolean().optional()
})
export type PiRpcToolEvent = z.infer<typeof piRpcToolEventSchema>

export const piRpcPromptReplySchema = z.object({
  type: z.literal('response'),
  command: z.literal('prompt'),
  success: z.boolean(),
  error: z.string().optional(),
  data: z
    .object({
      disposition: z.enum(['started', 'queued', 'handled']).optional(),
      agentInvoked: z.boolean().optional()
    })
    .optional()
})

export const piRpcDialogSchema = z.discriminatedUnion('method', [
  z.object({
    id: z.union([z.string(), z.number()]),
    method: z.literal('confirm'),
    title: z.string(),
    message: z.string().optional()
  }),
  z.object({
    id: z.union([z.string(), z.number()]),
    method: z.literal('select'),
    title: z.string(),
    message: z.string().optional(),
    options: z.array(z.string()).min(1).max(64)
  }),
  z.object({
    id: z.union([z.string(), z.number()]),
    method: z.enum(['input', 'editor']),
    title: z.string(),
    message: z.string().optional(),
    placeholder: z.string().optional(),
    prefill: z.string().optional()
  })
])
export type PiRpcDialog = z.infer<typeof piRpcDialogSchema>

const count = z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER)
const amount = z.number().nonnegative().finite()
export const strictPiThinkingLevelSchema = z.enum(['off', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max'])
const rates = { input: amount, output: amount, cacheRead: amount, cacheWrite: amount }
const compatBooleanFields = ['supportsStore', 'supportsDeveloperRole', 'supportsReasoningEffort', 'supportsUsageInStreaming', 'supportsFinishReason', 'requiresToolResultName', 'requiresAssistantAfterToolResult', 'requiresThinkingAsText', 'requiresReasoningContentOnAssistantMessages', 'supportsOpenAIGrammarTools', 'supportsStrictMode', 'sendSessionAffinityHeaders', 'supportsLongCacheRetention', 'supportsMaxOutputTokens'] as const
const strictCompat = z.strictObject({
  ...Object.fromEntries(compatBooleanFields.map(key => [key, z.boolean().optional()])),
  maxTokensField: z.enum(['max_completion_tokens', 'max_tokens']).optional(),
  thinkingFormat: z.enum(['openai', 'openrouter', 'deepseek', 'together', 'baseten', 'zai', 'qwen', 'chat-template', 'qwen-chat-template', 'string-thinking', 'ant-ling']).optional(),
  chatTemplateKwargs: z.strictObject({}).optional(), chatTemplateArgs: z.strictObject({}).optional(),
  cacheControlFormat: z.literal('anthropic').optional(), sessionAffinityFormat: z.enum(['headers', 'safety_identifier']).optional(),
  vllmPriority: z.number().finite().optional(), allowedFallbackModels: z.tuple([]).optional(),
  openRouterRouting: z.strictObject({}).optional(), vercelGatewayRouting: z.strictObject({}).optional()
})
export const strictPiModelSchema = z.strictObject({
  id: z.string().min(1), name: z.string(), api: z.string().min(1), provider: z.string().min(1), baseUrl: z.string().url(),
  input: z.array(z.enum(['text', 'image'])), type: z.literal('chat').optional(), reasoning: z.boolean(),
  inputLimits: z.strictObject({ maxRequestBytes: amount.optional(), images: z.strictObject({
    resize: z.strictObject({ maxWidth: amount.optional(), maxHeight: amount.optional(), maxBytes: amount.optional(), jpegQuality: amount.optional() }).optional(),
    maxPerMessage: amount.optional(), maxPerRequest: amount.optional()
  }).optional() }).optional(),
  cost: z.strictObject({ ...rates, tiers: z.array(z.strictObject({ ...rates, inputTokensAbove: amount })).optional() }),
  headers: z.record(z.string(), z.string()).optional(),
  thinkingLevelMap: z.strictObject(Object.fromEntries(strictPiThinkingLevelSchema.options.map(key => [key, z.string().nullable().optional()]))).optional(),
  promptCache: z.strictObject({ short: amount.optional(), long: amount.optional() }).optional(),
  contextWindow: amount, maxTokens: amount, samplingParams: z.strictObject({}).optional(), compat: strictCompat.optional()
})
export type StrictPiModel = z.infer<typeof strictPiModelSchema>
export const strictPiUsageSchema = z.strictObject({ ...rates, totalTokens: amount, cost: z.strictObject({ ...rates, total: amount }), cacheWrite1h: amount.optional(), reasoning: amount.optional() })
const strictText = z.strictObject({ type: z.literal('text'), text: z.string(), textSignature: z.string().optional() })
const strictThinking = z.strictObject({ type: z.literal('thinking'), thinking: z.string(), thinkingSignature: z.string().optional(), redacted: z.boolean().optional() })
const strictUser = z.strictObject({ role: z.literal('user'), content: z.union([z.string(), z.array(strictText)]), timestamp: amount })
const strictSystem = z.strictObject({ role: z.literal('system'), content: z.union([z.string(), z.array(strictText)]), timestamp: amount,
  sections: z.record(z.string(), z.string().nullable()).optional(), toolsAdded: z.tuple([]).optional(), toolsRemoved: z.tuple([]).optional() })
const strictAssistant = z.strictObject({
  role: z.literal('assistant'), content: z.array(z.union([strictText, strictThinking])), api: z.string(), provider: z.string(), model: z.string(),
  usage: strictPiUsageSchema, stopReason: z.enum(['pending', 'stop', 'length', 'aborted']), timestamp: amount,
  responseModel: z.string().optional(), responseId: z.string().optional(), providerThinkingLevel: z.string().optional(),
  thinkingLevel: strictPiThinkingLevelSchema.optional(), rawStopReason: z.enum(['stop', 'end', 'length']).optional(), endTurn: z.boolean().optional(),
  diagnostics: z.tuple([]).optional(), errorMessage: z.enum(['Request aborted', 'Request was aborted']).optional()
}).superRefine((message, ctx) => {
  if (message.stopReason === 'aborted' ? message.errorMessage === undefined : message.errorMessage !== undefined) ctx.addIssue({ code: 'custom', message: 'native_pi_error_unreviewed' })
})
export const strictPiMessageSchema = z.union([strictUser, strictSystem, strictAssistant])
export type StrictPiMessage = z.infer<typeof strictPiMessageSchema>
const entryCommon = { id: z.string().regex(/^(?:[a-f0-9]{8}|[a-f0-9]{8}-[a-f0-9-]{27})$/), parentId: z.string().nullable(), timestamp: z.iso.datetime() }
export const strictPiNativeHeaderSchema = z.strictObject({ type: z.literal('session'), version: z.literal(3), id: z.uuidv7(), timestamp: z.iso.datetime(), cwd: z.string().min(1) })
export const strictPiNativeEntrySchema = z.discriminatedUnion('type', [
  z.strictObject({ ...entryCommon, type: z.literal('message'), message: strictPiMessageSchema }),
  z.strictObject({ ...entryCommon, type: z.literal('thinking_level_change'), thinkingLevel: strictPiThinkingLevelSchema }),
  z.strictObject({ ...entryCommon, type: z.literal('model_change'), provider: z.string(), modelId: z.string() }),
  z.strictObject({ ...entryCommon, type: z.literal('session_info'), name: z.string().optional() }),
  z.strictObject({ ...entryCommon, type: z.literal('label'), targetId: z.string(), label: z.string().optional() })
])
export type StrictPiNativeEntry = z.infer<typeof strictPiNativeEntrySchema>
export const strictPiStateSchema = z.strictObject({
  model: strictPiModelSchema.optional(), thinkingLevel: strictPiThinkingLevelSchema, isStreaming: z.boolean(), isCompacting: z.boolean(),
  steeringMode: z.enum(['all', 'one-at-a-time']), followUpMode: z.enum(['all', 'one-at-a-time']), sessionFile: z.string().optional(),
  sessionId: z.uuidv7(), sessionName: z.string().optional(), autoCompactionEnabled: z.boolean(), messageCount: count, pendingMessageCount: count
})
const strictDelta = z.discriminatedUnion('type', [
  z.strictObject({ type: z.literal('text_start'), contentIndex: count }), z.strictObject({ type: z.literal('thinking_start'), contentIndex: count }),
  z.strictObject({ type: z.literal('text_delta'), contentIndex: count, delta: z.string() }), z.strictObject({ type: z.literal('thinking_delta'), contentIndex: count, delta: z.string() }),
  z.strictObject({ type: z.literal('text_end'), contentIndex: count, content: z.string() }), z.strictObject({ type: z.literal('thinking_end'), contentIndex: count, content: z.string() })
])
const strictStats = z.strictObject({ sessionId: z.uuidv7(), sessionFile: z.string().optional(), userMessages: count, assistantMessages: count,
  toolCalls: z.literal(0), toolResults: z.literal(0), totalMessages: count,
  tokens: z.strictObject({ input: amount, output: amount, cacheRead: amount, cacheWrite: amount, total: amount }), cost: amount,
  contextUsage: z.strictObject({ tokens: amount.nullable(), contextWindow: amount, percent: amount.nullable() }).optional() })
const response = <const C extends string, T extends z.ZodType>(command: C, data: T) => z.strictObject({ id: z.string().min(1), type: z.literal('response'), command: z.literal(command), success: z.literal(true), data })
const noDataResponse = <const C extends string>(command: C) => z.strictObject({ id: z.string().min(1), type: z.literal('response'), command: z.literal(command), success: z.literal(true) })
export const strictPiRawSchema = z.union([
  ...(['agent_start', 'turn_start', 'agent_settled'] as const).map(type => z.strictObject({ type: z.literal(type) })),
  z.strictObject({ type: z.literal('queue_update'), steering: z.array(z.string()), followUp: z.array(z.string()) }),
  z.strictObject({ type: z.literal('message_start'), message: strictPiMessageSchema }), z.strictObject({ type: z.literal('message_end'), message: strictPiMessageSchema }),
  z.strictObject({ type: z.literal('message_update'), usage: strictPiUsageSchema, assistantMessageEvent: strictDelta }),
  z.strictObject({ type: z.literal('turn_end'), message: strictPiMessageSchema, toolResults: z.tuple([]) }),
  z.strictObject({ type: z.literal('agent_end'), messages: z.array(strictPiMessageSchema), willRetry: z.literal(false) }),
  z.strictObject({ type: z.literal('thinking_level_changed'), level: strictPiThinkingLevelSchema }),
  z.strictObject({ type: z.literal('session_info_changed'), name: z.string().optional() }),
  z.strictObject({ type: z.literal('entry_appended'), entry: strictPiNativeEntrySchema }),
  response('prompt', z.strictObject({ disposition: z.enum(['started', 'queued']) })),
  ...(['abort', 'set_auto_compaction', 'set_auto_retry', 'set_thinking_level'] as const).map(noDataResponse),
  response('get_state', strictPiStateSchema), response('set_model', strictPiModelSchema),
  response('get_available_models', z.strictObject({ models: z.array(strictPiModelSchema) })),
  response('get_available_thinking_levels', z.strictObject({ levels: z.array(strictPiThinkingLevelSchema) })),
  response('get_last_assistant_text', z.strictObject({ text: z.string().nullable() })), response('get_session_stats', strictStats)
])
export type StrictPiRaw = z.infer<typeof strictPiRawSchema>

export const strictPiOutboundSchema = z.union([
  z.strictObject({ type: z.literal('prompt'), id: z.string().min(1), message: z.string().min(1), images: z.tuple([]).optional(), streamingBehavior: z.enum(['steer', 'followUp']).optional() }),
  ...(['abort', 'get_state', 'get_available_models', 'get_available_thinking_levels', 'get_last_assistant_text', 'get_session_stats'] as const).map(type => z.strictObject({ type: z.literal(type), id: z.string().min(1) })),
  ...(['set_auto_compaction', 'set_auto_retry'] as const).map(type => z.strictObject({ type: z.literal(type), id: z.string().min(1), enabled: z.literal(false) })),
  z.strictObject({ type: z.literal('set_model'), id: z.string().min(1), provider: z.string().min(1), modelId: z.string().min(1) }),
  z.strictObject({ type: z.literal('set_thinking_level'), id: z.string().min(1), level: strictPiThinkingLevelSchema })
])
