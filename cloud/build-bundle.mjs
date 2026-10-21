// Generate cloud/account-bundle.mjs from its sources.
// Run: node cloud/build-bundle.mjs
// Verify committed bundle: node cloud/build-bundle.mjs --check
import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

const url = name => new URL(name, import.meta.url);
const content = name => readFileSync(url(name), 'utf8');

export function buildBundle() {
  let bundle = content('./worker.mjs');
  const sources = [
    ['PAGE', 'PAGE', './account.html'],
    ['APP_CSS', 'CSS', './account.css'],
    ['APP_JS', 'JS', './account.js'],
  ];
  for (const [name, placeholder, filename] of sources) {
    const marker = `const ${name}=__${placeholder}__;`;
    if (bundle.split(marker).length !== 2) {
      throw new Error('Expected exactly one template placeholder: ' + marker);
    }
    bundle = bundle.replace(marker, () => `const ${name}=${JSON.stringify(content(filename))};`);
  }
  if (/__(PAGE|CSS|JS)__/.test(bundle)) throw new Error('Unexpanded Worker bundle placeholder');
  return bundle;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const expected = buildBundle(), filename = url('./account-bundle.mjs');
  if (process.argv.includes('--check')) {
    if (content('./account-bundle.mjs') !== expected) {
      console.error('Cloudflare Worker bundle is stale: run node cloud/build-bundle.mjs');
      process.exitCode = 1;
    } else console.log('Cloudflare Worker bundle matches its sources');
  } else {
    writeFileSync(filename, expected);
    console.log('Updated cloud/account-bundle.mjs');
  }
}
