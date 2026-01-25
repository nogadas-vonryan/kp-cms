import { defineStore } from 'pinia';

export type ToastLevel = 'info' | 'success' | 'warning' | 'error';

export interface ToastMessage {
  id: string;
  message: string;
  level: ToastLevel;
  timeoutMs?: number;
}

interface SystemState {
  isLoading: boolean;
  toasts: ToastMessage[];
}

export const useSystemStore = defineStore('system', {
  state: (): SystemState => ({
    isLoading: false,
    toasts: [],
  }),
  actions: {
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading;
    },
    pushToast(toast: ToastMessage) {
      const existing = this.toasts.find((t) => t.id === toast.id);
      if (existing) {
        this.toasts = this.toasts.map((t) => (t.id === toast.id ? toast : t));
        return;
      }
      this.toasts.push(toast);
    },
    dismissToast(id: string) {
      this.toasts = this.toasts.filter((toast) => toast.id !== id);
    },
    clearToasts() {
      this.toasts = [];
    },
  },
});
