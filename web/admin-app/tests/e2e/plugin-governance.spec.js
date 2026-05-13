import { expect, test } from '@playwright/test';
import {
  archivePlugin,
  bulkArchivePlugins,
  bulkRestorePlugins,
  disableCommunityPlugin,
  disablePlugin,
  enableCommunityPlugin,
  ensurePluginEnabled,
  injectFailedPluginHook,
  injectFailedPluginMigration,
  installPluginManifest,
  pluginAuditLogs,
  pluginHooks,
  pluginHealth,
  pluginHealthSummary,
  pluginMigrations,
  dryRunPluginManifest,
  validatePluginManifest,
  retryPluginMigration,
  setCategoryEnabled,
  restorePlugin,
  userPost,
  uniqueTitle,
} from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin governance center', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ page, request }) => {
    await retryPluginMigration(request, 'qa', 'qa_questions').catch(() => {});
    await retryPluginMigration(request, 'qa', 'qa_answers').catch(() => {});
    await ensurePluginEnabled(request, 'qa');
    await enableCommunityPlugin(request, 1, 'qa');
    await seedAdminSession(page);
  });

  test('opens plugin center and filters plugin list', async ({ page }) => {
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
    await expect(page.getByTestId('plugin-stats')).toContainText('全部插件');
    await expect(page.getByText('问答插件')).toBeVisible();

    await page.getByTestId('plugin-search').fill('qa');
    await expect(page.getByText('问答插件')).toBeVisible();
    await expect(page.getByText('文档插件')).toBeHidden();
  });

  test('shows health summary and supports manifest validate dry-run and install dialogs', async ({ page, request }) => {
    const code = `e2e_plugin_${Date.now()}`;
    const contentType = `e2e_type_${Date.now()}`;
    const manifest = buildManifest(code, contentType);
    let installed = false;

    try {
      await page.goto('/admin-next/plugins/list');
      await expect(page.getByTestId('plugin-health-summary')).toBeVisible();
      await expect(page.getByTestId('plugin-health-summary').locator('.health-card')).toHaveCount(9);

      const summary = await pluginHealthSummary(request);
      expect(summary.summary).toBeTruthy();
      await expect(page.getByTestId('plugin-health-summary').locator('.health-card').first()).toBeVisible();

      await page.goto('/admin-next/plugins/install');
      await expect(page.getByTestId('plugin-install-page')).toBeVisible();
      await page.getByTestId('plugin-manifest-validate').click();
      await expect(page.getByTestId('plugin-manifest-panel')).toBeVisible();
      await page.getByTestId('plugin-manifest-input').fill(manifest);
      await page.getByTestId('plugin-manifest-submit').click();
      await expect(page.getByTestId('plugin-result-summary')).toBeVisible();
      await expect(page.getByTestId('plugin-result-summary')).toContainText('manifest');
      await expect(page.getByTestId('plugin-result-summary')).toContainText('已禁用');
      await page.getByTestId('plugin-result-close').click();

      await page.getByTestId('plugin-manifest-dry-run').click();
      await expect(page.getByTestId('plugin-manifest-panel')).toBeVisible();
      await page.getByTestId('plugin-manifest-input').fill(manifest);
      await page.getByTestId('plugin-manifest-submit').click();
      await expect(page.getByTestId('plugin-result-summary')).toBeVisible();
      await expect(page.getByTestId('plugin-result-summary')).toContainText('manifest');
      await page.getByTestId('plugin-result-close').click();

      await page.getByTestId('plugin-manifest-install').click();
      await expect(page.getByTestId('plugin-manifest-panel')).toBeVisible();
      await page.getByTestId('plugin-manifest-input').fill(manifest);
      // install wizard: validate -> dry-run preview -> confirm -> install result
      await page.getByTestId('plugin-manifest-submit').click();
      await expect(page.getByTestId('plugin-result-summary')).toBeVisible();
      await page.getByTestId('plugin-manifest-submit').click();
      await expect(page.getByTestId('plugin-result-panel')).toBeVisible();
      await page.getByTestId('plugin-manifest-submit').click();
      await page.getByTestId('plugin-manifest-submit').click();
      await expect(page.getByTestId('plugin-result-summary')).toBeVisible();
      await expect(page.getByTestId('plugin-result-summary')).toContainText('已禁用');
      installed = true;
      await page.getByTestId('plugin-result-close').click();
    } finally {
      if (installed) {
        await archivePlugin(request, code).catch(() => {});
      }
    }
  });

  test('shows upgrade dry-run compatibility matrix for an existing plugin', async ({ page, request }) => {
    const code = `e2e_upgrade_preview_${Date.now()}`;
    let installed = false;
    try {
      await installPluginManifest(request, buildManifest(code, `${code}_content`));
      installed = true;

      await page.goto('/admin-next/plugins/install');
      await expect(page.getByTestId('plugin-install-page')).toBeVisible();

      const upgradeManifest = JSON.parse(buildManifest(code, `${code}_content`));
      upgradeManifest.version = '9.9.9';
      const previewResponse = await request.post(`/api/v1/admin/plugins/${code}/upgrade/dry-run`, {
        headers: { Authorization: 'Bearer devhub-admin-1', 'Content-Type': 'application/json' },
        data: { manifest: upgradeManifest },
      });
      expect(previewResponse.ok()).toBeTruthy();
      const preview = await previewResponse.json();
      expect(preview.current_version).toBeTruthy();
      expect(preview.new_version).toBe('9.9.9');
      expect(preview.compatibility_status).toBeTruthy();
      expect(JSON.stringify(preview)).toContain('dependency_summary');
    } finally {
      if (installed) {
        await archivePlugin(request, code).catch(() => {});
        await restorePlugin(request, code).catch(() => {});
      }
    }
  });

  test('opens plugin detail tabs and shows schema validation errors', async ({ page }) => {
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
    const qaRow = page.getByRole('row', { name: /问答插件/ });
    await expect(qaRow).toBeVisible();
    await qaRow.getByRole('button', { name: '详情' }).first().click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();

    for (const tabName of ['概览', '内容类型', '权限', '菜单', '配置', 'Hook', '路由', '审计']) {
      await page.getByRole('tab', { name: tabName }).click();
      await expect(page.getByRole('tab', { name: tabName })).toHaveAttribute('aria-selected', 'true');
    }

    await page.getByRole('tab', { name: '配置' }).click();
    await expect(page.getByRole('button', { name: '配置模型' })).toBeVisible();
    await expect(page.getByRole('button', { name: '最终生效配置（只读）' })).toBeVisible();
    const globalPanel = page.getByTestId('plugin-global-config-panel');
    await expect(globalPanel).toBeVisible();
    await expect(globalPanel.getByTestId('plugin-global-config-clear')).toBeVisible();

    await globalPanel.getByTestId('plugin-global-config-clear').click();
    await expect(page.getByTestId('schema-error-box')).toContainText('required');
    await expect(page.getByTestId('plugin-global-config-save')).toBeDisabled();
  });

  test('archives plugin and shows archived state with restore entry', async ({ page, request }) => {
    try {
      await page.goto('/admin-next/plugins/list');
      await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
      await page.getByRole('row', { name: /问答插件/ }).getByRole('button', { name: '更多' }).click();
      await page.getByTestId('plugin-archive-qa').click();
      const dialog = page.getByRole('dialog');
      await expect(dialog).toContainText('历史内容');
      await expect(dialog).toContainText('当前启用子站');
      await page.getByRole('button', { name: '确认归档' }).click();
      await expect(page.locator('.el-message').filter({ hasText: '插件已归档' }).first()).toBeVisible();
      await page.reload();
      await expect(page.getByText('已归档').first()).toBeVisible();
      await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
      const qaRow = page.getByRole('row', { name: /问答插件/ });
      await expect(qaRow).toBeVisible();
      await qaRow.getByRole('button', { name: '详情' }).first().click();
      await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();
      await page.getByRole('tab', { name: '概览' }).click();
      await expect(page.getByText('归档时间')).toBeVisible();
      await page.getByRole('button', { name: 'Close this dialog' }).click();
      await expect(page.getByTestId('plugin-detail-drawer')).toBeHidden();
      await page.getByRole('row', { name: /问答插件/ }).getByRole('button', { name: '更多' }).click();
      await page.getByTestId('plugin-restore-qa').click();
      await expect(page.getByRole('dialog')).toContainText('不会自动启用');
      await page.getByRole('button', { name: '确认恢复' }).click();
      await expect(page.locator('.el-message').filter({ hasText: '插件已恢复为已禁用状态' }).first()).toBeVisible();
      await page.reload();
      await page.getByRole('row', { name: /问答插件/ }).getByRole('button', { name: '更多' }).click();
      await expect(page.getByTestId('plugin-enable-qa')).toBeVisible();
    } finally {
      await restorePlugin(request, 'qa').catch(() => {});
      await ensurePluginEnabled(request, 'qa').catch(() => {});
      await enableCommunityPlugin(request, 1, 'qa').catch(() => {});
    }
  });

  test('supports bulk archive and restore actions with result summary', async ({ page, request }) => {
    try {
      await ensurePluginEnabled(request, 'projects');
      await ensurePluginEnabled(request, 'jobs');
      await page.goto('/admin-next/plugins/list');
      const rows = page.locator('[data-testid="admin-plugins-page"] .el-table__body-wrapper .el-table__row');
      await rows.filter({ hasText: '开源项目插件' }).locator('.el-checkbox__inner').click();
      await rows.filter({ hasText: '招聘插件' }).locator('.el-checkbox__inner').click();

      await page.getByTestId('plugin-bulk-archive').click();
      await page.getByRole('button', { name: '确认' }).last().click();
      await expect(page.getByTestId('plugin-result-summary')).toBeVisible();
      await expect(page.getByTestId('plugin-result-summary')).toContainText('成功项');
      await page.getByTestId('plugin-result-close').click();

      await page.reload();
      await rows.filter({ hasText: '开源项目插件' }).locator('.el-checkbox__inner').click();
      await rows.filter({ hasText: '招聘插件' }).locator('.el-checkbox__inner').click();

      await page.getByTestId('plugin-bulk-restore').click();
      await page.getByRole('button', { name: '确认' }).last().click();
      await expect(page.getByTestId('plugin-result-summary')).toBeVisible();
      await expect(page.getByTestId('plugin-result-summary')).toContainText('成功项');
      await page.getByTestId('plugin-result-close').click();
    } finally {
      await ensurePluginEnabled(request, 'projects').catch(() => {});
      await ensurePluginEnabled(request, 'jobs').catch(() => {});
    }
  });

  test('shows impact before global disable and blocks community enable when globally disabled', async ({ page, request }) => {
    try {
      await page.goto('/admin-next/plugins/list');
      await page.getByRole('row', { name: /问答插件/ }).getByRole('button', { name: '更多' }).click();
      await page.getByTestId('plugin-disable-qa').click();
      await expect(page.getByRole('dialog')).toContainText('历史内容详情页和 SEO 不受影响');
      await expect(page.getByRole('dialog')).toContainText('当前启用子站');
      await expect(page.getByRole('dialog')).toContainText('将阻止发布的板块');
      await page.getByRole('button', { name: '确认禁用' }).click();
      await expect(page.getByText('插件状态已更新')).toBeVisible();

      await page.goto('/admin-next/communities');
      await page.getByTestId('community-plugins-1').click();
      await expect(page.getByTestId('community-plugin-drawer')).toBeVisible();
      await expect(page.getByText('该插件已被全局禁用')).toBeVisible();
    } finally {
      await ensurePluginEnabled(request, 'qa');
    }
  });

  test('opens community plugin config and blocks invalid schema values', async ({ page }) => {
    await page.goto('/admin-next/communities');
    await page.getByTestId('community-plugins-1').click();
    await expect(page.getByTestId('community-plugin-drawer')).toBeVisible();
    await expect(page.getByTestId('community-plugin-drawer').getByText('子站已启用').first()).toBeVisible();

    await page.getByTestId('community-plugin-config-qa').click();
    await expect(page.getByTestId('community-plugin-config-dialog')).toBeVisible();
    await expect(page.getByTestId('plugin-json-editor').getByText('子站配置')).toBeVisible();
    await page.getByTestId('json-clear-object').click();
    await expect(page.getByTestId('schema-error-box')).toContainText('required');
    await expect(page.getByTestId('community-plugin-config-save')).toBeDisabled();
  });

  test('renders enhanced config_schema auto form with grouping, enum, boolean, number, required and sensitive masking', async ({ page, request }) => {
    const ts = Date.now();
    const code = `e2e_form_${ts}`;
    const contentType = `e2e_type_${ts}`;
    const manifest = buildManifest(code, contentType);
    let installed = false;

    try {
      const m = JSON.parse(manifest);
      m.name = 'E2E 表单增强插件';
      m.config_schema = {
        type: 'object',
        required: ['title_template', 'enable_feature'],
        properties: {
          title_template: { type: 'string', title: '标题模板', description: '用于 SEO 标题渲染', default: '{title}', minLength: 3, 'x-group': 'SEO' },
          seo_mode: { type: 'string', title: 'SEO 模式', enum: ['simple', 'advanced'], enumNames: ['简单', '高级'], default: 'simple', 'x-group': 'SEO' },
          enable_feature: { type: 'boolean', title: '启用功能', description: '是否启用该功能', default: true, 'x-group': '基础配置', 'x-labels': { true: '开启', false: '关闭' } },
          max_items: { type: 'integer', title: '最大条目数', description: '限制为整数', minimum: 1, maximum: 10, default: 3, 'x-group': '基础配置' },
          api_token: { type: 'string', title: 'API Token', description: '敏感字段，默认脱敏', format: 'password', default: 'secret_token', 'x-group': '敏感字段' },
        },
      };
      const enhancedManifest = JSON.stringify(m, null, 2);

      await installPluginManifest(request, enhancedManifest);
      installed = true;
      await ensurePluginEnabled(request, code);

      await page.goto('/admin-next/plugins/list');
      await page.getByTestId(`plugin-detail-${code}`).click();
      const drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: '配置' }).click();

      const editor = drawer.getByTestId('plugin-json-editor');
      await expect(editor).toBeVisible();
      await expect(editor.getByTestId('config-mode-toggle')).toBeVisible();

      // group cards (schema group title is used as data-testid suffix)
      await expect(drawer.getByTestId('config-group-SEO')).toBeVisible();
      await expect(drawer.getByTestId('config-group-基础配置')).toBeVisible();

      // enum renders select
      await expect(drawer.getByTestId('config-field-seo_mode')).toBeVisible();
      await drawer.getByTestId('config-field-seo_mode').click();
      await expect(page.getByRole('option', { name: '简单' })).toBeVisible();
      await page.keyboard.press('Escape');

      // boolean renders switch
      await expect(drawer.getByTestId('config-field-enable_feature')).toBeVisible();

      // integer min/max
      await expect(drawer.getByTestId('config-field-max_items')).toBeVisible();

      // sensitive masking toggle exists
      await expect(drawer.getByTestId('config-sensitive-toggle-api_token')).toBeVisible();

      // diff preview exists and masks sensitive field
      await expect(drawer.getByTestId('plugin-config-preview')).toBeVisible();
      await expect(drawer.getByTestId('plugin-config-preview')).toContainText('api_token');
      await expect(drawer.getByTestId('plugin-config-preview')).toContainText('******');
    } finally {
      if (installed) await archivePlugin(request, code).catch(() => {});
    }
  });

  test('opens generic plugin content page with filters', async ({ page }) => {
    await page.goto('/admin-next/plugins/content');
    await expect(page.getByTestId('plugin-content-hub-page')).toBeVisible();
    await page.getByTestId('plugin-content-hub-open-qa').click();
    await expect(page).toHaveURL(/\/admin-next\/qa/);
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await expect(page.getByTestId('plugin-content-type-count')).toBeVisible();
    await expect(page.getByTestId('plugin-content-health')).toBeVisible();
    await expect(page.getByTestId('plugin-content-community-filter')).toBeVisible();
    await expect(page.getByTestId('plugin-content-status-filter')).toBeVisible();
    await page.getByTestId('plugin-content-back').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/content/);
  });

  test('failed migration blocks enablement until retry succeeds', async ({ page, request }) => {
    const errorText = `E2E forced migration failure ${Date.now()}`;
    try {
      await disablePlugin(request, 'qa');
      await disableCommunityPlugin(request, 1, 'qa');
      await injectFailedPluginMigration(request, 'qa', 'qa_questions', errorText);

      const globalEnable = await request.post('/api/v1/admin/plugins/qa/enable', {
        headers: { Authorization: 'Bearer devhub-admin-1' },
      });
      expect(globalEnable.ok()).toBeFalsy();
      expect(await globalEnable.text()).toContain('plugin_migration_failed');

      const communityEnable = await enableCommunityPlugin(request, 1, 'qa');
      expect(communityEnable.ok()).toBeFalsy();
      expect(await communityEnable.text()).toContain('plugin_migration_failed');

      await page.goto('/admin-next/plugins/list');
      await page.getByTestId('plugin-detail-qa').click();
      const drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: '迁移' }).click();
      const migrationsPanel = drawer.getByLabel('迁移');
      await expect(migrationsPanel.getByRole('cell', { name: '失败' })).toBeVisible();
      await expect(migrationsPanel.getByText(errorText)).toBeVisible();
      await migrationsPanel.getByRole('button', { name: '重试' }).first().click();
      await expect(page.getByText('迁移重试完成')).toBeVisible();
      await expect(migrationsPanel.getByRole('cell', { name: '成功' }).first()).toBeVisible();

      const migrations = await pluginMigrations(request, 'qa');
      expect(migrations.summary.failed).toBe(0);

      await ensurePluginEnabled(request, 'qa');
      const restored = await enableCommunityPlugin(request, 1, 'qa');
      expect(restored.ok()).toBeTruthy();

      const audits = await pluginAuditLogs(request, 'qa', { action: 'plugin.migration', page_size: '50' });
      const actions = (audits.items || []).map((item) => item.action).join('\n');
      expect(actions).toContain('plugin.migration.failed');
      expect(actions).toContain('plugin.migration.retry');
      expect(actions).toContain('plugin.migration.success');
    } finally {
      await retryPluginMigration(request, 'qa', 'qa_questions').catch(() => {});
      await retryPluginMigration(request, 'qa', 'qa_answers').catch(() => {});
      await ensurePluginEnabled(request, 'qa').catch(() => {});
      await enableCommunityPlugin(request, 1, 'qa').catch(() => {});
    }
  });

  test('hook failure injection records blocking and non-blocking failures', async ({ page, request }) => {
    const blockingError = `E2E blocking hook failure ${Date.now()}`;
    const nonBlockingError = `E2E non-blocking hook failure ${Date.now()}`;

    const createQuestion = async (titlePrefix) =>
      userPost(request, '/api/v1/topics', {
        community_id: 1,
        community_slug: 'php',
        category_id: 102,
        content_type: 'question',
        title: uniqueTitle(titlePrefix),
        summary: 'E2E HookBus 自动化测试摘要',
        content: '这是一段用于 HookBus E2E 自动化测试的正文内容，长度满足发布校验。',
        tags: [],
      });

    try {
      await setCategoryEnabled(request, 102, true);

      await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', blockingError);
      const blocked = await createQuestion('E2E Hook Blocked Topic');
      expect(blocked.ok()).toBeFalsy();
      expect(await blocked.text()).toContain(blockingError);

      await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', '', true);

      await injectFailedPluginHook(request, 'qa', 'AfterCreateContent', 'non_blocking', nonBlockingError);
      const created = await createQuestion('E2E Hook NonBlocking Topic');
      expect(created.ok()).toBeTruthy();

      const hookStats = await pluginHooks(request, 'qa');
      const hookText = JSON.stringify(hookStats);
      expect(hookText).toContain('BeforeCreateContent');
      expect(hookText).toContain('AfterCreateContent');
      expect(hookText).toContain(blockingError);
      expect(hookText).toContain(nonBlockingError);

      const blockedAudits = await pluginAuditLogs(request, 'qa', { action: 'plugin.hook.blocked', page_size: '50' });
      expect(JSON.stringify(blockedAudits)).toContain(blockingError);
      const failedAudits = await pluginAuditLogs(request, 'qa', { action: 'plugin.hook.failed', page_size: '50' });
      expect(JSON.stringify(failedAudits)).toContain(nonBlockingError);

      await page.goto('/admin-next/plugins/list');
      await page.getByTestId('plugin-detail-qa').click();
      const drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: 'Hook' }).click();
      await expect(drawer.getByRole('cell', { name: blockingError }).first()).toBeVisible();
      await expect(drawer.getByRole('cell', { name: nonBlockingError }).first()).toBeVisible();
    } finally {
      await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', '', true).catch(() => {});
      await injectFailedPluginHook(request, 'qa', 'AfterCreateContent', 'non_blocking', '', true).catch(() => {});
    }
  });

  test('shows plugin health states and filters plugin audit logs', async ({ page, request }) => {
    const blockingError = `E2E health hook error ${Date.now()}`;
    const migrationError = `E2E health migration error ${Date.now()}`;
    try {
      await retryPluginMigration(request, 'docs', 'docs_spaces').catch(() => {});
      await retryPluginMigration(request, 'docs', 'docs_documents').catch(() => {});
      await ensurePluginEnabled(request, 'docs');
      await retryPluginMigration(request, 'qa', 'qa_questions').catch(() => {});
      await ensurePluginEnabled(request, 'qa');

      await page.goto('/admin-next/plugins/list');
      await expect(page.getByText('文档插件')).toBeVisible();
      await page.getByTestId('plugin-detail-docs').click();
      let drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: '运行状态' }).click();
      await expect(drawer.getByText('健康').first()).toBeVisible();
      await page.getByRole('button', { name: 'Close this dialog' }).click();

      await injectFailedPluginMigration(request, 'qa', 'qa_questions', migrationError);
      await page.reload();
      await page.getByTestId('plugin-detail-qa').click();
      drawer = page.getByTestId('plugin-detail-drawer');
      await page.getByRole('tab', { name: '运行状态' }).click();
      await expect(drawer.getByText('异常').first()).toBeVisible();
      await expect(drawer.getByText(/failed migration|失败迁移|迁移/).first()).toBeVisible();
      await retryPluginMigration(request, 'qa', 'qa_questions');

      const badConfig = await request.put('/api/v1/admin/plugins/qa/config', {
        headers: { Authorization: 'Bearer devhub-admin-1', 'Content-Type': 'application/json' },
        data: { config_json: { allow_anonymous_answer: 'bad', default_question_status: 'draft' } },
      });
      expect(badConfig.ok()).toBeFalsy();
      expect(await badConfig.text()).toMatch(/boolean|enum|允许范围|类型/);

      await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', blockingError);
      const blocked = await userPost(request, '/api/v1/topics', {
        community_id: 1,
        community_slug: 'php',
        category_id: 102,
        content_type: 'question',
        title: uniqueTitle('E2E Health Hook Topic'),
        summary: 'E2E 插件健康状态测试摘要',
        content: '这是一段用于插件健康状态 E2E 自动化测试的正文内容，长度满足发布校验。',
        tags: [],
      });
      expect(blocked.ok()).toBeFalsy();
      expect(await blocked.text()).toContain(blockingError);

      await page.reload();
      await page.getByTestId('plugin-detail-qa').click();
      drawer = page.getByTestId('plugin-detail-drawer');
      await page.getByRole('tab', { name: '运行状态' }).click();
      await expect(drawer.getByRole('row', { name: /Hook 状态.*警告|Hook 状态.*异常/ }).first()).toBeVisible();
      await expect(drawer.getByText(blockingError).first()).toBeVisible();

      await page.getByRole('tab', { name: '审计' }).click();
      const allAudits = await pluginAuditLogs(request, 'qa', {
        action: 'plugin.hook.blocked',
        plugin_code: 'qa',
        page_size: '50',
      });
      expect(JSON.stringify(allAudits)).toContain(blockingError);

      await page.getByTestId('plugin-audit-action-filter').locator('input').fill('plugin.hook.blocked');
      await page.getByTestId('plugin-audit-metadata-filter').locator('input').fill(blockingError);
      await page.getByRole('button', { name: '查询' }).click();
      await expect(drawer.getByText(/暂无审计记录|plugin\.hook\.blocked/).first()).toBeVisible();

      const audits = await pluginAuditLogs(request, 'qa', {
        action: 'plugin.hook.blocked',
        metadata: blockingError,
        plugin_code: 'qa',
        page_size: '20',
      });
      expect(JSON.stringify(audits)).toContain(blockingError);
    } finally {
      await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', '', true).catch(() => {});
      await retryPluginMigration(request, 'qa', 'qa_questions').catch(() => {});
      await ensurePluginEnabled(request, 'qa').catch(() => {});
      await enableCommunityPlugin(request, 1, 'qa').catch(() => {});
    }
  });

  function buildManifest(code, contentType) {
    return JSON.stringify(
      {
        code,
        name: 'E2E Manifest Plugin',
        version: '1.0.0',
        description: 'E2E 使用的 manifest 示例',
        compatible_core_version: '>=1.3.4',
        is_system: false,
        content_types: [
          {
            type: contentType,
            name: 'E2E 内容',
            plugin_code: code,
            create_permission: `${code}.create`,
          },
        ],
        permissions: [
          {
            code: `${code}.create`,
            name: '创建 E2E 内容',
            scope: 'community',
          },
        ],
        menus: [],
        routes: [],
        hooks: [],
        config_schema: { type: 'object', properties: {} },
        migrations: [],
      },
      null,
      2,
    );
  }
});
