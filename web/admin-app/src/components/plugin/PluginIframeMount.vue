<template>
  <div ref="host" class="plugin-iframe-mount">
    <div v-if="error" class="empty-state">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';

type MountArea = 'admin' | 'frontend';

const props = defineProps<{
  pluginCode: string;
  area?: MountArea;
  communitySlug?: string;
  height?: string;
}>();

const host = ref<HTMLElement | null>(null);
const error = ref('');

let scriptLoading: Promise<void> | null = null;

function loadSharedHelper(): Promise<void> {
  if (typeof window === 'undefined') return Promise.resolve();
  const w = window as any;
  if (w.DevHubOfficialPluginMountHost) return Promise.resolve();
  if (scriptLoading) return scriptLoading;

  scriptLoading = new Promise<void>((resolve, reject) => {
    const existing = document.querySelector('script[data-devhub-official-plugin-host-helper]') as HTMLScriptElement | null;
    if (existing) {
      existing.addEventListener('load', () => resolve());
      existing.addEventListener('error', () => reject(new Error('load_failed')));
      return;
    }
    const script = document.createElement('script');
    script.src = '/plugins/assets/devhub-plugin-mount-host.js';
    script.defer = true;
    script.setAttribute('data-devhub-official-plugin-host-helper', '1');
    script.addEventListener('load', () => resolve());
    script.addEventListener('error', () => reject(new Error('load_failed')));
    document.head.appendChild(script);
  });
  return scriptLoading;
}

async function doMount() {
  if (!host.value) return;
  error.value = '';

  const h = host.value as HTMLElement & { __devhubPluginUnmount?: () => void };
  if (typeof h.__devhubPluginUnmount === 'function') {
    h.__devhubPluginUnmount();
  }
  h.innerHTML = '';

  h.dataset.devhubPluginMount = '1';
  h.dataset.pluginCode = props.pluginCode;
  h.dataset.area = props.area || 'admin';
  if (props.communitySlug) h.dataset.communitySlug = props.communitySlug;
  else delete h.dataset.communitySlug;
  if (props.height) h.style.height = props.height;

  try {
    await loadSharedHelper();
    const w = window as any;
    if (!w.DevHubOfficialPluginMountHost || typeof w.DevHubOfficialPluginMountHost.mount !== 'function') {
      throw new Error('helper_unavailable');
    }
    w.DevHubOfficialPluginMountHost.mount(h, { pluginCode: props.pluginCode, area: props.area || 'admin', communitySlug: props.communitySlug || '' });
  } catch (e: any) {
    error.value = '加载失败';
  }
}

onMounted(() => {
  void doMount();
});

watch(
  () => [props.pluginCode, props.communitySlug, props.area],
  () => {
    void doMount();
  },
);

onBeforeUnmount(() => {
  const h = host.value as (HTMLElement & { __devhubPluginUnmount?: () => void }) | null;
  if (typeof h?.__devhubPluginUnmount === 'function') {
    h.__devhubPluginUnmount();
  }
});
</script>

<style scoped>
.plugin-iframe-mount {
  width: 100%;
}
.empty-state {
  color: var(--el-text-color-secondary);
  padding: 10px 12px;
}
</style>
