<template>
  <section class="plugin-page" data-testid="plugin-search-index-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">运行排障</div>
        <h2>搜索索引</h2>
        <p class="muted">聚合 searchable 内容类型与索引重建入口。当前版本不展示外部搜索引擎配置为已实现能力。</p>
      </div>
    </div>

    <el-alert
      type="warning"
      show-icon
      :closable="false"
      class="mb"
      title="当前版本限制：尚未形成独立的搜索索引治理页与重建入口（仅保留基础实现与调用点）。本页只提供限制说明，避免出现不可用的假按钮。"
    />

    <el-card shadow="never">
      <template #header>
        <strong>后续建议</strong>
      </template>
      <ul class="muted">
        <li>v1.5 建议引入异步重建队列 + 可观测的重建结果记录，再开放插件级/内容类型级重建入口。</li>
        <li>当前阶段如遇搜索问题，请结合审计日志与 Hook 执行记录排查。</li>
      </ul>
    </el-card>
  </section>
</template>

<script setup>
</script>

