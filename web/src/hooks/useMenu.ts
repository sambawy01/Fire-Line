import { useQuery } from '@tanstack/react-query';
import { menuApi } from '../lib/api';

export function useMenuItems(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['menu', 'items', locationId, from, to],
    queryFn: () => menuApi.getItems(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useMenuSummary(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['menu', 'summary', locationId, from, to],
    queryFn: () => menuApi.getSummary(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}
