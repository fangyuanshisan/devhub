<template>
  <el-collapse class="admin-technical-details" :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)" :data-testid="testid">
    <el-collapse-item v-for="block in safeBlocks" :key="block.name || block.title" :name="block.name || block.title" :title="block.title || '技术详情'">
      <pre class="admin-technical-details__pre">{{ formatJSON(block.value) }}</pre>
    </el-collapse-item>
  </el-collapse>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  blocks: { type: Array, default: () => [] },
  modelValue: { type: [String, Array], default: () => [] },
  redact: { type: Boolean, default: true },
  testid: { type: String, default: '' },
});
defineEmits(['update:modelValue']);

const safeBlocks = computed(() => props.blocks.filter((block) => block && block.value !== undefined && block.value !== null));

function formatJSON(value) {
  const safeValue = props.redact ? redactSensitive(value) : value;
  if (typeof safeValue === 'string') return safeValue;
  try {
    return JSON.stringify(safeValue ?? {}, null, 2);
  } catch {
    return String(safeValue ?? '');
  }
}

function redactSensitive(value, depth = 0) {
  if (depth > 8) return '[REDACTED_DEPTH]';
  if (Array.isArray(value)) return value.map((item) => redactSensitive(item, depth + 1));
  if (!value || typeof value !== 'object') return value;
  return Object.fromEntries(Object.entries(value).map(([key, raw]) => {
    const normalized = String(key || '').toLowerCase();
    const isSafeRef = normalized.endsWith('_ref') || normalized === 'ref' || normalized.endsWith('_status') || normalized.endsWith('_key_id') || normalized.endsWith('_masked') || normalized === 'masked_value';
    const sensitive = !isSafeRef && (
      normalized.includes('encrypted_value')
      || normalized.includes('authorization')
      || normalized.includes('secret')
      || normalized.includes('token')
      || normalized.includes('password')
      || normalized.includes('ciphertext')
      || normalized.includes('hash')
    );
    return [key, sensitive ? '[REDACTED]' : redactSensitive(raw, depth + 1)];
  }));
}
</script>
