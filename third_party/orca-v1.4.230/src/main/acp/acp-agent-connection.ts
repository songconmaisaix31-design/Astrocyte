import type { PipedProcessSpawner } from '@orca/process-host/process-spec'
import { spawnProcess } from '@orca/process-host'
import {
  spawnManagedProviderProcess,
  type ManagedProviderProcess,
  type ProviderProcessExit
} from '../provider-process/managed-provider-process'
import type { ProviderProcessLaunch, SwaOpenCodeToolsPolicy } from '../provider-process/provider-process-launch'
import { SwaProviderProcessTree } from '../provider-process/swa-provider-process-tree'
import { AcpAuthRequiredError, AcpConnectionClosedError, AcpInputAdmissionError } from './acp-errors'
import { resolveAcpPeerOptions, type AcpPeerOptions } from './acp-peer-limits'
import { AcpSessionRuntime, type AcpSessionRuntimeOptions } from './acp-session-runtime'

/** How an agent that signs in through a browser says so on its output. */
export type AcpSignInWatch = {
  /** A sign-in line's valid link (the connection ends signed out), null for a malformed one,
   *  undefined for any other line. */
  read(line: string, source: 'stdout' | 'stderr'): string | null | undefined
  /** The agent's answer on its sign-in: a session started, or it asked to sign in. */
  answered?(signedIn: boolean): void
  /** Receives each valid link instead, and the connection stays open for the sign-in to finish. */
  link?(link: string): void
}

// A full sign-in link with its framing; the stderr tail kept for diagnostics stays smaller.
const SIGN_IN_LINE_MAX_CHARS = 32_768

export type AcpAgentConnectionOptions = Omit<AcpSessionRuntimeOptions, 'peer'> & {
  peer?: Omit<AcpPeerOptions, 'closeOnInputEnd'>
  /** Process exit evidence, including expected closes and processless spawn failures. */
  onExit?: (error: Error, context: { expected: boolean; exit: ProviderProcessExit }) => void
  /** Any stdout or stderr chunk from the child. */
  onOutput?: () => void
  signIn?: AcpSignInWatch
}

/** The execution host owns the child and protocol lifetime; the adapter owns turns and Stop. */
export function createAcpAgentConnection(
  launch: ProviderProcessLaunch,
  options: AcpAgentConnectionOptions = {},
  spawnImpl: PipedProcessSpawner = spawnProcess
): AcpAgentConnection {
  return new AcpAgentConnection(launch, options, spawnImpl)
}

export class AcpAgentConnection extends AcpSessionRuntime {
  private readonly managed: ManagedProviderProcess
  private readonly lifecycle: { closing: boolean; error?: Error }
  private readonly diagnoseLifecycle: (message: string) => void
  private readonly tree?: SwaProviderProcessTree
  private readonly swaPolicy?: SwaOpenCodeToolsPolicy
  private swaSessionId?: string
  readonly spawned: Promise<void>

  constructor(
    launch: ProviderProcessLaunch,
    options: AcpAgentConnectionOptions = {},
    spawnImpl: PipedProcessSpawner = spawnProcess
  ) {
    // Validate before spawning so invalid limits cannot leave an unowned child.
    const peer = resolveAcpPeerOptions({ ...options.peer, closeOnInputEnd: false })
    const { signIn } = options
    const tree = launch.swaJobIPC ? new SwaProviderProcessTree() : undefined
    if (launch.swaOpenCodePolicy && !tree) throw new Error('The restricted OpenCode process has no Job')
    if (tree) launch.swaOpenCodePolicy?.attach(tree)
    let signedOut = (): void => {}
    let unreadable = (): void => {}
    // A sign-in line is never kept or logged; only a whole, valid one is an answer.
    const signInLine = (line: string, source: 'stdout' | 'stderr', cut = false): boolean => {
      const link = signIn?.read(line, source)
      if (link === undefined) {
        return false
      }
      if (link !== null && !cut) {
        if (signIn?.link) {
          signIn.link(link)
        } else {
          signedOut()
        }
      } else if (source === 'stdout') {
        unreadable()
      }
      return true
    }
    let managed: ManagedProviderProcess
    try { managed = spawnManagedProviderProcess(launch, {
      spawnImpl,
      site: 'acp-agent-teardown',
      ...(options.onOutput ? { onOutput: options.onOutput } : {}),
      ...(launch.swaOpenCodePolicy
        ? { onStderrLine: (line: string, cut: boolean) => launch.swaOpenCodePolicy!.stderr(line, cut), stderrLineMaxChars: SIGN_IN_LINE_MAX_CHARS }
        : signIn
        ? {
            onStderrLine: (line: string, cut: boolean) => signInLine(line, 'stderr', cut),
            stderrLineMaxChars: SIGN_IN_LINE_MAX_CHARS
          }
        : {})
    }) } catch (error) { tree?.closeBeforeSpawnFailure(); throw error }
    tree?.bind(managed.child, () => ({ observed: managed.rootExitObserved, processless: managed.processless }), true)
    const lifecycle: { closing: boolean; error?: Error } = { closing: false }
    const diagnose = (message: string): void => {
      try {
        options.onDiagnostic?.(message)
      } catch {
        /* Diagnostics cannot interrupt cleanup. */
      }
    }
    super(managed.child.stdout, managed.child.stdin, {
      ...options,
      peer,
      onOtherLine: (line, cut) => signInLine(line, 'stdout', cut),
      onAuthAnswered: (signedIn) => signIn?.answered?.(signedIn),
      onClose: (error) => {
        lifecycle.error ??= error
        if (lifecycle.closing || !tree && managed.rootVerdict === 'exited') {
          return
        }
        void managed.close(tree).then(
          (result) => {
            if (result.root !== 'exited' || tree && result.tree !== 'exited') {
              diagnose('ACP agent exit was not proven after transport failure')
            }
          },
          (failure: unknown) => diagnose(`ACP agent cleanup failed: ${String(failure)}`)
        )
        options.onClose?.(error)
      }
    })
    this.managed = managed
    this.tree = tree
    this.swaPolicy = launch.swaOpenCodePolicy
    this.lifecycle = lifecycle
    this.diagnoseLifecycle = diagnose
    const { child } = managed
    this.spawned = new Promise<void>((resolve) => {
      if (child.pid !== undefined) {
        resolve()
        return
      }
      const done = (): void => {
        child.removeListener('spawn', done)
        child.removeListener('error', done)
        resolve()
      }
      child.once('spawn', done)
      child.once('error', done)
    }).then(() => tree?.ready())
    void this.spawned.catch(() => undefined)
    if (this.swaPolicy) this.subscribe(event => {
      try { this.swaPolicy!.trace('session/update', event) }
      catch { void this.close(new Error('The restricted OpenCode trace could not be retained')) }
    })
    child.on('error', (error) => super.close(error))
    signedOut = () => {
      signIn?.answered?.(false)
      super.close(new AcpAuthRequiredError('The agent asked to sign in in a browser'))
    }
    // A broken sign-in line on the protocol stream is a broken agent, not a sign-in answer.
    unreadable = () =>
      super.close(new AcpConnectionClosedError('The agent printed a sign-in link Orca cannot read'))
    child.stderr.on('error', (error) => super.close(error))
    managed.onExit((exit) => {
      const error =
        lifecycle.error ??
        new AcpConnectionClosedError(launch.swaOpenCodePolicy ? 'The restricted OpenCode process exited' : managed.stderrTail().trim() || `${launch.command} exited`)
      super.close(error)
      try {
        options.onExit?.(error, { expected: lifecycle.closing, exit })
      } catch (failure) {
        diagnose(`ACP exit listener failed: ${String(failure)}`)
      }
    })
  }

  get pid(): number | undefined {
    return this.managed.child.pid
  }

  get exited(): boolean {
    return this.managed.rootVerdict === 'exited'
  }

  get rootVerdict(): ManagedProviderProcess['rootVerdict'] {
    return this.managed.rootVerdict
  }

  /** Retained evidence from the last close attempt; a null tree means no observation. */
  get lastCloseResult(): Readonly<ManagedProviderProcess['lastCloseResult']> {
    return this.managed.lastCloseResult
  }

  /** Reports retained cleanup uncertainty; false does not prove descendant exit. */
  get processTreeUnproven(): boolean {
    const result = this.lastCloseResult
    return this.exited && (this.tree ? result?.tree !== 'exited' : result?.tree === 'unverifiable' || result?.tree === 'live')
  }

  stderrTail(): string {
    return this.swaPolicy ? 'Restricted OpenCode native process' : this.managed.stderrTail().trim()
  }

  async verifySwaSession(sessionId: string): Promise<void> {
    if (!this.swaPolicy) return
    await this.swaPolicy.ready()
    await this.swaPolicy.verifySession(sessionId)
    this.swaSessionId = sessionId
  }

  override async prompt(prompt: Parameters<AcpSessionRuntime['prompt']>[0], meta?: Parameters<AcpSessionRuntime['prompt']>[1], beforeSend?: () => void): ReturnType<AcpSessionRuntime['prompt']> {
    if (this.swaPolicy) {
      try {
        if (!this.swaSessionId || prompt.some(item => item.type !== 'text') || prompt.some(item => item.type === 'text' && /^\s*\//u.test(item.text))) throw new Error('Restricted OpenCode input changed')
        await this.verifySwaSession(this.swaSessionId)
        beforeSend?.()
      } catch (error) { throw new AcpInputAdmissionError('Restricted OpenCode input was rejected before native delivery', { cause: error }) }
      this.swaPolicy.trace('session/prompt', { sessionId: this.swaSessionId, prompt })
    } else beforeSend?.()
    const response = await super.prompt(prompt, meta)
    this.swaPolicy?.trace('session/prompt/response', { sessionId: this.swaSessionId, response })
    return response
  }

  onExit(listener: (exit: ProviderProcessExit) => void): void {
    this.managed.onExit((exit) => {
      if (this.tree && this.tree.treeVerdict !== 'exited') return
      try {
        listener(exit)
      } catch (failure) {
        this.diagnoseLifecycle(`ACP exit listener failed: ${String(failure)}`)
      }
    })
  }

  pauseReading(): void {
    this.managed.child.stdout.pause()
  }

  resumeReading(): void {
    this.managed.child.stdout.resume()
  }

  override close(error?: Error): Promise<boolean> {
    this.lifecycle.closing ||= this.managed.rootVerdict !== 'exited'
    this.drainNotifications(error)
    return this.managed
      .close(this.tree)
      .then((result) => result.root === 'exited' && (!this.tree || result.tree === 'exited'))
      .finally(() => super.close(error))
  }
}
