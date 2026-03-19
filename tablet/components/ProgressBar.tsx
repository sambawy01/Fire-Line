import { View, Text, StyleSheet } from 'react-native';

interface ProgressBarProps {
  counted: number;
  total: number;
}

export default function ProgressBar({ counted, total }: ProgressBarProps) {
  const pct = total > 0 ? Math.min(counted / total, 1) : 0;
  const pctDisplay = Math.round(pct * 100);

  return (
    <View style={styles.container}>
      <View style={styles.track}>
        <View style={[styles.fill, { width: `${pctDisplay}%` }]} />
      </View>
      <Text style={styles.label}>
        {counted} of {total} items counted
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 6,
  },
  track: {
    height: 10,
    borderRadius: 5,
    backgroundColor: '#2a2a4a',
    overflow: 'hidden',
  },
  fill: {
    height: '100%',
    borderRadius: 5,
    backgroundColor: '#22c55e',
  },
  label: {
    fontSize: 13,
    color: '#aaaaaa',
    textAlign: 'center',
  },
});
