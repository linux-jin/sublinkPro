import { useCallback, useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Snackbar from '@mui/material/Snackbar';

import { useAuth } from 'contexts/AuthContext';
import { useTranslation } from 'react-i18next';
import { formatBytes } from 'views/airports/utils';
import Socks5Settings from './components/Socks5Settings';
import { closeAllSocks5Connections, closeSocks5Connection, getSocks5Connections, getSocks5Status, probeSocks5Health } from 'api/settings';

function formatElapsed(value) {
  if (!value) return '-';
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000));
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}

function healthColor(status) {
  if (status === 'healthy') return 'success';
  if (status === 'unhealthy') return 'error';
  if (status === 'checking') return 'warning';
  return 'default';
}

function Socks5Monitor({ showMessage }) {
  const { t } = useTranslation();
  const [status, setStatus] = useState(null);
  const [connections, setConnections] = useState([]);
  const [probing, setProbing] = useState(false);
  const load = useCallback(async () => {
    try {
      const [statusResponse, connectionsResponse] = await Promise.all([getSocks5Status(), getSocks5Connections()]);
      setStatus(statusResponse.data);
      setConnections(Array.isArray(connectionsResponse.data) ? connectionsResponse.data : []);
    } catch (error) {
      if (error.response?.status !== 404) showMessage(error.response?.data?.msg || error.message, 'error');
    }
  }, [showMessage]);
  useEffect(() => {
    load();
    const timer = setInterval(() => {
      if (document.visibilityState === 'visible') load();
    }, 3000);
    return () => clearInterval(timer);
  }, [load]);
  const disconnect = async (id) => {
    try {
      await closeSocks5Connection(id);
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    }
  };
  const probeHealth = async () => {
    setProbing(true);
    try {
      const response = await probeSocks5Health();
      showMessage(
        t(response.data?.started === false ? 'settings.socks5.messages.healthProbeRunning' : 'settings.socks5.messages.healthProbeStarted')
      );
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    } finally {
      setProbing(false);
    }
  };
  const disconnectAll = async () => {
    try {
      await closeAllSocks5Connections();
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    }
  };
  const stats = status?.stats || {};
  const health = status?.health || { nodes: [] };
  return (
    <Card variant="outlined" sx={{ mt: 2 }}>
      <CardHeader
        title={t('settings.socks5.monitor.title')}
        subheader={t('settings.socks5.monitor.subheader')}
        action={
          <Stack direction="row" spacing={1}>
            <Button onClick={probeHealth} disabled={probing || health.running || !status?.config?.running}>
              {probing || health.running ? t('settings.socks5.monitor.probing') : t('settings.socks5.monitor.probeNow')}
            </Button>
            <Button color="error" onClick={disconnectAll} disabled={!connections.length}>
              {t('settings.socks5.monitor.disconnectAll')}
            </Button>
          </Stack>
        }
      />
      <CardContent>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
          <Chip label={t('settings.socks5.monitor.activeConnections', { count: stats.activeConnections ?? 0 })} />
          <Chip label={t('settings.socks5.monitor.totalConnections', { count: stats.totalConnections ?? 0 })} />
          <Chip label={t('settings.socks5.monitor.successfulConnections', { count: stats.successfulConnections ?? 0 })} color="success" />
          <Chip label={t('settings.socks5.monitor.failedConnections', { count: stats.failedConnections ?? 0 })} color="error" />
          <Chip label={t('settings.socks5.monitor.uploadBytes', { value: formatBytes(stats.uploadBytes ?? 0) })} />
          <Chip label={t('settings.socks5.monitor.downloadBytes', { value: formatBytes(stats.downloadBytes ?? 0) })} />
        </Stack>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ mt: 1.5 }}>
          <Chip label={t('settings.socks5.monitor.healthyNodes', { count: health.healthyNodes ?? 0 })} color="success" />
          <Chip label={t('settings.socks5.monitor.unhealthyNodes', { count: health.unhealthyNodes ?? 0 })} color="error" />
          <Chip label={t('settings.socks5.monitor.checkingNodes', { count: health.checkingNodes ?? 0 })} color="warning" />
          <Chip label={t('settings.socks5.monitor.unknownNodes', { count: health.unknownNodes ?? 0 })} />
        </Stack>
        <Box sx={{ mt: 2 }}>
          {connections.length ? (
            connections.map((connection) => (
              <Stack
                key={connection.id}
                direction={{ xs: 'column', md: 'row' }}
                spacing={1}
                alignItems={{ md: 'center' }}
                sx={{ py: 1, borderBottom: 1, borderColor: 'divider' }}
              >
                <Typography sx={{ flex: 1 }}>
                  {connection.target || t('settings.socks5.monitor.connecting')} ·{' '}
                  {connection.nodeName || t('settings.socks5.monitor.waitingNode')}
                </Typography>
                <Typography variant="caption">
                  {connection.clientAddress} · {connection.phase} · ↑ {formatBytes(connection.uploadBytes ?? 0)} · ↓{' '}
                  {formatBytes(connection.downloadBytes ?? 0)} ·{' '}
                  {t('settings.socks5.monitor.duration', { value: formatElapsed(connection.startedAt) })} ·{' '}
                  {t('settings.socks5.monitor.idle', { value: formatElapsed(connection.lastActivity) })}
                </Typography>
                <Button size="small" color="error" onClick={() => disconnect(connection.id)}>
                  {t('settings.socks5.monitor.disconnect')}
                </Button>
              </Stack>
            ))
          ) : (
            <Typography color="text.secondary">{t('settings.socks5.monitor.empty')}</Typography>
          )}
        </Box>
        <Box sx={{ mt: 2 }}>
          <Typography variant="subtitle1" sx={{ mb: 1 }}>
            {t('settings.socks5.monitor.nodeHealth')}
          </Typography>
          {health.truncated && (
            <Alert severity="info" sx={{ mb: 1 }}>
              {t('settings.socks5.monitor.healthTruncated', { shown: health.nodes?.length ?? 0, total: health.totalNodes ?? 0 })}
            </Alert>
          )}
          {health.nodes?.length ? (
            health.nodes.map((node) => (
              <Stack
                key={`${node.nodeId}-${node.nodeName}`}
                direction={{ xs: 'column', md: 'row' }}
                spacing={1}
                alignItems={{ md: 'center' }}
                sx={{ py: 1, borderBottom: 1, borderColor: 'divider' }}
              >
                <Typography sx={{ flex: 1 }}>{node.nodeName || `#${node.nodeId}`}</Typography>
                <Chip
                  size="small"
                  color={healthColor(node.status)}
                  label={t(`settings.socks5.monitor.healthStatus.${node.status || 'unknown'}`)}
                />
                <Typography variant="caption">
                  {node.latencyMs > 0 ? `${node.latencyMs} ms` : '-'} ·{' '}
                  {t('settings.socks5.monitor.failures', { count: node.consecutiveFailures ?? 0 })}
                  {node.cooldownUntil
                    ? ` · ${t('settings.socks5.monitor.cooldownUntil', { value: new Date(node.cooldownUntil).toLocaleString() })}`
                    : ''}
                </Typography>
                {node.lastError && (
                  <Typography variant="caption" color="error" sx={{ maxWidth: 360 }} noWrap title={node.lastError}>
                    {node.lastError}
                  </Typography>
                )}
              </Stack>
            ))
          ) : (
            <Typography color="text.secondary">{t('settings.socks5.monitor.noHealthData')}</Typography>
          )}
        </Box>
      </CardContent>
    </Card>
  );
}

export default function Socks5GatewayPage() {
  const { user } = useAuth();
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const isAdmin = [user?.role, ...(user?.roles || [])].some((role) => String(role || '').toLowerCase() === 'admin');

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const handleSnackbarClose = useCallback(() => {
    setSnackbar((previous) => ({ ...previous, open: false }));
  }, []);

  if (!isAdmin) {
    return <Navigate to="/system/settings" replace />;
  }

  return (
    <>
      <Socks5Settings showMessage={showMessage} />
      <Socks5Monitor showMessage={showMessage} />
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleSnackbarClose}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </>
  );
}
