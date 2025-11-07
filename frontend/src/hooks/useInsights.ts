import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';

export function useInsights() {
  return useQuery({
    queryKey: ['insights'],
    queryFn: api.getInsights,
    staleTime: 5 * 60 * 1000, // 5 minutes - insights are generated periodically
  });
}

export function useGenerateInsights() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ plugin, period }: { plugin: string; period: { start: string; end: string } }) =>
      api.generateInsights(plugin, period),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['insights'] });
    },
  });
}
