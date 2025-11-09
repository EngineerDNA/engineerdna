import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { QUERY_STALE_TIME_MS } from '../constants';

export function useDashboardTemplates(role?: string) {
  return useQuery({
    queryKey: ['dashboard-templates', role],
    queryFn: () => api.getDashboardTemplates(role),
    staleTime: QUERY_STALE_TIME_MS,
  });
}

export function useCreateDashboardFromTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ templateId, name }: { templateId: string; name?: string }) =>
      api.createDashboardFromTemplate(templateId, name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });
}
