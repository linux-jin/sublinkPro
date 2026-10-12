import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
// Load the standalone ESM utility without changing the app's package type.
const code = await readFile(new URL('../src/utils/cronValidation.js', import.meta.url), 'utf8');
const { validateCronExpression } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);

test('five-field schedules accepted by the preview and backend', () => {
  for (const cron of ['0 3 * * *', '*/30 * * * *', '0 0 * * 0', '0,30 1-3 * * 1-5', '0 0 */1 * *', '0 0 1-31/2 * *']) {
    assert.equal(validateCronExpression(cron), true, cron);
  }
});
test('reject zero steps before the preview parser can loop forever', () => {
  for (const cron of [
    '*/0 * * * *',
    '0-59/0 * * * *',
    '0 0 * * 7',
    '0 0 31-1 * *',
    '0,,1 * * * *',
    '60 * * * *',
    '0 0 0 * *',
    '0 0 * * * *',
    'CRON_TZ=UTC 0 0 * * *',
    '',
    null
  ]) {
    assert.equal(validateCronExpression(cron), false, String(cron));
  }
});
