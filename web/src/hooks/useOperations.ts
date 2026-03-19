import { useQuery } from '@tanstack/react-query';
import { operationsApi } from '../lib/api';

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
