import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useEvents(params?: Parameters<typeof api.getEvents>[0]) {
  return useQuery({
    queryKey: ['events', params],
    queryFn: () => api.getEvents(params),
  });
}
