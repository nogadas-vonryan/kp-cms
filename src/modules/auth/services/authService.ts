import api from '@/core/api/client';
import { useAuthStore } from '@/modules/auth/store';
import type { AuthResponse, LoginRequest, RegisterRequest } from '@/types';

export const AuthService = {
  async register(payload: RegisterRequest) {
    const response = await api.post<AuthResponse>('/auth/register', payload);
    const authStore = useAuthStore();
    authStore.setAuth(response.data);
    return response.data;
  },

  async login(payload: LoginRequest) {
    const response = await api.post<AuthResponse>('/auth/login', payload);
    const authStore = useAuthStore();
    authStore.setAuth(response.data);
    return response.data;
  },

  async getCurrentUser() {
    const response = await api.get<AuthResponse>('/auth/me');
    return response.data;
  },

  async logout() {
    try {
      await api.post('/auth/logout');
    } finally {
      const authStore = useAuthStore();
      authStore.logout();
    }
  },
};
