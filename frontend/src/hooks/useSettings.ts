import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: api.getSettings,
    staleTime: 60000, // Settings don't change often, cache for 1 minute
  });
}

export function useRole() {
  return useQuery({
    queryKey: ['settings', 'role'],
    queryFn: api.getUserRole,
    staleTime: 60000,
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (role: string) => api.setUserRole(role),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['settings', 'role'] });
      await queryClient.invalidateQueries({ queryKey: ['settings'] });
      await queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });
}

export function usePrimaryDashboard() {
  return useQuery({
    queryKey: ['settings', 'primary-dashboard'],
    queryFn: api.getPrimaryDashboard,
    staleTime: 60000,
  });
}

export function useUpdatePrimaryDashboard() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (dashboardId: string) => api.setPrimaryDashboard(dashboardId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['settings', 'primary-dashboard'] });
      await queryClient.invalidateQueries({ queryKey: ['settings'] });
    },
  });
}
