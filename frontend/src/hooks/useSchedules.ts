import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { CreateScheduleRequest, UpdateScheduleRequest } from '../api/types';

export function useSchedules() {
  return useQuery({
    queryKey: ['schedules'],
    queryFn: api.getSchedules,
  });
}

export function useSchedule(id: string) {
  return useQuery({
    queryKey: ['schedules', id],
    queryFn: () => api.getSchedule(id),
    enabled: !!id,
  });
}

export function useCreateSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: CreateScheduleRequest) => api.createSchedule(req),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['schedules'] });
    },
  });
}

export function useUpdateSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateScheduleRequest }) =>
      api.updateSchedule(id, req),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['schedules'] });
    },
  });
}

export function useDeleteSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.deleteSchedule(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['schedules'] });
    },
  });
}

export function useTriggerSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.triggerSchedule(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['schedules'] });
      await queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });
}
