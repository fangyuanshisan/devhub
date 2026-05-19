<template>
  <section class="plugin-page" data-testid="plugin-versions-page">
    <div class="page-hero">
      <div>
        <p class="eyebrow">插件包治理</p>
        <h1>插件版本仓库</h1>
        <p class="muted">聚合已安装插件、本地包、上传包和远程只读索引版本；远程索引只展示元数据，不下载、不安装、不执行代码。</p>
      </div>
      <el-button :loading="loading" data-testid="plugin-versions-refresh" @click="load">刷新</el-button>
    </div>

    <el-alert
      class="mb"
      type="info"
      show-icon
      :closable="false"
      title="版本仓库不会自动升级；升级前必须重新校验 checksum、签名、依赖、Core 兼容和风险报告。"
    />

    <div class="filter-panel mb" data-testid="plugin-versions-filters">
      <el-input v-model="filters.keyword" clearable placeholder="搜索插件 code / 名称 / 版本" @keyup.enter="load" />
      <el-select v-model="filters.source" clearable placeholder="来源">
        <el-option label="已安装" value="installed" />
        <el-option label="本地插件包" value="local_package" />
        <el-option label="上传包" value="uploaded_package" />
        <el-option label="远程索引" value="remote_index" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
    </div>

    <el-card shadow="never" data-testid="plugin-version-list">
      <el-table :data="items" stripe size="small" v-loading="loading">
        <el-table-column prop="plugin_code" label="插件" min-width="180">
          <template #default="{ row }">
            <strong>{{ row.plugin_code }}</strong>
            <div class="muted tiny">{{ row.plugin_name || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="installed_version" label="已安装版本" width="130" />
        <el-table-column prop="latest_local_version" label="最新本地版本" width="140" />
        <el-table-column prop="latest_remote_version" label="最新远程版本" width="140" />
        <el-table-column label="来源" min-width="160">
          <template #default="{ row }">
            <el-tag v-for="src in row.sources || []" :key="src" class="mr-xs" size="small" effect="plain">{{ sourceLabel(src) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="升级" width="120">
          <template #default="{ row }">
            <el-tag :type="row.update_available ? 'warning' : 'info'" effect="plain">{{ row.update_available ? '可对比' : '无新版本' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" data-testid="plugin-version-detail-btn" @click="openVersions(row)">查看版本</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="detailVisible" size="70%" title="插件版本详情" data-testid="plugin-version-detail-drawer">
      <section v-if="detail" data-testid="plugin-version-detail">
        <h3>{{ detail.plugin_code }} 版本列表</h3>
        <p class="muted">当前安装版本：{{ detail.installed_version || '-' }}</p>
        <el-table :data="detail.versions || []" size="small" stripe>
          <el-table-column prop="version" label="版本" width="120" />
          <el-table-column label="来源" width="150">
            <template #default="{ row }">{{ sourceLabel(row.source) }}</template>
          </el-table-column>
          <el-table-column label="状态" width="130">
            <template #default="{ row }">{{ pluginStatusText(row.status) }}</template>
          </el-table-column>
          <el-table-column label="风险 / 签名" min-width="220">
            <template #default="{ row }">
              <el-tag :type="riskType(row.risk_level)" size="small" effect="plain">{{ pluginRiskText(row.risk_level || 'low') }}</el-tag>
              <el-tag class="ml-xs" size="small" effect="plain">{{ pluginStatusText(row.signature_status) }}</el-tag>
              <el-tag class="ml-xs" size="small" effect="plain">{{ trustLevelLabel(row.trust_status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="package_path" label="路径 / 来源" min-width="260">
            <template #default="{ row }">{{ row.package_path || row.remote_source_id || row.readonly_message || '-' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="170" fixed="right">
            <template #default="{ row }">
              <el-tooltip v-if="row.readonly" content="远程索引版本只读，不能直接升级对比">
                <el-button text disabled data-testid="plugin-version-remote-readonly">只读</el-button>
              </el-tooltip>
              <el-button
                v-else
                text
                type="primary"
                :disabled="!row.is_upgrade_candidate"
                data-testid="plugin-upgrade-diff-btn"
                @click="openUpgradeDiff(row)"
              >
                升级差异
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>
    </el-drawer>

    <el-drawer v-model="diffVisible" size="76%" title="升级差异对比" data-testid="plugin-upgrade-diff-drawer">
      <section v-if="upgradeDiff" data-testid="plugin-upgrade-diff">
        <div class="diff-head">
          <div>
            <h3>{{ upgradeDiff.plugin_code }}：{{ upgradeDiff.current_version }} → {{ upgradeDiff.target_version }}</h3>
            <p class="muted">来源：{{ sourceLabel(upgradeDiff.source) }}；不会自动升级，不执行第三方代码 / SQL / 前端资产。</p>
          </div>
          <el-tag :type="riskType(upgradeDiff.risk_report?.level)" size="large">{{ pluginStatusText(upgradeDiff.status) }}</el-tag>
        </div>
        <el-descriptions :column="5" border class="mb">
          <el-descriptions-item label="新增">{{ upgradeDiff.summary?.added || 0 }}</el-descriptions-item>
          <el-descriptions-item label="移除">{{ upgradeDiff.summary?.removed || 0 }}</el-descriptions-item>
          <el-descriptions-item label="变更">{{ upgradeDiff.summary?.changed || 0 }}</el-descriptions-item>
          <el-descriptions-item label="高风险">{{ upgradeDiff.summary?.high_risk || 0 }}</el-descriptions-item>
          <el-descriptions-item label="阻断">{{ upgradeDiff.summary?.blocked || 0 }}</el-descriptions-item>
        </el-descriptions>

        <el-alert v-if="upgradeDiff.risk_report?.summary" class="mb" :type="riskType(upgradeDiff.risk_report.level)" show-icon :closable="false" :title="upgradeDiff.risk_report.summary" />

        <el-collapse v-model="openSections" data-testid="plugin-upgrade-diff-sections">
          <el-collapse-item v-for="section in upgradeDiff.diff_sections || []" :key="section.section" :name="section.section">
            <template #title>
              <strong>{{ section.title }}</strong>
              <el-tag class="ml-xs" size="small" :type="riskType(section.risk_level)" effect="plain">{{ pluginRiskText(section.risk_level) }}</el-tag>
            </template>
            <el-table :data="section.items || []" size="small" stripe>
              <el-table-column prop="path" label="路径" min-width="220" />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column label="风险" width="100">
                <template #default="{ row }"><el-tag :type="riskType(row.risk_level)" effect="plain">{{ pluginRiskText(row.risk_level) }}</el-tag></template>
              </el-table-column>
              <el-table-column prop="message" label="说明" min-width="220" />
              <el-table-column label="变更前" min-width="180">
                <template #default="{ row }"><code>{{ compact(row.before) }}</code></template>
              </el-table-column>
              <el-table-column label="变更后" min-width="180">
                <template #default="{ row }"><code>{{ compact(row.after) }}</code></template>
              </el-table-column>
            </el-table>
          </el-collapse-item>
        </el-collapse>

        <div class="mt">
          <el-button type="primary" :disabled="upgradeDiff.status === 'blocked'" data-testid="plugin-upgrade-submit-approval" @click="submitUpgradeApproval">
            提交升级审批
          </el-button>
          <span v-if="upgradeDiff.status === 'blocked'" class="muted ml-xs">存在阻断项，不能提交审批。</span>
        </div>
      </section>
    </el-drawer>
  </section>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { listPluginVersions, listPluginCodeVersions, dryRunPluginVersionUpgradeDiff, createPluginApproval } from '@/api/admin';
import { trustLevelLabel } from '@/i18n/formatters';
import { pluginRiskText, pluginStatusText } from '@/modules/plugins/statusText';

const loading = ref(false);
const items = ref([]);
const filters = ref({ keyword: '', source: '' });
const detailVisible = ref(false);
const detail = ref(null);
const diffVisible = ref(false);
const upgradeDiff = ref(null);
const selectedVersion = ref(null);
const openSections = ref([]);

const load = async () => {
  loading.value = true;
  try {
    const data = await listPluginVersions({ ...filters.value, page_size: 50 });
    items.value = data.items || [];
  } finally {
    loading.value = false;
  }
};

const openVersions = async (row) => {
  const data = await listPluginCodeVersions(row.plugin_code);
  detail.value = data;
  detailVisible.value = true;
};

const openUpgradeDiff = async (row) => {
  selectedVersion.value = row;
  const data = await dryRunPluginVersionUpgradeDiff(row.plugin_code, row.version, {
    source: row.source,
    package_path: row.package_path,
    remote_index_id: row.remote_index_id,
  });
  upgradeDiff.value = data;
  openSections.value = (data.diff_sections || []).map((item) => item.section);
  diffVisible.value = true;
};

const submitUpgradeApproval = async () => {
  if (!selectedVersion.value || !upgradeDiff.value) return;
  await createPluginApproval({
    action: 'upgrade',
    plugin_code: upgradeDiff.value.plugin_code,
    package_path: selectedVersion.value.package_path,
    reason: `升级差异确认：${upgradeDiff.value.current_version} -> ${upgradeDiff.value.target_version}`,
  });
  ElMessage.success('已提交升级审批，审批详情会保留预检 / 差异快照');
};

const riskType = (level) => {
  if (level === 'blocked' || level === 'high') return 'danger';
  if (level === 'warning' || level === 'medium') return 'warning';
  if (level === 'low' || level === 'ok') return 'success';
  return 'info';
};

const sourceLabel = (source) => {
  const value = String(source || '');
  if (value === 'installed') return '已安装';
  if (value === 'local_package') return '本地插件包';
  if (value === 'uploaded_package') return '上传包';
  if (value === 'remote_index') return '远程索引';
  return value || '-';
};

const compact = (value) => {
  if (value === null || value === undefined) return '-';
  if (typeof value === 'string') return value;
  return JSON.stringify(value);
};

onMounted(load);
</script>

<style scoped>
.plugin-page {
  padding: 16px;
}
.page-hero {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 14px;
}
.page-hero h1 {
  margin: 4px 0 8px;
  font-size: 24px;
}
.eyebrow,
.muted {
  color: #64748b;
}
.eyebrow {
  margin: 0;
  font-size: 13px;
}
.tiny {
  font-size: 12px;
}
.filter-panel {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) 180px auto;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
}
.diff-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.mb {
  margin-bottom: 14px;
}
.mt {
  margin-top: 16px;
}
.mr-xs {
  margin-right: 6px;
}
.ml-xs {
  margin-left: 6px;
}
code {
  white-space: normal;
  word-break: break-all;
}
</style>
