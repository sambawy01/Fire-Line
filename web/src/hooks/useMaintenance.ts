import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { maintenanceApi } from '../lib/api';

export function useEquipment(locationId: string | null, status?: string, category?: string) {
  return useQuery({
    queryKey: ['maintenance', 'equipment', locationId, status, category],
    queryFn: () => maintenanceApi.listEquipment(locationId ?? undefined, status, category),
    enabled: !!locationId,
  });
}

export function useEquipmentDetail(equipmentId: string | null) {
  return useQuery({
    queryKey: ['maintenance', 'equipment', 'detail', equipmentId],
    queryFn: () => maintenanceApi.getEquipment(equipmentId!),
    enabled: !!equipmentId,
  });
}

export function useMaintenanceTickets(locationId: string | null, status?: string, priority?: string) {
  return useQuery({
    queryKey: ['maintenance', 'tickets', locationId, status, priority],
    queryFn: () => maintenanceApi.listTickets(locationId ?? undefined, status, priority),
    enabled: !!locationId,
  });
}

export function useTicketDetail(ticketId: string | null) {
  return useQuery({
    queryKey: ['maintenance', 'tickets', 'detail', ticketId],
    queryFn: () => maintenanceApi.getTicket(ticketId!),
    enabled: !!ticketId,
  });
}

export function useOverdueEquipment(locationId: string | null) {
  return useQuery({
    queryKey: ['maintenance', 'overdue', locationId],
    queryFn: () => maintenanceApi.getOverdue(locationId ?? undefined),
    enabled: !!locationId,
  });
}

export function useMaintenanceStats(locationId: string | null) {
  return useQuery({
    queryKey: ['maintenance', 'stats', locationId],
    queryFn: () => maintenanceApi.getStats(locationId ?? undefined),
    enabled: !!locationId,
  });
}

export function useCreateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => maintenanceApi.createTicket(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useUpdateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => maintenanceApi.updateTicket(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useCompleteTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, resolution, actualCost }: { id: string; resolution: string; actualCost: number }) =>
      maintenanceApi.completeTicket(id, resolution, actualCost),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useAddMaintenanceLog() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ ticketId, data }: { ticketId: string; data: { action: string; notes?: string; cost?: number; performed_by?: string } }) =>
      maintenanceApi.addLog(ticketId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}
