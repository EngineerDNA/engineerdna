import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useLiveMetrics(refetchInterval = 30000) {
  return useQuery({
    queryKey: ['liveMetrics'],
    queryFn: api.getTodayMetrics,
    refetchInterval,
  });
}

export function useSprintBurndown(sprintId?: string, refetchInterval = 30000) {
  return useQuery({
    queryKey: ['sprintBurndown', sprintId],
    queryFn: () => api.getSprintBurndown(sprintId),
    refetchInterval,
    enabled: true,
  });
}

export function useTodayActivity(refetchInterval = 30000) {
  return useQuery({
    queryKey: ['todayActivity'],
    queryFn: api.getTodayActivity,
    refetchInterval,
  });
}
