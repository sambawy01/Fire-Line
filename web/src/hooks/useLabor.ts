import { useQuery } from '@tanstack/react-query';
import { laborApi } from '../lib/api';

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
