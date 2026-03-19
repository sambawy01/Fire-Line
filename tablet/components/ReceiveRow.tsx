import { useState, useEffect } from 'react';
import { View, Text, TextInput, TouchableOpacity, StyleSheet } from 'react-native';

interface ReceiveRowProps {
  ingredientName: string;
  orderedQty: number;
  orderedUnit: string;
  estimatedCost: number;
  receivedQty: number;
  receivedCost: number;
  verified: boolean;
  note: string;
  onChangeQty: (qty: number) => void;
  onChangeCost: (cost: number) => void;
  onNotReceived: () => void;
  onChangeNote: (note: string) => void;
}

type VarianceFlag = 'exact' | 'short' | 'over' | 'not_received';

function computeFlag(orderedQty: number, receivedQty: number): VarianceFlag {
  if (receivedQty === 0) return 'not_received';
  const ratio = receivedQty / orderedQty;
  if (ratio >= 0.98 && ratio <= 1.02) return 'exact';
  if (receivedQty < orderedQty) return 'short';
  return 'over';
}

export default function ReceiveRow({
  ingredientName,
  orderedQty,
  orderedUnit,
  estimatedCost,
  receivedQty,
  receivedCost,
  verified,
  note,
  onChangeQty,
  onChangeCost,
  onNotReceived,
  onChangeNote,
}: ReceiveRowProps) {
  const [rawQty, setRawQty] = useState(String(receivedQty));
  const [rawCost, setRawCost] = useState(receivedCost.toFixed(2));
  const [localNote, setLocalNote] = useState(note);

  // Keep local state in sync when store updates externally (e.g. markNotReceived)
  useEffect(() => {
    setRawQty(String(receivedQty));
  }, [receivedQty]);

  useEffect(() => {
    setRawCost(receivedCost.toFixed(2));
  }, [receivedCost]);

  useEffect(() => {
    setLocalNote(note);
  }, [note]);

  const handleQtyChange = (text: string) => {
    const cleaned = text.replace(/[^0-9.]/g, '');
    setRawQty(cleaned);
    const parsed = parseFloat(cleaned);
    if (!isNaN(parsed) && parsed >= 0) {
      onChangeQty(parsed);
    }
  };

  const handleCostChange = (text: string) => {
    const cleaned = text.replace(/[^0-9.]/g, '');
    setRawCost(cleaned);
    const parsed = parseFloat(cleaned);
    if (!isNaN(parsed) && parsed >= 0) {
      onChangeCost(parsed);
    }
  };

  const handleNoteChange = (text: string) => {
    setLocalNote(text);
    onChangeNote(text);
  };

  const parsedQty = parseFloat(rawQty) || 0;
  const flag = computeFlag(orderedQty, parsedQty);

  const varianceColor =
    flag === 'exact' ? '#22c55e' : flag === 'not_received' ? '#e94560' : '#f59e0b';
  const varianceIcon = flag === 'exact' ? '✓' : flag === 'not_received' ? '✗' : '⚠';
  const varianceLabel =
    flag === 'exact'
      ? 'OK'
      : flag === 'not_received'
        ? 'Not received'
        : flag === 'short'
          ? 'Short'
          : 'Over';

  const formattedEstCost = `$${estimatedCost.toFixed(2)}`;
  const formattedOrdered = `Ordered: ${orderedQty} ${orderedUnit} @ ${formattedEstCost}`;

  return (
    <View style={styles.container}>
      {/* Header row: name + variance badge */}
      <View style={styles.headerRow}>
        <View style={styles.nameBlock}>
          <Text style={styles.ingredientName} numberOfLines={2}>
            {ingredientName}
          </Text>
          <Text style={styles.orderedRef}>{formattedOrdered}</Text>
        </View>
        <View style={[styles.varianceBadge, { borderColor: varianceColor }]}>
          <Text style={[styles.varianceIcon, { color: varianceColor }]}>{varianceIcon}</Text>
          <Text style={[styles.varianceLabel, { color: varianceColor }]}>{varianceLabel}</Text>
        </View>
      </View>

      {/* Input row: qty + cost */}
      <View style={styles.inputRow}>
        <View style={styles.inputGroup}>
          <Text style={styles.inputLabel}>Received Qty</Text>
          <View style={styles.qtyRow}>
            <TextInput
              style={[
                styles.qtyInput,
                verified && flag === 'exact' && styles.qtyInputExact,
                verified && flag !== 'exact' && flag !== 'not_received' && styles.qtyInputWarning,
                flag === 'not_received' && styles.qtyInputError,
              ]}
              value={rawQty}
              onChangeText={handleQtyChange}
              keyboardType="decimal-pad"
              placeholder="0"
              placeholderTextColor="#444"
              returnKeyType="done"
              blurOnSubmit
              selectTextOnFocus
              accessibilityLabel={`Received quantity for ${ingredientName}`}
            />
            <Text style={styles.unit}>{orderedUnit}</Text>
          </View>
        </View>

        <View style={styles.inputGroup}>
          <Text style={styles.inputLabel}>Unit Cost ($)</Text>
          <TextInput
            style={styles.costInput}
            value={rawCost}
            onChangeText={handleCostChange}
            keyboardType="decimal-pad"
            placeholder="0.00"
            placeholderTextColor="#444"
            returnKeyType="done"
            blurOnSubmit
            selectTextOnFocus
            accessibilityLabel={`Unit cost for ${ingredientName}`}
          />
        </View>

        <TouchableOpacity
          style={styles.notReceivedButton}
          onPress={onNotReceived}
          accessibilityLabel={`Mark ${ingredientName} as not received`}
          hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
        >
          <Text style={styles.notReceivedText}>Not{'\n'}Received</Text>
        </TouchableOpacity>
      </View>

      {/* Note input */}
      <TextInput
        style={styles.noteInput}
        value={localNote}
        onChangeText={handleNoteChange}
        placeholder="Note (optional)…"
        placeholderTextColor="#444"
        returnKeyType="done"
        blurOnSubmit
        accessibilityLabel={`Note for ${ingredientName}`}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#16213e',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#1e1e3a',
    gap: 12,
  },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: 12,
  },
  nameBlock: {
    flex: 1,
    gap: 3,
  },
  ingredientName: {
    fontSize: 16,
    fontWeight: '700',
    color: '#e0e0e0',
  },
  orderedRef: {
    fontSize: 13,
    color: '#888888',
  },
  varianceBadge: {
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1.5,
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 6,
    minWidth: 72,
    gap: 2,
  },
  varianceIcon: {
    fontSize: 18,
    fontWeight: '700',
  },
  varianceLabel: {
    fontSize: 11,
    fontWeight: '600',
  },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    gap: 12,
  },
  inputGroup: {
    gap: 4,
  },
  inputLabel: {
    fontSize: 11,
    color: '#888888',
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  qtyRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  qtyInput: {
    width: 100,
    height: 52,
    borderRadius: 8,
    borderWidth: 1.5,
    borderColor: '#2a2a4a',
    backgroundColor: '#0f0f1e',
    color: '#ffffff',
    fontSize: 22,
    fontWeight: '700',
    textAlign: 'center',
    paddingHorizontal: 8,
  },
  qtyInputExact: {
    borderColor: '#22c55e',
    backgroundColor: '#0a1f0a',
  },
  qtyInputWarning: {
    borderColor: '#f59e0b',
    backgroundColor: '#1a1500',
  },
  qtyInputError: {
    borderColor: '#e94560',
    backgroundColor: '#1f0a0f',
  },
  unit: {
    fontSize: 14,
    color: '#888',
    minWidth: 32,
  },
  costInput: {
    width: 100,
    height: 52,
    borderRadius: 8,
    borderWidth: 1.5,
    borderColor: '#2a2a4a',
    backgroundColor: '#0f0f1e',
    color: '#ffffff',
    fontSize: 20,
    fontWeight: '600',
    textAlign: 'center',
    paddingHorizontal: 8,
  },
  notReceivedButton: {
    height: 52,
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 8,
    borderWidth: 1.5,
    borderColor: '#e94560',
    justifyContent: 'center',
    alignItems: 'center',
  },
  notReceivedText: {
    color: '#e94560',
    fontSize: 12,
    fontWeight: '700',
    textAlign: 'center',
    lineHeight: 16,
  },
  noteInput: {
    height: 40,
    borderRadius: 6,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    backgroundColor: '#0f0f1e',
    color: '#cccccc',
    fontSize: 14,
    paddingHorizontal: 10,
  },
});
