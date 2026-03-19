import { useQuery } from '@tanstack/react-query';
import { vendorApi } from '../lib/api';

export function useVendors(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['vendor', 'list', locationId, from, to],
    queryFn: () => vendorApi.getVendors(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useVendorSummary(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['vendor', 'summary', locationId, from, to],
    queryFn: () => vendorApi.getSummary(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}
