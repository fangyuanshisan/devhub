<template>
  <section class="plugin-page" data-testid="plugins-legacy-page">
    <el-alert
      title="插件中心已按功能分页"
      type="info"
      show-icon
      :closable="false"
      class="mb"
    />
  </section>
</template>

<script setup>
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

onMounted(() => {
  router.replace('/plugins/overview');
});
</script>
