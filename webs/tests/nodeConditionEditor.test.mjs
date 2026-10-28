import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
const asModule = (code) => `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`;
const converter = asModule(await readFile(new URL('../src/utils/unlockConditionConverter.js', import.meta.url), 'utf8'));
const editorCode = (await readFile(new URL('../src/utils/nodeConditionEditor.js', import.meta.url), 'utf8')).replace(
  "'./unlockConditionConverter'",
  JSON.stringify(converter)
);
const { createNodeConditionValue, serializeNodeConditionValue } = await import(asModule(editorCode));

test('shared editor preserves ordinary tag-compatible conditions and value types', () => {
  const original = {
    logic: 'or',
    conditions: [
      { field: 'protocol', operator: 'equals', value: 'vless' },
      { field: 'delay_time', operator: 'less_than', value: 120 }
    ]
  };
  assert.deepEqual(JSON.parse(serializeNodeConditionValue(createNodeConditionValue(JSON.stringify(original)))), original);
});
test('same-provider unlock groups keep their saved semantics on a shared-editor round trip', () => {
  const original = {
    logic: 'and',
    conditions: [
      { field: 'unlock_provider', operator: 'equals', value: 'openai' },
      { field: 'unlock_status', operator: 'equals', value: 'available' },
      { field: 'unlock_keyword', operator: 'contains', value: 'US' },
      { field: 'protocol', operator: 'equals', value: 'vless' }
    ]
  };
  const editor = createNodeConditionValue(JSON.stringify(original));
  assert.equal(editor.conditions[0].field, 'unlock_condition');
  assert.deepEqual(JSON.parse(serializeNodeConditionValue(editor)), original);
});
test('new optional editor defaults are isolated and do not reuse mutable condition objects', () => {
  const first = createNodeConditionValue('', 'quality_status'),
    second = createNodeConditionValue('', 'quality_status');
  first.conditions[0].value = 'success';
  assert.equal(second.conditions[0].value, '');
  assert.equal(createNodeConditionValue().conditions[0].field, 'protocol');
});

const stub = asModule(
  `export const getNodeConditionFieldMetas=()=>globalThis.__conditionTestMetadata || []; export const getUnlockProviderOptions=()=>[]; export const getUnlockStatusOptions=()=>[]; export const resolveNodeConditionOptionSource=()=>[];`
);
const optionsCode = (await readFile(new URL('../src/utils/nodeConditionOptions.js', import.meta.url), 'utf8'))
  .replace("'views/nodes/utils'", JSON.stringify(stub))
  .replace("'./fraudScore'", JSON.stringify(asModule('export const QUALITY_STATUS_OPTIONS=[];')));
const { getNodeConditionFieldsForSurface } = await import(asModule(optionsCode));
test('candidate controls exclude check-dependent fields, member controls retain them', () => {
  const candidates = getNodeConditionFieldsForSurface('smart-group-candidate').map((f) => f.value);
  assert(candidates.includes('protocol'));
  assert(candidates.includes('source'));
  assert(!candidates.includes('speed'));
  assert(!candidates.includes('tags'));
  assert(!candidates.includes('unlock_condition'));
  const members = getNodeConditionFieldsForSurface('smart-group-member').map((f) => f.value);
  assert(members.includes('quality_status'));
  assert(members.includes('unlock_condition'));
});
test('server field surfaces constrain smart-group controls without changing legacy tag fields', () => {
  try {
    globalThis.__conditionTestMetadata = [
      { value: 'protocol', surfaces: ['tag', 'smart-group-candidate', 'smart-group-member'] },
      { value: 'quality_status', surfaces: ['tag', 'smart-group-member'] },
      { value: 'delay_time', surfaces: ['tag', 'smart-group-member'] }
    ];
    assert.deepEqual(
      getNodeConditionFieldsForSurface('smart-group-candidate').map((f) => f.value),
      ['protocol']
    );
    assert(getNodeConditionFieldsForSurface('tag').some((f) => f.value === 'delay_time'));
    assert.equal(getNodeConditionFieldsForSurface('smart-group-candidate')[0].labelKey, 'nodeConditions.fields.protocol');
  } finally {
    delete globalThis.__conditionTestMetadata;
  }
});
