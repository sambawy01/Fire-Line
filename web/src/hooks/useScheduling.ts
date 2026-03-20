import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { schedulingApi } from '../lib/api';

export function useSchedule(locationId: string | null, weekStart: string) {
  return useQuery({
    queryKey: ['scheduling', 'schedule', locationId, weekStart],
    queryFn: () => schedulingApi.getSchedule(locationId!, weekStart),
    enabled: !!locationId && !!weekStart,
    staleTime: 30_000,
    retry: false,
  });
}

export function useForecast(locationId: string | null, date: string) {
  return useQuery({
    queryKey: ['scheduling', 'forecast', locationId, date],
    queryFn: () => schedulingApi.forecast(locationId!, date),
    enabled: !!locationId && !!date,
    staleTime: 60_000,
  });
}

export function useSwaps(locationId: string | null) {
  return useQuery({
    queryKey: ['scheduling', 'swaps', locationId],
    queryFn: () => schedulingApi.swaps(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useOvertimeRisk(locationId: string | null, weekStart: string) {
  return useQuery({
    queryKey: ['scheduling', 'overtime-risk', locationId, weekStart],
    queryFn: () => schedulingApi.overtimeRisk(locationId!, weekStart),
    enabled: !!locationId && !!weekStart,
    staleTime: 30_000,
  });
}

export function useLaborCost(scheduleId: string | null) {
  return useQuery({
    queryKey: ['scheduling', 'cost', scheduleId],
    queryFn: () => schedulingApi.cost(scheduleId!),
    enabled: !!scheduleId,
    staleTime: 30_000,
  });
}

export function useGenerateSchedule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ locationId, weekStart }: { locationId: string; weekStart: string }) =>
      schedulingApi.generate(locationId, weekStart),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scheduling', 'schedule'] });
    },
  });
}

export function usePublishSchedule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => schedulingApi.publish(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scheduling', 'schedule'] });
    },
  });
}

export function useReviewSwap() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved }: { id: string; approved: boolean }) =>
      schedulingApi.reviewSwap(id, approved),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scheduling', 'swaps'] });
    },
  });
}
