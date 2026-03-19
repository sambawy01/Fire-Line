import { useEffect, useRef } from 'react';
import { Tabs } from 'expo-router';
import { Text } from 'react-native';
import { useAuthStore } from '../../stores/auth';

function TabIcon({ label }: { label: string }) {
  return <Text style={{ fontSize: 20 }}>{label}</Text>;
}

export default function TabsLayout() {
  const { checkTimeout, touchActivity } = useAuthStore();
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    // Check for inactivity every 30 seconds
    intervalRef.current = setInterval(() => {
      checkTimeout();
    }, 30_000);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [checkTimeout]);

  return (
    <Tabs
      screenListeners={{
        state: () => {
          touchActivity();
        },
      }}
      screenOptions={{
        tabBarStyle: {
          backgroundColor: '#16213e',
          borderTopColor: '#0f3460',
          borderTopWidth: 1,
          height: 64,
        },
        tabBarActiveTintColor: '#e94560',
        tabBarInactiveTintColor: '#666',
        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '600',
          marginBottom: 6,
        },
        headerStyle: {
          backgroundColor: '#1a1a2e',
        },
        headerTintColor: '#ffffff',
        headerTitleStyle: {
          fontWeight: '700',
          fontSize: 18,
        },
      }}
    >
      <Tabs.Screen
        name="count"
        options={{
          title: 'Count',
          tabBarIcon: () => <TabIcon label="📋" />,
        }}
      />
      <Tabs.Screen
        name="waste"
        options={{
          title: 'Waste',
          tabBarIcon: () => <TabIcon label="🗑️" />,
        }}
      />
      <Tabs.Screen
        name="receive"
        options={{
          title: 'Receive',
          tabBarIcon: () => <TabIcon label="📦" />,
        }}
      />
      <Tabs.Screen
        name="kds"
        options={{
          title: 'KDS',
          tabBarIcon: () => <TabIcon label="🖥️" />,
        }}
      />
      <Tabs.Screen
        name="clock"
        options={{
          title: 'Clock',
          tabBarIcon: () => <TabIcon label="⏱️" />,
        }}
      />
    </Tabs>
  );
}
