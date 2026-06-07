<template>
  <div class="plugin-config-keys" data-testid="plugin-config-keys-page">
    <div class="page-header">
      <div>
        <h1>插件配置密钥</h1>
        <p class="desc">
          仅展示 key_id，不展示密钥明文；支持轮换预检与受控重新加密（默认不处理配置历史版本）。不支持 KMS/Vault/自动定时轮换。
        </p>
      </div>
      <div class="actions">
        <el-button :loading="loading" @click="loadStatus">刷新</el-button>
      </div>
    </div>

    <el-card shadow="never" class="status-card">
      <template #header>
        <div class="card-header">
          <span>密钥状态</span>
          <PluginStatusTag :value="status?.status" testid="plugin-config-keys-status-tag" />
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="current_key_id">{{ status?.current_key_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="key_count">{{ status?.key_count ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="legacy_v1_supported">{{ status?.legacy_v1_supported ? 'true' : 'false' }}</el-descriptions-item>
        <el-descriptions-item label="loaded_key_ids">
          <span v-if="(status?.loaded_key_ids || []).length">{{ (status?.loaded_key_ids || []).join(', ') }}</span>
          <span v-else>-</span>
        </el-descriptions-item>
      </el-descriptions>
      <el-alert
        v-if="(status?.warnings || []).length"
        type="warning"
        show-icon
        class="mt12"
        :title="status.warnings[0]"
      />
    </el-card>

    <el-card shadow="never" class="rotation-card">
      <template #header>
        <div class="card-header">
          <span>轮换预检 / 重新加密</span>
        </div>
      </template>

      <PluginPackageBoundaryNotice
        title="预检不会修改任何配置；重新加密只会把可解密的旧配置转换为当前受支持的加密格式。"
        description="不执行第三方代码，不执行外部 SQL，不动态加载前端资产；页面不会展示密钥明文、密文或真实密钥，本轮不支持 KMS/Vault/自动定时轮换。"
      />

      <div class="form-row">
        <el-select v-model="form.scope" placeholder="scope" style="width: 160px">
          <el-option label="all" value="all" />
          <el-option label="plugin" value="plugin" />
          <el-option label="community" value="community" />
        </el-select>
        <el-input v-model="form.plugin_code" placeholder="plugin_code（scope=plugin）" clearable style="width: 220px" />
        <el-input v-model="form.community_id" placeholder="community_id（scope=community）" clearable style="width: 220px" />
        <el-switch v-model="form.include_config_versions" disabled />
        <span class="hint">include_config_versions（默认不支持）</span>
        <el-button type="primary" :loading="dryRunLoading" data-testid="plugin-config-keys-rotation-dryrun" @click="dryRun">执行轮换预检</el-button>
      </div>

      <div v-if="dryRunResult" class="result">
        <div class="result-header">
          <PluginStatusTag :value="dryRunResult.status" testid="plugin-config-keys-rotation-status" />
          <span class="ml8">current_key_id: {{ dryRunResult.current_key_id }}</span>
        </div>

        <el-descriptions :column="3" border class="mt12">
          <el-descriptions-item label="total_sensitive_values">{{ dryRunResult.summary?.total_sensitive_values ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="already_current">{{ dryRunResult.summary?.already_current ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="needs_reencrypt">{{ dryRunResult.summary?.needs_reencrypt ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="legacy_v1">{{ dryRunResult.summary?.legacy_v1 ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="decrypt_failed">{{ dryRunResult.summary?.decrypt_failed ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="missing_key">{{ dryRunResult.summary?.missing_key ?? 0 }}</el-descriptions-item>
        </el-descriptions>

        <el-alert
          v-if="(dryRunResult.errors || []).length"
          type="error"
          show-icon
          class="mt12"
          :title="dryRunResult.errors[0]?.message || 'blocked'"
          :description="dryRunResult.errors[0]?.suggestion || ''"
        />
        <el-alert
          v-else-if="(dryRunResult.warnings || []).length"
          type="warning"
          show-icon
          class="mt12"
          :title="dryRunResult.warnings[0]"
        />

        <div class="mt12 actions">
          <el-button
            type="danger"
            :disabled="!canManage || dryRunResult.status === 'blocked'"
            :loading="reencryptLoading"
            data-testid="plugin-config-keys-reencrypt"
            @click="doReencrypt"
          >
            重新加密（使用当前 key）
          </el-button>
          <el-button :disabled="!dryRunResult.items?.length" @click="showItems = !showItems">
            {{ showItems ? '隐藏明细' : '查看明细' }} ({{ (dryRunResult.items || []).length }})
          </el-button>
          <el-alert v-if="!canManage" type="warning" show-icon title="缺少 plugin.manage：只能查看，不能执行重新加密。" class="ml12" />
        </div>

        <el-table v-if="showItems" :data="dryRunResult.items || []" class="mt12" height="420" data-testid="plugin-config-keys-items-table">
          <el-table-column prop="plugin_code" label="plugin_code" width="180" />
          <el-table-column prop="scope" label="scope" width="110" />
          <el-table-column prop="community_id" label="community_id" width="120" />
          <el-table-column prop="field_path" label="field_path" min-width="220" />
          <el-table-column prop="cipher_version" label="cipher" width="100" />
          <el-table-column prop="key_id" label="key_id" min-width="160" />
          <el-table-column prop="status" label="状态" width="150">
            <template #default="{ row }">{{ genericStatusLabel(row.status) }}</template>
          </el-table-column>
          <el-table-column prop="message" label="说明" min-width="260" />
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import PluginPackageBoundaryNotice from './components/PluginPackageBoundaryNotice.vue';
import PluginStatusTag from './components/PluginStatusTag.vue';
import { dryRunPluginConfigKeyRotation, getPluginConfigKeyStatus, reencryptPluginConfigKeys } from '@/api/plugins';
import { genericStatusLabel } from '@/i18n/formatters';
import { pluginMessageText } from '@/modules/plugins/statusText';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const canManage = computed(() => auth.can('plugin.manage'));

const loading = ref(false);
const status = ref(null);

const form = ref({
  scope: 'all',
  plugin_code: '',
  community_id: '',
  include_config_versions: false,
});

const dryRunLoading = ref(false);
const reencryptLoading = ref(false);
const dryRunResult = ref(null);
const showItems = ref(false);

const loadStatus = async () => {
  loading.value = true;
  try {
    status.value = await getPluginConfigKeyStatus();
  } catch (e) {
    const error = apiError(e);
    status.value = {
      current_key_id: '',
      loaded_key_ids: [],
      legacy_v1_supported: false,
      key_count: 0,
      status: 'blocked',
      warnings: [error],
    };
    ElMessage.error(error);
  } finally {
    loading.value = false;
  }
};

const dryRun = async () => {
  dryRunLoading.value = true;
  showItems.value = false;
  try {
    const payload = {
      scope: form.value.scope,
      plugin_code: form.value.plugin_code || '',
      community_id: Number(form.value.community_id || 0) || 0,
      include_config_versions: false,
    };
    dryRunResult.value = await dryRunPluginConfigKeyRotation(payload);
  } catch (e) {
    const error = apiError(e);
    dryRunResult.value = {
      status: 'blocked',
      current_key_id: status.value?.current_key_id || '',
      summary: {
        total_sensitive_values: 0,
        already_current: 0,
        needs_reencrypt: 0,
        legacy_v1: 0,
        decrypt_failed: 0,
        missing_key: 1,
      },
      items: [],
      warnings: [],
      errors: [{ message: error, suggestion: '请在启动环境中配置插件配置加密密钥后重启服务，再回到本页重试。' }],
    };
    ElMessage.error(error);
  } finally {
    dryRunLoading.value = false;
  }
};

function apiError(error) {
  const data = error?.response?.data?.error || error?.response?.data || {};
  return pluginMessageText(data, error?.message || '请求失败');
}

const doReencrypt = async () => {
  if (!canManage.value) {
    ElMessage.warning('缺少 plugin.manage');
    return;
  }
  if (!dryRunResult.value) {
    ElMessage.warning('请先执行预检');
    return;
  }
  const currentKeyID = status.value?.current_key_id || dryRunResult.value.current_key_id || '';
  await ElMessageBox.confirm(
    `确认执行重新加密？将使用 current_key_id=${currentKeyID} 重写敏感字段密文；不会返回明文，也不会展示密文。`,
    '确认重新加密',
    { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' },
  );
  reencryptLoading.value = true;
  try {
    const payload = {
      scope: form.value.scope,
      plugin_code: form.value.plugin_code || '',
      community_id: Number(form.value.community_id || 0) || 0,
      include_config_versions: false,
      confirm_current_key_id: currentKeyID,
    };
    const data = await reencryptPluginConfigKeys(payload);
    ElMessage.success(`重新加密完成：更新 ${data.updated_count || 0} 条`);
    await loadStatus();
    await dryRun();
  } catch (e) {
    ElMessage.error(apiError(e) || '重新加密失败');
  } finally {
    reencryptLoading.value = false;
  }
};

onMounted(loadStatus);
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}
.desc {
  color: #666;
  margin: 6px 0 0 0;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  flex-wrap: wrap;
}
.hint {
  color: #888;
  font-size: 12px;
}
.mt12 {
  margin-top: 12px;
}
.ml8 {
  margin-left: 8px;
}
.ml12 {
  margin-left: 12px;
}
</style>
