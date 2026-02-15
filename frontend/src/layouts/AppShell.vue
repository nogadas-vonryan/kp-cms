<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { AuthService } from '@/modules/auth/services/authService';
import { computed, ref } from 'vue';
import { Menu, X, FileText, BarChart3, Settings, LogOut, Upload, Save, Users, Calendar, Computer } from 'lucide-vue-next';

const router = useRouter();
const authStore = useAuthStore();
const sidebarOpen = ref(false);

const isAdmin = computed(() => authStore.role === 'RoleAdmin');
const username = computed(() => authStore.user?.id || '');

async function logout() {
  await AuthService.logout();
  router.push('/login');
}

function navigateTo(path: string) {
  router.push(path);
  sidebarOpen.value = false;
}
</script>

<template>
  <div class="flex h-screen bg-gray-50">
    <!-- Sidebar Drawer -->
    <aside
      class="fixed inset-y-0 left-0 z-30 w-64 transform transition-transform duration-300 ease-out flex flex-col shadow-xl sm:static sm:z-auto sm:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Background with image and overlay (desktop only) -->
      <div class="absolute inset-0 hidden sm:block">
        <div class="absolute inset-0 bg-cover bg-center" style="background-image: url('/cover_image.jpg')"></div>
        <div class="absolute inset-0 bg-linear-to-b from-green-700/95 to-green-800/95" />
      </div>

      <!-- Mobile background (no image) -->
      <div class="absolute inset-0 sm:hidden bg-linear-to-b from-green-700 to-green-800" />
      <!-- Sidebar Header with Logo -->
      <div class="shrink-0 px-4 py-6 border-b border-green-600/30 relative z-10">
        <div class="flex items-center gap-4">
          <img 
            src="/brgy_logo.png" 
            alt="Barangay Logo" 
            class="w-12 h-12 rounded-full object-cover shrink-0"
          />
          <div class="min-w-0">
            <h2 class="text-xs font-bold text-white truncate">Katarungang Pambarangay</h2>
            <p class="text-[10px] text-green-100 truncate">Barangay Bagumbayan, Taguig City</p>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 overflow-y-auto py-6 px-3 relative z-10">
        <div class="space-y-1">
          <!-- Documents Link -->
          <button
            @click="navigateTo('/documents')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/documents' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <FileText :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Documents</span>
          </button>

          <!-- Inhabitants Link -->
          <button
            @click="navigateTo('/inhabitants')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/inhabitants' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <Users :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Inhabitants</span>
          </button>

          <!-- Calendar Link -->
          <button
            @click="navigateTo('/calendar')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/calendar' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <Calendar :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Calendar</span>
          </button>

          <!-- Reports Link (Admin only) -->
          <button
            v-if="isAdmin"
            @click="navigateTo('/reports/export')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path.startsWith('/reports') ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <BarChart3 :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Reports</span>
          </button>


          <button
            v-if="isAdmin"
            @click="navigateTo('/import')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/import' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <Upload :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Import</span>
          </button>

          <button
            v-if="isAdmin"
            @click="navigateTo('/backup')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/backup' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <Save :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Backup</span>
          </button>

          <!-- Admin Link (Admin only) -->
          <button
            v-if="isAdmin"
            @click="navigateTo('/admin')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/admin' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <Computer :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Admin</span>
          </button>

          <button
            @click="navigateTo('/settings')"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-green-600 transition-all duration-200 group"
            :class="$route.path === '/settings' ? 'bg-amber-500 text-white shadow-md' : ''"
          >
            <Settings :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
            <span class="text-sm font-medium">Settings</span>
          </button>
        </div>
      </nav>

      <!-- Logout Section -->
      <div class="shrink-0 border-t border-green-600/30 px-3 py-4 relative z-10">
        <button
          @click="logout"
          class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-green-100 hover:bg-red-600/20 hover:text-red-200 transition-all duration-200 group"
        >
          <LogOut :size="18" class="shrink-0 group-hover:scale-110 transition-transform duration-200" />
          <span class="text-sm font-medium">Logout</span>
        </button>
      </div>
    </aside>

    <!-- Overlay (Mobile Only) -->
    <transition name="fade">
      <div
        v-if="sidebarOpen"
        @click="sidebarOpen = false"
        class="fixed inset-0 z-20 bg-black/40 backdrop-blur-sm sm:hidden"
      />
    </transition>

    <!-- Main Content -->
    <main class="flex-1 overflow-y-auto">
      <div class="h-full flex flex-col">
        <!-- Header -->
        <header class="shrink-0 bg-white border-b border-gray-200 shadow-sm">
          <div class="h-14 px-4 sm:px-6 flex items-center justify-between">
            <!-- Mobile Menu Button -->
            <button
              @click="sidebarOpen = !sidebarOpen"
              aria-label="Toggle menu"
              class="sm:hidden p-2 -ml-2 hover:bg-gray-100 rounded-lg transition-colors duration-200"
            >
              <Menu v-if="!sidebarOpen" :size="20" class="text-gray-900" />
              <X v-else :size="20" class="text-gray-900" />
            </button>

            <!-- User Info -->
            <div class="flex items-center gap-4 ml-auto">
              <span class="text-sm text-gray-600 font-medium">{{ username }}</span>
            </div>
          </div>
        </header>

        <!-- Page Content -->
        <div class="flex-1 overflow-y-auto p-4 sm:p-8">
          <RouterView />
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 200ms ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
