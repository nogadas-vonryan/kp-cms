This is a comprehensive architectural plan for a **Vue 3 + TypeScript + Vite** frontend tailored to your `Archivist Server` API.

This plan prioritizes **modularity** (so you can rip out or add features without breaking the core), **security** (matching your API's strict role requirements), and **extensibility** (preparing for those future plugins like PDF editing).

---

### 1. High-Level Architecture

We will use a **Feature-Based Architecture**. Instead of grouping files by type (e.g., all views in one folder, all components in another), we group them by **Domain** (e.g., `Auth`, `Documents`, `System`). This makes the codebase "modular by default."

**Core Stack:**

* **Framework:** Vue 3 (Script Setup)
* **Build Tool:** Vite
* **Language:** TypeScript (Strict typing is crucial for a stable plugin system)
* **State:** Pinia (Modular stores)
* **Networking:** Axios (with interceptors for your AuthMiddleware)

---

### 2. Directory Structure

This structure separates the "Core" application from the "Modules" (features). Future plugins (Calendar, Wacom) will simply be new folders in `modules/`.

```text
src/
├── core/                   # The "Brain" of the app
│   ├── api/                # Axios instance & interceptors
│   ├── auth/               # RBAC logic, Role guards, user state
│   ├── plugins/            # The "Plugin Loader" logic (for future)
│   └── ui/                 # Base UI components (Buttons, Inputs, Modals)
├── modules/                # Feature Modules
│   ├── auth/               # Login screens, Profile views
│   ├── documents/          # Document list, file upload, detailed view
│   └── admin/              # Conflict resolution, server reload tools
├── layouts/                # App shells (SidebarLayout, EmptyLayout)
├── types/                  # Shared TypeScript interfaces (Document, etc.)
└── main.ts

```

---

### 3. The "Core" Layer (Security & API)

Since your API requires strict Auth headers and Role checks, we centralize this logic so you never accidentally make an insecurity call.

#### A. The API Client (Secure wrapper)

This wrapper automatically injects the token and handles 401/403 errors globally.

```typescript
// src/core/api/client.ts
import axios from 'axios';
import { useAuthStore } from '@/modules/auth/store';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL, // e.g., http://localhost:8080
});

// Request Interceptor: Auto-attach Token
api.interceptors.request.use((config) => {
  const authStore = useAuthStore();
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`;
  }
  return config;
});

// Response Interceptor: Auto-logout on 401
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore();
      authStore.logout(); // Redirect to login
    }
    return Promise.reject(error);
  }
);

export default api;

```

#### B. Access Control (RBAC) Composable

Instead of checking `user.role === 'admin'` everywhere, use a semantic composable. This makes it easy to add "Editors" or "Viewers" later.

```typescript
// src/core/auth/usePermission.ts
import { useAuthStore } from '@/modules/auth/store';

export function usePermission() {
  const auth = useAuthStore();

  function can(action: 'create' | 'edit' | 'delete' | 'admin_tools') {
    if (!auth.user) return false;
    
    // Admin has access to everything
    if (auth.user.role === 'RoleAdmin') return true;

    // Standard user restrictions
    switch (action) {
      case 'create': return false; // Per your API, only Admin posts documents
      case 'delete': return false;
      default: return true; // View is public/standard
    }
  }

  return { can };
}

```

**Usage in Template:**

```html
<button v-if="can('delete')" @click="deleteDoc">Delete</button>

```

---

### 4. Modular Feature Design (The "Document" Module)

Every module should be self-contained. The `documents` module will handle the logic for `/documents`.

**File: `src/modules/documents/services/documentService.ts**`
(Keep the API logic separate from the Vue components)

```typescript
import api from '@/core/api/client';
import type { Document, CreateDocumentRequest } from '@/types';

export const DocumentService = {
  getAll: (offset = 0, limit = 15) => 
    api.get<Document[]>('/documents', { params: { offset, limit } }),

  getOne: (uuid: string) => 
    api.get<Document>(`/documents/${uuid}`),

  // Only exposes admin methods if generic type allows, but API enforces security
  create: (data: CreateDocumentRequest) => 
    api.post<Document>('/documents', data),
    
  uploadFile: (uuid: string, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return api.post(`/documents/${uuid}/files`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
  }
};

```

---

### 5. Future-Proofing: The Plugin System

You mentioned PDF editing, Wacom integration, etc. To prevent your app from becoming a monolith, we design a **Slot-Based Plugin System**.

1. **Define a Plugin Interface:**
Create a standard way for a plugin to "register" itself (e.g., adding a tab to the Document Detail view).
2. **The Registry Store:**
A Pinia store that holds a list of active plugins.
3. **Dynamic Component Injection:**
In your `DocumentDetail.vue`, you leave "Slots" open for plugins to fill.

**Example Implementation:**

```typescript
// src/core/plugins/pluginRegistry.ts
import { defineStore } from 'pinia';
import { shallowRef, type Component } from 'vue';

export interface ArchivistPlugin {
  id: string;
  name: string;
  // Where does this plugin inject UI?
  locations: {
    documentTab?: Component; // e.g., "PDF Editor" tab
    sidebarItem?: Component; // e.g., "Calendar" link
  };
}

export const usePluginStore = defineStore('plugins', () => {
  const plugins = shallowRef<ArchivistPlugin[]>([]);

  function register(plugin: ArchivistPlugin) {
    plugins.value.push(plugin);
  }

  return { plugins, register };
});

```

**Future usage (e.g., adding PDF Editor):**
You wouldn't touch the core `DocumentDetail.vue` code. You would just `register()` the PDF plugin, and the UI automatically renders the new tab because `DocumentDetail.vue` iterates over `pluginStore.plugins`.

---

### 6. Security Checklist for Frontend

Since you have a public endpoint (`GET /`) and protected endpoints, follow these rules:

1. **Route Guards:** In `vue-router`, check strictly for `meta.requiresAuth` and `meta.requiresAdmin`.
2. **Sanitize Filenames:** When displaying `files` from your API, never trust the `file_name` to be safe HTML. Always escape it (Vue does this by default with `{{ }}`).
3. **Memory Management:** For "Video Streaming" or "PDF Editing" later, use `URL.createObjectURL()` carefully and **always** `revokeObjectURL()` on `onUnmounted` to prevent memory leaks in the browser.