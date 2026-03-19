import { useState, useMemo } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  StyleSheet,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { WasteIngredient, WasteReason, LogWasteInput } from '../stores/waste';

const REASONS: { key: WasteReason; label: string; color: string }[] = [
  { key: 'expired', label: 'Expired', color: '#ef4444' },
  { key: 'dropped', label: 'Dropped', color: '#f97316' },
  { key: 'overcooked', label: 'Overcooked', color: '#f59e0b' },
  { key: 'contaminated', label: 'Contaminated', color: '#dc2626' },
  { key: 'overproduction', label: 'Overproduction', color: '#3b82f6' },
  { key: 'other', label: 'Other', color: '#6b7280' },
];

interface WasteFormProps {
  ingredients: WasteIngredient[];
  onSubmit: (input: LogWasteInput) => Promise<void>;
}

export default function WasteForm({ ingredients, onSubmit }: WasteFormProps) {
  const [ingredientSearch, setIngredientSearch] = useState('');
  const [selectedIngredient, setSelectedIngredient] = useState<WasteIngredient | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [quantity, setQuantity] = useState('');
  const [reason, setReason] = useState<WasteReason | null>(null);
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const filteredIngredients = useMemo(() => {
    if (!ingredientSearch.trim()) return ingredients.slice(0, 20);
    const q = ingredientSearch.toLowerCase();
    return ingredients.filter((i) => i.name.toLowerCase().includes(q)).slice(0, 20);
  }, [ingredients, ingredientSearch]);

  const canSubmit =
    selectedIngredient !== null &&
    quantity.trim() !== '' &&
    parseFloat(quantity) > 0 &&
    reason !== null;

  const handleSelectIngredient = (ing: WasteIngredient) => {
    setSelectedIngredient(ing);
    setIngredientSearch(ing.name);
    setShowDropdown(false);
  };

  const handleSubmit = async () => {
    if (!canSubmit || !selectedIngredient || !reason) return;
    const qty = parseFloat(quantity);
    if (isNaN(qty) || qty <= 0) {
      Alert.alert('Invalid Quantity', 'Please enter a valid quantity greater than zero.');
      return;
    }
    setSubmitting(true);
    try {
      await onSubmit({
        ingredient_id: selectedIngredient.ingredient_id,
        quantity: qty,
        unit: selectedIngredient.unit,
        reason,
        note: note.trim(),
      });
      // Reset form
      setSelectedIngredient(null);
      setIngredientSearch('');
      setQuantity('');
      setReason(null);
      setNote('');
    } catch {
      // Error already handled upstream
    } finally {
      setSubmitting(false);
    }
  };

  const handleCameraPress = () => {
    Alert.alert('Coming Soon', 'Photo upload coming soon.');
  };

  return (
    <View style={styles.container}>
      <Text style={styles.sectionTitle}>Log Waste</Text>

      {/* Ingredient picker */}
      <View style={styles.field}>
        <Text style={styles.label}>Ingredient</Text>
        <View style={styles.searchContainer}>
          <TextInput
            style={styles.ingredientInput}
            value={ingredientSearch}
            onChangeText={(t) => {
              setIngredientSearch(t);
              setShowDropdown(true);
              if (selectedIngredient && t !== selectedIngredient.name) {
                setSelectedIngredient(null);
              }
            }}
            onFocus={() => setShowDropdown(true)}
            placeholder="Search ingredient…"
            placeholderTextColor="#555"
            returnKeyType="search"
          />
          {selectedIngredient && (
            <TouchableOpacity
              style={styles.clearBtn}
              onPress={() => {
                setSelectedIngredient(null);
                setIngredientSearch('');
                setShowDropdown(false);
              }}
            >
              <Text style={styles.clearBtnText}>✕</Text>
            </TouchableOpacity>
          )}
        </View>

        {showDropdown && filteredIngredients.length > 0 && !selectedIngredient && (
          <View style={styles.dropdown}>
            <FlatList
              data={filteredIngredients}
              keyExtractor={(i) => i.ingredient_id}
              keyboardShouldPersistTaps="handled"
              nestedScrollEnabled
              style={styles.dropdownList}
              renderItem={({ item }) => (
                <TouchableOpacity
                  style={styles.dropdownItem}
                  onPress={() => handleSelectIngredient(item)}
                  activeOpacity={0.75}
                >
                  <Text style={styles.dropdownItemName}>{item.name}</Text>
                  <Text style={styles.dropdownItemUnit}>
                    {item.unit} · {item.category}
                  </Text>
                </TouchableOpacity>
              )}
            />
          </View>
        )}
      </View>

      {/* Quantity + Unit */}
      <View style={styles.qtyRow}>
        <View style={[styles.field, styles.qtyField]}>
          <Text style={styles.label}>Quantity</Text>
          <TextInput
            style={styles.qtyInput}
            value={quantity}
            onChangeText={(t) => setQuantity(t.replace(/[^0-9.]/g, ''))}
            keyboardType="decimal-pad"
            placeholder="0"
            placeholderTextColor="#444"
            returnKeyType="done"
            blurOnSubmit
            selectTextOnFocus
          />
        </View>
        <View style={[styles.field, styles.unitField]}>
          <Text style={styles.label}>Unit</Text>
          <View style={styles.unitDisplay}>
            <Text style={styles.unitText}>
              {selectedIngredient ? selectedIngredient.unit : '—'}
            </Text>
          </View>
        </View>
      </View>

      {/* Reason chips */}
      <View style={styles.field}>
        <Text style={styles.label}>Reason</Text>
        <View style={styles.reasonRow}>
          {REASONS.map((r) => (
            <TouchableOpacity
              key={r.key}
              style={[
                styles.reasonChip,
                { borderColor: r.color },
                reason === r.key && { backgroundColor: r.color },
              ]}
              onPress={() => setReason(r.key)}
              activeOpacity={0.75}
            >
              <Text
                style={[
                  styles.reasonChipText,
                  { color: reason === r.key ? '#fff' : r.color },
                ]}
              >
                {r.label}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      {/* Note */}
      <View style={styles.field}>
        <Text style={styles.label}>Note (optional)</Text>
        <TextInput
          style={styles.noteInput}
          value={note}
          onChangeText={setNote}
          placeholder="e.g. found in walk-in expired"
          placeholderTextColor="#555"
          returnKeyType="done"
          blurOnSubmit
        />
      </View>

      {/* Actions */}
      <View style={styles.actionRow}>
        <TouchableOpacity style={styles.cameraBtn} onPress={handleCameraPress} activeOpacity={0.8}>
          <Text style={styles.cameraBtnText}>📷</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.submitBtn, !canSubmit && styles.submitBtnDisabled]}
          onPress={handleSubmit}
          disabled={!canSubmit || submitting}
          activeOpacity={0.8}
        >
          {submitting ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={styles.submitBtnText}>Log Waste</Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );
}

export { REASONS };

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#16213e',
    padding: 16,
    gap: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 2,
  },
  field: {
    gap: 6,
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: '#888',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  searchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  ingredientInput: {
    flex: 1,
    height: 52,
    backgroundColor: '#0f0f1e',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    paddingHorizontal: 14,
    fontSize: 18,
    color: '#ffffff',
  },
  clearBtn: {
    position: 'absolute',
    right: 12,
    height: 52,
    justifyContent: 'center',
    paddingHorizontal: 4,
  },
  clearBtnText: {
    color: '#888',
    fontSize: 16,
  },
  dropdown: {
    position: 'absolute',
    top: 80,
    left: 0,
    right: 0,
    zIndex: 100,
    backgroundColor: '#0f0f1e',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    maxHeight: 200,
    overflow: 'hidden',
  },
  dropdownList: {
    maxHeight: 200,
  },
  dropdownItem: {
    paddingHorizontal: 14,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#1e1e3a',
    minHeight: 52,
    justifyContent: 'center',
  },
  dropdownItemName: {
    fontSize: 16,
    color: '#ffffff',
    fontWeight: '500',
  },
  dropdownItemUnit: {
    fontSize: 13,
    color: '#888',
    marginTop: 2,
  },
  qtyRow: {
    flexDirection: 'row',
    gap: 12,
  },
  qtyField: {
    flex: 2,
  },
  unitField: {
    flex: 1,
  },
  qtyInput: {
    height: 52,
    backgroundColor: '#0f0f1e',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    paddingHorizontal: 14,
    fontSize: 22,
    fontWeight: '700',
    color: '#ffffff',
    textAlign: 'center',
  },
  unitDisplay: {
    height: 52,
    backgroundColor: '#0a0a1a',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#1e1e3a',
    justifyContent: 'center',
    alignItems: 'center',
  },
  unitText: {
    fontSize: 18,
    color: '#aaaaaa',
    fontWeight: '600',
  },
  reasonRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  reasonChip: {
    paddingHorizontal: 14,
    paddingVertical: 10,
    borderRadius: 24,
    borderWidth: 1.5,
    minHeight: 44,
    justifyContent: 'center',
    alignItems: 'center',
  },
  reasonChipText: {
    fontSize: 14,
    fontWeight: '600',
  },
  noteInput: {
    height: 48,
    backgroundColor: '#0f0f1e',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#2a2a4a',
    paddingHorizontal: 14,
    fontSize: 16,
    color: '#ffffff',
  },
  actionRow: {
    flexDirection: 'row',
    gap: 12,
    alignItems: 'center',
  },
  cameraBtn: {
    width: 52,
    height: 52,
    borderRadius: 10,
    backgroundColor: '#1a1a2e',
    borderWidth: 1,
    borderColor: '#2a2a4a',
    justifyContent: 'center',
    alignItems: 'center',
  },
  cameraBtnText: {
    fontSize: 22,
  },
  submitBtn: {
    flex: 1,
    height: 52,
    borderRadius: 10,
    backgroundColor: '#e94560',
    justifyContent: 'center',
    alignItems: 'center',
  },
  submitBtnDisabled: {
    opacity: 0.4,
  },
  submitBtnText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '700',
  },
});
