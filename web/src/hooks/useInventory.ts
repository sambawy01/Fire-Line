import { useQuery } from '@tanstack/react-query';
import { inventoryApi } from '../lib/api';

export function useUsage(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'usage', locationId],
    queryFn: () => inventoryApi.getUsage(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePARStatus(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'par', locationId],
    queryFn: () => inventoryApi.getPARStatus(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useVariances(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'variances', locationId],
    queryFn: () => inventoryApi.getVariances(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useWasteLogs(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'waste', locationId],
    queryFn: () => inventoryApi.getWasteLogs(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}
