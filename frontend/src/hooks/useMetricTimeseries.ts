import { useQuery } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';
import type { MetricTimeseriesRequest } from '../api/types';

export function useMetricTimeseries(params: MetricTimeseriesRequest | undefined) {
  return useQuery({
    queryKey: ['metric-timeseries', params],
    queryFn: () => dashboardsApi.getMetricTimeseries(params!),
    enabled: !!params,
    staleTime: 30000,
  });
}
