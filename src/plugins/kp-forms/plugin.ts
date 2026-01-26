import KPFormsTab from './KPFormsTab.vue';
import type { ArchivistPlugin } from '@/core/plugins/pluginRegistry';

export function kpFormsPlugin(): ArchivistPlugin {
  return {
    id: 'kpForms',
    name: 'KP Forms',
    version: '1.0.0',
    locations: {
      kpFormsTab: KPFormsTab
    },
    metadata: {
      description: 'Generate Katarungang Pambarangay PDF forms directly from a document.',
      icon: '📄',
      category: 'Forms'
    }
  };
}
