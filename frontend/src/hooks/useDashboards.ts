import { useQuery } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';
import { QUERY_STALE_TIME_MS } from '../constants';

export function useDashboards() {
  return useQuery({
    queryKey: ['dashboards'],
    queryFn: () => dashboardsApi.getDashboards(),
    staleTime: QUERY_STALE_TIME_MS,
  });
}
