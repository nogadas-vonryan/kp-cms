import { defineStore } from 'pinia';
import { shallowRef, type Component } from 'vue';

export interface PluginLocations {
  documentTab?: Component;
  sidebarItem?: Component;
}

export interface ArchivistPlugin {
  id: string;
  name: string;
  locations: PluginLocations;
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
