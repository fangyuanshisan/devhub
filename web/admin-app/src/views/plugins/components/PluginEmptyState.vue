<template>
  <section class="plugin-empty" :data-testid="testid">
    <el-empty :description="description">
      <template v-if="$slots.default">
        <slot />
      </template>
    </el-empty>
  </section>
</template>

<script setup>
defineProps({
  description: { type: String, default: '暂无数据' },
  testid: { type: String, default: '' },
});
</script>

<style scoped>
.plugin-empty {
  margin: 12px 0;
  padding: 12px 0;
  border-radius: 10px;
  border: 1px dashed #d7dce4;
  background: #fff;
}
</style>

