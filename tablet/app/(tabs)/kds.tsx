import { View, Text, StyleSheet } from 'react-native';

export default function KdsScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.icon}>🖥️</Text>
      <Text style={styles.title}>KDS</Text>
      <Text style={styles.subtitle}>Coming Soon</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  icon: {
    fontSize: 48,
    marginBottom: 8,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: '#ffffff',
  },
  subtitle: {
    fontSize: 18,
    color: '#aaaaaa',
  },
});
