import type { PipedProcessSpawner } from '@orca/process-host/process-spec'
import { spawnProcess } from '@orca/process-host'
import {
  spawnManagedProviderProcess,
  type ManagedProviderProcess,
  type ProviderProcessExit
} from '../provider-process/managed-provider-process'
import type { ProviderProcessLaunch } from '../provider-process/provider-process-launch'
import type { ProviderProcessCloseResult } from '../provider-process/provider-process-close'
import {
  JsonlRpcPeer,
  JsonlRpcStreamClosedError,
  type JsonlRpcPeerHandlers,
  type JsonlRpcRecord
} from './peer'
import { resolveJsonlRpcPeerOptions, type JsonlRpcPeerOptions } from './peer-limits'
import { SwaProviderProcessTree } from '../provider-process/swa-provider-process-tree'

export type JsonlRpcAgentConnectionOptions = JsonlRpcPeerHandlers & {
  allowedCommands?: ReadonlySet<string>
  peer?: JsonlRpcPeerOptions
  onExit?: (error: Error, context: { expected: boolean; exit: ProviderProcessExit }) => void
  /** Any stdout or stderr chunk from the child. */
  onOutput?: () => void
}

/** The owning runtime keeps process exit evidence separate from transport closure. */
export class JsonlRpcAgentConnection {
  private readonly managed: ManagedProviderProcess
  private readonly peer: JsonlRpcPeer
  private readonly ownedTree?: SwaProviderProcessTree
  private verifiedExit = false
  private readonly verifiedExitListeners = new Set<() => void>()
  private closing = false
  private finishObservedExit?: () => void
  private streamExitTimer?: ReturnType<typeof setTimeout>

  constructor(
    launch: ProviderProcessLaunch,
    private readonly options: JsonlRpcAgentConnectionOptions = {},
    spawnImpl: PipedProcessSpawner = spawnProcess
  ) {
    const peerOptions = resolveJsonlRpcPeerOptions(options.peer)
    if (launch.env?.SWA_WORKER_ROOT && !launch.swaJobIPC) throw new Error('SWA refuses SDK launch without parent Job admission')
    if (launch.swaJobIPC) this.ownedTree = new SwaProviderProcessTree()
    try {
      this.managed = spawnManagedProviderProcess(launch, {
        spawnImpl,
        site: 'jsonl-rpc-agent-teardown',
        ...(this.ownedTree ? { acceptClose: result => result.root === 'exited' && (result.tree === 'exited' || this.managed.processless) } : {}),
        ...(options.onOutput ? { onOutput: options.onOutput } : {})
      })
    } catch (error) {
      this.ownedTree?.closeBeforeSpawnFailure()
      throw error
    }
    const managed = this.managed
    this.ownedTree?.bind(managed.child, () => ({ observed: managed.rootExitObserved, processless: managed.processless }))
    this.peer = new JsonlRpcPeer(
      managed.child.stdout,
      managed.child.stdin,
      {
        ...options,
        onClose: (error) => {
          const fail = (): void => {
            if (this.closing || managed.rootVerdict === 'exited') {
              return
            }
            void managed.close(this.ownedTree).then(
              (result) => {
                if (result.root !== 'exited' || this.ownedTree && result.tree !== 'exited') {
                  this.diagnose('Agent exit was not proven after RPC failure')
                }
              },
              () => this.diagnose('Agent cleanup failed after RPC failure')
            )
            options.onClose?.(error)
          }
          if (
            error instanceof JsonlRpcStreamClosedError &&
            !this.closing &&
            managed.rootVerdict !== 'exited'
          ) {
            // EOF often precedes the exit carrying an import/startup failure's code.
            this.streamExitTimer = setTimeout(fail, 250)
            this.streamExitTimer.unref()
          } else {
            fail()
          }
        }
      },
      peerOptions
    )
    managed.child.on('error', this.onError)
    managed.child.stderr.on('error', this.onError)
    managed.onExit((exit) => {
      clearTimeout(this.streamExitTimer)
      const error = new Error(
        exit.processless
          ? 'Agent process could not start'
          : `Agent process exited (code ${exit.code ?? 'none'}, signal ${exit.signal ?? 'none'})`
      )
      this.peer.finishOnInputEnd(error)
      const stdout = managed.child.stdout
      let finished = false
      let timer: ReturnType<typeof setTimeout> | undefined
      const finish = (): void => {
        if (finished) {
          return
        }
        finished = true
        clearTimeout(timer)
        stdout.removeListener('end', finish)
        stdout.removeListener('close', finish)
        this.finishObservedExit = undefined
        this.peer.close(error)
        managed.child.removeListener('error', this.onError)
        const publish = (): void => {
          this.verifiedExit = true
          for (const listener of this.verifiedExitListeners) listener()
          this.verifiedExitListeners.clear()
          try { options.onExit?.(error, { expected: this.closing, exit }) }
          catch { this.diagnose('Agent exit observer failed') }
        }
        if (!this.ownedTree) publish()
        else void managed.close(this.ownedTree).then(result => {
          if (result.root === 'exited' && (result.tree === 'exited' || managed.processless)) publish()
          else this.diagnose('SWA SDK root exited while complete Job exit remains unknown')
        }, () => this.diagnose('SWA SDK Job exit query failed; retain its session and resources'))
      }
      this.finishObservedExit = finish
      if (
        this.closing ||
        exit.processless ||
        this.peer.closed ||
        stdout.destroyed ||
        stdout.readableEnded
      ) {
        finish()
      } else {
        stdout.once('end', finish)
        stdout.once('close', finish)
        // A descendant can inherit stdout after the provider root exits.
        timer = setTimeout(finish, 1_000)
        timer.unref()
      }
    })
  }

  get pid(): number | undefined {
    return this.managed.child.pid
  }
  get closed(): boolean {
    return this.peer.closed
  }
  get rootVerdict(): ManagedProviderProcess['rootVerdict'] {
    if (this.ownedTree && this.managed.rootVerdict === 'exited' && !this.managed.processless && this.ownedTree.treeVerdict !== 'exited') return 'unverifiable'
    return this.managed.rootVerdict
  }
  get processless(): boolean {
    return this.managed.processless
  }
  get lastCloseResult(): ManagedProviderProcess['lastCloseResult'] {
    return this.managed.lastCloseResult
  }

  onExit(listener: () => void): void {
    if (!this.ownedTree) this.managed.onExit(listener)
    else if (this.verifiedExit) listener()
    else this.verifiedExitListeners.add(listener)
  }

  async request(
    command: string,
    params: Record<string, unknown> = {},
    options: { timeoutMs?: number | null } = {}
  ): Promise<unknown> {
    if (this.options.allowedCommands && !this.options.allowedCommands.has(command)) {
      return Promise.reject(new Error(`SWA refuses Pi RPC command: ${command}`))
    }
    await this.ownedTree?.ready()
    return this.peer.request(command, params, options)
  }

  async send(record: JsonlRpcRecord): Promise<void> {
    if (this.options.allowedCommands && !this.options.allowedCommands.has(record.type)) {
      return Promise.reject(new Error(`SWA refuses Pi RPC command: ${record.type}`))
    }
    await this.ownedTree?.ready()
    return this.peer.send(record)
  }

  pauseReading(): void {
    this.peer.pauseReading()
  }

  resumeReading(): void {
    this.peer.resumeReading()
  }

  /** Stop sends the dialect's abort first, then closes this connection through the supervisor. */
  close(error?: Error): Promise<ProviderProcessCloseResult> {
    this.closing = true
    clearTimeout(this.streamExitTimer)
    this.finishObservedExit?.()
    this.peer.close(error)
    return this.managed.close(this.ownedTree)
  }

  private readonly onError = (error: Error): void => this.peer.close(error)
  private diagnose(message: string): void {
    try {
      this.options.onDiagnostic?.(message)
    } catch {
      /* Diagnostics cannot break cleanup. */
    }
  }
}
