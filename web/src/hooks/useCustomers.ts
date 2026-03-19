import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { customerApi } from '../lib/api';

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
