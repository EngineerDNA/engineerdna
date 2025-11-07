import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useSystemInfo() {
  return useQuery({
    queryKey: ['systemInfo'],
    queryFn: api.getSystemInfo,
  });
}
