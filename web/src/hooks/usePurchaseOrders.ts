import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { poApi } from '../lib/api';

export function usePOs(locationId: string | null, status?: string) {
  return useQuery({
    queryKey: ['po', 'list', locationId, status],
    queryFn: () => poApi.list(locationId!, status),
    enabled: !!locationId,
    staleTime: 15_000,
  });
}

export function usePO(poId: string | null) {
  return useQuery({
    queryKey: ['po', 'detail', poId],
    queryFn: () => poApi.get(poId!),
    enabled: !!poId,
  });
}

export function usePARBreaches(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'par-breaches', locationId],
    queryFn: () => poApi.parBreaches(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useApprovePO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (poId: string) => poApi.approve(poId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['po'] }),
  });
}

export function useCancelPO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (poId: string) => poApi.cancel(poId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['po'] }),
  });
}
