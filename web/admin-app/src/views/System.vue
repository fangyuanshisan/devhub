<template>
  <div class="system-page">
    <AdminPageHeader
      title="敏感配置与运行安全状态"
      description="按总览、当前生效配置、安全与密钥、外部服务策略、SecretCenter 和配置审计组织入口。root key 仍只读，token / secret / Authorization 不回显。"
      status="healthy"
      testid="system-security-page-header"
    >
      <template #actions>
        <el-button size="small" type="primary" plain data-testid="system-copy-diagnostics" @click="copyRedactedDiagnostics">复制脱敏诊断信息</el-button>
        <el-tag type="info" effect="plain">受控配置</el-tag>
      </template>
    </AdminPageHeader>

    <el-card shadow="never" class="security-status-card">
      <div v-if="sensitiveError" class="mt12">
        <el-alert type="error" show-icon :closable="false" :title="sensitiveError" />
      </div>

      <el-tabs v-else v-model="activeSection" class="system-tabs">
        <el-tab-pane label="总览" name="overview">
          <div class="dh-admin-metrics">
            <AdminMetricCard
              v-for="card in summaryCards"
              :key="card.key"
              :label="card.title"
              :value="card.statusText"
              :status="card.statusKey"
              :help="card.description"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="当前生效配置" name="effective">
          <div class="section-head">
            <div>
              <div class="section-title">当前生效配置</div>
              <div class="section-desc">这里展示当前运行时实际生效的 external_service、HTTP Allowlist 和 SecretCenter 脱敏元数据；非敏感字段明文展示，敏感字段只展示引用、状态和脱敏后缀。</div>
            </div>
            <el-button size="small" type="primary" @click="copyRedactedDiagnostics">复制脱敏诊断信息</el-button>
          </div>
          <div class="dh-admin-metrics mt12" data-testid="system-effective-metrics">
            <AdminMetricCard label="系统版本" :value="safeText(effectiveConfig?.devhub_version || effectiveConfig?.version, 'unknown')" />
            <AdminMetricCard label="store mode" :value="safeText(effectiveConfig?.store_mode, 'unknown')" />
            <AdminMetricCard label="root key 状态" :value="keyringState().text" :status="keyringState().type === 'success' ? 'healthy' : keyringState().type === 'warning' ? 'warning' : 'error'" />
            <AdminMetricCard label="SecretCenter 状态" :value="secretCenterState().text" :status="secretCenterState().type === 'success' ? 'healthy' : secretCenterState().type === 'warning' ? 'warning' : 'error'" />
            <AdminMetricCard label="HTTP Allowlist" :value="`${effectiveAllowlistRows.length} 条`" :status="effectiveAllowlistRows.length ? 'healthy' : 'warning'" />
          </div>
          <section class="plain-section mt12" data-testid="system-effective-runtime">
            <div class="card-head">
              <div>
                <div class="section-title">基础运行信息</div>
                <div class="section-desc">当前版本、存储模式、生成时间、root key 和 SecretCenter 状态均为只读信息。</div>
              </div>
              <el-tag effect="plain">{{ safeText(effectiveConfig?.generated_at) }}</el-tag>
            </div>
            <dl class="runtime-grid">
              <div><dt>devhub_version</dt><dd>{{ safeText(effectiveConfig?.devhub_version || effectiveConfig?.version, 'unknown') }}</dd></div>
              <div><dt>store_mode</dt><dd>{{ safeText(effectiveConfig?.store_mode, 'unknown') }}</dd></div>
              <div><dt>root_key_configured</dt><dd>{{ yesNo(rootKeyConfigured) }}</dd></div>
              <div><dt>root_key_source</dt><dd>{{ safeText(effectiveConfig?.root_key_status?.source || keyring?.source, 'env') }}</dd></div>
              <div><dt>secret_center_available</dt><dd>{{ yesNo(secretCenterAvailable) }}</dd></div>
              <div><dt>http_allowlist_source</dt><dd>{{ allowlistSourceName(effectiveConfig?.http_allowlist_source) }}</dd></div>
            </dl>
          </section>
          <AdminInlineHint
            class="mt12"
            type="info"
            description="诊断内容已由后端脱敏，可用于排障；其中不包含 token、secret、Authorization、root key 或密文字段。"
          />
          <section v-if="effectiveNextSteps.length" class="plain-section mt12" data-testid="system-effective-next-steps">
            <div class="section-title">下一步建议</div>
            <div class="next-step-list">
              <div v-for="step in effectiveNextSteps" :key="step" class="next-step-item">{{ step }}</div>
            </div>
          </section>
          <AdminTechnicalDetails class="mt12" :blocks="effectiveTechnicalBlocks" testid="system-effective-technical-details" />
          <section class="plain-section mt12" data-testid="system-effective-secretcenter">
            <div class="card-head">
              <div>
                <div class="section-title">SecretCenter 当前状态</div>
                <div class="section-desc">只展示引用层状态、namespace 计数和跳转入口；root key 不在后台保存、生成或复制。</div>
              </div>
              <AdminStatusTag :value="effectiveSecretCenter.status || 'warning'" />
            </div>
            <dl class="runtime-grid">
              <div><dt>secret_ref_count</dt><dd>{{ safeText(effectiveSecretCenter.secret_ref_count, '0') }}</dd></div>
              <div><dt>external_service</dt><dd>{{ safeText(effectiveSecretCenter.namespace_counts?.external_service, '0') }}</dd></div>
              <div><dt>webhook</dt><dd>{{ safeText(effectiveSecretCenter.namespace_counts?.webhook, '0') }}</dd></div>
              <div><dt>callback</dt><dd>{{ safeText(effectiveSecretCenter.namespace_counts?.callback, '0') }}</dd></div>
            </dl>
            <div class="quick-actions mt12">
              <el-button size="small" @click="activeSection = 'secret-center'">去 SecretCenter</el-button>
              <el-button size="small" @click="activeSection = 'keys'">root key 状态</el-button>
              <el-button size="small" @click="openAudit('secret_center')">Secret 审计</el-button>
            </div>
          </section>
          <section class="plain-section mt12" data-testid="system-effective-webhook-callback">
            <div class="card-head">
              <div>
                <div class="section-title">Webhook / Callback 安全摘要</div>
                <div class="section-desc">这里仅展示 Webhook Secret 与 Callback Token 的状态计数和治理入口，不展示明文、hash 或密文。</div>
              </div>
              <el-tag effect="plain">disabled/revoked {{ safeText(webhookSecurity.disabled_or_revoked_count, '0') }}</el-tag>
            </div>
            <dl class="runtime-grid">
              <div><dt>webhook_secret_total</dt><dd>{{ safeText(webhookSecurity.webhook_secret_total, '0') }}</dd></div>
              <div><dt>active_webhook_secrets</dt><dd>{{ safeText(webhookSecurity.active_webhook_secrets, '0') }}</dd></div>
              <div><dt>callback_token_total</dt><dd>{{ safeText(webhookSecurity.callback_token_total, '0') }}</dd></div>
              <div><dt>active_callback_tokens</dt><dd>{{ safeText(webhookSecurity.active_callback_tokens, '0') }}</dd></div>
              <div><dt>last_webhook_secret_used_at</dt><dd>{{ safeText(webhookSecurity.last_webhook_secret_used_at) }}</dd></div>
              <div><dt>last_callback_token_used_at</dt><dd>{{ safeText(webhookSecurity.last_callback_token_used_at) }}</dd></div>
            </dl>
            <div class="quick-actions mt12">
              <el-button size="small" @click="goWebhookSecurity('secrets')">Webhook 密钥</el-button>
              <el-button size="small" @click="goWebhookSecurity('callback_tokens')">Callback Token</el-button>
              <el-button size="small" @click="goWebhookSecurity('callback_requests')">回调请求</el-button>
              <el-button size="small" @click="openAudit('', '', 'webhook callback token_ref secret_ref')">查看审计</el-button>
            </div>
          </section>
          <section class="plain-section mt12">
            <div class="card-head">
              <div>
                <div class="section-title">external_service HTTP Allowlist</div>
                <div class="section-desc">系统默认、环境变量来源、后台配置来源与最终生效列表均为非敏感配置，可明文展示。</div>
              </div>
              <el-tag>{{ effectiveAllowlistRows.length }} 条生效</el-tag>
            </div>
            <div class="allowlist-groups">
              <div v-for="group in effectiveAllowlistGroups" :key="group.key" class="allowlist-group">
                <div class="allowlist-group-title">{{ group.title }}</div>
                <div v-if="group.items.length" class="allowlist-chips">
                  <el-tag v-for="item in group.items" :key="`${group.key}-${item.origin}`" effect="plain">{{ item.origin }}</el-tag>
                </div>
                <div v-else class="muted">暂无</div>
              </div>
            </div>
            <el-table class="mt12" :data="effectiveAllowlistRows" stripe border>
              <el-table-column prop="origin" label="Origin" min-width="240" />
              <el-table-column label="来源" width="130">
                <template #default="{ row }">{{ allowlistSourceName(row.source) }}</template>
              </el-table-column>
              <el-table-column label="生效来源" width="150">
                <template #default="{ row }">{{ allowlistSourceName(row.source) || '系统默认' }}</template>
              </el-table-column>
              <el-table-column label="操作" width="260" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="activeSection = 'external'">去外部服务策略</el-button>
                  <el-button link type="primary" @click="copyText(row.origin, '已复制 origin')">复制 origin</el-button>
                  <el-button link type="primary" @click="copyRedactedDiagnostics">复制诊断</el-button>
                  <el-button link type="primary" @click="openAudit('external_service.http_allowlist', row.origin)">查看审计</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
          <section class="plain-section mt12">
            <div class="card-head">
              <div>
                <div class="section-title">插件 external_service 运行配置</div>
                <div class="section-desc">endpoint_url、health_check_path、timeout_ms、failure_policy 明文展示；token 只展示 token_ref、状态、key_id 和脱敏值。</div>
              </div>
              <el-tag>{{ effectiveExternalServices.length }} 个服务</el-tag>
            </div>
            <el-table class="mt12" :data="effectiveExternalServices" stripe border>
              <el-table-column prop="plugin_code" label="插件" min-width="150">
                <template #default="{ row }">
                  <div>{{ row.plugin_name || row.plugin_code }}</div>
                  <div class="muted">{{ row.plugin_code }}</div>
                </template>
              </el-table-column>
              <el-table-column prop="endpoint_url" label="endpoint_url" min-width="220">
                <template #default="{ row }">{{ safeText(row.endpoint_url) }}</template>
              </el-table-column>
              <el-table-column prop="health_check_path" label="health_check_path" width="160">
                <template #default="{ row }">{{ safeText(row.health_check_path) }}</template>
              </el-table-column>
              <el-table-column prop="enabled" label="启用" width="80">
                <template #default="{ row }">{{ yesNo(row.enabled) }}</template>
              </el-table-column>
              <el-table-column prop="timeout_ms" label="timeout_ms" width="120">
                <template #default="{ row }">{{ safeText(row.timeout_ms) }}</template>
              </el-table-column>
              <el-table-column prop="failure_policy" label="failure_policy" width="140">
                <template #default="{ row }">{{ safeText(row.failure_policy) }}</template>
              </el-table-column>
              <el-table-column prop="auth_type" label="auth_type" width="110">
                <template #default="{ row }">{{ safeText(row.auth_type, 'none') }}</template>
              </el-table-column>
              <el-table-column prop="endpoint_origin" label="endpoint origin / allowlist" min-width="240">
                <template #default="{ row }">
                  <div class="mono-wrap">{{ safeText(row.endpoint_origin || row.endpoint_url) }}</div>
                  <div class="muted">{{ allowlistSourceName(row.http_allowlist_source) }} / {{ row.http_allowlist_matched ? '已匹配' : '未匹配' }}</div>
                  <div v-if="row.http_allowlist_message" class="muted">{{ row.http_allowlist_message }}</div>
                </template>
              </el-table-column>
              <el-table-column prop="config_source" label="配置来源" width="150">
                <template #default="{ row }">{{ configSourceLabel(row.config_source) }}</template>
              </el-table-column>
              <el-table-column prop="current_health" label="当前健康" width="120">
                <template #default="{ row }"><AdminStatusTag :value="safeText(row.current_health, 'unknown')" /></template>
              </el-table-column>
              <el-table-column prop="token_ref" label="token_ref" min-width="240">
                <template #default="{ row }">
                  <div class="secret-ref-cell">
                    <span class="secret-ref-text">{{ shortenRef(row.token_ref || '') || '未配置' }}</span>
                    <el-button v-if="row.token_ref" size="small" link type="primary" @click="copyText(row.token_ref, '已复制 token_ref')">复制</el-button>
                    <el-button v-if="row.token_ref" size="small" link type="primary" @click="openSecretByRef(row.token_ref)">查看 Secret</el-button>
                  </div>
                  <div v-if="row.token_namespace || row.token_name" class="muted">{{ safeText(row.token_namespace) }} / {{ safeText(row.token_name) }}</div>
                  <div class="muted">{{ row.token_message || '' }} 来源：{{ row.token_source || 'SecretCenter' }}</div>
                  <div v-if="row.next_steps?.length" class="muted">{{ row.next_steps[0] }}</div>
                </template>
              </el-table-column>
              <el-table-column prop="token_status" label="token 状态" width="120">
                <template #default="{ row }"><AdminStatusTag :value="row.token_status || 'unknown'" /></template>
              </el-table-column>
              <el-table-column prop="token_masked" label="脱敏值" width="130">
                <template #default="{ row }">{{ safeText(row.token_masked) }}</template>
              </el-table-column>
              <el-table-column prop="token_key_id" label="key_id" width="130">
                <template #default="{ row }">{{ safeText(row.token_key_id) }}</template>
              </el-table-column>
              <el-table-column prop="last_health_check_at" label="最近检查" width="170">
                <template #default="{ row }">{{ safeText(row.last_health_check_at) }}</template>
              </el-table-column>
              <el-table-column prop="last_success_at" label="最近成功" width="170">
                <template #default="{ row }">{{ safeText(row.last_success_at) }}</template>
              </el-table-column>
              <el-table-column prop="last_failure_at" label="最近失败" width="170">
                <template #default="{ row }">{{ safeText(row.last_failure_at) }}</template>
              </el-table-column>
              <el-table-column prop="last_error_summary" label="最近错误摘要" min-width="220">
                <template #default="{ row }">{{ safeText(row.last_error_summary) }}</template>
              </el-table-column>
              <el-table-column label="下一步" min-width="240">
                <template #default="{ row }">
                  <div v-if="row.next_steps?.length" class="next-step-list compact">
                    <div v-for="step in row.next_steps" :key="step" class="next-step-item">{{ step }}</div>
                  </div>
                  <span v-else class="muted">暂无阻断建议</span>
                </template>
              </el-table-column>
              <el-table-column label="排障入口" width="360" fixed="right">
                <template #default="{ row }">
                  <div class="quick-actions table-actions">
                    <el-button link type="primary" @click="goPluginExternalConfig(row.plugin_code)">去配置</el-button>
                    <el-button link type="primary" @click="runHealthCheck(row)">健康检查</el-button>
                    <el-button link type="primary" @click="goExternalExecutions(row.plugin_code)">运行记录</el-button>
                    <el-button v-if="row.token_ref" link type="primary" @click="openSecretByRef(row.token_ref)">Secret</el-button>
                    <el-button link type="primary" @click="openAudit('external_service', row.plugin_code)">审计</el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!effectiveExternalServices.length" description="暂无 external_service 运行配置" />
          </section>
        </el-tab-pane>

        <el-tab-pane label="安全与密钥" name="keys">
          <div class="status-grid single">
            <section class="status-panel">
              <div class="status-panel-head">
                <div>
                  <div class="status-panel-title">{{ keyringCard.title }}</div>
                  <div class="status-panel-desc">{{ keyringCard.description }}</div>
                </div>
                <el-tag :type="keyringCard.tagType" effect="light">{{ keyringCard.statusText }}</el-tag>
              </div>
              <dl class="status-fields">
                <div v-for="field in keyringCard.fields" :key="field.label" class="status-field">
                  <dt>{{ field.label }}</dt>
                  <dd>{{ field.value }}</dd>
                </div>
              </dl>
            </section>
          </div>
          <el-alert
            v-if="keyringWarnings.length"
            type="warning"
            show-icon
            class="mt12"
            :closable="false"
            :title="keyringWarnings[0]"
          />
          <el-alert
            type="info"
            show-icon
            class="mt12"
            :closable="false"
            title="root key 只能来自环境变量或外部 Secret 系统，后台不会保存、生成或修改。"
          />
          <div class="example-list mt12">
            <div v-for="example in envExampleBlocks" :key="example.title" class="code-panel">
              <div class="code-panel-head">
                <span>{{ example.title }}</span>
                <el-button size="small" link type="primary" @click="copyExample(example.code)">复制</el-button>
              </div>
              <pre><code>{{ example.code }}</code></pre>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="外部服务策略" name="external">
          <div class="status-grid single">
            <section class="status-panel">
              <div class="status-panel-head">
                <div>
                  <div class="status-panel-title">{{ allowlistCard.title }}</div>
                  <div class="status-panel-desc">{{ allowlistCard.description }}</div>
                </div>
                <el-tag :type="allowlistCard.tagType" effect="light">{{ allowlistCard.statusText }}</el-tag>
              </div>
              <dl class="status-fields">
                <div v-for="field in allowlistCard.fields" :key="field.label" class="status-field">
                  <dt>{{ field.label }}</dt>
                  <dd>{{ field.value }}</dd>
                </div>
              </dl>
            </section>
          </div>
          <div class="section-head mt16">
            <div>
              <div class="section-title">external_service HTTP Allowlist 管理</div>
              <div class="section-desc">系统默认和环境变量来源不可删除；后台配置来源可受控新增 / 删除。</div>
            </div>
            <el-button type="primary" size="small" @click="openAllowlistDialog">新增 Origin</el-button>
          </div>
          <el-alert
            type="warning"
            show-icon
            class="mt12"
            :closable="false"
            title="非 localhost HTTP 只建议用于本地开发或受控内网联调；生产环境建议使用 HTTPS。"
          />
          <div class="example-list mt12">
            <div class="code-panel">
              <div class="code-panel-head">
                <span>{{ externalAllowlistEnvExample.title }}</span>
                <el-button size="small" link type="primary" @click="copyExample(externalAllowlistEnvExample.code)">复制</el-button>
              </div>
              <pre><code>{{ externalAllowlistEnvExample.code }}</code></pre>
            </div>
          </div>
          <el-table class="mt12" :data="allowlistRows" stripe border>
            <el-table-column prop="origin" label="Origin" min-width="220" />
            <el-table-column label="来源" width="130">
              <template #default="{ row }">{{ allowlistSourceName(row.source) }}</template>
            </el-table-column>
            <el-table-column prop="usage" label="用途说明" min-width="240" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }"><el-tag size="small" type="success" effect="plain">{{ row.status || 'active' }}</el-tag></template>
            </el-table-column>
            <el-table-column label="更新时间" width="170">
              <template #default="{ row }">{{ row.updated_at || row.created_at || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="110" fixed="right">
              <template #default="{ row }">
                <el-button v-if="row.deletable" link type="danger" @click="removeAllowlist(row)">删除</el-button>
                <span v-else class="muted">不可删除</span>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="SecretCenter" name="secret-center">
          <div class="section-head">
            <div>
              <div class="section-title">敏感配置引用</div>
              <div class="section-desc">这里展示系统中已加密保存的 token、secret、Webhook 密钥等敏感配置引用。页面不会显示明文，只显示引用地址、所属业务和使用状态。</div>
            </div>
            <div class="section-actions">
              <el-button size="small" @click="loadSecrets">刷新列表</el-button>
              <el-button size="small" type="primary" plain @click="openAudit('secret_center')">查看审计</el-button>
            </div>
          </div>
          <section class="info-panel mt12">
            <div class="info-panel-title">SecretCenter 用来治理敏感配置引用</div>
            <div class="info-panel-desc">
              SecretCenter 用于加密保存运行时敏感配置，例如 external_service token、Webhook Secret、Callback Token。后台只显示 secret_ref / token_ref 引用，不显示真实明文；secret://... 是引用地址，不是敏感值本身。如果需要修改 external_service token，请到对应插件的 external_service 配置中轮换。
            </div>
          </section>
          <div class="status-grid single mt12">
            <section class="status-panel">
              <div class="status-panel-head">
                <div>
                  <div class="status-panel-title">{{ secretCenterCard.title }}</div>
                  <div class="status-panel-desc">{{ secretCenterCard.description }}</div>
                </div>
                <AdminStatusTag :value="secretCenterCard.statusKey" />
              </div>
              <dl class="status-fields">
                <div v-for="field in secretCenterCard.fields" :key="field.label" class="status-field">
                  <dt>{{ field.label }}</dt>
                  <dd>{{ field.value }}</dd>
                </div>
              </dl>
            </section>
          </div>
          <div class="filter-bar mt12">
            <el-select v-model="secretFilters.type" size="small" placeholder="类型" style="width: 160px">
              <el-option label="全部" value="all" />
              <el-option label="外部服务" value="external_service" />
              <el-option label="Webhook Secret" value="webhook" />
              <el-option label="Callback Token" value="callback" />
              <el-option label="插件配置敏感字段" value="plugin_config" />
              <el-option label="测试数据" value="test" />
              <el-option label="其他" value="other" />
            </el-select>
            <el-select v-model="secretFilters.status" size="small" placeholder="状态" style="width: 140px">
              <el-option label="全部" value="all" />
              <el-option label="正常" value="active" />
              <el-option label="已禁用" value="disabled" />
              <el-option label="已吊销" value="revoked" />
              <el-option label="未知" value="unknown" />
            </el-select>
            <el-input
              v-model="secretFilters.keyword"
              size="small"
              clearable
              placeholder="搜索引用 / 所属业务 / 名称 / key_id"
              class="secret-search"
            />
          </div>
          <el-alert
            v-if="secretTestCount"
            type="info"
            show-icon
            class="mt12"
            :closable="false"
            title="已识别到测试数据引用。测试数据会打上“测试”标签，不会自动删除；如需清理，请按插件包或测试数据清理说明单独处理。"
          />
          <el-table class="mt12" :data="filteredSecretRows" stripe border>
            <el-table-column label="名称 / 用途" min-width="220">
              <template #default="{ row }">
                <div>{{ safeText(row.display_name || row.description || row.name) }}</div>
                <div class="muted">{{ safeText(row.usage, '运行时敏感配置引用') }}</div>
              </template>
            </el-table-column>
            <el-table-column label="敏感配置引用" min-width="280">
              <template #default="{ row }">
                <div class="secret-ref-cell">
                  <el-tooltip :content="row.ref" placement="top">
                    <span class="secret-ref-text">{{ shortenRef(row.ref) }}</span>
                  </el-tooltip>
                  <el-tag v-if="isTestSecret(row)" size="small" type="warning" effect="plain">测试</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="所属业务" min-width="150">
              <template #default="{ row }">
                <div>{{ businessLabel(row) }}</div>
                <div v-if="businessHint(row)" class="muted">{{ businessHint(row) }}</div>
              </template>
            </el-table-column>
            <el-table-column label="类型" min-width="150">
              <template #default="{ row }">{{ secretTypeLabel(row) }}</template>
            </el-table-column>
            <el-table-column label="当前状态" width="110">
              <template #default="{ row }">
                <AdminStatusTag :value="normalizedSecretStatus(row.status)" />
              </template>
            </el-table-column>
            <el-table-column prop="key_id" label="加密密钥 ID" width="150">
              <template #default="{ row }">{{ safeText(row.key_id) }}</template>
            </el-table-column>
            <el-table-column prop="last_used_at" label="最近使用时间" width="170">
              <template #default="{ row }">{{ safeText(row.last_used_at) }}</template>
            </el-table-column>
            <el-table-column prop="rotated_at" label="最近轮换" width="170">
              <template #default="{ row }">{{ safeText(row.rotated_at) }}</template>
            </el-table-column>
            <el-table-column prop="masked_value" label="脱敏值" width="130">
              <template #default="{ row }">{{ safeText(row.masked_value) }}</template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" width="170">
              <template #default="{ row }">{{ safeText(row.updated_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="360" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openSecretDetail(row)">查看详情</el-button>
                <el-button link type="primary" @click="copySecretRef(row)">复制引用</el-button>
                <el-button link type="primary" @click="openSecretAudit(row)">查看审计</el-button>
                <el-dropdown trigger="click" @command="(command) => handleSecretAction(command, row)">
                  <el-button link type="primary">更多</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="source" :disabled="!canJumpSecretSource(row)">跳转到来源配置</el-dropdown-item>
                      <el-dropdown-item command="rotate" :disabled="!canRotateSecret(row)">轮换</el-dropdown-item>
                      <el-dropdown-item command="disable" :disabled="isTestSecret(row) || row.status === 'disabled' || row.status === 'revoked'">禁用</el-dropdown-item>
                      <el-dropdown-item command="revoke" :disabled="isTestSecret(row) || row.status === 'revoked'">吊销</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>
          <el-empty
            v-if="!filteredSecretRows.length"
            description="暂无敏感配置引用。当插件保存 external_service token、创建 Webhook Secret 或 Callback Token 后，这里会显示对应 secret_ref / token_ref。"
          />
        </el-tab-pane>

        <el-tab-pane label="配置审计" name="audit">
          <section class="plain-section">
            <div class="card-head">
              <div>
                <div class="section-title">admin_logs / 配置变更记录</div>
                <div class="section-desc">集中查看 SecretCenter、external_service HTTP allowlist 和敏感配置治理相关记录。</div>
              </div>
              <el-tag>{{ logTotal }} 条</el-tag>
            </div>
            <el-form :inline="true" :model="logQuery">
              <el-form-item><el-select v-model="logQuery.type" style="width: 130px"><el-option label="全部" value="all" /><el-option label="治理" value="audit" /><el-option label="运营" value="operation" /><el-option label="系统" value="system" /><el-option label="登录" value="login" /></el-select></el-form-item>
              <el-form-item><el-input v-model="logQuery.action" placeholder="动作" clearable /></el-form-item>
              <el-form-item><el-input v-model="logQuery.actor" placeholder="操作人" clearable /></el-form-item>
              <el-form-item><el-input v-model="logQuery.target" placeholder="对象 / secret_ref" clearable /></el-form-item>
              <el-form-item><el-input v-model="logQuery.metadata" placeholder="namespace / source_id / plugin_code" clearable /></el-form-item>
              <el-form-item><el-button @click="loadLogs">筛选</el-button></el-form-item>
            </el-form>
            <el-table :data="logList" stripe>
              <el-table-column prop="site" label="站点" width="100" />
              <el-table-column prop="actor" label="操作人" width="120" />
              <el-table-column label="身份" width="110"><template #default="{ row }">{{ actorTypeName(row.actor_type) }}</template></el-table-column>
              <el-table-column prop="role" label="角色" width="120" />
              <el-table-column prop="type" label="类型" width="90" />
              <el-table-column prop="action" label="动作" min-width="220" />
              <el-table-column prop="target" label="对象" min-width="180" />
              <el-table-column prop="created_at" label="时间" width="170" />
            </el-table>
            <el-pagination v-model:current-page="logQuery.page" v-model:page-size="logQuery.page_size" class="pager" layout="total, prev, pager, next" :total="logTotal" @change="loadLogs" />
          </section>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="allowlistDialog.visible" title="新增 HTTP Allowlist Origin" width="520px">
      <el-form label-width="90px">
        <el-form-item label="Origin">
          <el-input v-model="allowlistForm.origin" placeholder="http://172.17.0.1:18081" />
        </el-form-item>
        <el-form-item label="用途说明">
          <el-input v-model="allowlistForm.usage" type="textarea" :rows="3" placeholder="例如：本地 Docker receiver 联调" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="allowlistForm.risk_confirmed">
            我确认该 HTTP origin 仅用于本地开发或受控内网联调，生产环境优先使用 HTTPS。
          </el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="allowlistDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="allowlistDialog.saving" @click="submitAllowlist">确认新增</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="diagnosticDialog.visible" title="手动复制脱敏诊断信息" width="720px">
      <el-alert type="info" show-icon :closable="false" title="浏览器剪贴板不可用时，可从下方文本框手动复制。内容已经脱敏，不包含 token / secret / Authorization / root key / 密文字段。" />
      <el-input class="mt12" type="textarea" :rows="14" readonly :model-value="diagnosticDialog.text" />
      <template #footer>
        <el-button @click="diagnosticDialog.visible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="secretDetail.visible" title="敏感配置引用详情" size="620px">
      <template v-if="secretDetail.record">
        <el-alert
          v-if="isTestSecret(secretDetail.record)"
          class="mb12"
          type="warning"
          show-icon
          :closable="false"
          title="该引用被识别为测试 / fixture / seed 数据，危险操作默认隐藏；请到测试数据清理流程处理。"
        />
        <el-descriptions :column="1" border>
          <el-descriptions-item label="名称">{{ safeText(secretDetail.record.display_name || secretDetail.record.name) }}</el-descriptions-item>
          <el-descriptions-item label="namespace">{{ safeText(secretDetail.record.namespace) }}</el-descriptions-item>
          <el-descriptions-item label="name / key">{{ safeText(secretDetail.record.name) }}</el-descriptions-item>
          <el-descriptions-item label="敏感配置引用">
            <div class="secret-ref-cell">
              <el-tooltip :content="secretDetail.record.ref" placement="top">
                <span class="secret-ref-text">{{ secretDetail.record.ref }}</span>
              </el-tooltip>
              <el-button size="small" link type="primary" @click="copySecretRef(secretDetail.record)">复制引用</el-button>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="所属业务">{{ businessLabel(secretDetail.record) }}</el-descriptions-item>
          <el-descriptions-item label="关联对象">{{ safeText(secretDetail.record.associated_with) }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ secretTypeLabel(secretDetail.record) }}</el-descriptions-item>
          <el-descriptions-item label="usage_type">{{ safeText(secretDetail.record.usage_type) }}</el-descriptions-item>
          <el-descriptions-item label="当前状态"><AdminStatusTag :value="normalizedSecretStatus(secretDetail.record.status)" /></el-descriptions-item>
          <el-descriptions-item label="脱敏值">{{ safeText(secretDetail.record.masked_value) }}</el-descriptions-item>
          <el-descriptions-item label="加密密钥 ID">{{ safeText(secretDetail.record.key_id) }}</el-descriptions-item>
          <el-descriptions-item label="最近使用时间">{{ safeText(secretDetail.record.last_used_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ safeText(secretDetail.record.updated_at) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ safeText(secretDetail.record.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="轮换时间">{{ safeText(secretDetail.record.rotated_at) }}</el-descriptions-item>
          <el-descriptions-item label="created_by">{{ safeText(secretDetail.record.created_by) }}</el-descriptions-item>
          <el-descriptions-item label="updated_by">{{ safeText(secretDetail.record.updated_by) }}</el-descriptions-item>
          <el-descriptions-item label="用途说明">{{ safeText(secretDetail.record.description) }}</el-descriptions-item>
          <el-descriptions-item label="是否可用">{{ yesNo(secretDetail.record.available) }}</el-descriptions-item>
          <el-descriptions-item label="是否可解密">{{ decryptableLabel(secretDetail.record) }}</el-descriptions-item>
          <el-descriptions-item label="来源类型">{{ safeText(secretDetail.source?.type || secretDetail.record.source_type) }}</el-descriptions-item>
          <el-descriptions-item label="来源说明">{{ safeText(secretDetail.source?.label || sourceEntryLabel(secretDetail.record)) }}</el-descriptions-item>
          <el-descriptions-item label="source_id">{{ safeText(secretDetail.source?.source_id || secretDetail.record.source_id) }}</el-descriptions-item>
          <el-descriptions-item label="source_code">{{ safeText(secretDetail.source?.source_code || secretDetail.record.source_code) }}</el-descriptions-item>
          <el-descriptions-item label="来源插件">{{ safeText(secretDetail.source?.plugin_code) }}</el-descriptions-item>
          <el-descriptions-item label="来源配置项">{{ safeText(secretDetail.source?.config_entry) }}</el-descriptions-item>
          <el-descriptions-item label="来源治理页">{{ safeText(secretDetail.source?.management_page) }}</el-descriptions-item>
        </el-descriptions>
        <section class="plain-section mt12">
          <div class="section-title">使用关系</div>
          <div class="section-desc">基于当前存储配置解析，不通过字符串猜测业务归属。</div>
          <el-table class="mt12" :data="secretDetail.usages" stripe border>
            <el-table-column label="类型" width="130"><template #default="{ row }">{{ safeText(row.label || row.type) }}</template></el-table-column>
            <el-table-column label="插件" width="150"><template #default="{ row }">{{ safeText(row.plugin_name || row.plugin_code) }}</template></el-table-column>
            <el-table-column label="配置项" min-width="190"><template #default="{ row }">{{ safeText(row.config_entry || row.management_page) }}</template></el-table-column>
            <el-table-column label="状态" width="120"><template #default="{ row }">{{ safeText(row.current_health || row.status) }}</template></el-table-column>
          </el-table>
        </section>
        <el-alert
          class="mt12"
          type="info"
          show-icon
          :closable="false"
          title="该页面只展示敏感配置引用和元数据，不显示明文。真实值仅在服务端内部按需解密使用。"
        />
        <AdminTechnicalDetails class="mt12" :blocks="secretDetailTechnicalBlocks" testid="secret-detail-technical-details" />
        <div class="drawer-actions">
          <el-button type="primary" plain :disabled="!canJumpSecretSource(secretDetail.record)" @click="goSecretSource(secretDetail.record)">跳转到来源配置</el-button>
          <el-button plain @click="openSecretAudit(secretDetail.record)">查看审计</el-button>
          <el-button plain :disabled="!canRotateSecret(secretDetail.record)" @click="rotateSecretPlaceholder(secretDetail.record)">
            {{ rotateSecretButtonText(secretDetail.record) }}
          </el-button>
          <el-button
            v-if="!isTestSecret(secretDetail.record)"
            type="warning"
            plain
            :disabled="secretDetail.record.status === 'disabled' || secretDetail.record.status === 'revoked'"
            @click="disableSecret(secretDetail.record)"
          >
            禁用
          </el-button>
          <el-button
            v-if="!isTestSecret(secretDetail.record)"
            type="danger"
            plain
            :disabled="secretDetail.record.status === 'revoked'"
            @click="revokeSecret(secretDetail.record)"
          >
            吊销
          </el-button>
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="secretActionDialog.visible" :title="secretActionTitle" width="620px">
      <template v-if="secretActionDialog.preview">
        <el-alert
          :type="secretActionDialog.preview.allowed ? 'warning' : 'error'"
          show-icon
          :closable="false"
          :title="secretActionDialog.preview.warning || secretActionDialog.preview.message || '请确认操作影响'"
        />
        <dl class="impact-grid mt12">
          <div><dt>影响插件</dt><dd>{{ secretActionDialog.preview.affected_plugins }}</dd></div>
          <div><dt>外部服务</dt><dd>{{ secretActionDialog.preview.affected_external_services }}</dd></div>
          <div><dt>Webhook</dt><dd>{{ secretActionDialog.preview.affected_webhooks }}</dd></div>
          <div><dt>Callback</dt><dd>{{ secretActionDialog.preview.affected_callbacks }}</dd></div>
          <div><dt>24h 使用</dt><dd>{{ secretActionDialog.preview.usage_count_last_24h || 0 }}</dd></div>
          <div><dt>7d 使用</dt><dd>{{ secretActionDialog.preview.usage_count_last_7d || 0 }}</dd></div>
        </dl>
        <el-table class="mt12" :data="secretActionDialog.preview.affected_business || []" stripe border>
          <el-table-column label="业务" min-width="160"><template #default="{ row }">{{ safeText(row.label || row.type) }}</template></el-table-column>
          <el-table-column label="插件" min-width="150"><template #default="{ row }">{{ safeText(row.plugin_name || row.plugin_code) }}</template></el-table-column>
          <el-table-column label="配置 / 目标" min-width="220"><template #default="{ row }">{{ safeText(row.endpoint_url || row.target_url || row.config_entry) }}</template></el-table-column>
        </el-table>
        <el-alert
          v-if="secretActionDialog.action === 'revoke'"
          class="mt12"
          type="error"
          show-icon
          :closable="false"
          title="吊销不是直接可恢复操作；需要通过来源配置重新写入或轮换新凭据。"
        />
        <el-input
          v-if="secretActionDialog.action === 'revoke'"
          v-model="secretActionDialog.confirmRef"
          class="mt12"
          placeholder="请输入完整 ref 以确认吊销"
        />
      </template>
      <template #footer>
        <el-button @click="secretActionDialog.visible = false">取消</el-button>
        <el-button
          :type="secretActionDialog.action === 'revoke' ? 'danger' : 'warning'"
          :disabled="!secretActionCanSubmit"
          :loading="secretActionDialog.submitting"
          @click="submitSecretAction"
        >
          {{ secretActionDialog.action === 'revoke' ? '确认吊销' : '确认禁用' }}
        </el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { auditLogs, createExternalServiceHTTPAllowlist, deleteExternalServiceHTTPAllowlist, disableSecretRef, externalServiceHTTPAllowlist, getSecretRefDetail, getSecretRefMetadata, listSecretRefs, previewDisableSecretRef, previewRevokeSecretRef, revokeSecretRef, runPluginExternalServiceHealthCheck, systemEffectiveConfig, systemSensitiveConfigStatus } from '@/api/admin';
import { AdminInlineHint, AdminMetricCard, AdminPageHeader, AdminStatusTag, AdminTechnicalDetails } from '@/components/admin';
const router = useRouter();
const route = useRoute();
const activeSection = ref(String(route.query.tab || 'overview'));
const logQuery = reactive({ site: 'portal', type: 'all', action: '', actor: '', target: '', metadata: '', page: 1, page_size: 10 });
const logList = ref([]);
const logTotal = ref(0);
const sensitiveStatus = ref(null);
const sensitiveError = ref('');
const allowlistStatus = ref(null);
const effectiveConfig = ref(null);
const secretList = ref([]);
const allowlistDialog = reactive({ visible: false, saving: false });
const allowlistForm = reactive({ origin: '', usage: '', risk_confirmed: false });
const diagnosticDialog = reactive({ visible: false, text: '' });
const secretFilters = reactive({ type: 'all', status: 'all', keyword: '' });
const secretDetail = reactive({ visible: false, record: null, source: null, usages: [] });
const secretActionDialog = reactive({ visible: false, action: '', record: null, preview: null, confirmRef: '', submitting: false });

const envExampleBlocks = [
  {
    title: '插件配置加密 root key 示例',
    code: 'DEVHUB_PLUGIN_CONFIG_KEYS=\'[{"id":"local-v1","key":"base64-xxx","primary":true}]\'',
  },
  {
    title: '兼容单 key 示例',
    code: 'DEVHUB_PLUGIN_CONFIG_KEY_ID=local-v1\nDEVHUB_PLUGIN_CONFIG_KEY=base64-xxx',
  },
];
const externalAllowlistEnvExample = {
  title: '环境变量 HTTP Allowlist 示例',
  code: 'DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build',
};

async function loadLogs() {
  const data = await auditLogs(logQuery);
  logList.value = data.items || [];
  logTotal.value = data.total || 0;
}
async function loadSecrets() {
  const data = await listSecretRefs({ page: 1, page_size: 100 });
  secretList.value = data.items || [];
}
function actorTypeName(type) {
  return { admin_user: '后台人员', moderator: '子站版主', system: '系统' }[type] || type || '-';
}
async function loadSensitiveStatus() {
  sensitiveError.value = '';
  try {
    const [status, allowlist, effective] = await Promise.all([
      systemSensitiveConfigStatus(),
      externalServiceHTTPAllowlist(),
      systemEffectiveConfig(),
    ]);
    sensitiveStatus.value = status;
    allowlistStatus.value = allowlist;
    effectiveConfig.value = effective;
  } catch (e) {
    sensitiveStatus.value = null;
    allowlistStatus.value = null;
    effectiveConfig.value = null;
    sensitiveError.value = e?.message || '敏感配置状态加载失败';
  }
}
async function load() { await loadSensitiveStatus(); await loadLogs(); await loadSecrets(); }
watch(activeSection, async (next) => {
  const query = { ...route.query };
  if (!next || next === 'overview') delete query.tab;
  else query.tab = next;
  await router.replace({ query });
});
watch(() => route.query.tab, (value) => {
  const next = String(value || 'overview');
  if (activeSection.value !== next) activeSection.value = next;
});
async function copyExample(text) {
  await copyText(text, '示例已复制');
}
async function copyText(text, successMessage = '已复制') {
  try {
    await navigator.clipboard.writeText(String(text || ''));
    ElMessage.success(successMessage);
  } catch (e) {
    ElMessage.warning('当前浏览器不支持复制，请手动选择代码');
  }
}

const keyring = computed(() => sensitiveStatus.value?.plugin_config_keyring || {});
const secretCenter = computed(() => sensitiveStatus.value?.secret_center || {});
const httpPolicy = computed(() => allowlistStatus.value?.policy || sensitiveStatus.value?.external_service_http_policy || {});
const allowlistRows = computed(() => allowlistStatus.value?.effective_allowlist || []);
const effectiveAllowlist = computed(() => effectiveConfig.value?.external_service_http_allowlist || allowlistStatus.value || {});
const effectiveAllowlistRows = computed(() => effectiveAllowlist.value?.effective_allowlist || []);
const effectiveExternalServices = computed(() => effectiveConfig.value?.external_services || []);
const effectiveSecretCenter = computed(() => effectiveConfig.value?.secret_center_status || secretCenter.value || {});
const webhookSecurity = computed(() => effectiveConfig.value?.webhook_callback_security || {});
const rootKeyConfigured = computed(() => (effectiveConfig.value?.root_key_status?.status || keyring.value.status) === 'ok');
const secretCenterAvailable = computed(() => ['ok', 'warning'].includes(String(effectiveSecretCenter.value.status || '').toLowerCase()));
const diagnosticText = computed(() => effectiveConfig.value?.diagnostic_text || '');
const effectiveNextSteps = computed(() => {
  const steps = Array.isArray(effectiveConfig.value?.next_steps) ? [...effectiveConfig.value.next_steps] : [];
  effectiveExternalServices.value.forEach((row) => {
    if (Array.isArray(row.next_steps)) steps.push(...row.next_steps);
  });
  return [...new Set(steps.map((item) => String(item || '').trim()).filter(Boolean))];
});
const effectiveTechnicalBlocks = computed(() => [
  { name: 'diagnostic_text', title: '脱敏诊断文本', value: diagnosticText.value || '暂无脱敏诊断内容' },
  { name: 'effective_config', title: '当前生效配置原始摘要', value: effectiveConfig.value || {} },
]);
const effectiveAllowlistGroups = computed(() => [
  { key: 'defaults', title: '系统默认来源', items: effectiveAllowlist.value?.defaults || [] },
  { key: 'env', title: '环境变量来源', items: effectiveAllowlist.value?.env_allowlist || [] },
  { key: 'admin', title: '后台配置来源', items: effectiveAllowlist.value?.admin_allowlist || [] },
  { key: 'effective', title: '最终生效列表', items: effectiveAllowlistRows.value },
]);
const envAllowlistCount = computed(() => (allowlistStatus.value?.env_allowlist || []).length);
const adminAllowlistCount = computed(() => (allowlistStatus.value?.admin_allowlist || []).length);
const effectiveAllowlistCount = computed(() => (allowlistStatus.value?.effective_allowlist || []).length);
const keyringWarnings = computed(() => Array.isArray(keyring.value.warnings) ? keyring.value.warnings : []);

function safeText(value, fallback = '暂无') {
  if (value === undefined || value === null || value === '') return fallback;
  return String(value);
}
function yesNo(value) {
  return value ? '是' : '否';
}
function keyringState() {
  const status = String(keyring.value.status || '').toLowerCase();
  const missing = keyringWarnings.value.some((item) => String(item || '').includes('缺少'));
  if (status === 'ok') return { text: '正常', type: 'success' };
  if (status === 'blocked' && missing) return { text: '未配置', type: 'warning' };
  if (status === 'warning') return { text: '未配置', type: 'warning' };
  return { text: status ? '异常' : '未配置', type: status ? 'danger' : 'warning' };
}
function secretCenterState() {
  const status = String(secretCenter.value.status || '').toLowerCase();
  if (status === 'ok') return { text: '正常', type: 'success' };
  if (status === 'warning') return { text: '未就绪', type: 'warning' };
  return { text: status ? '异常' : '未就绪', type: status ? 'danger' : 'warning' };
}
function allowlistState() {
  return envAllowlistCount.value || adminAllowlistCount.value
    ? { text: '已配置', type: 'success' }
    : { text: '未配置', type: 'warning' };
}
function allowlistSourceName(source) {
  return {
    system_default: '系统默认',
    default: '系统默认',
    env: '环境变量',
    environment: '环境变量',
    admin: '后台配置',
    admin_setting: '后台配置',
    merged: '合并来源',
    empty: '未命中',
    unknown: '未知',
  }[source] || source || '-';
}
function configSourceLabel(source) {
  return {
    'system default': '系统默认',
    system_default: '系统默认',
    'env var': '环境变量',
    env: '环境变量',
    admin_config: '后台配置',
    'admin config': '后台配置',
    plugin_runtime_config: '插件运行配置',
    'plugin runtime config': '插件运行配置',
    SecretCenter: 'SecretCenter',
    secret_center: 'SecretCenter',
    unknown: '未知',
  }[source] || source || '未知';
}
function openAudit(action = '', target = '', metadata = '') {
  activeSection.value = 'audit';
  logQuery.action = action;
  logQuery.target = target;
  logQuery.metadata = metadata;
  logQuery.page = 1;
  loadLogs();
}
function openAllowlistDialog() {
  allowlistForm.origin = '';
  allowlistForm.usage = '';
  allowlistForm.risk_confirmed = false;
  allowlistDialog.visible = true;
}
async function submitAllowlist() {
  allowlistDialog.saving = true;
  try {
    allowlistStatus.value = await createExternalServiceHTTPAllowlist({ ...allowlistForm });
    await loadSensitiveStatus();
    allowlistDialog.visible = false;
    ElMessage.success('HTTP allowlist 已新增');
  } catch (e) {
    ElMessage.error(e?.message || '新增失败');
  } finally {
    allowlistDialog.saving = false;
  }
}
async function removeAllowlist(row) {
  try {
    await ElMessageBox.confirm(`确认删除 ${row.origin}？删除后该 HTTP endpoint 将重新被安全策略拒绝。`, '删除后台 HTTP allowlist', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
    });
  } catch (e) {
    return;
  }
  allowlistStatus.value = await deleteExternalServiceHTTPAllowlist(row.id);
  await loadSensitiveStatus();
  ElMessage.success('HTTP allowlist 已删除');
}

const statusCards = computed(() => {
  const keyState = keyringState();
  const secretState = secretCenterState();
  const httpState = allowlistState();
  const keyReady = keyring.value.status === 'ok';
  return [
    {
      key: 'keyring',
      title: '启动级加密密钥',
      statusText: keyState.text,
      statusKey: keyState.type === 'success' ? 'healthy' : keyState.type === 'warning' ? 'warning' : 'error',
      tagType: keyState.type,
      description: '用于加密 Webhook Secret、Callback Token、external_service token 等敏感配置。root key 只能来自环境变量或外部 Secret 系统，不能在后台保存。',
      fields: [
        { label: '当前密钥 ID', value: safeText(keyring.value.current_key_id, '未配置') },
        { label: '可用密钥数', value: safeText(keyring.value.key_count, '0') },
        { label: '是否可创建 Secret', value: yesNo(keyReady) },
        { label: '是否可解密历史配置', value: yesNo(keyReady || keyring.value.legacy_v1_supported) },
      ],
      badge: keyring.value.restart_required ? '需重启' : '',
    },
    {
      key: 'secret-center',
      title: 'SecretCenter 引用层',
      statusText: secretState.text,
      statusKey: secretState.type === 'success' ? 'healthy' : secretState.type === 'warning' ? 'warning' : 'error',
      tagType: secretState.type,
      description: '敏感值通过 secret_ref / token_ref 引用，后台只展示引用和脱敏状态，不回显明文。',
      fields: [
        { label: 'Secret 引用数', value: safeText(secretCenter.value.secret_ref_count, '0') },
        { label: 'external_service 引用数', value: safeText(secretCenter.value.namespace_counts?.external_service, '0') },
        { label: '最近更新', value: safeText(secretCenter.value.last_updated_at) },
        { label: '最近使用', value: safeText(secretCenter.value.last_used_at) },
      ],
    },
    {
      key: 'http-allowlist',
      title: 'external_service HTTP 策略',
      statusText: httpState.text,
      statusKey: httpState.type === 'success' ? 'healthy' : 'warning',
      tagType: httpState.type,
      description: 'HTTPS 默认允许；非 localhost HTTP 必须进入环境变量或后台配置 allowlist 才会放行。',
      fields: [
        { label: '系统默认', value: 'localhost / 127.0.0.1 / ::1' },
        { label: '环境变量来源', value: `${envAllowlistCount.value} 条` },
        { label: '后台配置来源', value: `${adminAllowlistCount.value} 条` },
        { label: '最终生效', value: `${effectiveAllowlistCount.value} 条` },
        { label: '当前策略', value: httpPolicy.value.non_local_http_needs_allowlist === false ? '非 localhost HTTP 未强制放行' : '非 localhost HTTP 默认拒绝' },
      ],
    },
  ];
});
const keyringCard = computed(() => statusCards.value[0] || {});
const secretCenterCard = computed(() => statusCards.value[1] || {});
const allowlistCard = computed(() => statusCards.value[2] || {});
const summaryCards = computed(() => statusCards.value);
const testSecretPrefixes = ['s15smoke', 'e2e', 'fixture', 'test', 'demo', 'seed'];
const filteredSecretRows = computed(() => {
  const keyword = secretFilters.keyword.trim().toLowerCase();
  return secretList.value.filter((row) => {
    if (secretFilters.type !== 'all' && secretCategory(row) !== secretFilters.type) return false;
    if (secretFilters.status !== 'all' && normalizedSecretStatus(row.status) !== secretFilters.status) return false;
    if (!keyword) return true;
    return [row.ref, row.namespace, row.name, row.key_id, row.display_name, row.usage, row.associated_with]
      .some((value) => String(value || '').toLowerCase().includes(keyword));
  });
});
const secretTestCount = computed(() => secretList.value.filter((row) => isTestSecret(row)).length);
const secretActionTitle = computed(() => secretActionDialog.action === 'revoke' ? '吊销 Secret 影响预览' : '禁用 Secret 影响预览');
const secretActionCanSubmit = computed(() => {
  if (!secretActionDialog.preview?.allowed) return false;
  if (secretActionDialog.action === 'revoke') return secretActionDialog.confirmRef === secretActionDialog.record?.ref;
  return true;
});
const secretDetailTechnicalBlocks = computed(() => [
  { name: 'metadata', title: 'Secret 元数据（已脱敏）', value: secretDetail.record || {} },
  { name: 'source', title: '来源解析', value: secretDetail.source || {} },
  { name: 'usages', title: '使用关系', value: secretDetail.usages || [] },
]);

function normalizedSecretStatus(status) {
  const value = String(status || '').toLowerCase();
  if (['active', 'disabled', 'revoked'].includes(value)) return value;
  return 'unknown';
}
function secretStatusLabel(status) {
  return { active: '正常', disabled: '已禁用', revoked: '已吊销', unknown: '未知' }[normalizedSecretStatus(status)];
}
function secretStatusTag(status) {
  return { active: 'success', disabled: 'warning', revoked: 'danger', unknown: 'info' }[normalizedSecretStatus(status)];
}
function parseSecretRef(row) {
  const ref = String(row?.ref || '');
  const match = ref.match(/^secret:\/\/([^/]+)\/(.+)$/);
  return {
    namespace: row?.namespace || match?.[1] || '',
    name: row?.name || match?.[2] || '',
  };
}
function isTestSecret(row) {
  const parsed = parseSecretRef(row);
  const values = [parsed.namespace, parsed.name, row?.ref].map((v) => String(v || '').toLowerCase());
  return values.some((value) => testSecretPrefixes.some((prefix) => value === prefix || value.startsWith(`${prefix}_`) || value.startsWith(`${prefix}-`) || value.includes(`/${prefix}`)));
}
function secretCategory(row) {
  if (isTestSecret(row)) return 'test';
  const ns = String(parseSecretRef(row).namespace || '').toLowerCase();
  if (ns === 'external_service') return 'external_service';
  if (ns === 'webhook') return 'webhook';
  if (ns === 'callback') return 'callback';
  if (ns === 'plugin_config') return 'plugin_config';
  return 'other';
}
function businessLabel(row) {
  const ns = String(parseSecretRef(row).namespace || '').toLowerCase();
  if (isTestSecret(row)) return '测试数据';
  if (ns === 'external_service') return '外部服务';
  if (ns === 'webhook') return 'Webhook 密钥';
  if (ns === 'callback') return 'Callback Token';
  if (ns === 'plugin_config') return '插件配置敏感字段';
  return ns ? `${ns}（未知类型）` : '其他（未知类型）';
}
function businessHint(row) {
  const parsed = parseSecretRef(row);
  if (!parsed.name) return '';
  return parsed.name;
}
function secretTypeLabel(row) {
  const parsed = parseSecretRef(row);
  const name = String(parsed.name || '').toLowerCase();
  if (isTestSecret(row)) return '测试数据';
  if (secretCategory(row) === 'external_service' && name.endsWith('/token')) return '外部服务 token';
  if (secretCategory(row) === 'external_service') return '外部服务 Secret';
  if (secretCategory(row) === 'webhook') return 'Webhook Secret';
  if (secretCategory(row) === 'callback') return 'Callback Token';
  if (secretCategory(row) === 'plugin_config') return '插件配置敏感字段';
  return '其他 Secret';
}
function decryptableLabel(row) {
  const category = secretCategory(row);
  if (category === 'callback') return '仅保存 hash，不支持解密明文';
  if (category === 'webhook') return '密文由 Webhook 密钥流程管理，详情不读取明文';
  if (category === 'external_service') return row?.masked_value ? '可脱敏展示' : '暂无可展示脱敏值';
  return '暂无可展示脱敏值';
}
function sourceEntryLabel(row) {
  const category = secretCategory(row);
  if (category === 'external_service') return '插件详情 -> external_service 配置';
  if (category === 'webhook') return 'Webhook 治理 -> Webhook 密钥';
  if (category === 'callback') return 'Webhook 治理 -> Callback Token';
  if (category === 'plugin_config') return '插件配置页';
  return '暂无法定位来源配置，请根据 secret_ref 手动查找。';
}
function canJumpSecretSource(row) {
  if (isTestSecret(row)) return false;
  const sourceCanJump = row?.source?.can_jump;
  if (sourceCanJump === false) return false;
  const category = secretCategory(row);
  if (!['external_service', 'webhook', 'callback', 'plugin_config'].includes(category)) return false;
  return Boolean(pluginCodeFromSecret(row) || row?.source_code || row?.source_id || row?.ref);
}
function canRotateSecret(row) {
  const category = secretCategory(row);
  return category === 'external_service' || category === 'webhook' || category === 'callback' || category === 'plugin_config';
}
function rotateSecretButtonText(row) {
  const category = secretCategory(row);
  if (category === 'external_service') return '去来源配置轮换';
  if (category === 'webhook') return '去 Webhook Secret 页面轮换';
  if (category === 'callback') return '去 Callback Token 页面轮换';
  if (category === 'plugin_config') return '去插件配置页更新';
  if (isTestSecret(row)) return '测试数据不支持轮换';
  return '未知来源不支持轮换';
}
function shortenRef(ref) {
  const text = String(ref || '');
  if (text.length <= 54) return text;
  return `${text.slice(0, 28)}...${text.slice(-18)}`;
}
async function copySecretRef(row) {
  await copyText(row?.ref || '', '已复制引用地址；复制的是 ref，不是明文。');
}
async function copyRedactedDiagnostics() {
  const text = diagnosticText.value || JSON.stringify({
    devhub_version: effectiveConfig.value?.devhub_version || 'unknown',
    store_mode: effectiveConfig.value?.store_mode || 'unknown',
    external_services: effectiveExternalServices.value,
    http_allowlist_effective_list: (effectiveAllowlist.value?.policy || {}).effective_allowlist || [],
  }, null, 2);
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success('已复制脱敏诊断信息，可安全用于提交问题或排查。');
  } catch (e) {
    diagnosticDialog.text = text;
    diagnosticDialog.visible = true;
    ElMessage.warning('复制失败，请在弹窗中手动选择脱敏内容。');
  }
}
function tokenStatusLabel(status) {
  const value = String(status || '').toLowerCase();
  return {
    active: '正常',
    disabled: '已禁用',
    revoked: '已吊销',
    missing: '未配置',
    not_found: '不存在',
    not_required: '不需要',
  }[value] || value || '未知';
}
function tokenStatusTag(status) {
  const value = String(status || '').toLowerCase();
  if (value === 'active' || value === 'not_required') return 'success';
  if (value === 'missing' || value === 'disabled') return 'warning';
  if (value === 'revoked' || value === 'not_found') return 'danger';
  return 'info';
}
async function openSecretDetail(row) {
  let record = row;
  let detail = null;
  try {
    detail = await getSecretRefDetail(row.ref);
    record = detail.record || row;
  } catch (e) {
    try {
      record = await getSecretRefMetadata(row.ref);
    } catch (_) {
      record = row;
    }
  }
  secretDetail.record = record;
  secretDetail.source = detail?.source || null;
  secretDetail.usages = detail?.usages || [];
  secretDetail.visible = true;
}
async function openSecretByRef(ref) {
  activeSection.value = 'secret-center';
  const row = secretList.value.find((item) => item.ref === ref) || { ref };
  await openSecretDetail(row);
}
function openSecretAudit(row) {
  const metadata = [row?.namespace, row?.name, row?.source_type, row?.source_id, row?.source_code, pluginCodeFromSecret(row)]
    .map((item) => String(item || '').trim())
    .filter(Boolean)
    .join(' ');
  openAudit('secret_center', row?.ref || 'secret_center', metadata || row?.ref || '');
}
function pluginCodeFromSecret(row) {
  if (row?.associated_with && secretCategory(row) !== 'external_service') return String(row.associated_with || '');
  const name = parseSecretRef(row).name;
  return String(name || '').split('/').filter(Boolean)[0] || '';
}
function goSecretSource(row) {
  const category = secretCategory(row);
  const pluginCode = pluginCodeFromSecret(row);
  if (category === 'external_service') {
    router.push({ path: '/plugins/overview', query: { tab: 'list', plugin_code: pluginCode || undefined, detail_tab: 'runtime' } });
    return;
  }
  if (category === 'webhook') {
    router.push({ path: '/plugins/webhooks', query: { tab: 'secrets', sec_ref: row?.ref || undefined, sec_plugin_code: pluginCode || undefined } });
    return;
  }
  if (category === 'callback') {
    router.push({ path: '/plugins/webhooks', query: { tab: 'callback_tokens', cbtk_plugin_code: pluginCode || undefined } });
    return;
  }
  if (category === 'plugin_config') {
    router.push({ path: '/plugins/overview', query: { tab: 'config', plugin_code: pluginCode || undefined } });
    return;
  }
  ElMessage.info('暂无法定位来源配置，请根据 secret_ref 手动查找。');
}
function rotateSecretPlaceholder(row) {
  const category = secretCategory(row);
  if (category === 'external_service') {
    ElMessage.info('请到对应插件的 external_service 配置中替换 token；SecretCenter 不在此处回显或收集明文。');
    goSecretSource(row);
    return;
  }
  if (category === 'webhook' || category === 'callback') {
    ElMessage.info('请到 Webhook 治理中的对应凭据页面执行轮换；SecretCenter 这里只展示引用元数据。');
    goSecretSource(row);
    return;
  }
  if (category === 'plugin_config') {
    ElMessage.info('这是插件配置敏感字段，请到插件配置页重新输入并保存。');
    goSecretSource(row);
    return;
  }
  if (isTestSecret(row)) {
    ElMessage.info('测试 / fixture / seed Secret 不在 SecretCenter 提供轮换入口。');
    return;
  }
  ElMessage.info('未知来源不能安全轮换，请先查看审计或技术详情。');
}
async function disableSecret(row) {
  if (isTestSecret(row)) {
    ElMessage.warning('测试 / fixture / seed Secret 默认不允许在 SecretCenter 执行危险操作。');
    return;
  }
  await openSecretAction(row, 'disable');
}
async function revokeSecret(row) {
  if (isTestSecret(row)) {
    ElMessage.warning('测试 / fixture / seed Secret 默认不允许在 SecretCenter 执行危险操作。');
    return;
  }
  await openSecretAction(row, 'revoke');
}
async function openSecretAction(row, action) {
  try {
    const preview = action === 'revoke'
      ? await previewRevokeSecretRef(row.ref)
      : await previewDisableSecretRef(row.ref);
    secretActionDialog.visible = true;
    secretActionDialog.action = action;
    secretActionDialog.record = row;
    secretActionDialog.preview = preview;
    secretActionDialog.confirmRef = '';
  } catch (e) {
    ElMessage.error(e?.message || '影响预览失败');
  }
}
async function submitSecretAction() {
  if (!secretActionDialog.record) return;
  secretActionDialog.submitting = true;
  try {
    const ref = secretActionDialog.record.ref;
    if (secretActionDialog.action === 'revoke') {
      await revokeSecretRef(ref, { confirm_ref: secretActionDialog.confirmRef });
      ElMessage.success('Secret 引用已吊销');
    } else {
      await disableSecretRef(ref);
      ElMessage.success('Secret 引用已禁用');
    }
    secretActionDialog.visible = false;
    await loadSecrets();
    await loadSensitiveStatus();
    if (secretDetail.record?.ref === ref) await openSecretDetail({ ref });
  } catch (e) {
    ElMessage.error(e?.message || '操作失败');
  } finally {
    secretActionDialog.submitting = false;
  }
}
function goPluginExternalConfig(pluginCode) {
  router.push({ path: '/plugins/overview', query: { tab: 'list', plugin_code: pluginCode || undefined, detail_tab: 'runtime' } });
}
function goExternalExecutions(pluginCode) {
  router.push({ path: '/plugins/webhooks', query: { tab: 'external_service', ext_plugin_code: pluginCode || undefined } });
}
function goWebhookSecurity(tab) {
  router.push({ path: '/plugins/webhooks', query: { tab } });
}
async function runHealthCheck(row) {
  if (!row?.plugin_code) return;
  try {
    await runPluginExternalServiceHealthCheck(row.plugin_code);
    ElMessage.success('健康检查已执行，请查看运行记录。');
    await loadSensitiveStatus();
  } catch (e) {
    ElMessage.error(e?.message || '健康检查失败');
  }
}
function handleSecretAction(command, row) {
  if (command === 'source') return goSecretSource(row);
  if (command === 'rotate') return rotateSecretPlaceholder(row);
  if (command === 'disable') return disableSecret(row);
  if (command === 'revoke') return revokeSecret(row);
  return undefined;
}
load();
</script>

<style scoped>
.system-page { display: flex; flex-direction: column; gap: 16px; }
.mt12 { margin-top: 12px; }
.mb12 { margin-bottom: 12px; }
.mt16 { margin-top: 16px; }
.card-head { display: flex; align-items: center; justify-content: space-between; }
.security-status-card :deep(.el-card__body) { padding-top: 16px; }
.security-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.security-actions { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.security-title { font-size: 16px; font-weight: 650; color: #1f2937; }
.security-subtitle { margin-top: 6px; max-width: 820px; color: #667085; line-height: 1.6; }
.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}
.status-grid.single { grid-template-columns: minmax(0, 1fr); }
.status-panel {
  min-width: 0;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 14px;
  background: #fff;
}
.status-panel-head { display: flex; justify-content: space-between; gap: 12px; align-items: flex-start; }
.status-panel-title { font-size: 15px; font-weight: 650; color: #111827; }
.status-panel-desc { margin-top: 8px; color: #667085; line-height: 1.55; font-size: 13px; }
.status-fields { margin: 14px 0 0; display: grid; gap: 10px; }
.status-field { display: grid; grid-template-columns: minmax(100px, 128px) minmax(0, 1fr); gap: 8px; align-items: start; }
.status-field dt { color: #667085; font-size: 13px; }
.status-field dd { margin: 0; color: #1f2937; font-weight: 550; min-width: 0; word-break: break-word; }
.status-field dd.multi-value { white-space: pre-line; }
.status-note { margin-top: 12px; }
.system-tabs { margin-top: 4px; }
.section-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.section-title { color: #111827; font-weight: 650; font-size: 15px; }
.section-desc { margin-top: 6px; color: #667085; line-height: 1.5; }
.section-actions { display: flex; flex-wrap: wrap; gap: 8px; justify-content: flex-end; }
.runtime-grid {
  margin: 14px 0 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
}
.runtime-grid div {
  min-width: 0;
  border: 1px solid #eef2f6;
  border-radius: 8px;
  padding: 10px;
  background: #fbfcfe;
}
.runtime-grid dt { color: #667085; font-size: 12px; }
.runtime-grid dd { margin: 4px 0 0; color: #111827; font-weight: 650; word-break: break-word; }
.quick-actions { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.quick-actions :deep(.el-button + .el-button) { margin-left: 0; }
.table-actions { gap: 4px 8px; }
.info-panel {
  border: 1px solid #c7d7fe;
  border-radius: 8px;
  background: #f4f7ff;
  padding: 14px;
}
.info-panel-title { color: #1d4ed8; font-weight: 650; font-size: 14px; }
.info-panel-desc { margin-top: 8px; color: #344054; line-height: 1.6; font-size: 13px; }
.filter-bar { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
.allowlist-groups { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; margin-top: 12px; }
.allowlist-group { min-width: 0; border: 1px solid #eef2f6; border-radius: 8px; padding: 12px; background: #fbfcfe; }
.allowlist-group-title { color: #344054; font-weight: 650; margin-bottom: 8px; }
.allowlist-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.secret-search { width: min(360px, 100%); }
.secret-ref-cell { display: flex; align-items: center; gap: 8px; min-width: 0; }
.secret-ref-text { min-width: 0; word-break: break-all; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; }
.mono-wrap { min-width: 0; word-break: break-all; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; }
.drawer-actions { margin-top: 16px; display: flex; flex-wrap: wrap; gap: 8px; }
.impact-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; margin: 0; }
.impact-grid div { border: 1px solid #e5e7eb; border-radius: 8px; padding: 10px; background: #fbfcfe; }
.impact-grid dt { color: #667085; font-size: 12px; }
.impact-grid dd { margin: 4px 0 0; color: #111827; font-weight: 650; }
.next-step-list { display: grid; gap: 8px; margin-top: 10px; }
.next-step-list.compact { margin-top: 0; gap: 6px; }
.next-step-item {
  color: #344054;
  line-height: 1.45;
  font-size: 13px;
  word-break: break-word;
}
.plain-section { border: 1px solid #e5e7eb; border-radius: 8px; padding: 14px; min-height: 100%; }
.plain-section .el-form { margin-top: 14px; }
.muted { color: #98a2b3; font-size: 13px; }
.example-list { display: grid; gap: 12px; }
.code-panel { border: 1px solid #e5e7eb; border-radius: 8px; overflow: hidden; background: #f8fafc; }
.code-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 9px 12px;
  border-bottom: 1px solid #e5e7eb;
  color: #344054;
  font-weight: 600;
}
.code-panel pre { margin: 0; padding: 12px; overflow: auto; white-space: pre-wrap; word-break: break-word; }
.code-panel code { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; color: #111827; }
@media (max-width: 1024px) {
  .security-head { flex-direction: column; }
  .security-actions { justify-content: flex-start; }
  .status-grid { grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); }
  .status-field { grid-template-columns: 1fr; gap: 4px; }
  .section-head,
  .card-head { align-items: flex-start; flex-direction: column; }
  .runtime-grid { grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); }
}
</style>
