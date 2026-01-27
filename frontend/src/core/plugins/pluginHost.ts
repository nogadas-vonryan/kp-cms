import { defineComponent, h, computed, type PropType } from 'vue';
import type { ArchivistPlugin, PluginContext } from './pluginRegistry';

/**
 * PluginHost renders registered plugins at specific locations with context
 * Usage:
 * <PluginHost :plugins="plugins" location="documentTab" :context="pluginContext" />
 */
export default defineComponent({
  name: 'PluginHost',
  props: {
    plugins: {
      type: Array as PropType<ArchivistPlugin[]>,
      required: true
    },
    location: {
      type: String as PropType<keyof ArchivistPlugin['locations']>,
      required: true
    },
    context: {
      type: Object as PropType<PluginContext>,
      required: true
    },
    class: {
      type: String,
      default: ''
    }
  },
  setup(props) {
    const activePlugins = computed(() => {
      return props.plugins.filter(plugin => {
        const component = plugin.locations[props.location];
        return !!component;
      });
    });

    return () => {
      if (activePlugins.value.length === 0) {
        return null;
      }

      return h(
        'div',
        {
          class: `plugin-host ${props.class}`,
          'data-location': props.location
        },
        activePlugins.value.map(plugin => {
          const component = plugin.locations[props.location];
          if (!component) return null;

          return h(component, {
            key: plugin.id,
            plugin,
            context: props.context,
            class: `plugin-${plugin.id}`
          });
        })
      );
    };
  }
});
