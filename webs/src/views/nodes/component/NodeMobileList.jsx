import PropTypes from 'prop-types';
import { memo } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';

// components
import NodeCard from './NodeCard';
import { getNodeThemeTokens } from '../nodeTheme';

/**
 * 移动端节点卡片列表
 */
function NodeMobileList({ nodes, page, rowsPerPage, selectedNodeIds, tagColorMap, protocolMeta, onSelect, onViewDetails }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const isSelected = (node) => selectedNodeIds.has(node.ID);
  // 后端分页：nodes 已经是当前页数据，无需客户端切片

  return (
    <Stack spacing={2}>
      {nodes.length === 0 && (
        <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 3, color: tokens.secondaryText }}>
          {t('nodes.mobile.empty')}
        </Typography>
      )}
      {nodes.map((node) => (
        <NodeCard
          key={node.ID}
          node={node}
          isSelected={isSelected(node)}
          tagColorMap={tagColorMap}
          protocolMeta={protocolMeta}
          onSelect={onSelect}
          onViewDetails={onViewDetails}
        />
      ))}
    </Stack>
  );
}

NodeMobileList.propTypes = {
  nodes: PropTypes.array.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
  selectedNodeIds: PropTypes.instanceOf(Set).isRequired,
  tagColorMap: PropTypes.object,
  protocolMeta: PropTypes.array,
  onSelect: PropTypes.func.isRequired,
  onViewDetails: PropTypes.func.isRequired
};

export default memo(NodeMobileList);
