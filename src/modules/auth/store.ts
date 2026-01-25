import { defineStore } from 'pinia';
import type { AuthUser } from '@/types';
import { logger } from '@/core/utils/logger';

const CSRF_TOKEN_KEY = 'archivist-csrf-token';
const USER_KEY = 'archivist-user';

interface AuthState {
  csrfToken: string | null;
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
    csrfToken: typeof localStorage === 'undefined' ? null : localStorage.getItem(CSRF_TOKEN_KEY),
    user: readFromStorage<AuthUser>(USER_KEY),
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.csrfToken && state.user),
    role: (state) => state.user?.role ?? null,
    username: (state) => state.user?.id ?? null,
  },
  actions: {
    setAuth(payload: { csrf_token: string; user: AuthUser }) {
      this.setCsrfToken(payload.csrf_token);
      this.setUser(payload.user);
    },
    setCsrfToken(token: string | null) {
      this.csrfToken = token;
      if (typeof localStorage === 'undefined') return;
      if (token) {
        localStorage.setItem(CSRF_TOKEN_KEY, token);
      } else {
        localStorage.removeItem(CSRF_TOKEN_KEY);
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
      this.setCsrfToken(null);
      this.setUser(null);
    },
  },
});
