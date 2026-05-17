<template>
  <section class="plugin-webhooks" data-testid="plugin-webhooks-page">
    <div class="page-title">
      <div>
        <h2>Webhook 治理</h2>
        <p class="muted">
          v1.7.6：治理 non_blocking delivery 的重试/熔断 + HMAC 签名与 Secret 轮换，不执行第三方插件代码。
        </p>
      </div>
      <div class="page-actions">
        <el-button
          size="small"
          type="primary"
          :loading="retryDueLoading"
          data-testid="webhook-retry-due"
          @click="retryDue"
        >
          扫描重试到期 delivery
        </el-button>
      </div>
    </div>

    <el-tabs v-model="tab" class="page-tabs" data-testid="webhook-tabs">
      <el-tab-pane label="Deliveries" name="deliveries">
        <PluginFilterBar title="Delivery 列表" tip="状态筛选在页内 Tab/筛选区，不在侧边导航" testid="webhook-deliveries-filter">
          <template #actions>
            <el-button size="small" @click="refreshDeliveries">刷新</el-button>
          </template>
          <el-input v-model="deliveryFilters.plugin_code" size="small" placeholder="plugin_code" style="width: 160px" />
          <el-input v-model="deliveryFilters.hook_name" size="small" placeholder="hook_name" style="width: 200px" />
          <el-select v-model="deliveryFilters.status" size="small" placeholder="status" style="width: 180px">
            <el-option label="all" value="all" />
            <el-option label="pending" value="pending" />
            <el-option label="sending" value="sending" />
            <el-option label="success" value="success" />
            <el-option label="failed" value="failed" />
            <el-option label="retry_scheduled" value="retry_scheduled" />
            <el-option label="retry_exhausted" value="retry_exhausted" />
            <el-option label="circuit_open" value="circuit_open" />
            <el-option label="skipped" value="skipped" />
          </el-select>
          <el-button size="small" type="primary" data-testid="webhook-deliveries-search" @click="refreshDeliveries">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="deliveriesError" />

        <el-table
          v-loading="deliveriesLoading"
          :data="deliveries.items"
          stripe
          border
          size="small"
          data-testid="webhook-deliveries-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="plugin" width="140" />
          <el-table-column prop="hook_name" label="hook" width="220" />
          <el-table-column label="status" width="160">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-delivery-status" />
            </template>
          </el-table-column>
          <el-table-column label="attempt" width="120">
            <template #default="{ row }">
              <span data-testid="webhook-delivery-attempt">{{ row.attempt }}/{{ row.max_attempts }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="next_retry_at" label="next_retry_at" width="170" />
          <el-table-column prop="response_status" label="resp" width="90" />
          <el-table-column prop="retry_reason" label="reason" width="140" />
          <el-table-column prop="error_message" label="error" min-width="220" show-overflow-tooltip />
          <el-table-column label="actions" width="140" fixed="right">
            <template #default="{ row }">
              <el-button
                size="small"
                type="warning"
                plain
                :disabled="row.status === 'success'"
                data-testid="webhook-retry-button"
                @click="manualRetry(row)"
              >
                手动重试
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <PluginEmptyState
          v-if="!deliveriesLoading && deliveries.items.length === 0 && !deliveriesError"
          description="暂无 delivery 记录"
          testid="webhook-deliveries-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="deliveries.pagination.page || 1"
            :page-size="deliveries.pagination.page_size || 20"
            :total="deliveries.pagination.total || 0"
            @current-change="onDeliveryPageChange"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="Circuit Breakers" name="circuits">
        <PluginFilterBar title="熔断状态" tip="维度：plugin_code + target_url" testid="webhook-circuits-filter">
          <template #actions>
            <el-button size="small" @click="refreshCircuits">刷新</el-button>
          </template>
          <el-input v-model="circuitFilters.plugin_code" size="small" placeholder="plugin_code" style="width: 180px" />
          <el-select v-model="circuitFilters.status" size="small" placeholder="status" style="width: 180px">
            <el-option label="all" value="all" />
            <el-option label="closed" value="closed" />
            <el-option label="open" value="open" />
            <el-option label="half_open" value="half_open" />
          </el-select>
          <el-button size="small" type="primary" data-testid="webhook-circuits-search" @click="refreshCircuits">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="circuitsError" />

        <el-table
          v-loading="circuitsLoading"
          :data="circuits.items"
          stripe
          border
          size="small"
          data-testid="webhook-circuits-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="plugin" width="160" />
          <el-table-column prop="target_url" label="target_url" min-width="260" show-overflow-tooltip />
          <el-table-column label="status" width="140">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-circuit-status" />
            </template>
          </el-table-column>
          <el-table-column prop="failure_count" label="failures" width="110" />
          <el-table-column prop="next_probe_at" label="next_probe_at" width="170" />
          <el-table-column prop="last_error_message" label="last_error" min-width="220" show-overflow-tooltip />
          <el-table-column label="actions" width="220" fixed="right">
            <template #default="{ row }">
              <el-button
                size="small"
                type="success"
                plain
                data-testid="webhook-circuit-close"
                :disabled="row.status === 'closed'"
                @click="closeCircuit(row)"
              >
                手动恢复
              </el-button>
              <el-button
                size="small"
                type="danger"
                plain
                data-testid="webhook-circuit-open"
                @click="openCircuit(row)"
              >
                手动熔断
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <PluginEmptyState
          v-if="!circuitsLoading && circuits.items.length === 0 && !circuitsError"
          description="暂无熔断记录"
          testid="webhook-circuits-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="circuits.pagination.page || 1"
            :page-size="circuits.pagination.page_size || 20"
            :total="circuits.pagination.total || 0"
            @current-change="onCircuitPageChange"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="Secrets" name="secrets">
        <PluginFilterBar title="Webhook Secrets" tip="Secret 明文只会在创建/轮换时展示一次" testid="webhook-secrets-filter">
          <template #actions>
            <el-button size="small" @click="refreshSecrets">刷新</el-button>
            <el-button size="small" type="primary" data-testid="webhook-secret-create" @click="openCreateSecret">
              创建 Secret
            </el-button>
          </template>
          <el-input v-model="secretFilters.plugin_code" size="small" placeholder="plugin_code" style="width: 180px" />
          <el-input v-model="secretFilters.secret_ref" size="small" placeholder="secret_ref" style="width: 220px" />
          <el-select v-model="secretFilters.status" size="small" placeholder="status" style="width: 180px">
            <el-option label="all" value="all" />
            <el-option label="active" value="active" />
            <el-option label="previous" value="previous" />
            <el-option label="disabled" value="disabled" />
            <el-option label="revoked" value="revoked" />
            <el-option label="expired" value="expired" />
          </el-select>
          <el-button size="small" type="primary" data-testid="webhook-secrets-search" @click="refreshSecrets">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="secretsError" />

        <el-table
          v-loading="secretsLoading"
          :data="secrets.items"
          stripe
          border
          size="small"
          data-testid="webhook-secrets-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="plugin" width="160" />
          <el-table-column prop="secret_ref" label="secret_ref" min-width="220" show-overflow-tooltip />
          <el-table-column prop="target_url" label="target_url" min-width="260" show-overflow-tooltip />
          <el-table-column label="status" width="140">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-secret-status" />
            </template>
          </el-table-column>
          <el-table-column prop="grace_until" label="grace_until" width="170" />
          <el-table-column prop="last_used_at" label="last_used_at" width="170" />
          <el-table-column label="actions" width="320" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" plain data-testid="webhook-secret-rotate" @click="rotateSecret(row)">
                轮换
              </el-button>
              <el-button
                v-if="row.status === 'disabled'"
                size="small"
                type="success"
                plain
                data-testid="webhook-secret-enable"
                @click="enableSecret(row)"
              >
                恢复
              </el-button>
              <el-button
                v-else
                size="small"
                type="warning"
                plain
                data-testid="webhook-secret-disable"
                :disabled="row.status === 'revoked' || row.status === 'expired'"
                @click="disableSecret(row)"
              >
                禁用
              </el-button>
              <el-button
                size="small"
                type="danger"
                plain
                data-testid="webhook-secret-revoke"
                :disabled="row.status === 'revoked'"
                @click="revokeSecret(row)"
              >
                吊销
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <PluginEmptyState
          v-if="!secretsLoading && secrets.items.length === 0 && !secretsError"
          description="暂无 Webhook Secret"
          testid="webhook-secrets-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="secrets.pagination.page || 1"
            :page-size="secrets.pagination.page_size || 20"
            :total="secrets.pagination.total || 0"
            @current-change="onSecretPageChange"
          />
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="createSecretDialogVisible" title="创建 Webhook Secret" width="560px">
      <div class="muted" style="margin-bottom: 8px">
        Secret 明文只会在创建成功后展示一次，请立即复制保存。
      </div>
      <el-form label-width="100px">
        <el-form-item label="plugin_code">
          <el-input v-model="createSecretForm.plugin_code" data-testid="webhook-secret-form-plugin" />
        </el-form-item>
        <el-form-item label="target_url">
          <el-input v-model="createSecretForm.target_url" data-testid="webhook-secret-form-target" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createSecretDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createSecretLoading" data-testid="webhook-secret-form-submit" @click="submitCreateSecret">
          创建
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="secretOnceDialogVisible" title="Secret（仅展示一次）" width="640px">
      <div class="muted" style="margin-bottom: 8px">关闭后将无法再次查看，请立即复制。</div>
      <el-input v-model="secretOnceValue" type="textarea" :rows="4" readonly data-testid="webhook-secret-once-value" />
      <template #footer>
        <el-button type="primary" data-testid="webhook-secret-once-close" @click="closeSecretOnceDialog">我已保存</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import * as admin from '@/api/admin';

import PluginFilterBar from './components/PluginFilterBar.vue';
import PluginEmptyState from './components/PluginEmptyState.vue';
import PluginErrorAlert from './components/PluginErrorAlert.vue';
import PluginStatusTag from './components/PluginStatusTag.vue';
import { confirmDanger } from './components/useDangerConfirm';

const route = useRoute();
const router = useRouter();

const tab = ref(String(route.query.tab || 'deliveries'));
watch(tab, async (next) => {
  await router.replace({ query: { ...route.query, tab: next } });
});

const deliveries = ref({ items: [], pagination: { page: 1, page_size: 20, total: 0 } });
const deliveriesLoading = ref(false);
const deliveriesError = ref('');
const deliveryFilters = ref({
  plugin_code: String(route.query.plugin_code || ''),
  hook_name: String(route.query.hook_name || ''),
  status: String(route.query.status || 'all'),
  page: Number(route.query.page || 1),
  page_size: 20,
});

const circuits = ref({ items: [], pagination: { page: 1, page_size: 20, total: 0 } });
const circuitsLoading = ref(false);
const circuitsError = ref('');
const circuitFilters = ref({
  plugin_code: String(route.query.cb_plugin_code || ''),
  status: String(route.query.cb_status || 'all'),
  page: Number(route.query.cb_page || 1),
  page_size: 20,
});

const retryDueLoading = ref(false);

const secrets = ref({ items: [], pagination: { page: 1, page_size: 20, total: 0 } });
const secretsLoading = ref(false);
const secretsError = ref('');
const secretFilters = ref({
  plugin_code: String(route.query.sec_plugin_code || ''),
  status: String(route.query.sec_status || 'all'),
  secret_ref: String(route.query.sec_ref || ''),
  page: Number(route.query.sec_page || 1),
  page_size: 20,
});

const createSecretDialogVisible = ref(false);
const createSecretLoading = ref(false);
const createSecretForm = ref({ plugin_code: '', target_url: '' });
const secretOnceDialogVisible = ref(false);
const secretOnceValue = ref('');

function openCreateSecret() {
  createSecretForm.value = { plugin_code: secretFilters.value.plugin_code || '', target_url: '' };
  createSecretDialogVisible.value = true;
}

function closeSecretOnceDialog() {
  secretOnceDialogVisible.value = false;
  secretOnceValue.value = '';
}

async function submitCreateSecret() {
  createSecretLoading.value = true;
  try {
    const res = await admin.createWebhookSecret({ plugin_code: createSecretForm.value.plugin_code, target_url: createSecretForm.value.target_url });
    secretOnceValue.value = res?.secret || '';
    secretOnceDialogVisible.value = true;
    createSecretDialogVisible.value = false;
    if (tab.value === 'secrets') await refreshSecrets();
  } finally {
    createSecretLoading.value = false;
  }
}

async function refreshSecrets() {
  secretsLoading.value = true;
  secretsError.value = '';
  try {
    const params = { ...secretFilters.value };
    await router.replace({ query: { ...route.query, tab: tab.value, sec_plugin_code: params.plugin_code, sec_status: params.status, sec_ref: params.secret_ref, sec_page: params.page } });
    const res = await admin.listWebhookSecrets(params);
    secrets.value = res;
  } catch (e) {
    secretsError.value = e?.message || '加载失败';
  } finally {
    secretsLoading.value = false;
  }
}

function onSecretPageChange(page) {
  secretFilters.value.page = page;
  refreshSecrets();
}

async function rotateSecret(row) {
  await confirmDanger(`确认轮换 Secret #${row.id}？轮换后新 Secret 明文只展示一次。`, '轮换 Secret', { confirmButtonText: '轮换', cancelButtonText: '取消' });
  const res = await admin.rotateWebhookSecret(row.id);
  secretOnceValue.value = res?.secret || '';
  secretOnceDialogVisible.value = true;
  await refreshSecrets();
}

async function disableSecret(row) {
  await confirmDanger(`确认禁用 Secret #${row.id}？禁用后将无法用于签名发送。`, '禁用 Secret', { confirmButtonText: '禁用', cancelButtonText: '取消' });
  await admin.disableWebhookSecret(row.id);
  await refreshSecrets();
}

async function enableSecret(row) {
  await confirmDanger(`确认恢复 Secret #${row.id}？恢复后将成为 active，并可能禁用同 target_url 的其他 active。`, '恢复 Secret', { confirmButtonText: '恢复', cancelButtonText: '取消' });
  await admin.enableWebhookSecret(row.id);
  await refreshSecrets();
}

async function revokeSecret(row) {
  await confirmDanger(`确认吊销 Secret #${row.id}？吊销后将立即失效且不可恢复。`, '吊销 Secret', { confirmButtonText: '吊销', cancelButtonText: '取消' });
  await admin.revokeWebhookSecret(row.id);
  await refreshSecrets();
}

async function refreshDeliveries() {
  deliveriesLoading.value = true;
  deliveriesError.value = '';
  try {
    const params = { ...deliveryFilters.value };
    await router.replace({ query: { ...route.query, tab: tab.value, plugin_code: params.plugin_code, hook_name: params.hook_name, status: params.status, page: params.page } });
    const res = await admin.listWebhookDeliveries(params);
    deliveries.value = res;
  } catch (e) {
    deliveriesError.value = e?.message || '加载失败';
  } finally {
    deliveriesLoading.value = false;
  }
}

async function refreshCircuits() {
  circuitsLoading.value = true;
  circuitsError.value = '';
  try {
    const params = { ...circuitFilters.value };
    await router.replace({ query: { ...route.query, tab: tab.value, cb_plugin_code: params.plugin_code, cb_status: params.status, cb_page: params.page } });
    const res = await admin.listWebhookCircuitBreakers(params);
    circuits.value = res;
  } catch (e) {
    circuitsError.value = e?.message || '加载失败';
  } finally {
    circuitsLoading.value = false;
  }
}

async function manualRetry(row) {
  await confirmDanger(`确认重试 delivery #${row.id}？`, '手动重试', { confirmButtonText: '重试', cancelButtonText: '取消' });
  await admin.retryWebhookDelivery(row.id);
  await refreshDeliveries();
}

async function closeCircuit(row) {
  await confirmDanger(`确认恢复熔断 #${row.id}？`, '手动恢复熔断', { confirmButtonText: '恢复', cancelButtonText: '取消' });
  await admin.closeWebhookCircuitBreaker(row.id);
  await refreshCircuits();
}

async function openCircuit(row) {
  await confirmDanger(`确认手动打开熔断 #${row.id}？将暂停投递该 target_url。`, '手动熔断', { confirmButtonText: '打开熔断', cancelButtonText: '取消' });
  await admin.openWebhookCircuitBreaker(row.id);
  await refreshCircuits();
}

async function retryDue() {
  retryDueLoading.value = true;
  try {
    await admin.retryDueWebhookDeliveries({ limit: 50 });
    if (tab.value === 'deliveries') await refreshDeliveries();
  } finally {
    retryDueLoading.value = false;
  }
}

function onDeliveryPageChange(page) {
  deliveryFilters.value.page = page;
  refreshDeliveries();
}

function onCircuitPageChange(page) {
  circuitFilters.value.page = page;
  refreshCircuits();
}

onMounted(async () => {
  if (tab.value === 'circuits') await refreshCircuits();
  else if (tab.value === 'secrets') await refreshSecrets();
  else await refreshDeliveries();
});
</script>

<style scoped>
.plugin-webhooks {
  padding: 4px 0 12px;
}

.page-title {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0 6px;
}

.page-title h2 {
  margin: 0;
  font-size: 18px;
  line-height: 26px;
}

.muted {
  color: #6b7280;
  margin: 4px 0 0;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
