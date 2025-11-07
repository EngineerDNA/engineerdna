import { useQuery } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';

export function useDashboard(id: string | undefined) {
  return useQuery({
    queryKey: ['dashboard', id],
    queryFn: () => dashboardsApi.getDashboard(id!),
    enabled: !!id,
    staleTime: 30000,
  });
}
