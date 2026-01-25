import { defineStore } from 'pinia';
import type { AuthUser } from '@/types';
import { logger } from '@/core/utils/logger';

const TOKEN_KEY = 'archivist-token';
const USER_KEY = 'archivist-user';

interface AuthState {
  token: string | null;
  user: AuthUser | null;
}

function readFromStorage<T>(key: string): T | null {
  if (typeof localStorage === 'undefined') return null;
  const raw = localStorage.getItem(key);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as T;
  } catch (error) {
    logger.warn('Failed to parse stored auth payload', error);
    return null;
  }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: typeof localStorage === 'undefined' ? null : localStorage.getItem(TOKEN_KEY),
    user: readFromStorage<AuthUser>(USER_KEY),
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token && state.user),
    role: (state) => state.user?.role ?? null,
  },
  actions: {
    setAuth(payload: { token: string; user: AuthUser }) {
      this.setToken(payload.token);
      this.setUser(payload.user);
    },
    setToken(token: string | null) {
      this.token = token;
      if (typeof localStorage === 'undefined') return;
      if (token) {
        localStorage.setItem(TOKEN_KEY, token);
      } else {
        localStorage.removeItem(TOKEN_KEY);
      }
    },
    setUser(user: AuthUser | null) {
      this.user = user;
      if (typeof localStorage === 'undefined') return;
      if (user) {
        localStorage.setItem(USER_KEY, JSON.stringify(user));
      } else {
        localStorage.removeItem(USER_KEY);
      }
    },
    logout() {
      this.setToken(null);
      this.setUser(null);
    },
  },
});
