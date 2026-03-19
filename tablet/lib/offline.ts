import AsyncStorage from '@react-native-async-storage/async-storage';

export async function savePending(key: string, data: any[]): Promise<void> {
  await AsyncStorage.setItem(`pending_${key}`, JSON.stringify(data));
}

export async function loadPending(key: string): Promise<any[]> {
  const raw = await AsyncStorage.getItem(`pending_${key}`);
  return raw ? JSON.parse(raw) : [];
}

export async function clearPending(key: string): Promise<void> {
  await AsyncStorage.removeItem(`pending_${key}`);
}
