import { lazy } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router-dom';
import { Box, Paper, Stack, Tab, Tabs, Typography } from '@mui/material';
import LocalOfferOutlinedIcon from '@mui/icons-material/LocalOfferOutlined';
import FilterAltOutlinedIcon from '@mui/icons-material/FilterAltOutlined';
import Loadable from 'ui-component/Loadable';
import { getClassificationLocation, getClassificationTab } from './navigation';

const TagManagement = Loadable(lazy(() => import('views/tags')));
const SmartGroupsPage = Loadable(lazy(() => import('views/smart-groups')));

export default function NodeClassificationPage() {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const tab = getClassificationTab(location.search);

  const changeTab = (_, value) => {
    navigate(getClassificationLocation(location, value), { state: location.state, preventScrollReset: true });
  };

  return (
    <Stack spacing={2} useFlexGap sx={{ minWidth: 0 }}>
      <Paper variant="outlined" sx={{ px: { xs: 1, sm: 2 }, borderRadius: 2 }}>
        <Tabs value={tab} onChange={changeTab} aria-label={t('nodeClassification.title')} variant="scrollable" scrollButtons="auto">
          <Tab
            value="tags"
            id="classification-tab-tags"
            aria-controls="classification-panel-tags"
            icon={<LocalOfferOutlinedIcon fontSize="small" />}
            iconPosition="start"
            label={t('nodeClassification.tabs.tags')}
          />
          <Tab
            value="smart-groups"
            id="classification-tab-smart-groups"
            aria-controls="classification-panel-smart-groups"
            icon={<FilterAltOutlinedIcon fontSize="small" />}
            iconPosition="start"
            label={t('nodeClassification.tabs.smartGroups')}
          />
        </Tabs>
        <Typography variant="body2" color="text.secondary" sx={{ pt: 1, pb: 2 }}>
          {t(tab === 'tags' ? 'nodeClassification.descriptions.tags' : 'nodeClassification.descriptions.smartGroups')}
        </Typography>
      </Paper>
      <Box role="tabpanel" id="classification-panel-tags" aria-labelledby="classification-tab-tags" hidden={tab !== 'tags'}>
        {tab === 'tags' && <TagManagement />}
      </Box>
      <Box
        role="tabpanel"
        id="classification-panel-smart-groups"
        aria-labelledby="classification-tab-smart-groups"
        hidden={tab !== 'smart-groups'}
      >
        {tab === 'smart-groups' && <SmartGroupsPage />}
      </Box>
    </Stack>
  );
}
