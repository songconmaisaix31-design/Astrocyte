import { resolve } from 'node:path';
import { loadPublicPaperInputs } from './public-inputs.mjs';

const papers = await loadPublicPaperInputs(resolve(process.argv[2]));
for (const [name, input] of Object.entries(papers)) {
  console.log(JSON.stringify({ name, source_url: input.source.source_url, source_key: input.source.source_key, observed_version: input.source.observed_version, text_bytes: Buffer.byteLength(input.text), content_state: input.source.content_state }));
}
