import { useQuery } from '@tanstack/react-query';
import { operationsApi, opsCommandApi } from '../lib/api';

export function useOperationsSummary(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['operations', 'summary', locationId, from, to],
    queryFn: () => operationsApi.getSummary(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 15_000,
    refetchInterval: 15_000,
  });
}

export function useOperationsHourly(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['operations', 'hourly', locationId, from, to],
    queryFn: () => operationsApi.getHourly(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

// SP18 — Command Center hooks

export function useHealth(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'health', locationId],
    queryFn: () => opsCommandApi.getHealth(locationId!),
    enabled: !!locationId,
    staleTime: 10_000,
    refetchInterval: 10_000,
  });
}

export function useOverload(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'overload', locationId],
    queryFn: () => opsCommandApi.getOverload(locationId!),
    enabled: !!locationId,
    staleTime: 10_000,
    refetchInterval: 10_000,
  });
}

export function usePriorities(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'priorities', locationId],
    queryFn: () => opsCommandApi.getPriorities(locationId!),
    enabled: !!locationId,
    staleTime: 10_000,
    refetchInterval: 10_000,
  });
}

export function useRealtimeHorizon(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'horizon', 'realtime', locationId],
    queryFn: () => opsCommandApi.getRealtimeHorizon(locationId!),
    enabled: !!locationId,
    staleTime: 10_000,
    refetchInterval: 10_000,
  });
}

export function useShiftHorizon(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'horizon', 'shift', locationId],
    queryFn: () => opsCommandApi.getShiftHorizon(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useDailyHorizon(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'horizon', 'daily', locationId],
    queryFn: () => opsCommandApi.getDailyHorizon(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useWeeklyHorizon(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'horizon', 'weekly', locationId],
    queryFn: () => opsCommandApi.getWeeklyHorizon(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useStrategicHorizon(locationId: string | null) {
  return useQuery({
    queryKey: ['operations', 'horizon', 'strategic', locationId],
    queryFn: () => opsCommandApi.getStrategicHorizon(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}
