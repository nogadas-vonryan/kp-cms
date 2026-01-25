import { defineStore } from 'pinia';
import { shallowRef, type Component } from 'vue';
import type { Document, PermissionAction } from '@/types';

export interface PluginLocations {
  documentTab?: Component;
  documentCreate?: Component;
  sidebarItem?: Component;
  staffTab?: Component;
  schedulingTab?: Component;
}

export interface PluginContext {
  document?: Document;
  isAdmin: boolean;
  events: {
    onDocumentUpdate?: (doc: Document) => void;
    onFileAdd?: (file: any) => void;
    onError?: (error: string) => void;
  };
}

export interface ArchivistPlugin {
  id: string;
  name: string;
  version?: string;
  locations: PluginLocations;
  permissions?: PermissionAction[];
  metadata?: {
    description?: string;
    icon?: string;
    category?: string;
  };
}

export const usePluginStore = defineStore('plugins', () => {
  const plugins = shallowRef<ArchivistPlugin[]>([]);

  function register(plugin: ArchivistPlugin) {
    const exists = plugins.value.some((entry) => entry.id === plugin.id);
    if (exists) {
      plugins.value = plugins.value.map((entry) =>
        entry.id === plugin.id ? plugin : entry
      );
      return;
    }
    plugins.value.push(plugin);
  }

  function unregister(id: string) {
    plugins.value = plugins.value.filter((plugin) => plugin.id !== id);
  }

  return { plugins, register, unregister };
});
