import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { menuScoringApi } from '../lib/api';

export function useMenuScores(locationId: string | null) {
  return useQuery({
    queryKey: ['menu', 'scores', locationId],
    queryFn: () => menuScoringApi.getScores(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useDependencies(locationId: string | null) {
  return useQuery({
    queryKey: ['menu', 'dependencies', locationId],
    queryFn: () => menuScoringApi.getDependencies(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useCrossSell(locationId: string | null, limit = 10) {
  return useQuery({
    queryKey: ['menu', 'cross-sell', locationId, limit],
    queryFn: () => menuScoringApi.getCrossSell(locationId!, limit),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useScoreMenu() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (locationId: string) => menuScoringApi.triggerScore(locationId),
    onSuccess: (_data, locationId) => {
      queryClient.invalidateQueries({ queryKey: ['menu', 'scores', locationId] });
    },
  });
}

export function useSimulatePrice() {
  return useMutation({
    mutationFn: ({
      locationId,
      menuItemId,
      newPrice,
    }: {
      locationId: string;
      menuItemId: string;
      newPrice: number;
    }) => menuScoringApi.simulatePrice(locationId, menuItemId, newPrice),
  });
}

export function useSimulateRemoval() {
  return useMutation({
    mutationFn: ({
      locationId,
      menuItemId,
    }: {
      locationId: string;
      menuItemId: string;
    }) => menuScoringApi.simulateRemoval(locationId, menuItemId),
  });
}

export function useSimulateIngredientCost() {
  return useMutation({
    mutationFn: ({
      locationId,
      ingredientId,
      newCostPerUnit,
    }: {
      locationId: string;
      ingredientId: string;
      newCostPerUnit: number;
    }) => menuScoringApi.simulateIngredientCost(locationId, ingredientId, newCostPerUnit),
  });
}
