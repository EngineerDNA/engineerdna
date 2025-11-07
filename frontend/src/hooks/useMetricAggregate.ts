import { useQuery } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';
import type { MetricAggregateRequest } from '../api/types';

export function useMetricAggregate(params: MetricAggregateRequest | undefined) {
  return useQuery({
    queryKey: ['metric-aggregate', params],
    queryFn: () => dashboardsApi.getMetricAggregate(params!),
    enabled: !!params,
    staleTime: 30000,
  });
}
