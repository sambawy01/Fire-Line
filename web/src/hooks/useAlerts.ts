import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { alertsApi } from '../lib/api';

export function useAlertQueue(locationId: string | null, opts?: { limit?: number }) {
  return useQuery({
    queryKey: ['alerts', 'queue', locationId],
    queryFn: async () => {
      const { alerts } = await alertsApi.getQueue(locationId ?? undefined);
      return opts?.limit ? alerts.slice(0, opts.limit) : alerts;
    },
    enabled: !!locationId,
    staleTime: 10_000,
  });
}

export function useAlertCount(locationId: string | null) {
  return useQuery({
    queryKey: ['alerts', 'count', locationId],
    queryFn: () => alertsApi.getCount(),
    enabled: !!locationId,
    staleTime: 10_000,
  });
}

export function useAcknowledgeAlert() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (alertId: string) => alertsApi.acknowledge(alertId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alerts'] });
    },
  });
}

export function useResolveAlert() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (alertId: string) => alertsApi.resolve(alertId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alerts'] });
    },
  });
}
