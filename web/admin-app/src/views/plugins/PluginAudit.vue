<template>
  <section class="plugin-page" data-testid="plugin-audit-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">安装与治理</div>
        <h2>审计日志</h2>
        <p class="muted">插件相关审计入口聚合；最终审计列表仍在“治理审计”页面展示（带 query 跳转）。</p>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" class="mb" title="提示：此页不复制实现全量审计列表，只提供插件相关的快捷筛选入口。" />

    <div class="governance-grid" data-testid="plugin-audit-grid">
      <article v-for="card in cards" :key="card.key" class="governance-card">
        <div class="governance-card-head">
          <div>
            <h3>{{ card.title }}</h3>
            <p>{{ card.desc }}</p>
          </div>
        </div>
        <div class="row-actions">
          <el-button type="primary" plain :data-testid="`plugin-audit-open-${card.key}`" @click="openAudit(card.query)">打开审计</el-button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

const cards = computed(() => ([
  { key: 'lifecycle', title: '安装 / 升级 / 归档', desc: '插件生命周期变更审计（install/upgrade/archive/restore/enable/disable）', query: { target_type: 'plugin' } },
  { key: 'config', title: '配置变更', desc: '全局/子站插件配置保存、schema 校验失败等', query: { action_prefix: 'plugin.config' } },
  { key: 'dependencies', title: '依赖阻断', desc: '依赖缺失、版本不匹配、循环依赖导致的阻断', query: { action_prefix: 'plugin.dependency' } },
  { key: 'hooks', title: 'Hook 失败', desc: 'blocking/non-blocking Hook 失败、阻断与排障记录', query: { action_prefix: 'plugin.hook' } },
  { key: 'migrations', title: '迁移执行', desc: 'migration run/retry/success/failed 相关审计', query: { action_prefix: 'plugin.migration' } },
]));

function openAudit(query) {
  router.push({ path: '/audit-logs', query: query || {} });
}
</script>

