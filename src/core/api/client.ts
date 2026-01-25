import axios from 'axios';
import { useAuthStore } from '@/modules/auth/store';
import { config } from '@/core/config';
import { logger } from '@/core/utils/logger';
import { createAppError } from './errorHandler';

const api = axios.create({
  baseURL: config.apiUrl,
  timeout: config.apiTimeout,
});

// Request Interceptor: Auto-attach Token
api.interceptors.request.use((config) => {
  const authStore = useAuthStore();
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`;
  }
  return config;
});

// Response Interceptor: Auto-logout on 401 + error logging
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const appError = createAppError(error);
    logger.error('API request failed', appError);
    
    if (error.response?.status === 401) {
      const authStore = useAuthStore();
      authStore.logout(); // Redirect to login
    }
    return Promise.reject(error);
  }
);

export default api;
