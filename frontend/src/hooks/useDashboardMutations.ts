import { useMutation, useQueryClient } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';
import type { CreateDashboardRequest, UpdateDashboardRequest } from '../api/types';

export function useDashboardMutations() {
  const queryClient = useQueryClient();

  const createDashboard = useMutation({
    mutationFn: (data: CreateDashboardRequest) => dashboardsApi.createDashboard(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });

  const updateDashboard = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateDashboardRequest }) =>
      dashboardsApi.updateDashboard(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard', variables.id] });
    },
  });

  const deleteDashboard = useMutation({
    mutationFn: (id: string) => dashboardsApi.deleteDashboard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });

  const cloneDashboard = useMutation({
    mutationFn: ({ id, name }: { id: string; name?: string }) =>
      dashboardsApi.cloneDashboard(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });

  return {
    createDashboard,
    updateDashboard,
    deleteDashboard,
    cloneDashboard,
  };
}
