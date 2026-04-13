import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adapterApi } from '../lib/api';

export function useLoyverseStatus() {
  return useQuery({
    queryKey: ['adapters', 'loyverse', 'status'],
    queryFn: () => adapterApi.getLoyverseStatus(),
    staleTime: 15_000,
    refetchInterval: 30_000,
  });
}

export function useLoyverseSync() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => adapterApi.triggerLoyverseSync(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['adapters', 'loyverse', 'status'] });
    },
  });
}

export function useLoyverseImport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (days?: number) => adapterApi.triggerLoyverseImport(days),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['adapters', 'loyverse', 'status'] });
    },
  });
}

export function useLoyverseConnect() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { api_token: string; store_id: string; org_id: string; location_id: string }) =>
      adapterApi.connectLoyverse(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['adapters', 'loyverse', 'status'] });
    },
  });
}
