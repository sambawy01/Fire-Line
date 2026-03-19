import { useQuery } from '@tanstack/react-query';
import { reportsApi } from '../lib/api';

export function useDailyReport(locationId: string | null) {
  return useQuery({
    queryKey: ['reports', 'daily', locationId],
    queryFn: () => reportsApi.getDaily(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}
