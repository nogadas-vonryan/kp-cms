import { usePluginStore, type ArchivistPlugin } from './pluginRegistry';
import { logger } from '@/core/utils/logger';

export interface PluginSource {
  id: string;
  factory: () => ArchivistPlugin;
}

/**
 * Registers built-in plugins. Extend this to pull from config or remote manifests.
 */
export function registerCorePlugins(sources: PluginSource[] = []) {
  const store = usePluginStore();
  sources.forEach((source) => {
    try {
      const plugin = source.factory();
      store.register(plugin);
      logger.info(`Plugin registered`, source.id);
    } catch (err) {
      logger.error(`Failed to register plugin: ${source.id}`, err);
    }
  });
}
