import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography
} from '@mui/material';
import MainCard from 'ui-component/cards/MainCard';
import { getNodeCountries } from 'api/nodes';
import { getNodeCheckProfiles } from 'api/nodeCheck';
import {
  checkSmartGroup,
  createSmartGroup,
  deleteSmartGroup,
  getSmartGroupMembers,
  getSmartGroups,
  updateSmartGroup
} from 'api/smartGroups';
import { formatCountry } from 'utils/countryDisplay';

const emptyForm = { name: '', countries: '', maxDelay: 500, minSpeed: 0, maxAgeHours: 72 };
const parseCountries = (value) => (value || '').split(',').filter(Boolean);

export default function SmartGroupsPage() {
  const { t } = useTranslation();
  const [groups, setGroups] = useState([]);
  const [countries, setCountries] = useState([]);
  const [profiles, setProfiles] = useState([]);
  const [profileId, setProfileId] = useState('');
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState(emptyForm);
  const [view, setView] = useState(null);
  const [members, setMembers] = useState(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');

  const refresh = useCallback(async () => {
    try {
      const response = await getSmartGroups();
      setGroups(response.data || []);
    } catch (error) {
      setMessage(error.message || t('smartGroups.loadFailed'));
    }
  }, [t]);

  useEffect(() => {
    void refresh();
    void getNodeCountries()
      .then((response) => setCountries(response.data || []))
      .catch(() => {});
    void getNodeCheckProfiles()
      .then((response) => setProfiles(response.data || []))
      .catch(() => {});
  }, [refresh]);

  const save = async () => {
    setBusy(true);
    try {
      if (editing?.id) await updateSmartGroup(editing.id, form);
      else await createSmartGroup(form);
      setEditing(null);
      await refresh();
    } catch (error) {
      setMessage(error.message || t('smartGroups.saveFailed'));
    } finally {
      setBusy(false);
    }
  };

  const showMembers = async (group) => {
    setView(group);
    setMembers(null);
    try {
      const response = await getSmartGroupMembers(group.id);
      setMembers(response.data);
    } catch (error) {
      setMessage(error.message || t('smartGroups.loadFailed'));
    }
  };

  const runCheck = async (group) => {
    if (!profileId) {
      setMessage(t('smartGroups.chooseProfile'));
      return;
    }
    setBusy(true);
    try {
      const response = await checkSmartGroup(group.id, Number(profileId));
      setMessage(t('smartGroups.started', { count: response.data?.count || 0 }));
    } catch (error) {
      setMessage(error.message || t('smartGroups.checkFailed'));
    } finally {
      setBusy(false);
    }
  };

  const remove = async (group) => {
    if (!window.confirm(t('smartGroups.confirmDelete', { name: group.name }))) return;
    setBusy(true);
    try {
      await deleteSmartGroup(group.id);
      if (view?.id === group.id) setView(null);
      await refresh();
    } catch (error) {
      setMessage(error.message || t('smartGroups.deleteFailed'));
    } finally {
      setBusy(false);
    }
  };

  return (
    <MainCard title={t('smartGroups.title')}>
      <Stack spacing={2}>
        <Typography variant="body2">{t('smartGroups.description')}</Typography>
        {message && (
          <Alert severity="info" onClose={() => setMessage('')}>
            {message}
          </Alert>
        )}
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
          <Button
            variant="contained"
            onClick={() => {
              setForm({ ...emptyForm });
              setEditing({});
            }}
          >
            {t('smartGroups.add')}
          </Button>
          <TextField
            select
            size="small"
            label={t('smartGroups.profile')}
            value={profileId}
            onChange={(event) => setProfileId(event.target.value)}
            sx={{ minWidth: 220 }}
          >
            {profiles.map((profile) => (
              <MenuItem key={profile.id ?? profile.ID} value={profile.id ?? profile.ID}>
                {profile.name ?? profile.Name}
              </MenuItem>
            ))}
          </TextField>
          <Button variant="outlined" onClick={() => void refresh()}>
            {t('smartGroups.refresh')}
          </Button>
        </Stack>
        {groups.map((group) => (
          <Paper key={group.id} variant="outlined" sx={{ p: 2 }}>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems={{ md: 'center' }} justifyContent="space-between">
              <Box>
                <Typography variant="h4">{group.name}</Typography>
                <Typography variant="body2">
                  {parseCountries(group.countries).map(formatCountry).join(' / ')} ·{' '}
                  {t('smartGroups.rule', { delay: group.maxDelay || '∞', speed: group.minSpeed || 0, age: group.maxAgeHours || '∞' })}
                </Typography>
              </Box>
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <Button disabled={busy} onClick={() => void showMembers(group)}>
                  {t('smartGroups.members')}
                </Button>
                <Button disabled={busy} onClick={() => void runCheck(group)}>
                  {t('smartGroups.check')}
                </Button>
                <Button
                  disabled={busy}
                  onClick={() => {
                    setForm({ ...group });
                    setEditing(group);
                  }}
                >
                  {t('smartGroups.edit')}
                </Button>
                <Button color="error" disabled={busy} onClick={() => void remove(group)}>
                  {t('smartGroups.delete')}
                </Button>
              </Stack>
            </Stack>
          </Paper>
        ))}
        {view && (
          <Box>
            <Typography variant="h4">
              {view.name} — {t('smartGroups.members')} ({members?.count ?? '…'})
            </Typography>
            <Button onClick={() => void showMembers(view)}>{t('smartGroups.refresh')}</Button>
            {members && (
              <TableContainer component={Paper} variant="outlined">
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>{t('smartGroups.node')}</TableCell>
                      <TableCell>{t('smartGroups.sourceGroup')}</TableCell>
                      <TableCell>{t('smartGroups.country')}</TableCell>
                      <TableCell>{t('smartGroups.delay')}</TableCell>
                      <TableCell>{t('smartGroups.speed')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {(members.nodes || []).slice(0, 100).map((node) => (
                      <TableRow key={node.id}>
                        <TableCell>{node.id}</TableCell>
                        <TableCell>{node.name}</TableCell>
                        <TableCell>{node.group}</TableCell>
                        <TableCell>{formatCountry(node.country)}</TableCell>
                        <TableCell>{node.delay} ms</TableCell>
                        <TableCell>{node.speed} MB/s</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
            {members?.count > 100 && <Typography variant="caption">{t('smartGroups.firstHundred')}</Typography>}
          </Box>
        )}
      </Stack>
      <Dialog open={editing !== null} onClose={() => setEditing(null)} fullWidth maxWidth="sm">
        <DialogTitle>{editing?.id ? t('smartGroups.edit') : t('smartGroups.add')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <TextField
              label={t('smartGroups.name')}
              required
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
            />
            <Autocomplete
              multiple
              freeSolo
              options={countries}
              value={parseCountries(form.countries)}
              onChange={(_event, values) => setForm({ ...form, countries: values.map((value) => value.trim().toUpperCase()).join(',') })}
              getOptionLabel={(option) => formatCountry(option)}
              renderInput={(params) => (
                <TextField {...params} required label={t('smartGroups.country')} helperText={t('smartGroups.countryHint')} />
              )}
            />
            <TextField
              label={t('smartGroups.maxDelay')}
              type="number"
              value={form.maxDelay}
              onChange={(event) => setForm({ ...form, maxDelay: Number(event.target.value) })}
              helperText={t('smartGroups.zeroUnlimited')}
            />
            <TextField
              label={t('smartGroups.minSpeed')}
              type="number"
              value={form.minSpeed}
              onChange={(event) => setForm({ ...form, minSpeed: Number(event.target.value) })}
              helperText={t('smartGroups.speedHint')}
            />
            <TextField
              label={t('smartGroups.maxAge')}
              type="number"
              value={form.maxAgeHours}
              onChange={(event) => setForm({ ...form, maxAgeHours: Number(event.target.value) })}
              helperText={t('smartGroups.zeroUnlimited')}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditing(null)}>{t('smartGroups.cancel')}</Button>
          <Button disabled={busy} variant="contained" onClick={() => void save()}>
            {t('smartGroups.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
