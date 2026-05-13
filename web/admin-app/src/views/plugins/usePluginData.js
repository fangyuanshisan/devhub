import { ref } from 'vue';
import { pluginHealthSummary, plugins as fetchPlugins } from '@/api/admin';

export function usePluginData() {
  const items = ref([]);
  const healthSummary = ref({});
  const loading = ref(false);
  const error = ref('');

  async function load() {
    loading.value = true;
    error.value = '';
    try {
      const [list, health] = await Promise.all([
        fetchPlugins(),
        pluginHealthSummary().catch(() => null),
      ]);
      items.value = list?.items || [];
      healthSummary.value = health?.summary || {};
    } catch (e) {
      items.value = [];
      healthSummary.value = {};
      error.value = String(e?.message || e || '加载失败');
    } finally {
      loading.value = false;
    }
  }

  return {
    items,
    healthSummary,
    loading,
    error,
    load,
  };
}

