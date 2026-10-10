import type { components } from '../../api/schema';

type Agent = components['schemas']['LocalAgentV1'];
type Settings = components['schemas']['LocalProjectSettingsV1'];
type Action = keyof Agent['capabilities'];

/** A human may verify an unknown operation, but registration never establishes support. */
export function nativeActionAllowed(settings: Settings, agent: Agent | undefined, cli: string, action: Action, unavailable: boolean): boolean {
  if (!cli || cli !== settings.external_model_cli || agent?.id !== cli || !settings.allowed_actions.includes(action)) return false;
  if (unavailable && action !== 'stop') return false;
  const status = agent.capabilities[action].status;
  return status === 'supported' || (status === 'unknown' && agent.native_adapter_registered === true);
}
