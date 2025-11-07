import { useQuery } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';
import type { MetricCompareRequest } from '../api/types';

export function useMetricCompare(params: MetricCompareRequest | undefined) {
  return useQuery({
    queryKey: ['metric-compare', params],
    queryFn: () => dashboardsApi.getMetricCompare(params!),
    enabled: !!params && params.entity_ids.length > 0,
    staleTime: 30000,
  });
}
