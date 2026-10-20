import { homedir } from 'node:os'
import { join } from 'node:path'
import { agentSessionAccountHome } from '../../shared/agent-session-account-home'
import { supportsSupervisedProviderChildLocation } from '../provider-process/supervised-provider-child-location'
import type { StructuredAgentRuntimeMetadata } from '../runtime/structured-agent-runtime-registration-types'
import { PI_RPC_AGENT } from './rpc-agent-definition'
import { piRpcVersionSupported, resolvePiRpcCommand } from './rpc-launch-resolution'
import { createPiModelCatalogProbe } from './rpc-model-catalog-probe'
import { PI_SESSION_HISTORY } from './pi-session-history'
import { PI_TRANSCRIPT_IMPORT } from './pi-transcript-import'

export const PI_RPC_RUNTIME_METADATA: StructuredAgentRuntimeMetadata = {
  definition: PI_RPC_AGENT,
  modelCatalog: () => ({ kind: 'unavailable', reason: 'SWA selects its fixed SDK model at session construction' }),
  transcriptImport: PI_TRANSCRIPT_IMPORT,
  sessionHistory: PI_SESSION_HISTORY,
  supportsLocation: supportsSupervisedProviderChildLocation,
  resolveAccountHome: async ({ launchEnv }) =>
    agentSessionAccountHome(
      PI_RPC_AGENT,
      requiredPrivateAccount(launchEnv)
    )
}

function requiredPrivateAccount(env: NodeJS.ProcessEnv): string {
  const path = env.PI_CODING_AGENT_DIR
  if (!path) throw new Error('SWA Pi account directory is required')
  return path
}
