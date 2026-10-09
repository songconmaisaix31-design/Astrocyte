import { importMaterial } from './api.mjs';
import { distillation, opportunity, paperImport, sourceRef } from './fixtures.mjs';

export async function seedCandidate(api, locator) {
  const imported = await importMaterial(api, paperImport(locator));
  const refs = [sourceRef(imported.detail.material)];
  const records = [];
  for (const stage of ['content', 'topic', 'project']) {
    const result = await api.write('/distillations', distillation(stage, refs));
    records.push(result.distillation);
  }
  const candidate = await api.write('/opportunities', opportunity(refs, records.map(record => record.id)), { status: 201 });
  return { ...imported, refs, records, candidate };
}
