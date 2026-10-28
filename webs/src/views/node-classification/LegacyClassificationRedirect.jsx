import PropTypes from 'prop-types';
import { Navigate, useLocation } from 'react-router-dom';
import { getClassificationLocation } from './navigation';

export default function LegacyClassificationRedirect({ tab }) {
  const location = useLocation();
  return <Navigate to={getClassificationLocation(location, tab)} state={location.state} replace />;
}

LegacyClassificationRedirect.propTypes = { tab: PropTypes.oneOf(['tags', 'smart-groups']).isRequired };
