import api from '@/core/api/client';
import { useAuthStore } from '@/modules/auth/store';
import type { AuthResponse } from '@/types';

export interface LoginRequest {
  email: string;
  password: string;
}

export const AuthService = {
  async login(payload: LoginRequest) {
    const response = await api.post<AuthResponse>('/auth/login', payload);
    const authStore = useAuthStore();
    authStore.setAuth(response.data);
    return response.data;
  },

  async refresh() {
    const response = await api.post<AuthResponse>('/auth/refresh');
    const authStore = useAuthStore();
    authStore.setAuth(response.data);
    return response.data;
  },

  logout() {
    const authStore = useAuthStore();
    authStore.logout();
  },
};
