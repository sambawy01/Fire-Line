import { useEffect, useState, useMemo, useCallback } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  TextInput,
  StyleSheet,
  Alert,
  ActivityIndicator,
  FlatList,
} from 'react-native';
import { useCountStore, CountLine } from '../../stores/count';
import { useAuthStore } from '../../stores/auth';
import ProgressBar from '../../components/ProgressBar';
import CategoryGroup from '../../components/CategoryGroup';
import CountRow from '../../components/CountRow';

type Screen = 'start' | 'counting' | 'review';

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

function groupByCategory(lines: CountLine[]): Record<string, CountLine[]> {
  return lines.reduce<Record<string, CountLine[]>>((acc, line) => {
    const cat = line.category || 'Uncategorized';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(line);
    return acc;
  }, {});
}

function varianceColor(expected: number, counted: number | null): string | undefined {
  if (counted === null || expected === 0) return undefined;
  const pct = Math.abs((counted - expected) / expected);
  if (pct > 0.15) return '#ef4444'; // red > 15%
  if (pct > 0.10) return '#f59e0b'; // amber > 10%
  return undefined;
}

// ────────────────────────────────────────────────────────────────────────────
// Start Screen
// ────────────────────────────────────────────────────────────────────────────

function StartScreen({ onStart }: { onStart: (type: 'full' | 'spot') => void }) {
  return (
    <View style={styles.startContainer}>
      <Text style={styles.startTitle}>Inventory Count</Text>
      <Text style={styles.startSubtitle}>Select the type of count to begin</Text>

      <View style={styles.startButtonRow}>
        <TouchableOpacity
          style={[styles.startButton, styles.startButtonFull]}
          onPress={() => onStart('full')}
          activeOpacity={0.8}
        >
          <Text style={styles.startButtonIcon}>📋</Text>
          <Text style={styles.startButtonLabel}>Full Count</Text>
          <Text style={styles.startButtonDesc}>Count all inventory items</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.startButton, styles.startButtonSpot]}
          onPress={() => onStart('spot')}
          activeOpacity={0.8}
        >
          <Text style={styles.startButtonIcon}>🔍</Text>
          <Text style={styles.startButtonLabel}>Spot Check</Text>
          <Text style={styles.startButtonDesc}>Count a specific category</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Counting Screen
// ────────────────────────────────────────────────────────────────────────────

function CountingScreen({
  onReview,
  onCancel,
}: {
  onReview: () => void;
  onCancel: () => void;
}) {
  const { lines, progress, pendingSync, updateLine, syncLines, syncing } = useCountStore();
  const [search, setSearch] = useState('');

  const filteredLines = useMemo(() => {
    if (!search.trim()) return lines;
    const q = search.toLowerCase();
    return lines.filter(
      (l) =>
        l.name.toLowerCase().includes(q) ||
        l.category.toLowerCase().includes(q),
    );
  }, [lines, search]);

  const grouped = useMemo(() => groupByCategory(filteredLines), [filteredLines]);
  const categories = useMemo(() => Object.keys(grouped).sort(), [grouped]);

  const handleSync = useCallback(async () => {
    await syncLines();
  }, [syncLines]);

  return (
    <View style={styles.flex}>
      <ProgressBar counted={progress.counted} total={progress.total} />

      {/* Search bar + sync indicator */}
      <View style={styles.searchRow}>
        <TextInput
          style={styles.searchInput}
          value={search}
          onChangeText={setSearch}
          placeholder="Search ingredients…"
          placeholderTextColor="#555"
          returnKeyType="search"
          clearButtonMode="while-editing"
        />
        {pendingSync.length > 0 && (
          <TouchableOpacity style={styles.syncButton} onPress={handleSync} disabled={syncing}>
            {syncing ? (
              <ActivityIndicator size="small" color="#22c55e" />
            ) : (
              <Text style={styles.syncText}>Sync {pendingSync.length}</Text>
            )}
          </TouchableOpacity>
        )}
      </View>

      <ScrollView style={styles.flex} keyboardShouldPersistTaps="handled">
        {categories.map((cat) => {
          const catLines = grouped[cat];
          const countedInCat = catLines.filter((l) => l.counted_qty !== null).length;
          return (
            <CategoryGroup
              key={cat}
              category={cat}
              count={`${countedInCat}/${catLines.length}`}
            >
              {catLines.map((line) => (
                <CountRow
                  key={line.ingredient_id}
                  name={line.name}
                  unit={line.unit}
                  value={line.counted_qty}
                  onChangeValue={(qty) => updateLine(line.ingredient_id, qty)}
                  onChangeNote={(note) => updateLine(line.ingredient_id, line.counted_qty ?? 0, note)}
                />
              ))}
            </CategoryGroup>
          );
        })}
        <View style={styles.bottomPad} />
      </ScrollView>

      {/* Footer */}
      <View style={styles.footer}>
        <TouchableOpacity style={styles.footerCancelBtn} onPress={onCancel} activeOpacity={0.8}>
          <Text style={styles.footerCancelText}>Cancel</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.footerReviewBtn} onPress={onReview} activeOpacity={0.8}>
          <Text style={styles.footerReviewText}>Review Count</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Review Screen
// ────────────────────────────────────────────────────────────────────────────

function ReviewScreen({
  onSubmit,
  onBack,
  submitting,
}: {
  onSubmit: () => void;
  onBack: () => void;
  submitting: boolean;
}) {
  const { lines, progress } = useCountStore();

  const variantLines = lines.filter(
    (l) => l.counted_qty !== null && l.expected_qty > 0,
  );

  const renderItem = useCallback(
    ({ item: line }: { item: CountLine }) => {
      const color = varianceColor(line.expected_qty, line.counted_qty);
      return (
        <View style={[styles.reviewRow, color ? { borderLeftColor: color, borderLeftWidth: 3 } : undefined]}>
          <View style={styles.reviewNameCol}>
            <Text style={styles.reviewName}>{line.name}</Text>
            <Text style={styles.reviewCategory}>{line.category}</Text>
          </View>
          <View style={styles.reviewQtyGroup}>
            <View style={styles.reviewQtyBlock}>
              <Text style={styles.reviewQtyLabel}>Expected</Text>
              <Text style={styles.reviewQtyExpected}>
                {line.expected_qty} {line.unit}
              </Text>
            </View>
            <View style={styles.reviewQtyBlock}>
              <Text style={styles.reviewQtyLabel}>Counted</Text>
              <Text style={[styles.reviewQtyCounted, color ? { color } : undefined]}>
                {line.counted_qty !== null ? `${line.counted_qty} ${line.unit}` : '—'}
              </Text>
            </View>
          </View>
        </View>
      );
    },
    [],
  );

  return (
    <View style={styles.flex}>
      {/* Legend */}
      <View style={styles.reviewLegend}>
        <View style={styles.legendItem}>
          <View style={[styles.legendDot, { backgroundColor: '#f59e0b' }]} />
          <Text style={styles.legendText}>Variance &gt;10%</Text>
        </View>
        <View style={styles.legendItem}>
          <View style={[styles.legendDot, { backgroundColor: '#ef4444' }]} />
          <Text style={styles.legendText}>Variance &gt;15%</Text>
        </View>
        <Text style={styles.reviewProgress}>
          {progress.counted}/{progress.total} counted
        </Text>
      </View>

      <FlatList
        data={lines}
        keyExtractor={(item) => item.ingredient_id}
        renderItem={renderItem}
        style={styles.flex}
        ItemSeparatorComponent={() => <View style={styles.separator} />}
        keyboardShouldPersistTaps="handled"
      />

      <View style={styles.footer}>
        <TouchableOpacity style={styles.footerCancelBtn} onPress={onBack} activeOpacity={0.8}>
          <Text style={styles.footerCancelText}>Back</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.footerSubmitBtn, submitting && styles.footerBtnDisabled]}
          onPress={onSubmit}
          disabled={submitting}
          activeOpacity={0.8}
        >
          {submitting ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={styles.footerSubmitText}>Submit Count</Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Main Screen Component
// ────────────────────────────────────────────────────────────────────────────

export default function CountScreen() {
  const { activeCount, startCount, submitCount, resetCount, error } = useCountStore();
  const { touchActivity } = useAuthStore();

  const [screen, setScreen] = useState<Screen>('start');
  const [starting, setStarting] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Restore screen state if a count was already in progress
  useEffect(() => {
    if (activeCount && activeCount.status === 'open' && screen === 'start') {
      setScreen('counting');
    }
  }, []);

  const handleStart = useCallback(
    async (type: 'full' | 'spot') => {
      touchActivity();
      setStarting(true);
      try {
        await startCount(type);
        setScreen('counting');
      } catch {
        Alert.alert('Error', error ?? 'Could not start count. Please try again.');
      } finally {
        setStarting(false);
      }
    },
    [startCount, touchActivity, error],
  );

  const handleCancel = useCallback(() => {
    Alert.alert('Cancel Count', 'All unsaved changes will be lost. Are you sure?', [
      { text: 'Keep Counting', style: 'cancel' },
      {
        text: 'Cancel Count',
        style: 'destructive',
        onPress: () => {
          resetCount();
          setScreen('start');
        },
      },
    ]);
  }, [resetCount]);

  const handleReview = useCallback(() => {
    touchActivity();
    setScreen('review');
  }, [touchActivity]);

  const handleSubmit = useCallback(async () => {
    touchActivity();
    setSubmitting(true);
    try {
      await submitCount();
      Alert.alert('Count Submitted', 'Inventory count has been submitted successfully.', [
        {
          text: 'OK',
          onPress: () => {
            resetCount();
            setScreen('start');
          },
        },
      ]);
    } catch {
      Alert.alert('Error', error ?? 'Could not submit count. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }, [submitCount, resetCount, touchActivity, error]);

  if (starting) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color="#e94560" />
        <Text style={styles.loadingText}>Starting count…</Text>
      </View>
    );
  }

  return (
    <View style={styles.screenContainer}>
      {screen === 'start' && <StartScreen onStart={handleStart} />}
      {screen === 'counting' && (
        <CountingScreen onReview={handleReview} onCancel={handleCancel} />
      )}
      {screen === 'review' && (
        <ReviewScreen
          onSubmit={handleSubmit}
          onBack={() => setScreen('counting')}
          submitting={submitting}
        />
      )}
    </View>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Styles
// ────────────────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  screenContainer: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  flex: {
    flex: 1,
  },
  centered: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 16,
  },
  loadingText: {
    color: '#aaaaaa',
    fontSize: 16,
  },

  // Start screen
  startContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 32,
    gap: 16,
  },
  startTitle: {
    fontSize: 28,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 4,
  },
  startSubtitle: {
    fontSize: 16,
    color: '#888',
    marginBottom: 24,
  },
  startButtonRow: {
    flexDirection: 'row',
    gap: 20,
    flexWrap: 'wrap',
    justifyContent: 'center',
  },
  startButton: {
    width: 200,
    alignItems: 'center',
    borderRadius: 16,
    padding: 28,
    gap: 8,
    borderWidth: 1,
  },
  startButtonFull: {
    backgroundColor: '#16213e',
    borderColor: '#0f3460',
  },
  startButtonSpot: {
    backgroundColor: '#16213e',
    borderColor: '#e94560',
  },
  startButtonIcon: {
    fontSize: 36,
  },
  startButtonLabel: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
  },
  startButtonDesc: {
    fontSize: 13,
    color: '#888',
    textAlign: 'center',
  },

  // Search
  searchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 8,
    gap: 10,
    backgroundColor: '#16213e',
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
  },
  searchInput: {
    flex: 1,
    height: 44,
    backgroundColor: '#0f0f1e',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    paddingHorizontal: 14,
    fontSize: 16,
    color: '#ffffff',
  },
  syncButton: {
    height: 44,
    paddingHorizontal: 14,
    backgroundColor: '#0f1f0f',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#22c55e',
    justifyContent: 'center',
    alignItems: 'center',
    minWidth: 80,
  },
  syncText: {
    color: '#22c55e',
    fontSize: 14,
    fontWeight: '600',
  },
  bottomPad: {
    height: 24,
  },

  // Footer
  footer: {
    flexDirection: 'row',
    gap: 12,
    padding: 16,
    backgroundColor: '#16213e',
    borderTopWidth: 1,
    borderTopColor: '#0f3460',
  },
  footerCancelBtn: {
    flex: 1,
    height: 52,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#444',
    justifyContent: 'center',
    alignItems: 'center',
  },
  footerCancelText: {
    color: '#aaaaaa',
    fontSize: 16,
    fontWeight: '600',
  },
  footerReviewBtn: {
    flex: 2,
    height: 52,
    borderRadius: 10,
    backgroundColor: '#0f3460',
    justifyContent: 'center',
    alignItems: 'center',
  },
  footerReviewText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '700',
  },
  footerSubmitBtn: {
    flex: 2,
    height: 52,
    borderRadius: 10,
    backgroundColor: '#22c55e',
    justifyContent: 'center',
    alignItems: 'center',
  },
  footerSubmitText: {
    color: '#0a1a0a',
    fontSize: 16,
    fontWeight: '700',
  },
  footerBtnDisabled: {
    opacity: 0.5,
  },

  // Review
  reviewLegend: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 16,
    paddingHorizontal: 16,
    paddingVertical: 10,
    backgroundColor: '#16213e',
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  legendDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  legendText: {
    fontSize: 13,
    color: '#888',
  },
  reviewProgress: {
    marginLeft: 'auto',
    fontSize: 14,
    fontWeight: '600',
    color: '#aaaaaa',
  },
  reviewRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
    backgroundColor: '#1a1a2e',
    minHeight: 56,
    borderLeftWidth: 3,
    borderLeftColor: 'transparent',
  },
  reviewNameCol: {
    flex: 1,
    gap: 3,
  },
  reviewName: {
    fontSize: 16,
    color: '#e0e0e0',
    fontWeight: '500',
  },
  reviewCategory: {
    fontSize: 12,
    color: '#666',
    textTransform: 'uppercase',
  },
  reviewQtyGroup: {
    flexDirection: 'row',
    gap: 24,
  },
  reviewQtyBlock: {
    alignItems: 'flex-end',
    minWidth: 80,
  },
  reviewQtyLabel: {
    fontSize: 11,
    color: '#666',
    textTransform: 'uppercase',
    marginBottom: 2,
  },
  reviewQtyExpected: {
    fontSize: 15,
    color: '#888',
    fontWeight: '500',
  },
  reviewQtyCounted: {
    fontSize: 15,
    color: '#22c55e',
    fontWeight: '700',
  },
  separator: {
    height: 1,
    backgroundColor: '#1e1e3a',
  },
});
