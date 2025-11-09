import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';

export function usePlugins() {
  return useQuery({
    queryKey: ['plugins'],
    queryFn: api.getPlugins,
    staleTime: 5 * 60 * 1000, // 5 minutes - plugins change less frequently
  });
}

export function useConfigurePlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ name, config }: { name: string; config: Record<string, string> }) =>
      api.configurePlugin(name, config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plugins'] });
    },
  });
}

export function useSyncPlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) => api.syncPlugin(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
      queryClient.invalidateQueries({ queryKey: ['plugins'] });
    },
  });
}
