import { useQuery } from '@tanstack/react-query';
import { payrollApi } from '../lib/api';

export function usePayrollSummary(locationId: string | null, periodStart: string, periodEnd: string) {
  return useQuery({
    queryKey: ['payroll', 'summary', locationId, periodStart, periodEnd],
    queryFn: () => payrollApi.getSummary(locationId!, periodStart, periodEnd),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePayrollHistory(locationId: string | null, months = 6) {
  return useQuery({
    queryKey: ['payroll', 'history', locationId, months],
    queryFn: () => payrollApi.getHistory(locationId!, months),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}
