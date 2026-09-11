import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CircularProgress from '@mui/material/CircularProgress';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import FormControlLabel from '@mui/material/FormControlLabel';

import HubIcon from '@mui/icons-material/Hub';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import StopCircleIcon from '@mui/icons-material/StopCircle';

import { getNodeSelector } from 'api/nodes';
import { getSocks5Settings, stopSocks5, updateSocks5Settings } from 'api/settings';

const defaultConfig = {
  enabled: false,
  listenAddress: '127.0.0.1',
  port: 1080,
  username: '',
  hasPassword: false,
  maskedPassword: '',
  nodeId: 0,
  selection: 'best',
  requireAuth: true,
  running: false,
  boundAddress: ''
};

function messageFromError(t, error, fallbackKey) {
  const response = error.response?.data;
  if (response?.i18nKey) return t(response.i18nKey, response.i18nParams || {});
  return t(fallbackKey, { message: response?.msg || error.message });
}

export default function Socks5Settings({ showMessage }) {
  const { t } = useTranslation();
  const [config, setConfig] = useState(defaultConfig);
  const [form, setForm] = useState({ ...defaultConfig, password: '', clearPassword: false });
  const [nodes, setNodes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [stopping, setStopping] = useState(false);

  const syncConfig = useCallback((next) => {
    const value = { ...defaultConfig, ...(next || {}) };
    setConfig(value);
    setForm({ ...value, password: '', clearPassword: false });
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [settingsResponse, nodesResponse] = await Promise.all([getSocks5Settings(), getNodeSelector({ page: 1, pageSize: 100 })]);
      syncConfig(settingsResponse.data);
      const items = nodesResponse.data?.items || nodesResponse.data || [];
      setNodes(Array.isArray(items) ? items : []);
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.socks5.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  }, [showMessage, syncConfig, t]);

  useEffect(() => {
    load();
  }, [load]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const response = await updateSocks5Settings({
        enabled: Boolean(form.enabled),
        listenAddress: form.listenAddress.trim(),
        port: Number(form.port) || 1080,
        username: form.username.trim(),
        password: form.password,
        clearPassword: Boolean(form.clearPassword),
        nodeId: Number(form.nodeId) || 0,
        selection: form.selection,
        requireAuth: Boolean(form.requireAuth)
      });
      syncConfig(response.data);
      showMessage(t('settings.socks5.messages.saved'));
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.socks5.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleStop = async () => {
    setStopping(true);
    try {
      const response = await stopSocks5();
      syncConfig(response.data);
      showMessage(t('settings.socks5.messages.stopped'));
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.socks5.messages.stopFailed'), 'error');
    } finally {
      setStopping(false);
    }
  };

  const busy = loading || saving || stopping;
  const specificInvalid = form.selection === 'specific' && !Number(form.nodeId);
  const authInvalid = form.enabled && form.requireAuth && (!form.username.trim() || (!form.password && !config.hasPassword));

  return (
    <Card variant="outlined">
      <CardHeader avatar={<HubIcon color="primary" />} title={t('settings.socks5.title')} subheader={t('settings.socks5.subheader')} />
      <CardContent>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={28} />
          </Box>
        ) : (
          <Stack spacing={2.5}>
            <Alert severity="warning">{t('settings.socks5.alerts.security')}</Alert>
            <Alert severity={config.running ? 'success' : 'info'}>
              {config.running
                ? t('settings.socks5.status.running', { address: config.boundAddress || `${config.listenAddress}:${config.port}` })
                : t('settings.socks5.status.stopped')}
            </Alert>
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.enabled)}
                  onChange={(event) => setForm((prev) => ({ ...prev, enabled: event.target.checked }))}
                />
              }
              label={t('settings.socks5.form.enabled')}
            />
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label={t('settings.socks5.form.listenAddress')}
                value={form.listenAddress}
                onChange={(event) => setForm((prev) => ({ ...prev, listenAddress: event.target.value }))}
                helperText={t('settings.socks5.form.listenAddressHelper')}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.port')}
                value={form.port}
                onChange={(event) => setForm((prev) => ({ ...prev, port: event.target.value }))}
                slotProps={{ htmlInput: { min: 1, max: 65535 } }}
              />
            </Stack>
            <FormControl fullWidth>
              <InputLabel>{t('settings.socks5.form.selection')}</InputLabel>
              <Select
                label={t('settings.socks5.form.selection')}
                value={form.selection}
                onChange={(event) => setForm((prev) => ({ ...prev, selection: event.target.value }))}
              >
                <MenuItem value="best">{t('settings.socks5.form.selectionBest')}</MenuItem>
                <MenuItem value="random">{t('settings.socks5.form.selectionRandom')}</MenuItem>
                <MenuItem value="specific">{t('settings.socks5.form.selectionSpecific')}</MenuItem>
              </Select>
            </FormControl>
            {form.selection === 'specific' && (
              <FormControl fullWidth error={specificInvalid}>
                <InputLabel>{t('settings.socks5.form.node')}</InputLabel>
                <Select
                  label={t('settings.socks5.form.node')}
                  value={Number(form.nodeId) || 0}
                  onChange={(event) => setForm((prev) => ({ ...prev, nodeId: Number(event.target.value) }))}
                >
                  <MenuItem value={0}>{t('settings.socks5.form.nodePlaceholder')}</MenuItem>
                  {nodes.map((node) => (
                    <MenuItem value={node.id} key={node.id}>
                      {node.effectiveName || node.name || node.linkName || `#${node.id}`}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.requireAuth)}
                  onChange={(event) => setForm((prev) => ({ ...prev, requireAuth: event.target.checked }))}
                />
              }
              label={t('settings.socks5.form.requireAuth')}
            />
            {form.requireAuth && (
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  fullWidth
                  label={t('settings.socks5.form.username')}
                  value={form.username}
                  onChange={(event) => setForm((prev) => ({ ...prev, username: event.target.value }))}
                />
                <TextField
                  fullWidth
                  type="password"
                  label={t('settings.socks5.form.password')}
                  value={form.password}
                  onChange={(event) => setForm((prev) => ({ ...prev, password: event.target.value, clearPassword: false }))}
                  placeholder={config.maskedPassword || ''}
                  disabled={form.clearPassword}
                  helperText={config.hasPassword ? t('settings.socks5.form.passwordSaved') : t('settings.socks5.form.passwordHelper')}
                />
              </Stack>
            )}
            {config.hasPassword && form.requireAuth && (
              <FormControlLabel
                control={
                  <Switch
                    checked={form.clearPassword}
                    onChange={(event) => setForm((prev) => ({ ...prev, clearPassword: event.target.checked, password: '' }))}
                  />
                }
                label={t('settings.socks5.form.clearPassword')}
              />
            )}
            {authInvalid && <Alert severity="error">{t('settings.socks5.messages.authRequired')}</Alert>}
            <Stack direction="row" spacing={1.5} useFlexGap flexWrap="wrap">
              <Button variant="contained" startIcon={<SaveIcon />} onClick={handleSave} disabled={busy || specificInvalid || authInvalid}>
                {saving ? t('settings.socks5.actions.saving') : t('settings.socks5.actions.save')}
              </Button>
              <Button variant="outlined" startIcon={<RefreshIcon />} onClick={load} disabled={busy}>
                {t('settings.socks5.actions.refresh')}
              </Button>
              <Button
                color="error"
                variant="outlined"
                startIcon={<StopCircleIcon />}
                onClick={handleStop}
                disabled={busy || !config.running}
              >
                {stopping ? t('settings.socks5.actions.stopping') : t('settings.socks5.actions.stop')}
              </Button>
            </Stack>
          </Stack>
        )}
      </CardContent>
    </Card>
  );
}
