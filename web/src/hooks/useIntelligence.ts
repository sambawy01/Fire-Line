import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { intelligenceApi } from '../lib/api';

export function useCEOBriefing() {
  return useQuery({
    queryKey: ['intelligence', 'ceo-briefing'],
    queryFn: () => intelligenceApi.getCEOBriefing(),
    staleTime: 30_000,
  });
}

export function useIntelligenceAnomalies() {
  return useQuery({
    queryKey: ['intelligence', 'anomalies'],
    queryFn: () => intelligenceApi.getAnomalies(),
    staleTime: 30_000,
  });
}

export function useResolveAnomaly() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; status: string; notes: string }) =>
      intelligenceApi.resolveAnomaly(data.id, data.status, data.notes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['intelligence'] });
    },
  });
}
