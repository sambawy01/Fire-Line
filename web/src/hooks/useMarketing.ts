import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { marketingApi } from '../lib/api';

export function useCampaigns(locationId: string | null, status?: string) {
  return useQuery({
    queryKey: ['marketing', 'campaigns', locationId, status],
    queryFn: () => marketingApi.listCampaigns(locationId!, status),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useCampaignMetrics() {
  return useQuery({
    queryKey: ['marketing', 'analytics', 'campaigns'],
    queryFn: () => marketingApi.campaignMetrics(),
    staleTime: 60_000,
  });
}

export function useLoyaltyMembers(tier?: string) {
  return useQuery({
    queryKey: ['marketing', 'loyalty', 'members', tier],
    queryFn: () => marketingApi.loyaltyMembers(tier),
    staleTime: 60_000,
  });
}

export function useLoyaltyMetrics() {
  return useQuery({
    queryKey: ['marketing', 'analytics', 'loyalty'],
    queryFn: () => marketingApi.loyaltyMetrics(),
    staleTime: 60_000,
  });
}

export function useActivateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => marketingApi.activate(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['marketing', 'campaigns'] });
      qc.invalidateQueries({ queryKey: ['marketing', 'analytics'] });
    },
  });
}

export function usePauseCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => marketingApi.pause(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['marketing', 'campaigns'] });
      qc.invalidateQueries({ queryKey: ['marketing', 'analytics'] });
    },
  });
}

export function useCreateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => marketingApi.createCampaign(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['marketing', 'campaigns'] });
    },
  });
}
