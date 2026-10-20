import { runWithParentJob } from './provider-job-bootstrap.mjs'

await runWithParentJob(async () => {
  const session = await import('./pi-session.mjs')
  await session.runRestrictedSession()
})
