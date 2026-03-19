import { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { useAuthStore } from '../../stores/auth';
import PinPad from '../../components/PinPad';

const MAX_FAILURES = 5;
const LOCKOUT_SECONDS = 30;

export default function PinScreen() {
  const { locationName, pinVerify } = useAuthStore();
  const [error, setError] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [failures, setFailures] = useState(0);
  const [lockedOut, setLockedOut] = useState(false);
  const [lockoutRemaining, setLockoutRemaining] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const startLockout = () => {
    setLockedOut(true);
    setLockoutRemaining(LOCKOUT_SECONDS);
    setError(`Too many failed attempts. Try again in ${LOCKOUT_SECONDS}s.`);

    timerRef.current = setInterval(() => {
      setLockoutRemaining((prev) => {
        if (prev <= 1) {
          if (timerRef.current) clearInterval(timerRef.current);
          setLockedOut(false);
          setFailures(0);
          setError(undefined);
          return 0;
        }
        const next = prev - 1;
        setError(`Too many failed attempts. Try again in ${next}s.`);
        return next;
      });
    }, 1000);
  };

  const handlePinComplete = async (pin: string) => {
    if (lockedOut || loading) return;

    setLoading(true);
    setError(undefined);

    try {
      await pinVerify(pin);
      // Navigation handled by _layout.tsx on activeStaff update
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : 'Invalid PIN. Please try again.';

      const newFailures = failures + 1;
      setFailures(newFailures);

      if (newFailures >= MAX_FAILURES) {
        startLockout();
      } else {
        setError(`${msg} (${MAX_FAILURES - newFailures} attempt${MAX_FAILURES - newFailures !== 1 ? 's' : ''} remaining)`);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.locationName}>{locationName ?? 'FireLine'}</Text>
      <Text style={styles.heading}>Enter your PIN</Text>

      {loading ? (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#e94560" />
          <Text style={styles.loadingText}>Verifying…</Text>
        </View>
      ) : (
        <PinPad
          onComplete={handlePinComplete}
          error={error}
          disabled={lockedOut || loading}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  locationName: {
    fontSize: 16,
    color: '#aaaaaa',
    marginBottom: 8,
    textAlign: 'center',
  },
  heading: {
    fontSize: 28,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 36,
    textAlign: 'center',
  },
  loadingContainer: {
    alignItems: 'center',
    gap: 16,
  },
  loadingText: {
    color: '#aaaaaa',
    fontSize: 16,
  },
});
