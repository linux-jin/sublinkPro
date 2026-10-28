import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

// material-ui
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Divider from '@mui/material/Divider';

import NodeConditionEditor from 'components/NodeConditionEditor';
import { convertBackendToUI, convertUIToBackend } from 'utils/unlockConditionConverter';

export default function RuleDialog({ open, onClose, onSave, editingRule, tags }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const [name, setName] = useState('');
  const [tagName, setTagName] = useState('');
  const [enabled, setEnabled] = useState(true);
  const [triggerType, setTriggerType] = useState('subscription_update');
  const [logic, setLogic] = useState('and');
  const [conditions, setConditions] = useState([{ field: 'link_country', operator: 'equals', value: '' }]);

  useEffect(() => {
    if (editingRule) {
      setName(editingRule.name || '');
      setTagName(editingRule.tagName || '');
      setEnabled(editingRule.enabled !== false);
      setTriggerType(editingRule.triggerType || 'subscription_update');
      try {
        const parsed = JSON.parse(editingRule.conditions || '{}');
        setLogic(parsed.logic || 'and');
        // 转换后端格式到UI格式（合并解锁字段）
        const backendConditions =
          parsed.conditions?.length > 0 ? parsed.conditions : [{ field: 'link_country', operator: 'equals', value: '' }];
        const uiConditions = convertBackendToUI(backendConditions);
        setConditions(uiConditions);
      } catch {
        setLogic('and');
        setConditions([{ field: 'link_country', operator: 'equals', value: '' }]);
      }
    } else {
      setName('');
      setTagName('');
      setEnabled(true);
      setTriggerType('subscription_update');
      setLogic('and');
      setConditions([{ field: 'link_country', operator: 'equals', value: '' }]);
    }
  }, [editingRule, open]);

  const handleSave = () => {
    if (!name.trim() || !tagName) return;

    // 转换UI格式到后端格式（展开解锁条件）
    const backendConditions = convertUIToBackend(conditions);
    const conditionsJson = JSON.stringify({ logic, conditions: backendConditions });

    onSave({
      name: name.trim(),
      tagName: tagName,
      enabled,
      triggerType,
      conditions: conditionsJson
    });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth fullScreen={isMobile}>
      <DialogTitle>{editingRule ? t('tags.dialog.rule.editTitle') : t('tags.dialog.rule.addTitle')}</DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', gap: 2 }}>
            <TextField label={t('tags.fields.ruleName')} value={name} onChange={(e) => setName(e.target.value)} fullWidth required />
            <FormControl sx={{ minWidth: isMobile ? '100%' : 150 }} fullWidth={isMobile}>
              <InputLabel>{t('tags.fields.tag')}</InputLabel>
              <Select value={tagName} label={t('tags.fields.tag')} onChange={(e) => setTagName(e.target.value)}>
                {tags.map((tag) => (
                  <MenuItem key={tag.name} value={tag.name}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Box sx={{ width: 12, height: 12, borderRadius: '50%', backgroundColor: tag.color }} />
                      {tag.name}
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>

          <Box
            sx={{
              display: 'flex',
              flexDirection: isMobile ? 'column' : 'row',
              gap: 2,
              alignItems: isMobile ? 'stretch' : 'center'
            }}
          >
            <FormControl sx={{ minWidth: isMobile ? '100%' : 150 }} fullWidth={isMobile}>
              <InputLabel>{t('tags.fields.trigger')}</InputLabel>
              <Select value={triggerType} label={t('tags.fields.trigger')} onChange={(e) => setTriggerType(e.target.value)}>
                <MenuItem value="subscription_update">{t('tags.trigger.subscriptionUpdate')}</MenuItem>
                <MenuItem value="speed_test">{t('tags.trigger.speedTest')}</MenuItem>
              </Select>
            </FormControl>
            <FormControlLabel
              control={<Switch checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />}
              label={t('tags.dialog.rule.enableRule')}
            />
          </Box>

          <Divider sx={{ my: 1 }} />

          <NodeConditionEditor
            logic={logic}
            conditions={conditions}
            onChange={(next) => {
              setLogic(next.logic);
              setConditions(next.conditions);
            }}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={handleSave} disabled={!name.trim() || !tagName}>
          {t('common.save')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

RuleDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  editingRule: PropTypes.object,
  tags: PropTypes.array.isRequired
};
