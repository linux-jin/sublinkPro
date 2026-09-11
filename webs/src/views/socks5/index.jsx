import { useCallback, useState } from 'react';
import { Navigate } from 'react-router-dom';

import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';

import { useAuth } from 'contexts/AuthContext';
import Socks5Settings from './components/Socks5Settings';

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
