import { useId } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import { Box, Typography, FormControl, InputLabel, Select, MenuItem, TextField, Button, IconButton, Chip, Card } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import UnlockConditionInput from './UnlockConditionInput';
import {
  getNodeConditionFieldMeta,
  getNodeConditionFieldsForSurface,
  getNodeConditionValueOptions,
  isNodeConditionNumericField,
  isNodeConditionSelectField
} from 'utils/nodeConditionOptions';
import { UNLOCK_CONDITION_FIELD } from 'utils/unlockConditionConverter';

const operators = [
  { value: 'equals', labelKey: 'tags.dialog.rule.operators.equals', type: 'string' },
  { value: 'not_equals', labelKey: 'tags.dialog.rule.operators.notEquals', type: 'string' },
  { value: 'contains', labelKey: 'tags.dialog.rule.operators.contains', type: 'string' },
  { value: 'not_contains', labelKey: 'tags.dialog.rule.operators.notContains', type: 'string' },
  { value: 'regex', labelKey: 'tags.dialog.rule.operators.regex', type: 'string' },
  { value: 'greater_than', labelKey: 'tags.dialog.rule.operators.greaterThan', type: 'number' },
  { value: 'less_than', labelKey: 'tags.dialog.rule.operators.lessThan', type: 'number' },
  { value: 'greater_or_equal', labelKey: 'tags.dialog.rule.operators.greaterOrEqual', type: 'number' },
  { value: 'less_or_equal', labelKey: 'tags.dialog.rule.operators.lessOrEqual', type: 'number' }
];

// Controlled UI conditions retain incomplete unlock edits until the parent saves.
export default function NodeConditionEditor({
  logic,
  conditions,
  onChange,
  surface = 'tag',
  defaultField = 'link_country',
  title,
  helperText,
  maxConditions
}) {
  const { t } = useTranslation();
  const editorId = useId();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const setConditions = (next) => onChange({ logic, conditions: next });
  const setLogic = (next) => onChange({ logic: next, conditions });
  const handleAddCondition = () => {
    setConditions([...conditions, { field: defaultField, operator: 'equals', value: '' }]);
  };

  const handleRemoveCondition = (index) => {
    if (conditions.length > 1) {
      setConditions(conditions.filter((_, i) => i !== index));
    }
  };

  const handleConditionChange = (index, key, value) => {
    const newConditions = conditions.map((condition) => ({ ...condition }));
    const previousField = newConditions[index].field;
    newConditions[index][key] = value;

    if (key === 'field') {
      const isNumeric = isNodeConditionNumericField(value);
      const isSelectField = isNodeConditionSelectField(value);
      const fieldMeta = getNodeConditionFieldMeta(value);
      const allowedOperators = fieldMeta?.operators || [];
      const currentOp = newConditions[index].operator;
      const opInfo = operators.find((o) => o.value === currentOp);

      // 从解锁情况切换到其他字段，或从其他字段切换到解锁情况时，需要重置 value
      const wasPreviousUnlock = previousField === UNLOCK_CONDITION_FIELD;
      const isCurrentUnlock = value === UNLOCK_CONDITION_FIELD;

      // 处理解锁情况虚拟字段
      if (isCurrentUnlock) {
        newConditions[index].operator = 'equals';
        newConditions[index].value = {
          provider: '',
          status: '',
          keyword: '',
          providerOperator: 'equals',
          statusOperator: 'equals',
          keywordOperator: 'contains'
        };
      } else if (wasPreviousUnlock || isSelectField) {
        // 从解锁情况切换出来，或切换到下拉字段，重置为空字符串
        if (!allowedOperators.includes(currentOp)) {
          newConditions[index].operator = allowedOperators[0] || 'equals';
        }
        newConditions[index].value = '';
      } else if (isNumeric && opInfo?.type === 'string' && !['equals', 'not_equals'].includes(currentOp)) {
        newConditions[index].operator = allowedOperators[0] || 'greater_than';
      } else if (!isNumeric && opInfo?.type === 'number') {
        newConditions[index].operator = allowedOperators[0] || 'equals';
      }
    }

    setConditions(newConditions);
  };

  // 处理解锁情况复合值的变化
  const handleUnlockConditionChange = (index, subField, subValue) => {
    const newConditions = conditions.map((condition) => ({ ...condition }));
    if (!newConditions[index].value || typeof newConditions[index].value !== 'object') {
      newConditions[index].value = {
        provider: '',
        status: '',
        keyword: '',
        providerOperator: 'equals',
        statusOperator: 'equals',
        keywordOperator: 'contains'
      };
    }
    // subField 为 'value' 时，subValue 是整个对象
    if (subField === 'value') {
      newConditions[index].value = subValue;
    } else {
      newConditions[index].value = {
        ...newConditions[index].value,
        [subField]: subValue
      };
    }
    setConditions(newConditions);
  };

  const getAvailableOperators = (field) => {
    const fieldMeta = getNodeConditionFieldMeta(field);
    if (Array.isArray(fieldMeta?.operators) && fieldMeta.operators.length > 0) {
      return operators.filter((o) => fieldMeta.operators.includes(o.value));
    }
    const isNumeric = isNodeConditionNumericField(field);
    const isSelectField = isNodeConditionSelectField(field);
    if (isSelectField) {
      return operators.filter((o) => ['equals', 'not_equals'].includes(o.value));
    }
    if (isNumeric) {
      return operators;
    }
    return operators.filter((o) => o.type === 'string');
  };

  const getVisibleFields = () =>
    getNodeConditionFieldsForSurface(surface).filter(
      (field) => !field.hidden && !['unlock_provider', 'unlock_status', 'unlock_keyword'].includes(field.value)
    );
  return (
    <Box role="group" aria-label={title || t('tags.dialog.rule.conditions')}>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          mb: 2,
          flexWrap: 'wrap',
          gap: 1
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
          <Typography variant="subtitle1">{title || t('tags.dialog.rule.conditions')}</Typography>
          <Chip
            label={logic === 'and' ? t('tags.dialog.rule.logic.and') : t('tags.dialog.rule.logic.or')}
            size="small"
            onClick={() => setLogic(logic === 'and' ? 'or' : 'and')}
            sx={{ cursor: 'pointer' }}
          />
        </Box>
        <Button
          size="small"
          startIcon={<AddIcon />}
          onClick={handleAddCondition}
          disabled={Boolean(maxConditions && conditions.length >= maxConditions)}
        >
          {t('tags.dialog.rule.addCondition')}
        </Button>
      </Box>

      {conditions.map((cond, index) =>
        isMobile ? (
          <Card key={index} variant="outlined" sx={{ mb: 1.5, p: 1.5, position: 'relative' }}>
            <IconButton
              size="small"
              color="error"
              aria-label={t('nodeConditions.editor.removeCondition', { index: index + 1 })}
              onClick={() => handleRemoveCondition(index)}
              disabled={conditions.length === 1}
              sx={{ position: 'absolute', top: 8, right: 8 }}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5, pr: 4 }}>
              <FormControl size="small" fullWidth>
                <InputLabel id={`${editorId}-${index}-field`}>{t('tags.dialog.rule.field')}</InputLabel>
                <Select
                  labelId={`${editorId}-${index}-field`}
                  value={cond.field}
                  label={t('tags.dialog.rule.field')}
                  onChange={(e) => handleConditionChange(index, 'field', e.target.value)}
                >
                  {getVisibleFields().map((f) => (
                    <MenuItem key={f.value} value={f.value}>
                      {f.labelKey ? t(f.labelKey, f.label) : f.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <FormControl size="small" fullWidth>
                <InputLabel id={`${editorId}-${index}-operator`}>{t('tags.dialog.rule.operator')}</InputLabel>
                <Select
                  labelId={`${editorId}-${index}-operator`}
                  value={cond.operator}
                  label={t('tags.dialog.rule.operator')}
                  onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
                  disabled={cond.field === UNLOCK_CONDITION_FIELD}
                >
                  {getAvailableOperators(cond.field).map((op) => (
                    <MenuItem key={op.value} value={op.value}>
                      {t(op.labelKey)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              {cond.field === UNLOCK_CONDITION_FIELD ? (
                <UnlockConditionInput
                  value={cond.value}
                  onChange={(newValue) => handleUnlockConditionChange(index, 'value', newValue)}
                  isMobile={true}
                />
              ) : getNodeConditionValueOptions(cond.field) ? (
                <FormControl size="small" fullWidth>
                  <InputLabel id={`${editorId}-${index}-value`}>{t('tags.dialog.rule.value')}</InputLabel>
                  <Select
                    labelId={`${editorId}-${index}-value`}
                    value={cond.value}
                    label={t('tags.dialog.rule.value')}
                    onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                  >
                    {getNodeConditionValueOptions(cond.field).map((opt) => (
                      <MenuItem key={opt.value} value={opt.value}>
                        {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              ) : (
                <TextField
                  size="small"
                  label={t('tags.dialog.rule.value')}
                  value={cond.value}
                  onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                  fullWidth
                  type={isNodeConditionNumericField(cond.field) ? 'number' : 'text'}
                  placeholder={cond.operator === 'regex' ? t('tags.dialog.rule.regexPlaceholder') : ''}
                />
              )}
            </Box>
          </Card>
        ) : (
          <Box key={index} sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 1.5, alignItems: 'center' }}>
            <FormControl size="small" sx={{ minWidth: 140 }}>
              <InputLabel id={`${editorId}-${index}-field`}>{t('tags.dialog.rule.field')}</InputLabel>
              <Select
                labelId={`${editorId}-${index}-field`}
                value={cond.field}
                label={t('tags.dialog.rule.field')}
                onChange={(e) => handleConditionChange(index, 'field', e.target.value)}
              >
                {getVisibleFields().map((f) => (
                  <MenuItem key={f.value} value={f.value}>
                    {f.labelKey ? t(f.labelKey, f.label) : f.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControl size="small" sx={{ minWidth: 120 }}>
              <InputLabel id={`${editorId}-${index}-operator`}>{t('tags.dialog.rule.operator')}</InputLabel>
              <Select
                labelId={`${editorId}-${index}-operator`}
                value={cond.operator}
                label={t('tags.dialog.rule.operator')}
                onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
                disabled={cond.field === UNLOCK_CONDITION_FIELD}
              >
                {getAvailableOperators(cond.field).map((op) => (
                  <MenuItem key={op.value} value={op.value}>
                    {t(op.labelKey)}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            {cond.field === UNLOCK_CONDITION_FIELD ? (
              <UnlockConditionInput
                value={cond.value}
                onChange={(newValue) => handleUnlockConditionChange(index, 'value', newValue)}
                isMobile={false}
              />
            ) : getNodeConditionValueOptions(cond.field) ? (
              <FormControl size="small" sx={{ minWidth: 140 }}>
                <InputLabel id={`${editorId}-${index}-value`}>{t('tags.dialog.rule.value')}</InputLabel>
                <Select
                  labelId={`${editorId}-${index}-value`}
                  value={cond.value}
                  label={t('tags.dialog.rule.value')}
                  onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                >
                  {getNodeConditionValueOptions(cond.field).map((opt) => (
                    <MenuItem key={opt.value} value={opt.value}>
                      {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            ) : (
              <TextField
                size="small"
                label={t('tags.dialog.rule.value')}
                value={cond.value}
                onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                sx={{ flex: 1 }}
                type={isNodeConditionNumericField(cond.field) ? 'number' : 'text'}
                placeholder={cond.operator === 'regex' ? t('tags.dialog.rule.regexPlaceholder') : ''}
              />
            )}
            <IconButton
              size="small"
              color="error"
              aria-label={t('nodeConditions.editor.removeCondition', { index: index + 1 })}
              onClick={() => handleRemoveCondition(index)}
              disabled={conditions.length === 1}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Box>
        )
      )}

      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 1 }}>
        {helperText || t('tags.dialog.rule.example')}
      </Typography>
    </Box>
  );
}

NodeConditionEditor.propTypes = {
  logic: PropTypes.string.isRequired,
  conditions: PropTypes.arrayOf(PropTypes.object).isRequired,
  onChange: PropTypes.func.isRequired,
  surface: PropTypes.string,
  defaultField: PropTypes.string,
  title: PropTypes.string,
  helperText: PropTypes.string,
  maxConditions: PropTypes.number
};
