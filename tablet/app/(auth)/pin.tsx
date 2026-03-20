import { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ActivityIndicator,
  Animated,
} from 'react-native';
import { useAuthStore } from '../../stores/auth';
import PinPad from '../../components/PinPad';

const MAX_FAILURES = 5;
const LOCKOUT_SECONDS = 30;
const WELCOME_DURATION_MS = 3000;

type Trend = 'up' | 'down' | 'stable';

function trendArrow(trend: Trend): string {
  if (trend === 'up') return '↑';
  if (trend === 'down') return '↓';
  return '→';
}

function trendColor(trend: Trend): string {
  if (trend === 'up') return '#22c55e';   // green-500
  if (trend === 'down') return '#ef4444'; // red-500
  return '#9ca3af';                        // gray-400
}

interface WelcomeOverlayProps {
  displayName: string;
  staffPoints: number;
  pointsTrend: Trend;
  onDone: () => void;
}

function WelcomeOverlay({ displayName, staffPoints, pointsTrend, onDone }: WelcomeOverlayProps) {
  const opacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    Animated.sequence([
      Animated.timing(opacity, { toValue: 1, duration: 300, useNativeDriver: true }),
      Animated.delay(WELCOME_DURATION_MS - 600),
      Animated.timing(opacity, { toValue: 0, duration: 300, useNativeDriver: true }),
    ]).start(() => onDone());
  }, [opacity, onDone]);

  return (
    <Animated.View style={[styles.overlayContainer, { opacity }]}>
      <Text style={styles.welcomeGreeting}>Welcome back,</Text>
      <Text style={styles.welcomeName}>{displayName}</Text>
      <View style={styles.pointsRow}>
        <Text style={styles.pointsValue}>{staffPoints.toLocaleString()} pts</Text>
        <Text style={[styles.trendArrow, { color: trendColor(pointsTrend) }]}>
          {' '}{trendArrow(pointsTrend)}
        </Text>
      </View>
    </Animated.View>
  );
}

export default function PinScreen() {
  const { locationName, pinVerify } = useAuthStore();
  const [error, setError] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [failures, setFailures] = useState(0);
  const [lockedOut, setLockedOut] = useState(false);
  const [lockoutRemaining, setLockoutRemaining] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Welcome overlay state
  const [welcomeVisible, setWelcomeVisible] = useState(false);
  const [welcomeData, setWelcomeData] = useState<{
    displayName: string;
    staffPoints: number;
    pointsTrend: Trend;
  } | null>(null);

  // activeStaff for reading points after verify
  const activeStaff = useAuthStore((s) => s.activeStaff);

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
      // Read updated activeStaff from store immediately after verify
      const staff = useAuthStore.getState().activeStaff;
      if (staff) {
        setWelcomeData({
          displayName: staff.display_name,
          staffPoints: staff.staff_points,
          pointsTrend: staff.points_trend,
        });
        setWelcomeVisible(true);
      }
      // Navigation handled by _layout.tsx on activeStaff update after overlay hides
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

  // While welcome overlay is shown, render it instead of the PIN pad
  if (welcomeVisible && welcomeData) {
    return (
      <View style={styles.container}>
        <WelcomeOverlay
          displayName={welcomeData.displayName}
          staffPoints={welcomeData.staffPoints}
          pointsTrend={welcomeData.pointsTrend}
          onDone={() => setWelcomeVisible(false)}
        />
      </View>
    );
  }

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
  // Welcome overlay
  overlayContainer: {
    alignItems: 'center',
    gap: 12,
  },
  welcomeGreeting: {
    fontSize: 20,
    color: '#aaaaaa',
    textAlign: 'center',
  },
  welcomeName: {
    fontSize: 40,
    fontWeight: '800',
    color: '#ffffff',
    textAlign: 'center',
  },
  pointsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 8,
  },
  pointsValue: {
    fontSize: 26,
    fontWeight: '700',
    color: '#ffffff',
  },
  trendArrow: {
    fontSize: 26,
    fontWeight: '800',
  },
});
