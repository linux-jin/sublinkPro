import Box from '@mui/material/Box';

import { ReleaseLogPanel } from 'components/SystemUpdatePanels';

export default function SystemUpdatesPage() {
  return (
    <Box sx={{ pb: 3 }}>
      <ReleaseLogPanel />
    </Box>
  );
}
