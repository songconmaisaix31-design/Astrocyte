import { readFile } from 'node:fs/promises';
import { parse } from 'yaml';
import Ajv2020 from 'ajv/dist/2020.js';
import addFormats from 'ajv-formats';

// Same schema normalization as scripts/validate-examples.mjs, applied to live responses.
const spec = parse(await readFile(new URL('../../contracts/openapi.yaml', import.meta.url), 'utf8'));
const definitions = JSON.parse(JSON.stringify(spec.components.schemas).replaceAll('#/components/schemas/', '#/$defs/'));
const ajv = new Ajv2020({ allErrors: true, strict: false });
addFormats(ajv);
const validators = new Map();

export function validateResponse(method, path, status, data) {
  // Queries belong to the real request; OpenAPI operation keys contain only paths.
  const pathname = new URL(path, 'http://localhost').pathname;
  const route = Object.keys(spec.paths).find(candidate => new RegExp(`^${candidate.replace(/\{[^}]+\}/g, '[^/]+')}$`).test(pathname));
  const operation = spec.paths[route]?.[method.toLowerCase()];
  let response = operation?.responses?.[String(status)];
  if (response?.$ref) response = response.$ref.slice(2).split('/').reduce((value, key) => value[key], spec);
  const schema = response?.content?.['application/json']?.schema;
  if (!schema) throw new Error(`No documented JSON response for ${method} ${path}: ${status}`);
  const key = `${method} ${route} ${status}`;
  if (!validators.has(key)) {
    const normalized = JSON.parse(JSON.stringify(schema).replaceAll('#/components/schemas/', '#/$defs/'));
    validators.set(key, ajv.compile({ ...normalized, $defs: definitions }));
  }
  const validate = validators.get(key);
  if (!validate(data)) throw new Error(`${method} ${path} ${status} violates OpenAPI: ${ajv.errorsText(validate.errors)}`);
}
