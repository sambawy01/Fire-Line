import { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  RefreshControl,
} from 'react-native';
import { useAuthStore } from '../../stores/auth';
import { api } from '../../lib/api';

// ── types ─────────────────────────────────────────────────────────────────────

interface KDSTicketItem {
  ticket_item_id: string;
  item_name: string;
  quantity: number;
  station_type: string;
  status: 'pending' | 'cooking' | 'ready';
}

interface KDSTicket {
  ticket_id: string;
  order_number: string;
  channel: string;
  status: string;
  items: KDSTicketItem[];
  elapsed_secs: number;
  created_at: string;
}

// ── helpers ───────────────────────────────────────────────────────────────────

function fmtSecs(secs: number): string {
  const m = Math.floor(secs / 60);
  const s = secs % 60;
  return `${m}:${String(s).padStart(2, '0')}`;
}

function urgencyColor(secs: number): string {
  const mins = secs / 60;
  if (mins >= 10) return '#ef4444'; // red
  if (mins >= 5) return '#f59e0b';  // yellow
  return '#22c55e';                  // green
}

/** Pick station type from ELU ratings (highest score wins) */
function stationFromELU(eluRatings?: Record<string, number>): string {
  if (!eluRatings || Object.keys(eluRatings).length === 0) return 'grill';
  return Object.entries(eluRatings).sort((a, b) => b[1] - a[1])[0][0];
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
  catering: 'Catering',
  online: 'Online',
};

// ── elapsed timer ─────────────────────────────────────────────────────────────

function ElapsedTimer({ baseSecs }: { baseSecs: number }) {
  const [elapsed, setElapsed] = useState(baseSecs);

  useEffect(() => {
    setElapsed(baseSecs);
    const id = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(id);
  }, [baseSecs]);

  const color = urgencyColor(elapsed);

  return (
    <Text style={[styles.timerText, { color }]}>{fmtSecs(elapsed)}</Text>
  );
}

// ── ticket card ───────────────────────────────────────────────────────────────

function TicketCard({
  ticket,
  onBump,
}: {
  ticket: KDSTicket;
  onBump: (itemId: string, status: string) => Promise<void>;
}) {
  const borderColor = urgencyColor(ticket.elapsed_secs);

  return (
    <View style={[styles.card, { borderLeftColor: borderColor }]}>
      {/* header */}
      <View style={styles.cardHeader}>
        <View style={styles.orderRow}>
          <Text style={styles.orderNum}>#{ticket.order_number}</Text>
          <View style={styles.channelBadge}>
            <Text style={styles.channelText}>
              {CHANNEL_LABELS[ticket.channel] ?? ticket.channel}
            </Text>
          </View>
        </View>
        <ElapsedTimer baseSecs={ticket.elapsed_secs} />
      </View>

      {/* items */}
      <View style={styles.itemsList}>
        {ticket.items.map((item) => (
          <View key={item.ticket_item_id} style={styles.itemRow}>
            <View style={styles.itemInfo}>
              <Text style={styles.itemQty}>{item.quantity}×</Text>
              <Text style={styles.itemName} numberOfLines={2}>
                {item.item_name}
              </Text>
            </View>

            <View style={styles.itemActions}>
              {item.status === 'pending' && (
                <TouchableOpacity
                  style={[styles.actionBtn, styles.startBtn]}
                  onPress={() => onBump(item.ticket_item_id, 'cooking')}
                  activeOpacity={0.75}
                >
                  <Text style={styles.actionBtnText}>Start</Text>
                </TouchableOpacity>
              )}
              {item.status === 'cooking' && (
                <TouchableOpacity
                  style={[styles.actionBtn, styles.doneBtn]}
                  onPress={() => onBump(item.ticket_item_id, 'ready')}
                  activeOpacity={0.75}
                >
                  <Text style={styles.actionBtnText}>Done</Text>
                </TouchableOpacity>
              )}
              {item.status === 'ready' && (
                <View style={styles.readyBadge}>
                  <Text style={styles.readyText}>Ready</Text>
                </View>
              )}
            </View>
          </View>
        ))}
      </View>
    </View>
  );
}

// ── main screen ───────────────────────────────────────────────────────────────

export default function KdsScreen() {
  const { activeStaff, locationId } = useAuthStore();

  const [tickets, setTickets] = useState<KDSTicket[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stationType, setStationType] = useState<string>('grill');

  // Derive station from highest ELU rating
  useEffect(() => {
    // activeStaff doesn't have elu_ratings; we'd need to fetch the full profile.
    // For now fall back to a sensible default; if activeStaff has a role-based hint, use it.
    const role = activeStaff?.role?.toLowerCase() ?? '';
    let station = 'grill';
    if (role.includes('fry') || role.includes('fryer')) station = 'fry';
    else if (role.includes('salad') || role.includes('cold')) station = 'salad';
    else if (role.includes('dessert') || role.includes('pastry')) station = 'dessert';
    else if (role.includes('bev')) station = 'beverage';
    else if (role.includes('expo')) station = 'expo';
    setStationType(station);
  }, [activeStaff]);

  const fetchTickets = useCallback(async () => {
    if (!locationId) return;
    try {
      const data = await api.get<{ tickets: KDSTicket[] }>(
        `/operations/kds/station/${stationType}?location_id=${locationId}`
      );
      setTickets(data.tickets ?? []);
      setError(null);
    } catch (err: any) {
      setError(err.message ?? 'Failed to load tickets');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [locationId, stationType]);

  // Initial + auto-refresh every 5 seconds
  useEffect(() => {
    setLoading(true);
    fetchTickets();
    const id = setInterval(fetchTickets, 5_000);
    return () => clearInterval(id);
  }, [fetchTickets]);

  const handleRefresh = () => {
    setRefreshing(true);
    fetchTickets();
  };

  const handleBump = async (itemId: string, status: string) => {
    try {
      await api.put(`/operations/kds/items/${itemId}/bump`, { status });
      // Optimistically update local state
      setTickets((prev) =>
        prev.map((t) => ({
          ...t,
          items: t.items.map((i) =>
            i.ticket_item_id === itemId ? { ...i, status: status as any } : i
          ),
        }))
      );
    } catch {
      // Will sync on next auto-refresh
    }
  };

  // ── render ──────────────────────────────────────────────────────────────────

  if (loading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color="#e94560" />
        <Text style={styles.loadingText}>Loading tickets…</Text>
      </View>
    );
  }

  return (
    <View style={styles.root}>
      {/* station header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>
          {stationType.charAt(0).toUpperCase() + stationType.slice(1)} Station
        </Text>
        {activeStaff && (
          <Text style={styles.headerSub}>{activeStaff.display_name}</Text>
        )}
        <View style={styles.ticketCount}>
          <Text style={styles.ticketCountText}>{tickets.length} tickets</Text>
        </View>
      </View>

      {error && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      )}

      {tickets.length === 0 ? (
        <View style={styles.centered}>
          <Text style={styles.emptyIcon}>✅</Text>
          <Text style={styles.emptyTitle}>Kitchen Clear</Text>
          <Text style={styles.emptySub}>No active tickets at this station</Text>
        </View>
      ) : (
        <ScrollView
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={handleRefresh}
              tintColor="#e94560"
            />
          }
        >
          {tickets.map((ticket) => (
            <TicketCard key={ticket.ticket_id} ticket={ticket} onBump={handleBump} />
          ))}
        </ScrollView>
      )}
    </View>
  );
}

// ── styles ────────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#0f172a',
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  loadingText: {
    color: '#94a3b8',
    fontSize: 16,
    marginTop: 8,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 14,
    backgroundColor: '#1e293b',
    borderBottomWidth: 1,
    borderBottomColor: '#334155',
    gap: 10,
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ffffff',
    flex: 1,
  },
  headerSub: {
    fontSize: 14,
    color: '#94a3b8',
  },
  ticketCount: {
    backgroundColor: '#e94560',
    borderRadius: 12,
    paddingHorizontal: 10,
    paddingVertical: 4,
  },
  ticketCountText: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '700',
  },
  errorBanner: {
    backgroundColor: '#7f1d1d',
    paddingHorizontal: 16,
    paddingVertical: 10,
  },
  errorText: {
    color: '#fca5a5',
    fontSize: 14,
  },
  list: {
    padding: 16,
    gap: 16,
  },
  // ticket card
  card: {
    backgroundColor: '#1e293b',
    borderRadius: 16,
    borderLeftWidth: 5,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.3,
    shadowRadius: 4,
    elevation: 4,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  orderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  orderNum: {
    fontSize: 22,
    fontWeight: '800',
    color: '#ffffff',
  },
  channelBadge: {
    backgroundColor: '#334155',
    borderRadius: 8,
    paddingHorizontal: 8,
    paddingVertical: 3,
  },
  channelText: {
    fontSize: 12,
    color: '#94a3b8',
    fontWeight: '600',
  },
  timerText: {
    fontSize: 22,
    fontWeight: '800',
    fontVariant: ['tabular-nums'],
  },
  // items
  itemsList: {
    gap: 10,
  },
  itemRow: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#0f172a',
    borderRadius: 12,
    paddingHorizontal: 14,
    paddingVertical: 12,
    justifyContent: 'space-between',
  },
  itemInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    flex: 1,
  },
  itemQty: {
    fontSize: 18,
    fontWeight: '700',
    color: '#f97316',
    minWidth: 28,
  },
  itemName: {
    fontSize: 16,
    color: '#e2e8f0',
    fontWeight: '600',
    flex: 1,
  },
  itemActions: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginLeft: 8,
  },
  actionBtn: {
    borderRadius: 10,
    paddingHorizontal: 18,
    paddingVertical: 10,
    minWidth: 72,
    alignItems: 'center',
  },
  startBtn: {
    backgroundColor: '#f97316',
  },
  doneBtn: {
    backgroundColor: '#22c55e',
  },
  actionBtnText: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '700',
  },
  readyBadge: {
    backgroundColor: '#166534',
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 8,
  },
  readyText: {
    color: '#86efac',
    fontSize: 14,
    fontWeight: '700',
  },
  // empty state
  emptyIcon: {
    fontSize: 56,
    marginBottom: 8,
  },
  emptyTitle: {
    fontSize: 24,
    fontWeight: '700',
    color: '#ffffff',
  },
  emptySub: {
    fontSize: 16,
    color: '#64748b',
  },
});
