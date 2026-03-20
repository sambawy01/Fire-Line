import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { portfolioApi } from '../lib/api';

// ── Hierarchy ────────────────────────────────────────────────────────────────

export function useHierarchy() {
  return useQuery({
    queryKey: ['portfolio', 'hierarchy'],
    queryFn: () => portfolioApi.getHierarchy(),
    staleTime: 60_000,
  });
}

export function useCreateNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { parent_node_id?: string | null; name: string; node_type: string; location_id?: string | null }) =>
      portfolioApi.createNode(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio', 'hierarchy'] }),
  });
}

export function useUpdateNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ nodeId, name }: { nodeId: string; name: string }) =>
      portfolioApi.updateNode(nodeId, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio', 'hierarchy'] }),
  });
}

export function useDeleteNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (nodeId: string) => portfolioApi.deleteNode(nodeId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio', 'hierarchy'] }),
  });
}

// ── Aggregation ──────────────────────────────────────────────────────────────

export function useAggregateKPIs(nodeId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['portfolio', 'kpis', nodeId, from, to],
    queryFn: () => portfolioApi.aggregateKPIs(nodeId!, from, to),
    enabled: !!nodeId,
    staleTime: 30_000,
  });
}

export function useComparison(locationIds: string[], from?: string, to?: string) {
  return useQuery({
    queryKey: ['portfolio', 'compare', locationIds.join(','), from, to],
    queryFn: () => portfolioApi.compareLocations(locationIds, from, to),
    enabled: locationIds.length > 0,
    staleTime: 30_000,
  });
}

// ── Benchmarking ─────────────────────────────────────────────────────────────

export function useCalculateBenchmarks() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ from, to }: { from?: string; to?: string }) =>
      portfolioApi.calculateBenchmarks(from, to),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['portfolio', 'benchmarks'] });
      qc.invalidateQueries({ queryKey: ['portfolio', 'outliers'] });
    },
  });
}

export function useBenchmarks(from?: string, to?: string) {
  return useQuery({
    queryKey: ['portfolio', 'benchmarks', from, to],
    queryFn: () => portfolioApi.getBenchmarks(from, to),
    staleTime: 60_000,
  });
}

export function useOutliers(from?: string, to?: string) {
  return useQuery({
    queryKey: ['portfolio', 'outliers', from, to],
    queryFn: () => portfolioApi.getOutliers(from, to),
    staleTime: 60_000,
  });
}

// ── Best Practices ───────────────────────────────────────────────────────────

export function useDetectBestPractices() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => portfolioApi.detectBestPractices(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio', 'best-practices'] }),
  });
}

export function useBestPractices(status?: string) {
  return useQuery({
    queryKey: ['portfolio', 'best-practices', status],
    queryFn: () => portfolioApi.listBestPractices(status),
    staleTime: 60_000,
  });
}

export function useAdoptPractice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => portfolioApi.adoptPractice(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio', 'best-practices'] }),
  });
}

export function useDismissPractice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => portfolioApi.dismissPractice(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio', 'best-practices'] }),
  });
}
