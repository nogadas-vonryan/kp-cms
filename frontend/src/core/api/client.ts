import axios from 'axios';
import { useAuthStore } from '@/modules/auth/store';
import { config } from '@/core/config';
import { logger } from '@/core/utils/logger';
import { createAppError } from './errorHandler';
import type { Router } from 'vue-router';

let router: Router | null = null;

export function setApiRouter(r: Router) {
  router = r;
}

const api = axios.create({
  baseURL: config.apiUrl,
  timeout: config.apiTimeout,
  withCredentials: true, // Enable sending session cookies
});

// Request Interceptor: Auto-attach CSRF Token
api.interceptors.request.use((config) => {
  const authStore = useAuthStore();
  if (authStore.csrfToken) {
    config.headers['X-CSRF-Token'] = authStore.csrfToken;
  }
  return config;
});

// Response Interceptor: Auto-logout on 401 + error logging
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const appError = createAppError(error);
    logger.error('API request failed', appError);
    
    // Handle 401 Unauthorized - redirect to login
    if (error.response?.status === 401) {
      const authStore = useAuthStore();
      authStore.logout();
      if (router && router.currentRoute.value.meta.requiresAuth) {
        router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } });
      }
    }
    
    return Promise.reject(error);
  }
);

export default api;
