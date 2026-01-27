import api from '@/core/api/client';
import { useAuthStore } from '@/modules/auth/store';
import type { AuthResponse, LoginRequest, RegisterRequest, ServerAuthResponse } from '@/types';

function mapAuthResponse(raw: ServerAuthResponse): AuthResponse {
  return {
    csrf_token: raw.csrf_token,
    user: {
      id: raw.user.ID,
      role: raw.user.Role === 'admin' || raw.user.Role === 'RoleAdmin' ? 'RoleAdmin' : 'RoleUser',
    },
  };
}

export const AuthService = {
  async register(payload: RegisterRequest) {
    const response = await api.post<ServerAuthResponse>('/api/auth/register', payload);
    const authStore = useAuthStore();
    const mapped = mapAuthResponse(response.data);
    authStore.setAuth(mapped);
    return mapped;
  },

  async login(payload: LoginRequest) {
    const response = await api.post<ServerAuthResponse>('/api/auth/login', payload);
    const authStore = useAuthStore();
    const mapped = mapAuthResponse(response.data);
    authStore.setAuth(mapped);
    return mapped;
  },

  async getCurrentUser() {
    const response = await api.get<ServerAuthResponse>('/api/auth/me');
    return mapAuthResponse(response.data);
  },

  async logout() {
    try {
      await api.post('/api/auth/logout');
    } finally {
      const authStore = useAuthStore();
      authStore.logout();
    }
  },
};
