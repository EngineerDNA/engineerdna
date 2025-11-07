import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: api.getSettings,
    staleTime: 60000, // Settings don't change often, cache for 1 minute
  });
}
