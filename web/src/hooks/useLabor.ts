import { useQuery } from '@tanstack/react-query';
import { laborApi } from '../lib/api';
import type { PointEvent } from '../lib/api';

export function useLaborSummary(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['labor', 'summary', locationId, from, to],
    queryFn: () => laborApi.getSummary(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useLaborEmployees(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['labor', 'employees', locationId, from, to],
    queryFn: () => laborApi.getEmployees(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useProfiles(locationId: string | null) {
  return useQuery({
    queryKey: ['labor', 'profiles', locationId],
    queryFn: () => laborApi.getProfiles(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useLeaderboard(locationId: string | null) {
  return useQuery({
    queryKey: ['labor', 'leaderboard', locationId],
    queryFn: () => laborApi.getLeaderboard(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePointHistory(employeeId: string | null) {
  return useQuery({
    queryKey: ['labor', 'points', employeeId],
    queryFn: () => laborApi.getPointHistory(employeeId!),
    enabled: !!employeeId,
    staleTime: 30_000,
  });
}

// Suppress unused import warning — PointEvent is re-exported for consumers
export type { PointEvent };
