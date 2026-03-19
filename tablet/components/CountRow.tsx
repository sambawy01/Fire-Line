import { useState } from 'react';
import { View, Text, TextInput, StyleSheet } from 'react-native';

interface CountRowProps {
  name: string;
  unit: string;
  value: number | null;
  onChangeValue: (qty: number) => void;
  onChangeNote?: (note: string) => void;
}

export default function CountRow({
  name,
  unit,
  value,
  onChangeValue,
  onChangeNote,
}: CountRowProps) {
  const [rawInput, setRawInput] = useState(value !== null ? String(value) : '');
  const [note, setNote] = useState('');

  const handleQtyChange = (text: string) => {
    // Allow only numeric input with optional single decimal point
    const cleaned = text.replace(/[^0-9.]/g, '');
    setRawInput(cleaned);
    const parsed = parseFloat(cleaned);
    if (!isNaN(parsed) && parsed >= 0) {
      onChangeValue(parsed);
    }
  };

  const handleNoteChange = (text: string) => {
    setNote(text);
    onChangeNote?.(text);
  };

  const hasValue = value !== null;

  return (
    <View style={styles.container}>
      <View style={styles.mainRow}>
        <View style={styles.nameContainer}>
          <Text style={styles.name} numberOfLines={2}>
            {name}
          </Text>
          {onChangeNote !== undefined && (
            <TextInput
              style={styles.noteInput}
              value={note}
              onChangeText={handleNoteChange}
              placeholder="Note (optional)"
              placeholderTextColor="#555"
              returnKeyType="done"
              blurOnSubmit
            />
          )}
        </View>
        <View style={styles.inputContainer}>
          <TextInput
            style={[styles.qtyInput, hasValue && styles.qtyInputFilled]}
            value={rawInput}
            onChangeText={handleQtyChange}
            keyboardType="decimal-pad"
            placeholder="—"
            placeholderTextColor="#444"
            returnKeyType="done"
            blurOnSubmit
            selectTextOnFocus
            accessibilityLabel={`Quantity for ${name}`}
          />
          <Text style={styles.unit}>{unit}</Text>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    borderBottomWidth: 1,
    borderBottomColor: '#1e1e3a',
    backgroundColor: '#1a1a2e',
    paddingHorizontal: 16,
    paddingVertical: 10,
    minHeight: 56,
  },
  mainRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  nameContainer: {
    flex: 1,
    gap: 4,
  },
  name: {
    fontSize: 16,
    color: '#e0e0e0',
    fontWeight: '500',
  },
  noteInput: {
    fontSize: 13,
    color: '#aaaaaa',
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a4a',
    paddingVertical: 2,
    marginTop: 2,
  },
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  qtyInput: {
    width: 90,
    height: 52,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    backgroundColor: '#0f0f1e',
    color: '#ffffff',
    fontSize: 22,
    fontWeight: '700',
    textAlign: 'center',
    paddingHorizontal: 8,
  },
  qtyInputFilled: {
    borderColor: '#22c55e',
    backgroundColor: '#0f1f0f',
  },
  unit: {
    fontSize: 14,
    color: '#888',
    width: 36,
  },
});
