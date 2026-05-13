<template>
  <section class="plugin-page" data-testid="plugin-events-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">运行排障</div>
        <h2>事件通知</h2>
        <p class="muted">聚合插件事件/通知模板的展示入口。当前版本仅确认安全边界与不支持项，不提供 email/sms/webhook 的假入口。</p>
      </div>
    </div>

    <el-alert
      type="warning"
      show-icon
      :closable="false"
      class="mb"
      title="当前版本限制：事件/通知模板尚未形成完整平台化展示与管理页；本页仅作为入口与限制说明，不提供不可用的操作按钮。"
    />

    <el-card shadow="never">
      <template #header>
        <strong>当前可用与限制</strong>
      </template>
      <ul class="muted">
        <li>站内通知（in_app）为当前唯一明确支持方向；email/sms/webhook 均不在 v1.4.0 支持范围。</li>
        <li>如需排障，请优先使用：审计日志、Hook 排障、依赖兼容与 Readiness 诊断。</li>
      </ul>
    </el-card>
  </section>
</template>

<script setup>
</script>

