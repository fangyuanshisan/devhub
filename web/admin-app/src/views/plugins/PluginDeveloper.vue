<template>
  <section class="plugin-page" data-testid="plugin-developer-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">开发者</div>
        <h2>开发者工具</h2>
        <p class="muted">SDK 文档、manifest 示例与模板命令聚合。后台不直接生成文件（仅提供命令与文档入口）。</p>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" class="mb" title="安全边界：不支持插件市场、zip 上传、远程安装、动态加载、脚本沙箱或第三方代码执行。" />

    <el-card class="mb" shadow="never">
      <template #header>
        <strong>文档入口</strong>
      </template>
      <ul class="doc-links">
        <li><a href="/docs/PLUGIN_SDK.md" target="_blank" rel="noreferrer">PLUGIN_SDK.md</a></li>
        <li><a href="/docs/PLUGIN_TEMPLATE.md" target="_blank" rel="noreferrer">PLUGIN_TEMPLATE.md</a></li>
        <li><a href="/docs/examples/plugin-manifest-example.json" target="_blank" rel="noreferrer">plugin-manifest-example.json</a></li>
      </ul>
      <div class="muted">说明：以上链接依赖当前部署是否暴露 docs 静态目录；如生产不暴露，请以仓库文档为准。</div>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <strong>模板命令</strong>
      </template>
      <pre class="json-box">go run ./cmd/devhub plugin:new \
  --code demo_links \
  --name "Demo Links" \
  --content_type demo_link \
  --content_name "演示链接" \
  --output examples/plugins \
  --with_config \
  --with_hooks \
  --with_migration</pre>
      <div class="muted">说明：此命令仅生成声明型模板，不执行第三方代码；生成的 manifest 会复用后端现有 validator 校验。</div>
    </el-card>
  </section>
</template>

<script setup>
</script>

<style scoped>
.doc-links {
  margin: 0;
  padding-left: 18px;
}
.doc-links li {
  margin: 6px 0;
}
</style>

