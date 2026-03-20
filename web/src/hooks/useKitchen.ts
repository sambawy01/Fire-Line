import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { kitchenApi } from '../lib/api';

export function useCapacity(locationId: string | null) {
  return useQuery({
    queryKey: ['kitchen', 'capacity', locationId],
    queryFn: () => kitchenApi.capacity(locationId!),
    enabled: !!locationId,
    staleTime: 10_000,
    refetchInterval: 10_000,
  });
}

export function useKDSTickets(locationId: string | null) {
  return useQuery({
    queryKey: ['kitchen', 'tickets', locationId],
    queryFn: () => kitchenApi.tickets(locationId!),
    enabled: !!locationId,
    staleTime: 5_000,
    refetchInterval: 5_000,
  });
}

export function useKDSMetrics(locationId: string | null) {
  return useQuery({
    queryKey: ['kitchen', 'metrics', locationId],
    queryFn: () => kitchenApi.metrics(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useBumpItem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ itemId, status }: { itemId: string; status: string }) =>
      kitchenApi.bumpItem(itemId, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['kitchen', 'tickets'] });
      queryClient.invalidateQueries({ queryKey: ['kitchen', 'capacity'] });
    },
  });
}
