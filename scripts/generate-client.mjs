import { readFile, writeFile } from 'node:fs/promises';
import openapiTS, { astToString } from 'openapi-typescript';

const spec = new URL('../contracts/openapi.yaml', import.meta.url);
const output = new URL('../web/src/api/schema.d.ts', import.meta.url);
const generated = '// Generated from contracts/openapi.yaml. Run pnpm generate; do not edit.\n' +
  astToString(await openapiTS(spec));
if (process.argv.includes('--check')) {
  if (await readFile(output, 'utf8') !== generated) {
    throw new Error('Generated client schema drift: run pnpm generate and commit schema.d.ts');
  }
  console.log('OpenAPI generated types match source');
} else {
  await writeFile(output, generated);
  console.log('Generated web/src/api/schema.d.ts');
}
