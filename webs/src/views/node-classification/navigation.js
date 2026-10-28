export const NODE_CLASSIFICATION_PATH = '/subscription/classification';

export function getClassificationTab(search) {
  return new URLSearchParams(search).get('tab') === 'smart-groups' ? 'smart-groups' : 'tags';
}

export function getClassificationLocation(location, tab) {
  const params = new URLSearchParams(location.search);
  params.set('tab', tab === 'smart-groups' ? 'smart-groups' : 'tags');
  return { pathname: NODE_CLASSIFICATION_PATH, search: `?${params.toString()}`, hash: location.hash || '' };
}
