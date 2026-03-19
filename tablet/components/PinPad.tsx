import { useEffect, useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';

const PIN_LENGTH = 6;

interface PinPadProps {
  onComplete: (pin: string) => void;
  error?: string;
  disabled?: boolean;
}

const KEYS = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', '⌫'];

export default function PinPad({ onComplete, error, disabled }: PinPadProps) {
  const [pin, setPin] = useState('');

  // Clear pin when error changes to a non-empty value
  useEffect(() => {
    if (error) {
      setPin('');
    }
  }, [error]);

  const handleKey = (key: string) => {
    if (disabled) return;

    if (key === '⌫') {
      setPin((prev) => prev.slice(0, -1));
      return;
    }

    if (key === '') return;

    if (pin.length >= PIN_LENGTH) return;

    const next = pin + key;
    setPin(next);

    if (next.length === PIN_LENGTH) {
      // Small delay so the last dot fills before submitting
      setTimeout(() => {
        onComplete(next);
      }, 80);
    }
  };

  return (
    <View style={styles.container}>
      {/* PIN dots */}
      <View style={styles.dotsRow}>
        {Array.from({ length: PIN_LENGTH }).map((_, i) => (
          <View
            key={i}
            style={[styles.dot, i < pin.length ? styles.dotFilled : styles.dotEmpty]}
          />
        ))}
      </View>

      {/* Error message */}
      {error ? (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      ) : (
        <View style={styles.errorPlaceholder} />
      )}

      {/* Key grid */}
      <View style={styles.grid}>
        {KEYS.map((key, index) => {
          const isEmpty = key === '';
          const isBackspace = key === '⌫';

          return (
            <TouchableOpacity
              key={index}
              style={[
                styles.keyButton,
                isEmpty && styles.keyButtonInvisible,
                (disabled || (isEmpty)) && styles.keyButtonDisabled,
              ]}
              onPress={() => handleKey(key)}
              disabled={disabled || isEmpty}
              activeOpacity={0.7}
            >
              <Text
                style={[
                  styles.keyText,
                  isBackspace && styles.backspaceText,
                  disabled && styles.keyTextDisabled,
                ]}
              >
                {key}
              </Text>
            </TouchableOpacity>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    paddingHorizontal: 16,
  },
  dotsRow: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 16,
    marginBottom: 16,
  },
  dot: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
  },
  dotFilled: {
    backgroundColor: '#e94560',
    borderColor: '#e94560',
  },
  dotEmpty: {
    backgroundColor: 'transparent',
    borderColor: '#555',
  },
  errorContainer: {
    backgroundColor: '#3d0c11',
    borderRadius: 8,
    paddingVertical: 10,
    paddingHorizontal: 20,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#e94560',
    minWidth: 240,
    alignItems: 'center',
  },
  errorPlaceholder: {
    height: 46,
    marginBottom: 16,
  },
  errorText: {
    color: '#e94560',
    fontSize: 15,
    textAlign: 'center',
  },
  grid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    width: 3 * 80 + 2 * 12, // 3 columns × 80px + 2 gaps × 12px
    gap: 12,
  },
  keyButton: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#16213e',
    borderWidth: 1,
    borderColor: '#0f3460',
    justifyContent: 'center',
    alignItems: 'center',
  },
  keyButtonInvisible: {
    backgroundColor: 'transparent',
    borderColor: 'transparent',
  },
  keyButtonDisabled: {
    opacity: 0.5,
  },
  keyText: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: '600',
  },
  backspaceText: {
    fontSize: 20,
  },
  keyTextDisabled: {
    color: '#888',
  },
});
