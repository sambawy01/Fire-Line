import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { customerApi, guestApi } from '../lib/api';

// Legacy hooks
export function useCustomers(locationId: string | null) {
  return useQuery({
    queryKey: ['customers', 'list', locationId],
    queryFn: () => customerApi.getCustomers(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useCustomerSummary(locationId: string | null) {
  return useQuery({
    queryKey: ['customers', 'summary', locationId],
    queryFn: () => customerApi.getSummary(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useAnalyzeCustomers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (locationId: string) => customerApi.analyze(locationId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

// SP15 Guest Profile hooks
export function useGuests(locationId: string | null, sortBy?: string) {
  return useQuery({
    queryKey: ['guests', 'list', locationId, sortBy],
    queryFn: () => guestApi.list(locationId!, sortBy),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useGuest(id: string | null) {
  return useQuery({
    queryKey: ['guests', 'detail', id],
    queryFn: () => guestApi.get(id!),
    enabled: !!id,
    staleTime: 60_000,
  });
}

export function useSegments() {
  return useQuery({
    queryKey: ['guests', 'analytics', 'segments'],
    queryFn: () => guestApi.segments(),
    staleTime: 60_000,
  });
}

export function useChurnDist() {
  return useQuery({
    queryKey: ['guests', 'analytics', 'churn'],
    queryFn: () => guestApi.churn(),
    staleTime: 60_000,
  });
}

export function useCLVDist() {
  return useQuery({
    queryKey: ['guests', 'analytics', 'clv'],
    queryFn: () => guestApi.clv(),
    staleTime: 60_000,
  });
}

export function useRefreshAnalytics() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => guestApi.refresh(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['guests'] });
    },
  });
}
