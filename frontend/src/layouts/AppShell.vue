<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { AuthService } from '@/modules/auth/services/authService';
import { computed, ref } from 'vue';
import { Menu, X, FileText, BarChart3, Settings, LogOut } from 'lucide-vue-next';

const router = useRouter();
const authStore = useAuthStore();
const sidebarOpen = ref(false);

const isAdmin = computed(() => authStore.role === 'RoleAdmin');
const username = computed(() => authStore.user?.id || '');

async function logout() {
  await AuthService.logout();
  router.push('/login');
}

function closeSidebar() {
  sidebarOpen.value = false;
}

function navigateTo(path: string) {
  router.push(path);
  closeSidebar();
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="sticky top-0 z-40 bg-white border-b border-gray-200">
      <div class="px-4 py-4 flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900">KPCMS</h1>
        <div class="flex items-center gap-4">
          <span class="hidden sm:block text-sm text-gray-600">{{ username }}</span>
          <button
            @click="sidebarOpen = !sidebarOpen"
            class="sm:hidden p-2 hover:bg-gray-100 rounded-lg transition"
          >
            <Menu v-if="!sidebarOpen" :size="24" class="text-gray-900" />
            <X v-else :size="24" class="text-gray-900" />
          </button>
        </div>
      </div>
    </header>

    <!-- Content Container -->
    <div class="flex">
      <!-- Sidebar -->
      <aside
        class="fixed top-16 left-0 bottom-0 z-30 w-64 bg-white border-r border-gray-200 transform transition-transform duration-200 flex flex-col sm:sticky sm:top-16 sm:translate-x-0 sm:h-[calc(100vh-4rem)] overflow-hidden"
        :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <nav class="flex-1 pt-4 pb-4 overflow-y-auto">
          <div class="px-4 space-y-2">
            <!-- Documents Link -->
            <button
              @click="navigateTo('/documents')"
              class="w-full flex items-center gap-3 px-4 py-3 text-gray-700 hover:bg-gray-100 rounded-lg transition"
              :class="$route.path === '/documents' ? 'bg-blue-50 text-blue-600' : ''"
            >
              <FileText :size="20" />
              <span class="text-sm font-medium">Documents</span>
            </button>

            <!-- Reports Link (Admin only) -->
            <button
              v-if="isAdmin"
              @click="navigateTo('/reports/export')"
              class="w-full flex items-center gap-3 px-4 py-3 text-gray-700 hover:bg-gray-100 rounded-lg transition"
              :class="$route.path.startsWith('/reports') ? 'bg-blue-50 text-blue-600' : ''"
            >
              <BarChart3 :size="20" />
              <span class="text-sm font-medium">Reports</span>
            </button>

            <!-- Admin Link (Admin only) -->
            <button
              v-if="isAdmin"
              @click="navigateTo('/admin')"
              class="w-full flex items-center gap-3 px-4 py-3 text-gray-700 hover:bg-gray-100 rounded-lg transition"
              :class="$route.path === '/admin' ? 'bg-blue-50 text-blue-600' : ''"
            >
              <Settings :size="20" />
              <span class="text-sm font-medium">Admin</span>
            </button>
          </div>
        </nav>

        <!-- Logout Button -->
        <div class="shrink-0 px-4 py-4 border-t border-gray-200">
          <button
            @click="logout"
            class="w-full flex items-center gap-3 px-4 py-3 text-gray-700 hover:bg-gray-100 rounded-lg transition"
          >
            <LogOut :size="20" />
            <span class="text-sm font-medium">Logout</span>
          </button>
        </div>
      </aside>

      <!-- Overlay (Mobile) -->
      <div
        v-if="sidebarOpen"
        @click="closeSidebar"
        class="fixed inset-0 z-20 bg-black/50 sm:hidden"
      />

      <!-- Main Content -->
      <main class="flex-1 min-h-[calc(100vh-4rem)] px-4 py-6 sm:max-w-5xl sm:mx-auto sm:w-full">
        <RouterView />
      </main>
    </div>
  </div>
</template>
