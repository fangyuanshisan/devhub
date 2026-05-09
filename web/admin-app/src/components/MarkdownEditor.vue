<template>
  <div ref="editorEl" class="markdown-editor" />
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import Editor from '@toast-ui/editor';

const props = defineProps({ modelValue: { type: String, default: '' } });
const emit = defineEmits(['update:modelValue']);
const editorEl = ref();
let editor;

onMounted(() => {
  editor = new Editor({
    el: editorEl.value,
    height: '320px',
    initialEditType: 'markdown',
    previewStyle: 'vertical',
    initialValue: props.modelValue,
    events: {
      change: () => emit('update:modelValue', editor.getMarkdown()),
    },
  });
});

watch(
  () => props.modelValue,
  (value) => {
    if (editor && value !== editor.getMarkdown()) editor.setMarkdown(value || '');
  },
);

onBeforeUnmount(() => editor?.destroy());
</script>
