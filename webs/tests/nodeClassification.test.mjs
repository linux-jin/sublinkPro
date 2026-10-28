import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import { runInNewContext } from 'node:vm';

const code = await readFile(new URL('../src/views/node-classification/navigation.js', import.meta.url), 'utf8');
const { NODE_CLASSIFICATION_PATH, getClassificationTab, getClassificationLocation } = await import(
  `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`
);

test('classification defaults safely to tags and supports direct smart-group links', () => {
  for (const search of ['', '?tab=tags', '?tab=unknown', '?tab=', '?other=smart-groups']) {
    assert.equal(getClassificationTab(search), 'tags', search);
  }
  assert.equal(getClassificationTab('?tab=smart-groups'), 'smart-groups');
});

test('tab links preserve unrelated query parameters and fragments without mutating the location', () => {
  const location = Object.freeze({
    pathname: '/subscription/tags',
    search: '?lng=en-US&filter=a%26b&tab=tags&tab=unknown',
    hash: '#saved'
  });
  const result = getClassificationLocation(location, 'smart-groups');
  assert.equal(result.pathname, NODE_CLASSIFICATION_PATH);
  assert.equal(result.hash, '#saved');
  const params = new URLSearchParams(result.search);
  assert.equal(params.get('lng'), 'en-US');
  assert.equal(params.get('filter'), 'a&b');
  assert.deepEqual(params.getAll('tab'), ['smart-groups']);
  assert.equal(location.pathname, '/subscription/tags');
});

test('legacy route determines the destination tab even if an old link has a conflicting query', () => {
  for (const tab of ['tags', 'smart-groups']) {
    const result = getClassificationLocation({ search: '?tab=unknown' }, tab);
    assert.equal(result.pathname, NODE_CLASSIFICATION_PATH);
    assert.equal(getClassificationTab(result.search), tab);
    assert.equal(result.hash, '');
  }
});

test('generated tab links round-trip across repeated navigation', () => {
  let location = { search: '?lng=zh-CN', hash: '#nodes' };
  for (const tab of ['smart-groups', 'tags', 'smart-groups', 'tags']) {
    location = getClassificationLocation(location, tab);
    assert.equal(getClassificationTab(location.search), tab);
    assert.equal(new URLSearchParams(location.search).get('lng'), 'zh-CN');
    assert.equal(location.hash, '#nodes');
  }
});

test('classification has one menu entry and shared task-menu icons remain declared', async () => {
  const source = await readFile(new URL('../src/menu-items/subscription.js', import.meta.url), 'utf8');
  const executable = source
    .replace(/import\s*\{([^}]+)\}\s*from '@tabler\/icons-react';/, (_, names) =>
      names
        .split(',')
        .map((name) => name.trim())
        .filter(Boolean)
        .map((name) => `const ${name} = null;`)
        .join('\n')
    )
    .replace(/export\s*\{([^}]+)\};/, 'globalThis.definitions = {$1};');
  const context = {};
  runInNewContext(executable, context);
  const items = Object.values(context.definitions).flatMap((group) => group.children);
  assert.equal(items.filter((item) => item.url === NODE_CLASSIFICATION_PATH).length, 1);
  assert.equal(
    items.some((item) => ['/subscription/tags', '/subscription/smart-groups'].includes(item.url)),
    false
  );
  assert.equal(items.find((item) => item.id === 'tasks').url, '/system/tasks');
});
