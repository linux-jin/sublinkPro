// Numeric five-field Cron syntax supported by the generator. Server validation
// remains authoritative; notably robfig/cron uses Sunday 0, not 7.
export const validateCronExpression = (cron) => {
  if (typeof cron !== 'string') return false;
  const fields = cron.trim().split(/\s+/);
  if (fields.length !== 5) return false;
  const ranges = [
    [0, 59],
    [0, 23],
    [1, 31],
    [1, 12],
    [0, 6]
  ];
  return fields.every((field, index) =>
    field.split(',').every((part) => {
      const match = /^(\*|\d+(?:-\d+)?)(?:\/(\d+))?$/.exec(part);
      if (!match) return false;
      const [, base, step] = match;
      if (step !== undefined && Number(step) <= 0) return false;
      // The preview parser supports steps on wildcard/range bases only.
      if (step !== undefined && base !== '*' && !base.includes('-')) return false;
      if (base === '*') return true;
      const [start, end = start] = base.split('-').map(Number);
      return start >= ranges[index][0] && end <= ranges[index][1] && start <= end;
    })
  );
};
