import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vendorScoringApi } from '../lib/api';

export function useVendorScores(locationId: string | null) {
  return useQuery({
    queryKey: ['vendor', 'scores', locationId],
    queryFn: () => vendorScoringApi.getScores(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePriceAnomalies(locationId: string | null) {
  return useQuery({
    queryKey: ['vendor', 'price-anomalies', locationId],
    queryFn: () => vendorScoringApi.priceAnomalies(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePriceTrend(
  ingredientId: string | null,
  vendorName: string | null,
  months = 6
) {
  return useQuery({
    queryKey: ['vendor', 'price-trend', ingredientId, vendorName, months],
    queryFn: () => vendorScoringApi.priceTrend(ingredientId!, vendorName!, months),
    enabled: !!ingredientId && !!vendorName,
    staleTime: 60_000,
  });
}

export function useVendorRecommendation(
  locationId: string | null,
  ingredientId: string | null
) {
  return useQuery({
    queryKey: ['vendor', 'recommend', locationId, ingredientId],
    queryFn: () => vendorScoringApi.recommend(locationId!, ingredientId!),
    enabled: !!locationId && !!ingredientId,
    staleTime: 60_000,
  });
}

export function useVendorCompare(
  locationId: string | null,
  ingredientId: string | null
) {
  return useQuery({
    queryKey: ['vendor', 'compare', locationId, ingredientId],
    queryFn: () => vendorScoringApi.compare(locationId!, ingredientId!),
    enabled: !!locationId && !!ingredientId,
    staleTime: 60_000,
  });
}

export function useCalculateScores() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (locationId: string) => vendorScoringApi.calculateScores(locationId),
    onSuccess: (_data, locationId) => {
      queryClient.invalidateQueries({ queryKey: ['vendor', 'scores', locationId] });
    },
  });
}
