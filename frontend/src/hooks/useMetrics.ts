import { useQuery } from '@tanstack/react-query';
import type {
  EngineerScoresResponse,
  EngineerCostResponse,
  TeamCostResponse,
} from '../types/metrics';

export function useEngineerScores(engineerId: string, startDate: string, endDate: string) {
  return useQuery<EngineerScoresResponse>({
    queryKey: ['engineer-scores', engineerId, startDate, endDate],
    queryFn: async () => {
      const response = await fetch(
        `/api/performance/individual/${engineerId}?start_date=${startDate}&end_date=${endDate}`
      );
      if (!response.ok) {
        throw new Error('Failed to fetch engineer scores');
      }
      return response.json();
    },
    staleTime: 30000, // 30 seconds
  });
}

export function useTeamScores(teamId: string, startDate: string, endDate: string) {
  return useQuery<EngineerScoresResponse>({
    queryKey: ['team-scores', teamId, startDate, endDate],
    queryFn: async () => {
      const response = await fetch(
        `/api/performance/team/${teamId}?start_date=${startDate}&end_date=${endDate}`
      );
      if (!response.ok) {
        throw new Error('Failed to fetch team scores');
      }
      return response.json();
    },
    staleTime: 30000,
  });
}

export function useEngineerCost(engineerId: string) {
  return useQuery<EngineerCostResponse>({
    queryKey: ['engineer-cost', engineerId],
    queryFn: async () => {
      const response = await fetch(`/api/cost/engineer/${engineerId}`);
      if (!response.ok) {
        throw new Error('Failed to fetch engineer cost');
      }
      return response.json();
    },
    staleTime: 300000, // 5 minutes (costs don't change often)
  });
}

export function useTeamCost(teamId: string) {
  return useQuery<TeamCostResponse>({
    queryKey: ['team-cost', teamId],
    queryFn: async () => {
      const response = await fetch(`/api/cost/team/${teamId}`);
      if (!response.ok) {
        throw new Error('Failed to fetch team cost');
      }
      return response.json();
    },
    staleTime: 300000, // 5 minutes
  });
}
