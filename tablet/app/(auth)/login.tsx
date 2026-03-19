import { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from 'react-native';
import { useAuthStore } from '../../stores/auth';

export default function LoginScreen() {
  const { login, locations, selectLocation } = useAuthStore();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    if (!email.trim() || !password.trim()) {
      setError('Email and password are required.');
      return;
    }

    setError(null);
    setLoading(true);

    try {
      await login(email.trim(), password);
      // After login, if only 1 location auto-select it
      // locations is updated by the store; read from store after login
      const updatedLocations = useAuthStore.getState().locations;
      if (updatedLocations.length === 1) {
        selectLocation(updatedLocations[0].id, updatedLocations[0].name);
      }
      // Navigation handled by _layout.tsx
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : 'Login failed. Please try again.';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.flex}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView
        contentContainerStyle={styles.container}
        keyboardShouldPersistTaps="handled"
      >
        {/* Logo / Title */}
        <View style={styles.logoContainer}>
          <Text style={styles.logoText}>🔥 FireLine</Text>
          <Text style={styles.tagline}>Restaurant Operations</Text>
        </View>

        {/* Form */}
        <View style={styles.card}>
          <Text style={styles.cardTitle}>Manager Login</Text>

          <Text style={styles.label}>Email</Text>
          <TextInput
            style={styles.input}
            value={email}
            onChangeText={setEmail}
            placeholder="manager@example.com"
            placeholderTextColor="#666"
            autoCapitalize="none"
            autoCorrect={false}
            keyboardType="email-address"
            textContentType="emailAddress"
            editable={!loading}
          />

          <Text style={styles.label}>Password</Text>
          <TextInput
            style={styles.input}
            value={password}
            onChangeText={setPassword}
            placeholder="••••••••"
            placeholderTextColor="#666"
            secureTextEntry
            textContentType="password"
            editable={!loading}
            onSubmitEditing={handleLogin}
            returnKeyType="done"
          />

          {error ? (
            <View style={styles.errorContainer}>
              <Text style={styles.errorText}>{error}</Text>
            </View>
          ) : null}

          <TouchableOpacity
            style={[styles.loginButton, loading && styles.loginButtonDisabled]}
            onPress={handleLogin}
            disabled={loading}
            activeOpacity={0.8}
          >
            {loading ? (
              <ActivityIndicator color="#ffffff" />
            ) : (
              <Text style={styles.loginButtonText}>Login</Text>
            )}
          </TouchableOpacity>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  container: {
    flexGrow: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 24,
    paddingVertical: 48,
  },
  logoContainer: {
    alignItems: 'center',
    marginBottom: 40,
  },
  logoText: {
    fontSize: 40,
    fontWeight: '800',
    color: '#e94560',
    letterSpacing: 1,
  },
  tagline: {
    fontSize: 16,
    color: '#aaaaaa',
    marginTop: 6,
  },
  card: {
    backgroundColor: '#16213e',
    borderRadius: 16,
    padding: 28,
    width: '100%',
    maxWidth: 420,
    borderWidth: 1,
    borderColor: '#0f3460',
  },
  cardTitle: {
    fontSize: 22,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 24,
    textAlign: 'center',
  },
  label: {
    fontSize: 14,
    color: '#aaaaaa',
    marginBottom: 6,
    marginTop: 4,
  },
  input: {
    backgroundColor: '#0d0d1a',
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#0f3460',
    color: '#ffffff',
    fontSize: 20,
    paddingHorizontal: 16,
    paddingVertical: 14,
    marginBottom: 16,
  },
  errorContainer: {
    backgroundColor: '#3d0c11',
    borderRadius: 8,
    padding: 12,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#e94560',
  },
  errorText: {
    color: '#e94560',
    fontSize: 15,
    textAlign: 'center',
  },
  loginButton: {
    backgroundColor: '#e94560',
    borderRadius: 10,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 4,
  },
  loginButtonDisabled: {
    opacity: 0.6,
  },
  loginButtonText: {
    color: '#ffffff',
    fontSize: 18,
    fontWeight: '700',
  },
});
