import { useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, FlatList } from 'react-native';
import { Slot, useRouter, useSegments } from 'expo-router';
import { useAuthStore } from '../stores/auth';

function LocationPicker() {
  const { locations, selectLocation } = useAuthStore();

  return (
    <View style={styles.centerContainer}>
      <Text style={styles.title}>Select Location</Text>
      <FlatList
        data={locations}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.locationButton}
            onPress={() => selectLocation(item.id, item.name)}
          >
            <Text style={styles.locationButtonText}>{item.name}</Text>
          </TouchableOpacity>
        )}
        contentContainerStyle={styles.listContent}
      />
    </View>
  );
}

export default function RootLayout() {
  const { managerToken, locationId, activeStaff } = useAuthStore();
  const router = useRouter();
  const segments = useSegments();

  useEffect(() => {
    const inAuthGroup = segments[0] === '(auth)';

    if (!managerToken) {
      if (!inAuthGroup) {
        router.replace('/(auth)/login');
      }
      return;
    }

    if (!locationId) {
      // Stay at root to show location picker (handled below)
      return;
    }

    if (!activeStaff) {
      if (segments[0] !== '(auth)' || segments[1] !== 'pin') {
        router.replace('/(auth)/pin');
      }
      return;
    }

    if (inAuthGroup) {
      router.replace('/(tabs)/count');
    }
  }, [managerToken, locationId, activeStaff, segments, router]);

  // Show location picker when logged in but no location chosen
  if (managerToken && !locationId) {
    return <LocationPicker />;
  }

  return <Slot />;
}

const styles = StyleSheet.create({
  centerContainer: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    paddingTop: 80,
    paddingHorizontal: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: '#ffffff',
    textAlign: 'center',
    marginBottom: 32,
  },
  listContent: {
    paddingBottom: 40,
  },
  locationButton: {
    backgroundColor: '#16213e',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#0f3460',
    paddingVertical: 20,
    paddingHorizontal: 24,
    marginBottom: 12,
    alignItems: 'center',
  },
  locationButtonText: {
    color: '#e94560',
    fontSize: 20,
    fontWeight: '600',
  },
});
