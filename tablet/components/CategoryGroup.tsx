import { useState, ReactNode } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';

interface CategoryGroupProps {
  category: string;
  children: ReactNode;
  count?: string;
}

export default function CategoryGroup({ category, children, count }: CategoryGroupProps) {
  const [expanded, setExpanded] = useState(true);

  return (
    <View style={styles.container}>
      <TouchableOpacity
        style={styles.header}
        onPress={() => setExpanded((prev) => !prev)}
        activeOpacity={0.75}
        accessibilityRole="button"
        accessibilityLabel={`${category} category, ${expanded ? 'collapse' : 'expand'}`}
      >
        <View style={styles.headerLeft}>
          <Text style={styles.chevron}>{expanded ? '▾' : '▸'}</Text>
          <Text style={styles.categoryName}>{category}</Text>
        </View>
        {count !== undefined && (
          <View style={styles.badge}>
            <Text style={styles.badgeText}>{count}</Text>
          </View>
        )}
      </TouchableOpacity>
      {expanded && <View style={styles.content}>{children}</View>}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginBottom: 4,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#16213e',
    paddingHorizontal: 16,
    paddingVertical: 12,
    minHeight: 48,
    borderLeftWidth: 3,
    borderLeftColor: '#e94560',
  },
  headerLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  chevron: {
    fontSize: 18,
    color: '#e94560',
    width: 20,
    textAlign: 'center',
  },
  categoryName: {
    fontSize: 16,
    fontWeight: '700',
    color: '#ffffff',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  badge: {
    backgroundColor: '#0f3460',
    borderRadius: 12,
    paddingHorizontal: 10,
    paddingVertical: 3,
  },
  badgeText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#aaaaaa',
  },
  content: {
    backgroundColor: '#1a1a2e',
  },
});
