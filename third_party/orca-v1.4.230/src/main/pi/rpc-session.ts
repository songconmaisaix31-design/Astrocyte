import type {
  AgentSessionOptionsResult,
  AgentSessionSlashCommand
} from '../../shared/agent-session-wire'
import type {
  StructuredAgentSessionAcquireInput,
  StructuredAgentSessionLifecycleEvent,
  AgentSessionDispatchOutcome
} from '../native-chat/agent-session-wire/structured-agent-session-adapter'
import type { StructuredAgentSessionLogger } from '../native-chat/agent-session-wire/structured-agent-session-logger'
import { providerTimelineSink } from '../native-chat/agent-session-timeline/provider-timeline-plan'
import { JsonlRpcTimelineLane } from '../jsonl-rpc/timeline-lane'
import {
  JsonlRpcAgentConnection,
  type JsonlRpcAgentConnectionOptions
} from '../jsonl-rpc/agent-connection'
import type { ProviderProcessLaunch } from '../provider-process/provider-process-launch'
import { PiRpcTurns } from './rpc-turns'
import { PiRpcDialogCallbacks } from './rpc-dialog-callbacks'
import { piRpcStateSchema } from './rpc-protocol'
import {
  withLiveCatalogListing,
  type AgentModelCatalogLiveListing
} from '../native-chat/agent-model-catalog/agent-model-catalog-entry'

export type PiRpcConnection = Pick<
  JsonlRpcAgentConnection,
  | 'request'
  | 'send'
  | 'close'
  | 'pauseReading'
  | 'resumeReading'
  | 'pid'
  | 'closed'
  | 'rootVerdict'
  | 'processless'
  | 'lastCloseResult'
  | 'onExit'
>
export type PiRpcSessionDeps = {
  openConnection?: (
    launch: ProviderProcessLaunch,
    handlers: JsonlRpcAgentConnectionOptions
  ) => PiRpcConnection
  onLifecycle: (event: StructuredAgentSessionLifecycleEvent) => void
  onSettled: (input: {
    sessionId: string
    clientMessageId: string
    fence: number
    outcome: AgentSessionDispatchOutcome
  }) => void
  onIdle: (input: { sessionId: string }) => void
  logger: StructuredAgentSessionLogger
}

/** One child owns its transport, dialect, dialogs and timeline, including shutdown retries. */
export class PiRpcSession {
  readonly connection: PiRpcConnection
  readonly lane: JsonlRpcTimelineLane
  readonly turns: PiRpcTurns
  readonly dialogs: PiRpcDialogCallbacks
  readonly selected = new Map<string, string>()
  private startCatalogListing?: AgentModelCatalogLiveListing
  readonly skipped: string[] = []
  commands?: AgentSessionSlashCommand[]
  options?: AgentSessionOptionsResult
  requestedClose = false
  private publishedExit = false
  private failedCause?: Error
  private exitDelivery: Promise<void> = Promise.resolve()
  private releaseAfterExit?: () => void

  constructor(
    readonly input: StructuredAgentSessionAcquireInput,
    readonly generation: string,
    launch: ProviderProcessLaunch,
    private readonly deps: PiRpcSessionDeps,
    /** A new Pi session starts on its config's model; a resumed or forked one keeps its own. */
    readonly resolvesConfig = false
  ) {
    const sink = input.events && providerTimelineSink(input.events)
    if (!sink) {
      throw new Error('Pi structured chat requires the shared timeline sink')
    }
    this.lane = new JsonlRpcTimelineLane({
      sink,
      sessionId: input.identity.sessionId,
      agent: 'pi',
      generation,
      namespace: input.identity.sessionId,
      pauseReading: () => this.connection?.pauseReading(),
      resumeReading: () => this.connection?.resumeReading(),
      onFailed: (reason) => this.fail(new Error(reason)),
      onInputAccepted: (clientMessageId) =>
        this.settle(clientMessageId, {
          state: 'accepted',
          providerIdentity: { provider: 'orca', clientMessageId }
        })
    })
    this.turns = new PiRpcTurns({
      lane: this.lane,
      generation,
      send: (frame) => this.connection.send(frame),
      request: (command, params, options) => this.connection.request(command, params, options),
      settled: (id, outcome) => this.settle(id, outcome),
      idle: () => {
        this.dialogs.cancelAll()
        deps.onIdle({ sessionId: input.identity.sessionId })
      },
      failed: (error) => this.fail(error),
      diagnostic: (error) =>
        deps.logger.warn('Pi context usage could not be read', {
          scope: 'pi-context-usage',
          sessionId: input.identity.sessionId,
          error
        })
    })
    this.dialogs = new PiRpcDialogCallbacks(
      this.lane,
      (frame) => this.connection.send(frame),
      (error) => this.fail(error)
    )
    this.connection = (
      deps.openConnection ?? ((spec, handlers) => new JsonlRpcAgentConnection(spec, handlers))
    )(launch, {
      allowedCommands: new Set(['get_state', 'get_messages', 'get_session_stats', 'get_last_assistant_text', 'prompt', 'abort']),
      ...(input.onOutput ? { onOutput: input.onOutput } : {}),
      onRecord: (frame) => {
        if (frame.type === 'extension_ui_request' || frame.type.startsWith('tool_execution_')) {
          this.fail(new Error('SWA received an unapproved Pi tool or extension request'))
          return
        }
        this.turns.receive(frame)
      },
      onClose: (error) => this.fail(error),
      onDiagnostic: (message) =>
        deps.logger.warn(message, { scope: 'pi-rpc', sessionId: input.identity.sessionId }),
      onExit: (error) => {
        this.exitDelivery = this.exit(error)
          .catch((cause: unknown) =>
            deps.logger.error('Pi exit publication failed', {
              scope: 'pi-exit',
              sessionId: input.identity.sessionId,
              error: cause
            })
          )
          .finally(() => this.releaseAfterExit?.())
      }
    })
  }

  async start(): Promise<string> {
    const state = piRpcStateSchema.parse(await this.connection.request('get_state'))
    this.lane.apply(this.turns.context.setModel(state.model ?? undefined, Date.now()))
    this.options = fixedOptions(state)
    for (const [key, value] of Object.entries(this.input.options ?? {})) {
      if ((key !== 'model' || value !== 'openai-codex/gpt-6.1-sol') &&
          (key !== 'effort' || value !== 'xhigh')) throw new Error('SWA Pi options are fixed')
    }
    return state.sessionFile
  }

  /** What this child listed at its start, with what its config resolved; the host saves it once. */
  get startListing(): AgentModelCatalogLiveListing | undefined {
    return this.startCatalogListing
  }

  /** What the child reports now; only the start says what its config resolved. */
  async readOptions() {
    return withLiveCatalogListing(fixedOptions(piRpcStateSchema.parse(await this.connection.request('get_state'))))
  }

  async close(requested = true): ReturnType<PiRpcConnection['close']> {
    this.requestedClose ||= requested
    this.dialogs.cancelAll()
    return this.connection.close()
  }

  fail(error: Error): void {
    this.failedCause ??= error
    this.deps.logger.error('Pi RPC session failed', {
      scope: 'pi-rpc',
      sessionId: this.input.identity.sessionId,
      error
    })
    void this.close(false)
      .then((result) => {
        if (result.root !== 'exited') {
          this.deps.logger.warn('Pi process exit is unverifiable', {
            scope: 'pi-stop',
            sessionId: this.input.identity.sessionId
          })
        }
      })
      .catch((cause: unknown) =>
        this.deps.logger.error('Pi process cleanup failed', {
          scope: 'pi-stop',
          sessionId: this.input.identity.sessionId,
          error: cause
        })
      )
  }

  drainObservedExit(): Promise<void> {
    return this.exitDelivery
  }

  retire(): Promise<void> {
    return new Promise((resolve) => {
      this.releaseAfterExit = () => {
        this.releaseAfterExit = undefined
        void this.lane.drained().then(() => {
          this.lane.dispose()
          resolve()
        })
      }
      if (this.publishedExit) {
        void this.exitDelivery.then(() => this.releaseAfterExit?.())
      }
    })
  }

  private async exit(error: Error): Promise<void> {
    if (this.publishedExit) {
      return
    }
    this.publishedExit = true
    this.dialogs.cancelAll()
    this.turns.end()
    this.lane.finalize()
    this.lane.apply([
      { type: 'session.ended', verdict: { state: 'interrupted', completedAt: Date.now() } }
    ])
    this.lane.flush()
    let timer: ReturnType<typeof setTimeout> | undefined
    try {
      await Promise.race([
        this.lane.drained(),
        new Promise<void>((resolve) => {
          timer = setTimeout(() => {
            this.deps.logger.warn('Pi final timeline did not drain before exit publication', {
              scope: 'pi-exit-drain',
              sessionId: this.input.identity.sessionId
            })
            resolve()
          }, 2_000)
          timer.unref()
        })
      ])
    } finally {
      clearTimeout(timer)
    }
    this.deps.onLifecycle({
      type: 'ended',
      sessionId: this.input.identity.sessionId,
      fence: this.input.fence,
      acquisitionGeneration: this.generation,
      observedAt: Date.now(),
      cause: this.requestedClose ? 'requested-close' : 'unexpected-exit',
      reason: (this.failedCause ?? error).message
    })
  }

  private settle(clientMessageId: string, outcome: AgentSessionDispatchOutcome): void {
    this.deps.onSettled({
      sessionId: this.input.identity.sessionId,
      clientMessageId,
      fence: this.input.fence,
      outcome
    })
  }
}

function fixedOptions(state: ReturnType<typeof piRpcStateSchema.parse>): AgentSessionOptionsResult {
  if (state.model?.provider !== 'openai-codex' || state.model.id !== 'gpt-6.1-sol' || state.thinkingLevel !== 'xhigh') {
    throw new Error('SWA Pi model or thinking level changed')
  }
  const id = `${state.model.provider}/${state.model.id}`
  return {
    models: [{ id, label: state.model.name || id, isDefault: true, efforts: [{ value: 'xhigh', label: 'xhigh' }] }],
    current: { model: id, effort: state.thinkingLevel, confirmed: ['model', 'effort'] }
  }
}
