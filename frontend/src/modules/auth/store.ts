import { defineStore } from 'pinia';
import type { AuthUser, OAuthConnection } from '@/types';
import { logger } from '@/core/utils/logger';

const CSRF_TOKEN_KEY = 'archivist-csrf-token';
const USER_KEY = 'archivist-user';
const OAUTH_CONNECTIONS_KEY = 'archivist-oauth-connections';

interface AuthState {
  csrfToken: string | null;
  user: AuthUser | null;
  oauthConnections: OAuthConnection[];
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
    oauthConnections: readFromStorage<OAuthConnection[]>(OAUTH_CONNECTIONS_KEY) ?? [],
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.csrfToken && state.user),
    role: (state) => state.user?.role ?? null,
    username: (state) => state.user?.id ?? null,
    isGoogleConnected: (state) => {
      // Ensure oauthConnections is always treated as an array
      const connections = state.oauthConnections ?? [];
      return connections.some(
        conn => conn.provider === 'google' && conn.is_connected
      );
    },
    hasCalendarScope: (state) => {
      // Ensure oauthConnections is always treated as an array
      const connections = state.oauthConnections ?? [];
      const googleConn = connections.find(
        conn => conn.provider === 'google' && conn.is_connected
      );
      if (!googleConn) return false;
      return googleConn.scopes?.some(scope => scope.includes('calendar')) ?? false;
    },
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
      this.setOAuthConnections([]);
    },
    setOAuthConnections(connections: OAuthConnection[] | null) {
      // Ensure we always have an array, not null
      this.oauthConnections = connections ?? [];
      if (typeof localStorage === 'undefined') return;
      if (this.oauthConnections.length > 0) {
        localStorage.setItem(OAUTH_CONNECTIONS_KEY, JSON.stringify(this.oauthConnections));
      } else {
        localStorage.removeItem(OAUTH_CONNECTIONS_KEY);
      }
    },
    addOAuthConnection(connection: OAuthConnection) {
      const existingIndex = this.oauthConnections.findIndex(
        conn => conn.provider === connection.provider
      );
      if (existingIndex >= 0) {
        this.oauthConnections[existingIndex] = connection;
      } else {
        this.oauthConnections.push(connection);
      }
      this.setOAuthConnections([...this.oauthConnections]);
    },
    removeOAuthConnection(provider: string) {
      this.oauthConnections = this.oauthConnections.filter(
        conn => conn.provider !== provider
      );
      this.setOAuthConnections([...this.oauthConnections]);
    },
  },
});
