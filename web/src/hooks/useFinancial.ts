import { useQuery } from '@tanstack/react-query';
import { financialApi } from '../lib/api';

export function usePnL(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'pnl', locationId],
    queryFn: () => financialApi.getPnL(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useAnomalies(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'anomalies', locationId],
    queryFn: () => financialApi.getAnomalies(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}
