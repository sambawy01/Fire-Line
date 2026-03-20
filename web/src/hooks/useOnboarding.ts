import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { onboardingApi } from '../lib/api';

// ─── Queries ─────────────────────────────────────────────────────────────────

export function useOnboardingSession() {
  return useQuery({
    queryKey: ['onboarding', 'session'],
    queryFn: () => onboardingApi.getSession(),
    retry: false,
    staleTime: 0,
  });
}

export function useOnboardingInsights(locationId: string | null) {
  return useQuery({
    queryKey: ['onboarding', 'insights', locationId],
    queryFn: () => onboardingApi.getInsights(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useInferConcept(locationId: string | null) {
  return useQuery({
    queryKey: ['onboarding', 'concept', locationId],
    queryFn: () => onboardingApi.inferConcept(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}

export function useRecommendModules(priorities: string[]) {
  return useQuery({
    queryKey: ['onboarding', 'modules', priorities],
    queryFn: () => onboardingApi.recommendModules(priorities),
    enabled: priorities.length > 0,
    staleTime: 60_000,
  });
}

export function useOnboardingChecklist() {
  return useQuery({
    queryKey: ['onboarding', 'checklist'],
    queryFn: () => onboardingApi.getChecklist(),
    staleTime: 0,
  });
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export function useStartOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) => onboardingApi.startOnboarding(userId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useUpdateStep() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      sessionId,
      step,
      data,
    }: {
      sessionId: string;
      step: string;
      data: Record<string, unknown>;
    }) => onboardingApi.updateStep(sessionId, step, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding', 'session'] }),
  });
}

export function useCompleteChecklistItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (itemId: string) => onboardingApi.completeChecklistItem(itemId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding', 'checklist'] }),
  });
}
