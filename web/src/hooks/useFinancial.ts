import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financialApi } from '../lib/api';

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
