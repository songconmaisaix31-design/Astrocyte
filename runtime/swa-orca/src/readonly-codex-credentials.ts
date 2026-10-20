import { readFile } from 'node:fs/promises'
import { decodeJwt } from 'jose'
import { z } from 'zod'
import type { CredentialStore } from '@earendil-works/pi-ai'

export const PROVIDER = 'openai-codex'
export const MODEL = 'gpt-6.1-sol'
export const REQUEST_TIMEOUT_MS = 600_000
const REFRESH_MARGIN_MS = 300_000

const Source = z.object({
  auth_mode: z.literal('chatgpt'),
  tokens: z.object({
    access_token: z.string().min(1),
    refresh_token: z.string().min(1),
    account_id: z.string().min(1)
  })
})

const Claims = z.object({
  exp: z.number().int().positive(),
  'https://api.openai.com/auth': z.object({ chatgpt_account_id: z.string().min(1) })
})

export type CredentialObservation = {
  operation: 'read' | 'list' | 'modify' | 'delete'
  provider: string
  at: number
}

// 原账户仍由原客户端管理；SDK端口没有执行刷新回调的权限。
export function readonlyCodexCredentials(authFile: string) {
  const observations: CredentialObservation[] = []
  const observe = (operation: CredentialObservation['operation'], provider: string) => {
    observations.push({ operation, provider, at: Date.now() })
    if (provider !== PROVIDER) throw new Error('unapproved_credential_provider')
  }
  const readCredential = async () => {
    let source: z.infer<typeof Source>
    let claims: z.infer<typeof Claims>
    try {
      source = Source.parse(JSON.parse(await readFile(authFile, 'utf8')))
      claims = Claims.parse(decodeJwt(source.tokens.access_token))
    } catch (cause) {
      throw new Error('selected_codex_oauth_source_invalid', { cause })
    }
    const expires = claims.exp * 1000
    if (!Number.isSafeInteger(expires) || expires - Date.now() <= REQUEST_TIMEOUT_MS + REFRESH_MARGIN_MS) {
      throw new Error('selected_codex_oauth_validity_insufficient')
    }
    const accountId = claims['https://api.openai.com/auth'].chatgpt_account_id
    if (source.tokens.account_id !== accountId) throw new Error('selected_codex_oauth_account_mismatch')
    return {
      type: 'oauth' as const,
      access: source.tokens.access_token,
      refresh: source.tokens.refresh_token,
      expires,
      accountId
    }
  }
  const store: CredentialStore = {
    async read(provider, options) {
      options?.signal?.throwIfAborted()
      observe('read', provider)
      return readCredential()
    },
    async list(options) {
      options?.signal?.throwIfAborted()
      observe('list', PROVIDER)
      await readCredential()
      return [{ providerId: PROVIDER, type: 'oauth' as const }]
    },
    async modify(provider) {
      observe('modify', provider)
      throw new Error('selected_codex_oauth_read_only')
    },
    async delete(provider) {
      observe('delete', provider)
      throw new Error('selected_codex_oauth_read_only')
    }
  }
  return { store, observations }
}
