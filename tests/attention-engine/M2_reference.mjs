// Independent reference for the application's floor-initialised scenario.
// The original Python zero-initialised golden stays unchanged.
import { readFileSync, writeFileSync } from 'node:fs';
const original = JSON.parse(readFileSync(new URL('./M2_golden_stream.json', import.meta.url)));
const doc = { ...original, initial: { ...original.initial, attention: original.floor }, output: [] };
let { attention: a, momentum: m, confidence: c } = doc.initial;
for (const step of doc.input) {
  const next = Math.min(doc.ceiling, Math.max(doc.floor, a * Math.exp(-doc.lambda) + step.signal_sum));
  m = doc.alpha * m + (1 - doc.alpha) * (next - a);
  c = Math.max(0, Math.min(1, c + step.confidence_delta));
  a = next;
  doc.output.push({ ts_ms: step.ts_ms, attention: a, momentum: m, confidence: c });
}
writeFileSync(new URL('./M2_application_golden.json', import.meta.url), JSON.stringify(doc, null, 2) + '\n');
