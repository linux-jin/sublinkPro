import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTaskProgress } from 'contexts/TaskProgressContext';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  CircularProgress,
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
import { getNodeCountries, getNodeGroups } from 'api/nodes';
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

const emptyForm = { name: '', countries: '', keyword: '', sourceGroups: [], maxDelay: 0, minSpeed: 0, maxAgeHours: 72 };
const parseCountries = (value) => (value || '').split(',').filter(Boolean);
// Intl's region names provide a full locale-aware country/territory catalog rather than only countries already found on nodes.
const regionCodes = Array.from({ length: 26 * 26 }, (_, index) => String.fromCharCode(65 + Math.floor(index / 26), 65 + (index % 26)));
const nonCountryRegions = new Set(['EU', 'UN', 'EZ', 'QO', 'ZZ']);
const exclusionKeys = ['delayUnusable', 'delayOverLimit', 'delayStale', 'speedUnusable', 'speedBelowMin', 'speedStale'];

export default function SmartGroupsPage() {
  const { t, i18n } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();
  const [groups, setGroups] = useState([]);
  const [countries, setCountries] = useState([]);
  const [sourceGroups, setSourceGroups] = useState([]);
  const [profiles, setProfiles] = useState([]);
  const [profileId, setProfileId] = useState('');
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState(emptyForm);
  const [view, setView] = useState(null);
  const [members, setMembers] = useState(null);
  const [memberPage, setMemberPage] = useState(1);
  const [loadingMembers, setLoadingMembers] = useState(false);
  const [checkMessage, setCheckMessage] = useState('');
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');
  const displayNames = useMemo(
    () => new Intl.DisplayNames([i18n.resolvedLanguage || i18n.language], { type: 'region' }),
    [i18n.resolvedLanguage, i18n.language]
  );
  const countryOptions = useMemo(
    () =>
      [...new Set([...countries, ...regionCodes.filter((code) => !nonCountryRegions.has(code) && displayNames.of(code) !== code)])].sort(),
    [countries, displayNames]
  );
  const groupOptions = useMemo(
    () => [...new Set([...sourceGroups, ...(form.sourceGroups || [])])].sort(),
    [sourceGroups, form.sourceGroups]
  );
  const countryLabel = (code) => {
    const name = displayNames.of(code);
    return name && name !== code ? `${formatCountry(code)} — ${name}` : formatCountry(code);
  };

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
      .then((response) => {
        const items = response.data || [];
        setProfiles(items);
        setProfileId((current) => current || String(items[0]?.id ?? items[0]?.ID ?? ''));
      })
      .catch(() => {});
    void getNodeGroups()
      .then((response) => setSourceGroups(response.data || []))
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

  const showMembers = useCallback(
    async (group, page = 1) => {
      setView(group);
      setMemberPage(page);
      setLoadingMembers(true);
      try {
        const response = await getSmartGroupMembers(group.id, page);
        setMembers(response.data);
      } catch (error) {
        setCheckMessage(error.message || t('smartGroups.loadFailed'));
      } finally {
        setLoadingMembers(false);
      }
    },
    [t]
  );

  useEffect(() => {
    if (!view) return undefined;
    const onComplete = ({ taskType }) => {
      if (taskType === 'speed_test') void showMembers(view, memberPage);
    };
    registerOnComplete(onComplete);
    return () => unregisterOnComplete(onComplete);
  }, [view, memberPage, showMembers, registerOnComplete, unregisterOnComplete]);

  const runCheck = async (group, nodeId = 0) => {
    if (!profileId) {
      setCheckMessage(t('smartGroups.chooseProfile'));
      return;
    }
    setBusy(true);
    setCheckMessage('');
    try {
      const response = await checkSmartGroup(group.id, Number(profileId), nodeId);
      setCheckMessage(t('smartGroups.started', { count: response.data?.count || 0 }));
    } catch (error) {
      setCheckMessage(error.message || t('smartGroups.checkFailed'));
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
                  {group.keyword && ` · ${t('smartGroups.keyword')}: ${group.keyword}`}
                  {group.sourceGroups?.length > 0 && ` · ${t('smartGroups.sourceGroups')}: ${group.sourceGroups.join(' / ')}`}
                </Typography>
              </Box>
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <Button
                  disabled={busy}
                  onClick={() => {
                    setMembers(null);
                    setCheckMessage('');
                    void showMembers(group);
                  }}
                >
                  {t('smartGroups.members')}
                </Button>
                <Button
                  disabled={busy}
                  onClick={() => {
                    setForm({ ...group, sourceGroups: group.sourceGroups || [] });
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
      </Stack>
      <Dialog open={view !== null} onClose={() => setView(null)} fullWidth maxWidth="lg" fullScreen={isMobile}>
        <DialogTitle>
          {view?.name} — {t('smartGroups.members')}
        </DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Typography variant="body2">{t('smartGroups.dialogHint')}</Typography>
            {checkMessage && (
              <Alert severity="info" onClose={() => setCheckMessage('')}>
                {checkMessage}
              </Alert>
            )}
            {members && (
              <Typography variant="body2">
                {t('smartGroups.candidateSummary', { count: members.candidateCount ?? 0, healthy: members.count ?? 0 })}
                {members.count === 0 && ` ${t(members.candidateCount ? 'smartGroups.noHealthyHint' : 'smartGroups.noCandidatesHint')}`}
              </Typography>
            )}
            {members?.statusCounts && members.candidateCount > members.count && (
              <Typography variant="caption" color="text.secondary">
                {exclusionKeys
                  .filter((key) => members.statusCounts[key] > 0)
                  .map((key) => t(`smartGroups.exclusions.${key}`, { count: members.statusCounts[key] }))
                  .join(' · ')}
              </Typography>
            )}
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ sm: 'center' }}>
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
              <Button variant="contained" disabled={busy || !members?.candidateCount} onClick={() => void runCheck(view)}>
                {t('smartGroups.check')}
              </Button>
              <Button disabled={loadingMembers} onClick={() => void showMembers(view, memberPage)}>
                {t('smartGroups.refresh')}
              </Button>
            </Stack>
            {view?.minSpeed > 0 && profiles.find((profile) => String(profile.id ?? profile.ID) === String(profileId))?.mode === 'tcp' && (
              <Alert severity="warning">{t('smartGroups.tcpWarning')}</Alert>
            )}
            {loadingMembers && <CircularProgress size={24} />}
            {members?.candidateCount === 0 && <Alert severity="info">{t('smartGroups.noCandidatesHint')}</Alert>}
            {members?.candidateCount > 0 && (
              <TableContainer component={Paper} variant="outlined" sx={{ maxHeight: isMobile ? 'none' : 520 }}>
                <Table size="small" stickyHeader>
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>{t('smartGroups.node')}</TableCell>
                      <TableCell>{t('smartGroups.sourceGroup')}</TableCell>
                      <TableCell>{t('smartGroups.country')}</TableCell>
                      <TableCell>{t('smartGroups.delay')}</TableCell>
                      <TableCell>{t('smartGroups.speed')}</TableCell>
                      <TableCell>{t('smartGroups.status')}</TableCell>
                      <TableCell>{t('smartGroups.check')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {(members.candidateNodes || []).map((node) => (
                      <TableRow key={node.id}>
                        <TableCell>{node.id}</TableCell>
                        <TableCell>{node.name}</TableCell>
                        <TableCell>{node.group || '—'}</TableCell>
                        <TableCell>
                          {formatCountry(node.country)}
                          {node.countrySource === 'name' && (
                            <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                              {t('smartGroups.nameInferred')}
                            </Typography>
                          )}
                        </TableCell>
                        <TableCell>{node.delay > 0 ? `${node.delay} ms` : '—'}</TableCell>
                        <TableCell>{node.speed > 0 ? `${node.speed} MB/s` : '—'}</TableCell>
                        <TableCell>{node.reason ? t(`smartGroups.reasons.${node.reason}`) : t('smartGroups.available')}</TableCell>
                        <TableCell>
                          <Button size="small" disabled={busy || !profileId} onClick={() => void runCheck(view, node.id)}>
                            {t('smartGroups.checkOne')}
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
            {members && members.candidateCount > (members.pageSize || 30) && (
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={2}>
                <Button disabled={loadingMembers || memberPage <= 1} onClick={() => void showMembers(view, memberPage - 1)}>
                  {t('smartGroups.previous')}
                </Button>
                <Typography variant="body2">
                  {t('smartGroups.page', { page: memberPage, total: Math.ceil(members.candidateCount / (members.pageSize || 30)) })}
                </Typography>
                <Button
                  disabled={loadingMembers || memberPage * (members.pageSize || 30) >= members.candidateCount}
                  onClick={() => void showMembers(view, memberPage + 1)}
                >
                  {t('smartGroups.next')}
                </Button>
              </Stack>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setView(null)}>{t('smartGroups.close')}</Button>
        </DialogActions>
      </Dialog>
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
              options={countryOptions}
              value={parseCountries(form.countries)}
              onChange={(_event, values) => setForm({ ...form, countries: values.map((value) => value.trim().toUpperCase()).join(',') })}
              getOptionLabel={countryLabel}
              renderInput={(params) => (
                <TextField {...params} required label={t('smartGroups.country')} helperText={t('smartGroups.countryHint')} />
              )}
            />
            <TextField
              label={t('smartGroups.keyword')}
              value={form.keyword || ''}
              onChange={(event) => setForm({ ...form, keyword: event.target.value })}
              helperText={t('smartGroups.keywordHint')}
            />
            <Autocomplete
              multiple
              options={groupOptions}
              value={form.sourceGroups || []}
              onChange={(_event, values) => setForm({ ...form, sourceGroups: values })}
              renderInput={(params) => (
                <TextField {...params} label={t('smartGroups.sourceGroups')} helperText={t('smartGroups.sourceGroupsHint')} />
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
