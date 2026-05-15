<template>
  <el-alert
    class="plugin-boundary-notice"
    type="info"
    show-icon
    :closable="false"
    :title="title"
    :description="description"
    data-testid="plugin-package-boundary-notice"
  />
</template>

<script setup>
defineProps({
  title: {
    type: String,
    default: '安全边界：不执行第三方代码，不执行外部 SQL，不动态加载前端资产。',
  },
  description: {
    type: String,
    default: '远程索引只读；zip 上传只是 staging，promote 只进入本地插件仓库，安装仍需 dry-run、审批和后端校验。',
  },
});
</script>

<style scoped>
.plugin-boundary-notice {
  margin-bottom: 12px;
}
</style>
