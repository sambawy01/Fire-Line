import { useEffect, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { useWasteStore, WasteLog, LogWasteInput } from '../../stores/waste';
import { useAuthStore } from '../../stores/auth';
import WasteForm, { REASONS } from '../../components/WasteForm';

// ────────────────────────────────────────────────────────────────────────────
// Reason badge colour map
// ────────────────────────────────────────────────────────────────────────────

const reasonColorMap: Record<string, string> = {
  expired: '#ef4444',
  dropped: '#f97316',
  overcooked: '#f59e0b',
  contaminated: '#dc2626',
  overproduction: '#3b82f6',
  other: '#6b7280',
};

function reasonLabel(reason: string): string {
  return REASONS.find((r) => r.key === reason)?.label ?? reason;
}

// ────────────────────────────────────────────────────────────────────────────
// Waste log row
// ────────────────────────────────────────────────────────────────────────────

function WasteLogRow({ item }: { item: WasteLog }) {
  const color = reasonColorMap[item.reason] ?? '#6b7280';

  const timeStr = (() => {
    try {
      const d = new Date(item.logged_at);
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return item.logged_at;
    }
  })();

  return (
    <View style={rowStyles.container}>
      <View style={rowStyles.main}>
        <Text style={rowStyles.name} numberOfLines={1}>
          {item.ingredient_name}
        </Text>
        <Text style={rowStyles.qty}>
          {item.quantity} {item.unit}
        </Text>
      </View>
      <View style={rowStyles.meta}>
        <View style={[rowStyles.badge, { backgroundColor: color + '22', borderColor: color }]}>
          <Text style={[rowStyles.badgeText, { color }]}>{reasonLabel(item.reason)}</Text>
        </View>
        <Text style={rowStyles.loggedBy}>{item.logged_by_name}</Text>
        <Text style={rowStyles.time}>{timeStr}</Text>
      </View>
      {item.note ? <Text style={rowStyles.note}>{item.note}</Text> : null}
    </View>
  );
}

const rowStyles = StyleSheet.create({
  container: {
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#1a1a2e',
    borderBottomWidth: 1,
    borderBottomColor: '#1e1e3a',
    minHeight: 56,
    gap: 6,
  },
  main: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  name: {
    fontSize: 16,
    fontWeight: '600',
    color: '#e0e0e0',
    flex: 1,
    marginRight: 12,
  },
  qty: {
    fontSize: 16,
    fontWeight: '700',
    color: '#ffffff',
  },
  meta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  badge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
    borderWidth: 1,
  },
  badgeText: {
    fontSize: 12,
    fontWeight: '700',
    textTransform: 'uppercase',
    letterSpacing: 0.3,
  },
  loggedBy: {
    fontSize: 13,
    color: '#888',
  },
  time: {
    fontSize: 13,
    color: '#666',
    marginLeft: 'auto',
  },
  note: {
    fontSize: 13,
    color: '#888',
    fontStyle: 'italic',
  },
});

// ────────────────────────────────────────────────────────────────────────────
// Main Waste Screen
// ────────────────────────────────────────────────────────────────────────────

export default function WasteScreen() {
  const { todaysLogs, ingredients, loading, error, loadLogs, loadIngredients, logWaste } =
    useWasteStore();
  const { touchActivity } = useAuthStore();

  useEffect(() => {
    loadIngredients();
    loadLogs();
  }, []);

  const handleRefresh = useCallback(async () => {
    touchActivity();
    await loadLogs();
  }, [loadLogs, touchActivity]);

  const handleLogWaste = useCallback(
    async (input: LogWasteInput) => {
      touchActivity();
      await logWaste(input);
    },
    [logWaste, touchActivity],
  );

  const renderItem = useCallback(({ item }: { item: WasteLog }) => {
    return <WasteLogRow item={item} />;
  }, []);

  const ListHeader = (
    <View>
      <WasteForm ingredients={ingredients} onSubmit={handleLogWaste} />

      {/* Today's waste header */}
      <View style={styles.feedHeader}>
        <Text style={styles.feedTitle}>Today's Waste</Text>
        {loading && <ActivityIndicator size="small" color="#e94560" />}
        <Text style={styles.feedCount}>{todaysLogs.length} entries</Text>
      </View>

      {error ? (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      ) : null}

      {!loading && todaysLogs.length === 0 && (
        <View style={styles.emptyState}>
          <Text style={styles.emptyIcon}>✅</Text>
          <Text style={styles.emptyText}>No waste logged today</Text>
        </View>
      )}
    </View>
  );

  return (
    <View style={styles.container}>
      <FlatList
        data={todaysLogs}
        keyExtractor={(item) => item.waste_id}
        renderItem={renderItem}
        ListHeaderComponent={ListHeader}
        style={styles.flex}
        keyboardShouldPersistTaps="handled"
        refreshControl={
          <RefreshControl
            refreshing={loading}
            onRefresh={handleRefresh}
            tintColor="#e94560"
            colors={['#e94560']}
          />
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  flex: {
    flex: 1,
  },
  feedHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#16213e',
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
    gap: 10,
  },
  feedTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#ffffff',
    flex: 1,
  },
  feedCount: {
    fontSize: 14,
    color: '#888',
  },
  errorBanner: {
    backgroundColor: '#3d0c11',
    borderWidth: 1,
    borderColor: '#e94560',
    margin: 12,
    borderRadius: 8,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  errorText: {
    color: '#e94560',
    fontSize: 14,
  },
  emptyState: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 48,
    gap: 10,
  },
  emptyIcon: {
    fontSize: 32,
  },
  emptyText: {
    fontSize: 16,
    color: '#888',
  },
});
