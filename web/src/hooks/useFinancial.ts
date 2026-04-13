import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financialApi } from '../lib/api';
import type { PnL } from '../lib/api';

/**
 * Fetch the last 7 days of daily PnL for sparkline data.
 * Returns an array of { day: number, value: number } where value is net_revenue in piasters.
 */
export function useWeeklyRevenue(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'weekly-revenue', locationId],
    queryFn: async () => {
      const days: { day: number; value: number }[] = [];
      const now = new Date();
      const promises: Promise<PnL>[] = [];

      for (let i = 6; i >= 0; i--) {
        const date = new Date(now);
        date.setDate(date.getDate() - i);
        const dateStr = date.toISOString().split('T')[0];
        promises.push(financialApi.getPnL(locationId!, dateStr, dateStr));
      }

      const results = await Promise.allSettled(promises);
      results.forEach((result, idx) => {
        days.push({
          day: idx,
          value: result.status === 'fulfilled' ? Math.round((result.value.net_revenue ?? 0) / 100) : 0,
        });
      });

      return days;
    },
    enabled: !!locationId,
    staleTime: 5 * 60_000, // 5 minutes — historical data doesn't change often
  });
}

export function usePnL(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'pnl', locationId],
    queryFn: () => financialApi.getPnL(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useAnomalies(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'anomalies', locationId],
    queryFn: () => financialApi.getAnomalies(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useBudgetVariance(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'budget-variance', locationId],
    queryFn: () => financialApi.budgetVariance(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useCostCenters(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'cost-centers', locationId],
    queryFn: () => financialApi.costCenters(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useTxAnomalies(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'tx-anomalies', locationId],
    queryFn: () => financialApi.txAnomalies(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePeriodComparison(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'period-comparison', locationId],
    queryFn: () => financialApi.periodComparison(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useListBudgets(locationId: string | null, periodType?: string) {
  return useQuery({
    queryKey: ['financial', 'budgets', locationId, periodType],
    queryFn: () => financialApi.listBudgets(locationId!, periodType),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useCreateBudget() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => financialApi.createBudget(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['financial', 'budgets'] });
      queryClient.invalidateQueries({ queryKey: ['financial', 'budget-variance'] });
    },
  });
}
