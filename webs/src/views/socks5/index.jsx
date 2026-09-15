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
import Socks5Settings from './components/Socks5Settings';
import { closeAllSocks5Connections, closeSocks5Connection, getSocks5Connections, getSocks5Status } from 'api/settings';

function Socks5Monitor({ showMessage }) {
  const { t } = useTranslation();
  const [status, setStatus] = useState(null);
  const [connections, setConnections] = useState([]);
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
  const disconnectAll = async () => {
    try {
      await closeAllSocks5Connections();
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    }
  };
  const stats = status?.stats || {};
  return (
    <Card variant="outlined" sx={{ mt: 2 }}>
      <CardHeader
        title={t('settings.socks5.monitor.title')}
        subheader={t('settings.socks5.monitor.subheader')}
        action={
          <Button color="error" onClick={disconnectAll} disabled={!connections.length}>
            {t('settings.socks5.monitor.disconnectAll')}
          </Button>
        }
      />
      <CardContent>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
          <Chip label={t('settings.socks5.monitor.activeConnections', { count: stats.activeConnections ?? 0 })} />
          <Chip label={t('settings.socks5.monitor.totalConnections', { count: stats.totalConnections ?? 0 })} />
          <Chip label={t('settings.socks5.monitor.successfulConnections', { count: stats.successfulConnections ?? 0 })} color="success" />
          <Chip label={t('settings.socks5.monitor.failedConnections', { count: stats.failedConnections ?? 0 })} color="error" />
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
                  {connection.clientAddress} · {connection.phase}
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
