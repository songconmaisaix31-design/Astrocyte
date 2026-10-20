import { cpus, freemem } from 'node:os'
import { performance } from 'node:perf_hooks'
import { setTimeout } from 'node:timers/promises'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'

function totals() {
  const values = cpus()
  if (!values.length || values.length > 64) throw new OrchestrationError('resource_refused', 'CPU sampling requires one observed processor group with at most 64 logical processors')
  return values.reduce((sum, cpu) => ({ count: sum.count + 1, idle: sum.idle + cpu.times.idle, total: sum.total + Object.values(cpu.times).reduce((a, b) => a + b, 0) }), { count: 0, idle: 0, total: 0 })
}

export async function admitWorkerResources() {
  const samples = []
  const started = performance.now()
  for (let index = 0; index < 2; index++) {
    if (index) await setTimeout(1500)
    const before = totals()
    await setTimeout(1000)
    const after = totals()
    const total = after.total - before.total, idle = after.idle - before.idle
    if (before.count !== after.count || total <= 0 || idle < 0 || idle > total) throw new OrchestrationError('resource_refused', 'The actual blocking CPU sample is unavailable')
    samples.push({ utc: new Date().toISOString(), elapsed_seconds: (performance.now() - started) / 1000, cpu_percent: 100 * (total - idle) / total, available_bytes: freemem() })
  }
  const interval = samples[1].elapsed_seconds - samples[0].elapsed_seconds
  const accepted = interval >= 2 && interval <= 5 && samples.every(row => row.cpu_percent < 85 && row.available_bytes >= 4 * 1024 ** 3)
  process.stderr.write(JSON.stringify({ type: 'swa_worker_resource_admission', samples, interval_seconds: interval, accepted }) + '\n')
  if (!accepted) throw new OrchestrationError('resource_refused', 'Fresh Worker samples require CPU below 85% and available physical memory at least 4 GiB, 2–5 seconds apart')
  return samples
}
