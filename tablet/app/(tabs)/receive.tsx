import { useEffect, useCallback, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Alert,
  ScrollView,
} from 'react-native';
import { useReceiveStore } from '../../stores/receive';
import type { PurchaseOrder, POLine, Discrepancy } from '../../stores/receive';
import ReceiveRow from '../../components/ReceiveRow';
import ProgressBar from '../../components/ProgressBar';

// ─── Helpers ────────────────────────────────────────────────────────────────

function formatCents(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function daysSince(iso: string): number {
  const ms = Date.now() - new Date(iso).getTime();
  return Math.floor(ms / (1000 * 60 * 60 * 24));
}

// ─── Phase 1: Pending list ───────────────────────────────────────────────────

function PendingList() {
  const { pendingPOs, loadPending, startReceiving, loading, error } = useReceiveStore();

  useEffect(() => {
    loadPending();
  }, [loadPending]);

  const handleSelect = useCallback(
    async (po: PurchaseOrder) => {
      await startReceiving(po.purchase_order_id);
    },
    [startReceiving],
  );

  if (loading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color="#e94560" />
        <Text style={styles.loadingText}>Loading deliveries…</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.centered}>
        <Text style={styles.errorText}>{error}</Text>
        <TouchableOpacity style={styles.retryButton} onPress={loadPending}>
          <Text style={styles.retryButtonText}>Retry</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (pendingPOs.length === 0) {
    return (
      <View style={styles.centered}>
        <Text style={styles.emptyIcon}>📦</Text>
        <Text style={styles.emptyTitle}>No Pending Deliveries</Text>
        <Text style={styles.emptySubtitle}>All purchase orders have been received.</Text>
      </View>
    );
  }

  const renderItem = ({ item }: { item: PurchaseOrder }) => {
    const days = daysSince(item.approved_at);
    const daysLabel = days === 0 ? 'Today' : days === 1 ? '1 day ago' : `${days} days ago`;

    return (
      <TouchableOpacity
        style={styles.poCard}
        onPress={() => handleSelect(item)}
        accessibilityLabel={`Receive delivery from ${item.vendor_name}`}
      >
        <View style={styles.poCardHeader}>
          <Text style={styles.vendorName}>{item.vendor_name}</Text>
          <Text style={styles.poAge}>{daysLabel}</Text>
        </View>
        <View style={styles.poCardFooter}>
          <Text style={styles.poMeta}>
            {item.line_count} {item.line_count === 1 ? 'item' : 'items'}
          </Text>
          <Text style={styles.poEstTotal}>{formatCents(item.total_estimated)}</Text>
        </View>
        <View style={styles.poChevron}>
          <Text style={styles.poChevronText}>›</Text>
        </View>
      </TouchableOpacity>
    );
  };

  return (
    <FlatList
      data={pendingPOs}
      keyExtractor={(item) => item.purchase_order_id}
      renderItem={renderItem}
      contentContainerStyle={styles.listContent}
      style={styles.list}
    />
  );
}

// ─── Phase 2: Line-by-line receiving ────────────────────────────────────────

function ReceivingPhase() {
  const {
    activePO,
    receivedLines,
    updateLine,
    markNotReceived,
    getProgress,
    phase,
    error,
    loading,
  } = useReceiveStore();
  const setPhase = useReceiveStore((s) => s.phase);

  const progress = getProgress();
  const allVerified = progress.verified === progress.total && progress.total > 0;

  const handleReview = () => {
    useReceiveStore.setState({ phase: 'review' });
  };

  if (loading || !activePO) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color="#e94560" />
        <Text style={styles.loadingText}>Loading PO…</Text>
      </View>
    );
  }

  const renderLine = ({ item }: { item: POLine }) => {
    const entry = receivedLines[item.po_line_id] ?? {
      received_qty: item.ordered_qty,
      received_unit_cost: item.estimated_unit_cost,
      note: '',
      verified: false,
    };

    return (
      <ReceiveRow
        ingredientName={item.ingredient_name}
        orderedQty={item.ordered_qty}
        orderedUnit={item.ordered_unit}
        estimatedCost={item.estimated_unit_cost}
        receivedQty={entry.received_qty}
        receivedCost={entry.received_unit_cost}
        verified={entry.verified}
        note={entry.note}
        onChangeQty={(qty) =>
          updateLine(item.po_line_id, qty, entry.received_unit_cost, entry.note)
        }
        onChangeCost={(cost) =>
          updateLine(item.po_line_id, entry.received_qty, cost, entry.note)
        }
        onChangeNote={(note) =>
          updateLine(item.po_line_id, entry.received_qty, entry.received_unit_cost, note)
        }
        onNotReceived={() => markNotReceived(item.po_line_id)}
      />
    );
  };

  return (
    <View style={styles.phaseContainer}>
      {/* Header */}
      <View style={styles.receivingHeader}>
        <View>
          <Text style={styles.receivingVendor}>{activePO.vendor_name}</Text>
          <Text style={styles.receivingPoId}>
            PO #{activePO.purchase_order_id.slice(0, 8).toUpperCase()}
          </Text>
        </View>
      </View>

      {error && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorBannerText}>{error}</Text>
        </View>
      )}

      {/* Progress */}
      <ProgressBar counted={progress.verified} total={progress.total} />

      {/* Line items */}
      <FlatList
        data={activePO.lines}
        keyExtractor={(item) => item.po_line_id}
        renderItem={renderLine}
        style={styles.list}
        contentContainerStyle={{ paddingBottom: 100 }}
        keyboardShouldPersistTaps="handled"
      />

      {/* Review button */}
      <View style={styles.bottomBar}>
        <TouchableOpacity
          style={[styles.primaryButton, !allVerified && styles.primaryButtonDisabled]}
          onPress={handleReview}
          disabled={!allVerified}
          accessibilityLabel="Review and submit receiving"
        >
          <Text style={styles.primaryButtonText}>
            {allVerified
              ? 'Review & Submit'
              : `${progress.total - progress.verified} items remaining`}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

// ─── Phase 3: Review & Submit ────────────────────────────────────────────────

function ReviewPhase() {
  const { activePO, getDiscrepancies, submitReceiving, reset, submitting, error } =
    useReceiveStore();
  const receivedLines = useReceiveStore((s) => s.receivedLines);
  const [submitted, setSubmitted] = useState(false);

  const discrepancies = getDiscrepancies();
  const totalItems = activePO?.lines.length ?? 0;

  const handleBack = () => {
    useReceiveStore.setState({ phase: 'receiving' });
  };

  const handleSubmit = async () => {
    try {
      await submitReceiving();
      setSubmitted(true);
    } catch {
      // error is stored in the store, displayed below
    }
  };

  if (submitted) {
    return (
      <View style={styles.centered}>
        <Text style={styles.successIcon}>✓</Text>
        <Text style={styles.successTitle}>Receiving Complete</Text>
        <Text style={styles.successSubtitle}>
          {activePO?.vendor_name ?? 'Delivery'} has been recorded.
        </Text>
        <TouchableOpacity
          style={[styles.primaryButton, { marginTop: 32 }]}
          onPress={reset}
          accessibilityLabel="Return to deliveries list"
        >
          <Text style={styles.primaryButtonText}>Back to Deliveries</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const renderDiscrepancy = ({ item }: { item: Discrepancy }) => {
    const isShort = item.flag === 'short';
    const isOver = item.flag === 'over';
    const isNotReceived = item.flag === 'not_received';

    const flagColor = isNotReceived ? '#e94560' : '#f59e0b';
    const flagLabel = isNotReceived ? 'Not Received' : isShort ? 'Short' : 'Over';

    return (
      <View style={styles.discrepancyRow}>
        <View style={styles.discrepancyLeft}>
          <Text style={styles.discrepancyName}>{item.ingredient_name}</Text>
          <Text style={styles.discrepancyDetail}>
            Ordered: {item.ordered_qty} {item.ordered_unit} → Received: {item.received_qty}{' '}
            {item.ordered_unit}
          </Text>
        </View>
        <View style={[styles.flagBadge, { backgroundColor: flagColor + '22', borderColor: flagColor }]}>
          <Text style={[styles.flagBadgeText, { color: flagColor }]}>{flagLabel}</Text>
        </View>
      </View>
    );
  };

  return (
    <View style={styles.phaseContainer}>
      <ScrollView contentContainerStyle={styles.reviewContent} keyboardShouldPersistTaps="handled">
        {/* Summary */}
        <View style={styles.reviewSummaryCard}>
          <Text style={styles.reviewSummaryTitle}>Summary</Text>
          <View style={styles.reviewSummaryRow}>
            <Text style={styles.reviewSummaryLabel}>Items Verified</Text>
            <Text style={styles.reviewSummaryValue}>{totalItems}</Text>
          </View>
          <View style={styles.reviewSummaryRow}>
            <Text style={styles.reviewSummaryLabel}>Discrepancies</Text>
            <Text
              style={[
                styles.reviewSummaryValue,
                discrepancies.length > 0 && { color: '#f59e0b' },
              ]}
            >
              {discrepancies.length}
            </Text>
          </View>
        </View>

        {discrepancies.length > 0 && (
          <>
            <Text style={styles.discrepanciesHeading}>Discrepancies</Text>
            {discrepancies.map((item) => (
              <View key={item.po_line_id}>
                {renderDiscrepancy({ item })}
              </View>
            ))}
          </>
        )}

        {error && (
          <View style={styles.errorBanner}>
            <Text style={styles.errorBannerText}>{error}</Text>
          </View>
        )}
      </ScrollView>

      {/* Action buttons */}
      <View style={styles.bottomBar}>
        <TouchableOpacity
          style={styles.secondaryButton}
          onPress={handleBack}
          accessibilityLabel="Back to edit receiving"
        >
          <Text style={styles.secondaryButtonText}>Back to Edit</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.primaryButton, { flex: 1 }, submitting && styles.primaryButtonDisabled]}
          onPress={handleSubmit}
          disabled={submitting}
          accessibilityLabel="Complete receiving"
        >
          {submitting ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.primaryButtonText}>Complete Receiving</Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );
}

// ─── Root screen ─────────────────────────────────────────────────────────────

export default function ReceiveScreen() {
  const phase = useReceiveStore((s) => s.phase);

  return (
    <View style={styles.screen}>
      {phase === 'list' && <PendingList />}
      {phase === 'receiving' && <ReceivingPhase />}
      {phase === 'review' && <ReviewPhase />}
    </View>
  );
}

// ─── Styles ──────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 32,
    gap: 12,
  },
  loadingText: {
    color: '#888',
    fontSize: 16,
    marginTop: 12,
  },
  errorText: {
    color: '#e94560',
    fontSize: 15,
    textAlign: 'center',
  },
  retryButton: {
    marginTop: 12,
    paddingVertical: 12,
    paddingHorizontal: 28,
    borderRadius: 8,
    backgroundColor: '#e94560',
  },
  retryButtonText: {
    color: '#fff',
    fontWeight: '700',
    fontSize: 15,
  },
  emptyIcon: {
    fontSize: 56,
    marginBottom: 8,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#e0e0e0',
  },
  emptySubtitle: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
    marginTop: 4,
  },
  list: {
    flex: 1,
  },
  listContent: {
    padding: 16,
    gap: 12,
  },
  // PO card
  poCard: {
    backgroundColor: '#16213e',
    borderRadius: 12,
    padding: 18,
    borderWidth: 1,
    borderColor: '#0f3460',
    position: 'relative',
    minHeight: 80,
  },
  poCardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 8,
  },
  vendorName: {
    fontSize: 18,
    fontWeight: '700',
    color: '#e0e0e0',
    flex: 1,
    marginRight: 8,
  },
  poAge: {
    fontSize: 13,
    color: '#888',
  },
  poCardFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  poMeta: {
    fontSize: 14,
    color: '#aaa',
  },
  poEstTotal: {
    fontSize: 16,
    fontWeight: '600',
    color: '#e0e0e0',
  },
  poChevron: {
    position: 'absolute',
    right: 16,
    top: '50%',
  },
  poChevronText: {
    fontSize: 28,
    color: '#e94560',
    lineHeight: 32,
  },
  // Phase containers
  phaseContainer: {
    flex: 1,
  },
  receivingHeader: {
    backgroundColor: '#16213e',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  receivingVendor: {
    fontSize: 18,
    fontWeight: '700',
    color: '#e0e0e0',
  },
  receivingPoId: {
    fontSize: 13,
    color: '#888',
    marginTop: 2,
  },
  errorBanner: {
    backgroundColor: '#3a0a12',
    borderLeftWidth: 4,
    borderLeftColor: '#e94560',
    paddingHorizontal: 16,
    paddingVertical: 10,
    marginHorizontal: 16,
    marginTop: 8,
    borderRadius: 4,
  },
  errorBannerText: {
    color: '#e94560',
    fontSize: 14,
  },
  bottomBar: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: '#16213e',
    borderTopWidth: 1,
    borderTopColor: '#0f3460',
    padding: 16,
    flexDirection: 'row',
    gap: 12,
  },
  primaryButton: {
    backgroundColor: '#e94560',
    borderRadius: 10,
    paddingVertical: 16,
    alignItems: 'center',
    justifyContent: 'center',
    flex: 1,
    minHeight: 52,
  },
  primaryButtonDisabled: {
    backgroundColor: '#4a2030',
  },
  primaryButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  secondaryButton: {
    borderRadius: 10,
    paddingVertical: 16,
    paddingHorizontal: 20,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1.5,
    borderColor: '#e94560',
    minHeight: 52,
  },
  secondaryButtonText: {
    color: '#e94560',
    fontSize: 16,
    fontWeight: '700',
  },
  // Review phase
  reviewContent: {
    padding: 16,
    gap: 16,
    paddingBottom: 120,
  },
  reviewSummaryCard: {
    backgroundColor: '#16213e',
    borderRadius: 12,
    padding: 18,
    borderWidth: 1,
    borderColor: '#0f3460',
    gap: 10,
  },
  reviewSummaryTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#e0e0e0',
    marginBottom: 4,
  },
  reviewSummaryRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  reviewSummaryLabel: {
    fontSize: 15,
    color: '#aaa',
  },
  reviewSummaryValue: {
    fontSize: 18,
    fontWeight: '700',
    color: '#e0e0e0',
  },
  discrepanciesHeading: {
    fontSize: 15,
    fontWeight: '700',
    color: '#f59e0b',
    marginTop: 4,
  },
  discrepancyRow: {
    backgroundColor: '#16213e',
    borderRadius: 10,
    padding: 14,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#1e1e3a',
    gap: 12,
  },
  discrepancyLeft: {
    flex: 1,
    gap: 3,
  },
  discrepancyName: {
    fontSize: 15,
    fontWeight: '600',
    color: '#e0e0e0',
  },
  discrepancyDetail: {
    fontSize: 13,
    color: '#888',
  },
  flagBadge: {
    borderWidth: 1,
    borderRadius: 6,
    paddingHorizontal: 10,
    paddingVertical: 5,
  },
  flagBadgeText: {
    fontSize: 12,
    fontWeight: '700',
  },
  // Success state
  successIcon: {
    fontSize: 64,
    color: '#22c55e',
    marginBottom: 8,
  },
  successTitle: {
    fontSize: 24,
    fontWeight: '700',
    color: '#e0e0e0',
  },
  successSubtitle: {
    fontSize: 15,
    color: '#888',
    textAlign: 'center',
  },
});
