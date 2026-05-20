<template>
  <section class="plugin-webhooks" data-testid="plugin-webhooks-page">
    <div class="page-title">
      <div>
        <h2>Webhook 治理</h2>
        <p class="muted">
          默认聚焦投递是否成功、失败在哪里、是否需要重试或熔断恢复；密钥、Token 和回调请求收进高级治理。
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
          扫描到期重试
        </el-button>
      </div>
    </div>

    <el-tabs v-model="tab" class="page-tabs" data-testid="webhook-tabs">
      <el-tab-pane label="总览" name="overview">
        <div class="webhook-overview-grid">
          <el-card shadow="never">
            <template #header>投递概况</template>
            <div class="webhook-stat-grid">
              <button class="webhook-stat" type="button" @click="tab = 'deliveries'">
                <span>投递记录</span>
                <strong>{{ deliveryRows.length }}</strong>
              </button>
              <button class="webhook-stat" type="button" @click="tab = 'exceptions'">
                <span>等待重试</span>
                <strong>{{ retryRows.length }}</strong>
              </button>
              <button class="webhook-stat" type="button" @click="tab = 'exceptions'">
                <span>熔断中</span>
                <strong>{{ openCircuitCount }}</strong>
              </button>
            </div>
            <ul class="overview-list">
              <li>Webhook 只做 non-blocking 投递、重试、熔断和回调治理，不执行第三方插件代码。</li>
              <li>失败先看“异常处理”，再进入高级治理检查密钥、Token 或回调请求。</li>
              <li>Secret、Callback Token 和 external_service token 明文不会进入列表或执行记录。</li>
            </ul>
          </el-card>
          <el-card shadow="never">
            <template #header>下一步操作</template>
            <div class="overview-actions">
              <el-button size="small" type="primary" plain @click="tab = 'deliveries'">查看投递记录</el-button>
              <el-button size="small" type="warning" plain @click="tab = 'exceptions'">处理异常</el-button>
              <el-button size="small" type="primary" plain @click="tab = 'advanced'">高级治理</el-button>
            </div>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="投递记录" name="deliveries">
        <PluginFilterBar title="投递记录" tip="记录每次 non_blocking Webhook 投递结果；异常记录会汇总到“异常处理”。" testid="webhook-deliveries-filter">
          <template #actions>
            <el-button size="small" @click="refreshDeliveries">刷新</el-button>
          </template>
          <el-input v-model="deliveryFilters.plugin_code" size="small" placeholder="插件编码" style="width: 160px" />
          <el-input v-model="deliveryFilters.hook_name" size="small" placeholder="Hook 名称" style="width: 200px" />
          <el-select v-model="deliveryFilters.status" size="small" placeholder="状态" style="width: 180px">
            <el-option label="全部" value="all" />
            <el-option label="待处理" value="pending" />
            <el-option label="发送中" value="sending" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="等待重试" value="retry_scheduled" />
            <el-option label="重试耗尽" value="retry_exhausted" />
            <el-option label="熔断中" value="circuit_open" />
            <el-option label="已跳过" value="skipped" />
          </el-select>
          <el-button size="small" type="primary" data-testid="webhook-deliveries-search" @click="refreshDeliveries">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="deliveriesError" />

        <el-table
          v-loading="deliveriesLoading"
          :data="deliveryRows"
          stripe
          border
          size="small"
          data-testid="webhook-deliveries-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="插件" width="140" />
          <el-table-column prop="hook_name" label="Hook" width="220" />
          <el-table-column label="状态" width="160">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-delivery-status" />
            </template>
          </el-table-column>
          <el-table-column label="尝试次数" width="120">
            <template #default="{ row }">
              <span data-testid="webhook-delivery-attempt">{{ row.attempt }}/{{ row.max_attempts }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="next_retry_at" label="下次重试" width="170" />
          <el-table-column prop="response_status" label="响应" width="90" />
          <el-table-column prop="retry_reason" label="重试原因" width="140" />
          <el-table-column prop="error_message" label="错误信息" min-width="220" show-overflow-tooltip />
          <el-table-column label="操作" width="140" fixed="right">
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
          v-if="!deliveriesLoading && deliveryRows.length === 0 && !deliveriesError"
          description="暂无投递记录"
          testid="webhook-deliveries-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="deliveriesPage.page || 1"
            :page-size="deliveriesPage.page_size || 20"
            :total="deliveriesPage.total || 0"
            @current-change="onDeliveryPageChange"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="异常处理" name="exceptions">
        <div class="webhook-overview-grid">
          <el-card shadow="never">
            <template #header>需要处理</template>
            <ul class="overview-list">
              <li>等待重试：检查远端服务是否恢复，必要时手动重试。</li>
              <li>熔断中：确认目标服务和签名配置，恢复后手动关闭熔断。</li>
              <li>认证失败：进入高级治理检查 Webhook 密钥或回调 Token。</li>
            </ul>
            <div class="overview-actions">
              <el-button size="small" type="primary" :loading="retryDueLoading" @click="retryDue">扫描到期重试</el-button>
              <el-button size="small" plain @click="tab = 'secrets'">检查 Webhook 密钥</el-button>
              <el-button size="small" plain @click="tab = 'callback_tokens'">检查回调 Token</el-button>
            </div>
          </el-card>
        </div>

        <PluginErrorAlert :message="deliveriesError || circuitsError" />

        <el-table
          v-loading="deliveriesLoading"
          :data="retryRows"
          stripe
          border
          size="small"
          data-testid="webhook-retry-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="插件" width="140" />
          <el-table-column prop="hook_name" label="Hook" width="220" />
          <el-table-column label="状态" width="160">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-retry-status" />
            </template>
          </el-table-column>
          <el-table-column label="尝试次数" width="120">
            <template #default="{ row }">{{ row.attempt }}/{{ row.max_attempts }}</template>
          </el-table-column>
          <el-table-column prop="next_retry_at" label="下次重试" width="170" />
          <el-table-column prop="retry_reason" label="重试原因" min-width="180" />
          <el-table-column prop="error_message" label="错误信息" min-width="220" show-overflow-tooltip />
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="warning" plain :disabled="row.status === 'success'" @click="manualRetry(row)">手动重试</el-button>
            </template>
          </el-table-column>
        </el-table>

        <PluginEmptyState
          v-if="!deliveriesLoading && retryRows.length === 0 && !deliveriesError"
          description="暂无等待重试的投递记录"
          testid="webhook-retry-empty"
        />

        <el-table
          v-loading="circuitsLoading"
          :data="openCircuitRows"
          stripe
          border
          size="small"
          class="mt"
          data-testid="webhook-circuits-table"
        >
          <el-table-column prop="plugin_code" label="插件" width="160" />
          <el-table-column prop="target_url" label="目标 URL" min-width="260" show-overflow-tooltip />
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-circuit-status" />
            </template>
          </el-table-column>
          <el-table-column prop="failure_count" label="失败次数" width="110" />
          <el-table-column prop="last_error_message" label="最近错误" min-width="220" show-overflow-tooltip />
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="success" plain :disabled="row.status === 'closed'" @click="closeCircuit(row)">手动恢复</el-button>
              <el-button size="small" type="danger" plain @click="openCircuit(row)">手动熔断</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="高级治理" name="advanced">
        <div class="webhook-overview-grid">
          <el-card shadow="never">
            <template #header>低频治理入口</template>
            <p class="muted">事件、密钥、Token、回调请求和原始技术字段仍可访问，但不再默认占据 Webhook 主视图。</p>
            <div class="overview-actions">
              <el-button size="small" type="primary" plain @click="tab = 'events'">事件通知</el-button>
              <el-button size="small" type="primary" plain @click="tab = 'secrets'">Webhook 密钥</el-button>
              <el-button size="small" type="primary" plain @click="tab = 'callback_tokens'">回调 Token</el-button>
              <el-button size="small" type="primary" plain @click="tab = 'callback_requests'">回调请求</el-button>
              <el-button size="small" type="primary" plain @click="tab = 'circuits'">熔断状态</el-button>
            </div>
            <el-alert
              class="mt"
              type="warning"
              show-icon
              :closable="false"
              title="高级治理不会展示 Secret、Token、Authorization Header 或敏感 payload 明文。"
            />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="事件通知" name="events">
        <PluginFilterBar title="Webhook 事件" tip="事件记录用于追踪投递链路，不展示敏感 payload 明文" testid="webhook-events-filter">
          <template #actions>
            <el-button size="small" @click="refreshEvents">刷新</el-button>
          </template>
          <el-input v-model="eventFilters.plugin_code" size="small" placeholder="插件编码" style="width: 160px" />
          <el-input v-model="eventFilters.hook_name" size="small" placeholder="Hook 名称" style="width: 200px" />
          <el-select v-model="eventFilters.status" size="small" placeholder="状态" style="width: 180px">
            <el-option label="全部" value="all" />
            <el-option label="待处理" value="pending" />
            <el-option label="投递中" value="delivering" />
            <el-option label="已投递" value="delivered" />
            <el-option label="失败" value="failed" />
            <el-option label="熔断中" value="circuit_open" />
            <el-option label="已跳过" value="skipped" />
          </el-select>
          <el-input v-model="eventFilters.community_id" size="small" placeholder="子站 ID" style="width: 140px" />
          <el-button size="small" type="primary" data-testid="webhook-events-search" @click="refreshEvents">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="eventsError" />

        <el-table
          v-loading="eventsLoading"
          :data="eventRows"
          stripe
          border
          size="small"
          data-testid="webhook-events-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="event_id" label="事件 ID" min-width="220" show-overflow-tooltip />
          <el-table-column prop="plugin_code" label="插件" width="140" />
          <el-table-column prop="hook_name" label="Hook" width="220" />
          <el-table-column prop="community_id" label="子站" width="110" />
          <el-table-column label="状态" width="160">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-event-status" />
            </template>
          </el-table-column>
          <el-table-column prop="occurred_at" label="发生时间" width="170" />
          <el-table-column prop="created_at" label="创建时间" width="170" />
        </el-table>

        <PluginEmptyState
          v-if="!eventsLoading && eventRows.length === 0 && !eventsError"
          description="暂无 Webhook 事件"
          testid="webhook-events-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="eventsPage.page || 1"
            :page-size="eventsPage.page_size || 20"
            :total="eventsPage.total || 0"
            @current-change="onEventPageChange"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="Webhook 密钥" name="secrets">
        <PluginFilterBar title="Webhook 密钥" tip="密钥明文只会在创建/轮换时展示一次；表格只显示 secret_ref。" testid="webhook-secrets-filter">
          <template #actions>
            <el-button size="small" @click="refreshSecrets">刷新</el-button>
            <el-button size="small" type="primary" data-testid="webhook-secret-create" @click="openCreateSecret">
              创建密钥
            </el-button>
          </template>
          <el-input v-model="secretFilters.plugin_code" size="small" placeholder="插件编码" style="width: 180px" />
          <el-input v-model="secretFilters.secret_ref" size="small" placeholder="密钥引用 secret_ref" style="width: 220px" />
          <el-select v-model="secretFilters.status" size="small" placeholder="状态" style="width: 180px">
            <el-option label="全部" value="all" />
            <el-option label="启用中" value="active" />
            <el-option label="上一版本" value="previous" />
            <el-option label="已停用" value="disabled" />
            <el-option label="已吊销" value="revoked" />
            <el-option label="已过期" value="expired" />
          </el-select>
          <el-button size="small" type="primary" data-testid="webhook-secrets-search" @click="refreshSecrets">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="secretsError" />

        <el-table
          v-loading="secretsLoading"
          :data="secretRows"
          stripe
          border
          size="small"
          data-testid="webhook-secrets-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="插件" width="160" />
          <el-table-column prop="secret_ref" label="密钥引用" min-width="220" show-overflow-tooltip />
          <el-table-column prop="target_url" label="目标 URL" min-width="260" show-overflow-tooltip />
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="webhook-secret-status" />
            </template>
          </el-table-column>
          <el-table-column prop="grace_until" label="宽限期至" width="170" />
          <el-table-column prop="last_used_at" label="最近使用" width="170" />
          <el-table-column label="操作" width="320" fixed="right">
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
          v-if="!secretsLoading && secretRows.length === 0 && !secretsError"
          description="暂无 Webhook 密钥"
          testid="webhook-secrets-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="secretsPage.page || 1"
            :page-size="secretsPage.page_size || 20"
            :total="secretsPage.total || 0"
            @current-change="onSecretPageChange"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="回调 Token" name="callback_tokens">
        <PluginFilterBar title="回调 Token" tip="回调 Token 不等于管理员权限；明文只会在创建/轮换时展示一次。" testid="callback-tokens-filter">
          <template #actions>
            <el-button size="small" @click="refreshCallbackTokens">刷新</el-button>
            <el-button size="small" type="primary" data-testid="callback-token-create" @click="openCreateCallbackToken">
              创建回调 Token
            </el-button>
          </template>
          <el-input v-model="callbackTokenFilters.plugin_code" size="small" placeholder="插件编码" style="width: 180px" />
          <el-select v-model="callbackTokenFilters.status" size="small" placeholder="状态" style="width: 180px">
            <el-option label="全部" value="all" />
            <el-option label="启用中" value="active" />
            <el-option label="已停用" value="disabled" />
            <el-option label="已吊销" value="revoked" />
            <el-option label="已过期" value="expired" />
          </el-select>
          <el-input v-model="callbackTokenFilters.scope" size="small" placeholder="权限范围，例如 config.read" style="width: 220px" />
          <el-button size="small" type="primary" data-testid="callback-tokens-search" @click="refreshCallbackTokens">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="callbackTokensError" />

        <el-table
          v-loading="callbackTokensLoading"
          :data="callbackTokenRows"
          stripe
          border
          size="small"
          data-testid="callback-tokens-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="plugin_code" label="插件" width="160" />
          <el-table-column prop="token_ref" label="Token 引用" min-width="220" show-overflow-tooltip />
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="callback-token-status" />
            </template>
          </el-table-column>
          <el-table-column prop="expires_at" label="过期时间" width="170" />
          <el-table-column prop="last_used_at" label="最近使用" width="170" />
          <el-table-column prop="last_used_ip" label="最近 IP" width="140" />
          <el-table-column prop="scopes_json" label="权限范围" min-width="220" show-overflow-tooltip />
          <el-table-column prop="community_scope_json" label="子站范围" min-width="220" show-overflow-tooltip />
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button
                size="small"
                type="primary"
                plain
                data-testid="callback-token-rotate"
                :disabled="row.status === 'revoked'"
                @click="rotateCallbackToken(row)"
              >
                轮换
              </el-button>
              <el-button
                size="small"
                type="warning"
                plain
                data-testid="callback-token-disable"
                :disabled="row.status === 'disabled' || row.status === 'revoked'"
                @click="disableCallbackToken(row)"
              >
                禁用
              </el-button>
              <el-button
                size="small"
                type="success"
                plain
                data-testid="callback-token-enable"
                :disabled="row.status !== 'disabled'"
                @click="enableCallbackToken(row)"
              >
                恢复
              </el-button>
              <el-button
                size="small"
                type="danger"
                plain
                data-testid="callback-token-revoke"
                :disabled="row.status === 'revoked'"
                @click="revokeCallbackToken(row)"
              >
                吊销
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <PluginEmptyState
          v-if="!callbackTokensLoading && callbackTokenRows.length === 0 && !callbackTokensError"
          description="暂无回调 Token"
          testid="callback-tokens-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="callbackTokensPage.page || 1"
            :page-size="callbackTokensPage.page_size || 20"
            :total="callbackTokensPage.total || 0"
            @current-change="onCallbackTokenPageChange"
          />
        </div>

        <el-dialog v-model="createCallbackTokenDialogVisible" title="创建回调 Token" width="620px">
          <el-form :model="createCallbackTokenForm" label-width="120px">
            <el-form-item label="插件编码">
              <el-input v-model="createCallbackTokenForm.plugin_code" placeholder="official_announcement" />
            </el-form-item>
            <el-form-item label="名称">
              <el-input v-model="createCallbackTokenForm.name" placeholder="公告插件回调 Token" />
            </el-form-item>
            <el-form-item label="权限范围">
              <el-select v-model="createCallbackTokenForm.scopes" multiple placeholder="选择权限范围" style="width: 100%">
                <el-option label="config.read" value="config.read" />
                <el-option label="audit.write" value="audit.write" />
              </el-select>
            </el-form-item>
            <el-form-item label="子站范围">
              <el-input v-model="createCallbackTokenForm.community_scope_text" placeholder="例如：1,2" />
            </el-form-item>
            <el-form-item label="过期时间">
              <el-input v-model="createCallbackTokenForm.expires_at" placeholder="RFC3339，例如：2027-01-01T00:00:00Z（可选）" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="createCallbackTokenDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="createCallbackTokenLoading" data-testid="callback-token-create-confirm" @click="submitCreateCallbackToken">创建</el-button>
          </template>
        </el-dialog>

        <el-dialog v-model="callbackTokenPlaintextDialogVisible" title="回调 Token（只展示一次）" width="560px">
          <div class="muted">请立即复制保存。关闭后无法再次查看。</div>
          <el-input v-model="callbackTokenPlaintext" type="textarea" :rows="3" readonly data-testid="callback-token-plaintext" />
          <template #footer>
            <el-button type="primary" @click="callbackTokenPlaintextDialogVisible = false">我已保存</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="回调请求" name="callback_requests">
        <PluginFilterBar title="回调请求" tip="插件服务回调受控 Core API 的请求记录（不保存 Token 明文）" testid="callback-requests-filter">
          <template #actions>
            <el-button size="small" @click="refreshCallbackRequests">刷新</el-button>
          </template>
          <el-input v-model="callbackRequestFilters.plugin_code" size="small" placeholder="插件编码" style="width: 180px" />
          <el-input v-model="callbackRequestFilters.token_ref" size="small" placeholder="Token 引用 token_ref" style="width: 220px" />
          <el-select v-model="callbackRequestFilters.status" size="small" placeholder="状态" style="width: 180px">
            <el-option label="全部" value="all" />
            <el-option label="已接受" value="accepted" />
            <el-option label="已拒绝" value="rejected" />
            <el-option label="失败" value="failed" />
          </el-select>
          <el-input v-model="callbackRequestFilters.request_id" size="small" placeholder="request_id" style="width: 220px" />
          <el-button size="small" type="primary" data-testid="callback-requests-search" @click="refreshCallbackRequests">查询</el-button>
        </PluginFilterBar>

        <PluginErrorAlert :message="callbackRequestsError" />

        <el-table
          v-loading="callbackRequestsLoading"
          :data="callbackRequestRows"
          stripe
          border
          size="small"
          data-testid="callback-requests-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="created_at" label="创建时间" width="170" />
          <el-table-column prop="plugin_code" label="插件" width="160" />
          <el-table-column prop="token_ref" label="Token 引用" min-width="220" show-overflow-tooltip />
          <el-table-column prop="api_path" label="API 路径" min-width="220" show-overflow-tooltip />
          <el-table-column prop="method" label="方法" width="90" />
          <el-table-column prop="scope_required" label="权限范围" width="140" />
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <PluginStatusTag :value="row.status" testid="callback-request-status" />
            </template>
          </el-table-column>
          <el-table-column prop="response_status" label="响应" width="90" />
          <el-table-column prop="error_code" label="错误码" width="160" />
          <el-table-column prop="error_message" label="错误信息" min-width="240" show-overflow-tooltip />
        </el-table>

        <PluginEmptyState
          v-if="!callbackRequestsLoading && callbackRequestRows.length === 0 && !callbackRequestsError"
          description="暂无回调请求"
          testid="callback-requests-empty"
        />

        <div class="pager">
          <el-pagination
            layout="prev, pager, next"
            :current-page="callbackRequestsPage.page || 1"
            :page-size="callbackRequestsPage.page_size || 20"
            :total="callbackRequestsPage.total || 0"
            @current-change="onCallbackRequestPageChange"
          />
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="createSecretDialogVisible" title="创建 Webhook 密钥" width="560px">
      <div class="muted" style="margin-bottom: 8px">
        密钥明文只会在创建成功后展示一次，请立即复制保存。
      </div>
      <el-form label-width="100px">
        <el-form-item label="插件编码">
          <el-input v-model="createSecretForm.plugin_code" data-testid="webhook-secret-form-plugin" />
        </el-form-item>
        <el-form-item label="目标 URL">
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

    <el-dialog v-model="secretOnceDialogVisible" title="Webhook 密钥（仅展示一次）" width="640px">
      <div class="muted" style="margin-bottom: 8px">关闭后将无法再次查看，请立即复制。</div>
      <el-input v-model="secretOnceValue" type="textarea" :rows="4" readonly data-testid="webhook-secret-once-value" />
      <template #footer>
        <el-button type="primary" data-testid="webhook-secret-once-close" @click="closeSecretOnceDialog">我已保存</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import * as admin from '@/api/admin';

import PluginFilterBar from './components/PluginFilterBar.vue';
import PluginEmptyState from './components/PluginEmptyState.vue';
import PluginErrorAlert from './components/PluginErrorAlert.vue';
import PluginStatusTag from './components/PluginStatusTag.vue';
import { confirmDanger } from './components/useDangerConfirm';

const route = useRoute();
const router = useRouter();

const normalizeMainTab = (value) => {
  const name = String(value || 'overview').replace(/_/g, '-');
  if (name === 'retry' || name === 'circuits') return 'exceptions';
  if (name === 'callback-tokens') return 'callback_tokens';
  if (name === 'callback-requests') return 'callback_requests';
  if (['overview', 'deliveries', 'exceptions', 'advanced', 'events', 'secrets', 'callback_tokens', 'callback_requests'].includes(name)) return name;
  return 'overview';
};

const tab = ref(normalizeMainTab(route.query.tab));
watch(tab, async (next) => {
  const normalized = normalizeMainTab(next);
  if (normalized !== next) {
    tab.value = normalized;
    return;
  }
  await router.replace({ query: { ...route.query, tab: next } });
});

watch(() => route.query.tab, (value) => {
  const next = normalizeMainTab(value);
  if (tab.value !== next) tab.value = next;
});

const events = ref({ items: [], pagination: { page: 1, page_size: 20, total: 0 } });
const eventsLoading = ref(false);
const eventsError = ref('');
const eventFilters = ref({
  plugin_code: String(route.query.evt_plugin_code || ''),
  hook_name: String(route.query.evt_hook_name || ''),
  status: String(route.query.evt_status || 'all'),
  community_id: String(route.query.evt_community_id || ''),
  page: Number(route.query.evt_page || 1),
  page_size: 20,
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

// ===== v1.7.7: callback tokens + requests =====
const callbackTokens = ref({ items: [], pagination: { page: 1, page_size: 20, total: 0 } });
const callbackTokensLoading = ref(false);
const callbackTokensError = ref('');
const callbackTokenFilters = ref({
  plugin_code: String(route.query.cbtk_plugin_code || ''),
  status: String(route.query.cbtk_status || 'all'),
  scope: String(route.query.cbtk_scope || ''),
  page: Number(route.query.cbtk_page || 1),
  page_size: 20,
});

const createCallbackTokenDialogVisible = ref(false);
const createCallbackTokenLoading = ref(false);
const createCallbackTokenForm = ref({
  plugin_code: '',
  name: '',
  scopes: ['config.read', 'audit.write'],
  community_scope_text: '',
  expires_at: '',
});
const callbackTokenPlaintextDialogVisible = ref(false);
const callbackTokenPlaintext = ref('');

const callbackRequests = ref({ items: [], pagination: { page: 1, page_size: 20, total: 0 } });
const callbackRequestsLoading = ref(false);
const callbackRequestsError = ref('');
const callbackRequestFilters = ref({
  plugin_code: String(route.query.cbr_plugin_code || ''),
  token_ref: String(route.query.cbr_token_ref || ''),
  status: String(route.query.cbr_status || 'all'),
  request_id: String(route.query.cbr_request_id || ''),
  page: Number(route.query.cbr_page || 1),
  page_size: 20,
});

const emptyPage = { page: 1, page_size: 20, total: 0 };
const normalizeListResponse = (value, fallbackPageSize = 20) => {
  const source = value && typeof value === 'object' ? value : {};
  const items = Array.isArray(source.items) ? source.items.filter(Boolean) : [];
  const pagination = source.pagination && typeof source.pagination === 'object' ? source.pagination : {};
  return {
    ...source,
    items,
    pagination: {
      page: Number(pagination.page || source.page || 1),
      page_size: Number(pagination.page_size || source.page_size || fallbackPageSize),
      total: Number(pagination.total ?? source.total ?? items.length),
    },
  };
};

const listRows = (state) => Array.isArray(state.value?.items) ? state.value.items.filter(Boolean) : [];
const listPage = (state) => state.value?.pagination || emptyPage;

const eventRows = computed(() => listRows(events));
const eventsPage = computed(() => listPage(events));
const deliveryRows = computed(() => listRows(deliveries));
const deliveriesPage = computed(() => listPage(deliveries));
const circuitRows = computed(() => listRows(circuits));
const circuitsPage = computed(() => listPage(circuits));
const openCircuitRows = computed(() => circuitRows.value.filter((row) => ['open', 'half_open', 'circuit_open'].includes(String(row.status || '').toLowerCase())));
const openCircuitCount = computed(() => openCircuitRows.value.length);
const secretRows = computed(() => listRows(secrets));
const secretsPage = computed(() => listPage(secrets));
const callbackTokenRows = computed(() => listRows(callbackTokens));
const callbackTokensPage = computed(() => listPage(callbackTokens));
const callbackRequestRows = computed(() => listRows(callbackRequests));
const callbackRequestsPage = computed(() => listPage(callbackRequests));
const retryRows = computed(() => deliveryRows.value.filter((row) => ['retry_scheduled', 'retry_exhausted'].includes(String(row.status || '').toLowerCase())));

async function refreshEvents() {
  eventsLoading.value = true;
  eventsError.value = '';
  try {
    const params = { ...eventFilters.value };
    await router.replace({
      query: {
        ...route.query,
        tab: tab.value,
        evt_plugin_code: params.plugin_code,
        evt_hook_name: params.hook_name,
        evt_status: params.status,
        evt_community_id: params.community_id,
        evt_page: params.page,
      },
    });
    const res = await admin.listWebhookEvents(params);
    events.value = normalizeListResponse(res, eventFilters.value.page_size);
  } catch (e) {
    eventsError.value = e?.message || '加载失败';
  } finally {
    eventsLoading.value = false;
  }
}

function onEventPageChange(page) {
  eventFilters.value.page = page;
  refreshEvents();
}

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
    secrets.value = normalizeListResponse(res, secretFilters.value.page_size);
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

function openCreateCallbackToken() {
  createCallbackTokenForm.value = {
    plugin_code: callbackTokenFilters.value.plugin_code || '',
    name: '',
    scopes: ['config.read', 'audit.write'],
    community_scope_text: '',
    expires_at: '',
  };
  createCallbackTokenDialogVisible.value = true;
}

function parseCommunityScope(text) {
  return String(text || '')
    .split(',')
    .map((x) => Number(String(x).trim()))
    .filter((n) => Number.isFinite(n) && n > 0);
}

async function submitCreateCallbackToken() {
  createCallbackTokenLoading.value = true;
  try {
    const payload = {
      plugin_code: createCallbackTokenForm.value.plugin_code,
      name: createCallbackTokenForm.value.name,
      scopes: createCallbackTokenForm.value.scopes || [],
      community_scope: parseCommunityScope(createCallbackTokenForm.value.community_scope_text),
      expires_at: createCallbackTokenForm.value.expires_at || '',
    };
    const res = await admin.createPluginCallbackToken(payload);
    callbackTokenPlaintext.value = res?.token || '';
    callbackTokenPlaintextDialogVisible.value = true;
    createCallbackTokenDialogVisible.value = false;
    if (tab.value === 'callback_tokens') await refreshCallbackTokens();
  } finally {
    createCallbackTokenLoading.value = false;
  }
}

async function refreshCallbackTokens() {
  callbackTokensLoading.value = true;
  callbackTokensError.value = '';
  try {
    const params = { ...callbackTokenFilters.value };
    await router.replace({
      query: {
        ...route.query,
        tab: tab.value,
        cbtk_plugin_code: params.plugin_code,
        cbtk_status: params.status,
        cbtk_scope: params.scope,
        cbtk_page: params.page,
      },
    });
    const res = await admin.listPluginCallbackTokens(params);
    callbackTokens.value = normalizeListResponse(res, callbackTokenFilters.value.page_size);
  } catch (e) {
    callbackTokensError.value = e?.message || '加载失败';
  } finally {
    callbackTokensLoading.value = false;
  }
}

function onCallbackTokenPageChange(page) {
  callbackTokenFilters.value.page = page;
  refreshCallbackTokens();
}

async function rotateCallbackToken(row) {
  await confirmDanger(`确认轮换回调 Token #${row.id}？新 Token 明文只展示一次，请立即保存。`, '轮换回调 Token', {
    confirmButtonText: '轮换',
    cancelButtonText: '取消',
  });
  const res = await admin.rotatePluginCallbackToken(row.id);
  callbackTokenPlaintext.value = res?.token || '';
  callbackTokenPlaintextDialogVisible.value = true;
  await refreshCallbackTokens();
}

async function disableCallbackToken(row) {
  await confirmDanger(`确认禁用回调 Token #${row.id}？禁用后插件服务将无法回调 Core API。`, '禁用回调 Token', {
    confirmButtonText: '禁用',
    cancelButtonText: '取消',
  });
  await admin.disablePluginCallbackToken(row.id);
  await refreshCallbackTokens();
}

async function enableCallbackToken(row) {
  await confirmDanger(`确认恢复回调 Token #${row.id}？`, '恢复回调 Token', { confirmButtonText: '恢复', cancelButtonText: '取消' });
  await admin.enablePluginCallbackToken(row.id);
  await refreshCallbackTokens();
}

async function revokeCallbackToken(row) {
  await confirmDanger(`确认吊销回调 Token #${row.id}？吊销后将立即失效且不可恢复。`, '吊销回调 Token', {
    confirmButtonText: '吊销',
    cancelButtonText: '取消',
  });
  await admin.revokePluginCallbackToken(row.id, {});
  await refreshCallbackTokens();
}

async function refreshCallbackRequests() {
  callbackRequestsLoading.value = true;
  callbackRequestsError.value = '';
  try {
    const params = { ...callbackRequestFilters.value };
    await router.replace({
      query: {
        ...route.query,
        tab: tab.value,
        cbr_plugin_code: params.plugin_code,
        cbr_token_ref: params.token_ref,
        cbr_status: params.status,
        cbr_request_id: params.request_id,
        cbr_page: params.page,
      },
    });
    const res = await admin.listPluginCallbackRequests(params);
    callbackRequests.value = normalizeListResponse(res, callbackRequestFilters.value.page_size);
  } catch (e) {
    callbackRequestsError.value = e?.message || '加载失败';
  } finally {
    callbackRequestsLoading.value = false;
  }
}

function onCallbackRequestPageChange(page) {
  callbackRequestFilters.value.page = page;
  refreshCallbackRequests();
}

async function rotateSecret(row) {
  await confirmDanger(`确认轮换 Webhook 密钥 #${row.id}？轮换后新密钥明文只展示一次，请立即保存。`, '轮换 Webhook 密钥', { confirmButtonText: '轮换', cancelButtonText: '取消' });
  const res = await admin.rotateWebhookSecret(row.id);
  secretOnceValue.value = res?.secret || '';
  secretOnceDialogVisible.value = true;
  await refreshSecrets();
}

async function disableSecret(row) {
  await confirmDanger(`确认禁用 Webhook 密钥 #${row.id}？禁用后将无法用于签名发送。`, '禁用 Webhook 密钥', { confirmButtonText: '禁用', cancelButtonText: '取消' });
  await admin.disableWebhookSecret(row.id);
  await refreshSecrets();
}

async function enableSecret(row) {
  await confirmDanger(`确认恢复 Webhook 密钥 #${row.id}？恢复后将成为启用中，并可能停用同目标 URL 的其他启用密钥。`, '恢复 Webhook 密钥', { confirmButtonText: '恢复', cancelButtonText: '取消' });
  await admin.enableWebhookSecret(row.id);
  await refreshSecrets();
}

async function revokeSecret(row) {
  await confirmDanger(`确认吊销 Webhook 密钥 #${row.id}？吊销后将立即失效且不可恢复。`, '吊销 Webhook 密钥', { confirmButtonText: '吊销', cancelButtonText: '取消' });
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
    deliveries.value = normalizeListResponse(res, deliveryFilters.value.page_size);
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
    circuits.value = normalizeListResponse(res, circuitFilters.value.page_size);
  } catch (e) {
    circuitsError.value = e?.message || '加载失败';
  } finally {
    circuitsLoading.value = false;
  }
}

async function manualRetry(row) {
  await confirmDanger(`确认重试投递记录 #${row.id}？`, '手动重试', { confirmButtonText: '重试', cancelButtonText: '取消' });
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
    if (tab.value === 'deliveries' || tab.value === 'retry') await refreshDeliveries();
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
  tab.value = normalizeMainTab(tab.value);
  if (tab.value === 'overview') {
    await refreshDeliveries();
    await refreshCircuits();
    return;
  }
  if (tab.value === 'events') await refreshEvents();
  else if (tab.value === 'secrets') await refreshSecrets();
  else if (tab.value === 'callback_tokens') await refreshCallbackTokens();
  else if (tab.value === 'callback_requests') await refreshCallbackRequests();
  else if (tab.value === 'exceptions') {
    await refreshDeliveries();
    await refreshCircuits();
  }
  else await refreshDeliveries();
});

watch(tab, async (next) => {
  const normalized = normalizeMainTab(next);
  if (normalized !== next) {
    tab.value = normalized;
    return;
  }
  if (next === 'overview') {
    await refreshDeliveries();
    await refreshCircuits();
    return;
  }
  if (next === 'events') await refreshEvents();
  else if (next === 'secrets') await refreshSecrets();
  else if (next === 'callback_tokens') await refreshCallbackTokens();
  else if (next === 'callback_requests') await refreshCallbackRequests();
  else if (next === 'deliveries') await refreshDeliveries();
  else if (next === 'exceptions') {
    deliveryFilters.value.status = 'all';
    await refreshDeliveries();
    await refreshCircuits();
  }
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

.webhook-overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 12px;
  margin-top: 8px;
}

.overview-list {
  margin: 0;
  padding-left: 18px;
  color: #4b5563;
  line-height: 1.8;
}

.overview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.webhook-stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
  margin-bottom: 12px;
}

.webhook-stat {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
  padding: 10px;
  text-align: left;
  cursor: pointer;
}

.webhook-stat span {
  display: block;
  color: #64748b;
  font-size: 12px;
}

.webhook-stat strong {
  display: block;
  margin-top: 4px;
  color: #0f172a;
  font-size: 22px;
  line-height: 28px;
}

.mt {
  margin-top: 12px;
}
</style>
