import { convertBackendToUI, convertUIToBackend } from './unlockConditionConverter';

export function createNodeConditionValue(raw = '', defaultField = 'protocol') {
  const fallback = { logic: 'and', conditions: [{ field: defaultField, operator: 'equals', value: '' }] };
  try {
    const parsed = JSON.parse(raw || '{}');
    if (!Array.isArray(parsed.conditions) || parsed.conditions.length === 0) return fallback;
    return { logic: parsed.logic || 'and', conditions: convertBackendToUI(parsed.conditions) };
  } catch {
    return fallback;
  }
}

export function serializeNodeConditionValue(value) {
  return JSON.stringify({ logic: value.logic, conditions: convertUIToBackend(value.conditions) });
}
