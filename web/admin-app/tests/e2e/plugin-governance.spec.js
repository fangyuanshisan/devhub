import { expect, test } from '@playwright/test';
import {
  disableCommunityPlugin,
  disablePlugin,
  enableCommunityPlugin,
  ensurePluginEnabled,
  injectFailedPluginHook,
  injectFailedPluginMigration,
  pluginAuditLogs,
  pluginHooks,
  pluginMigrations,
  retryPluginMigration,
  setCategoryEnabled,
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
    await page.goto('/admin-next/plugins');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
    await expect(page.getByTestId('plugin-stats')).toContainText('全部插件');
    await expect(page.getByText('问答插件')).toBeVisible();

    await page.getByPlaceholder('搜索 code / name').fill('qa');
    await expect(page.getByText('问答插件')).toBeVisible();
    await expect(page.getByText('文档插件')).toBeHidden();
  });

  test('opens plugin detail tabs and shows schema validation errors', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await page.getByTestId('plugin-detail-qa').click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();

    for (const tabName of ['概览', '内容类型', '权限', '菜单', '配置', 'Hooks', '路由', '审计']) {
      await page.getByRole('tab', { name: tabName }).click();
      await expect(page.getByRole('tab', { name: tabName })).toHaveAttribute('aria-selected', 'true');
    }

    await page.getByRole('tab', { name: '配置' }).click();
    await expect(page.getByRole('button', { name: 'config_schema' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'resolved_config' })).toBeVisible();

    await page.getByTestId('json-clear-object').click();
    await expect(page.getByTestId('schema-error-box')).toContainText('required');
    await expect(page.getByTestId('plugin-global-config-save')).toBeDisabled();
  });

  test('shows impact before global disable and blocks community enable when globally disabled', async ({ page, request }) => {
    try {
      await page.goto('/admin-next/plugins');
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
    await expect(page.getByTestId('plugin-json-editor').getByText('子站 config_json')).toBeVisible();
    await page.getByTestId('json-clear-object').click();
    await expect(page.getByTestId('schema-error-box')).toContainText('required');
    await expect(page.getByTestId('community-plugin-config-save')).toBeDisabled();
  });

  test('opens generic plugin content page with filters', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await page.getByTestId('plugin-manage-qa').click();
    await expect(page).toHaveURL(/\/admin-next\/qa/);
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await expect(page.getByText('content_type：')).toBeVisible();
    await expect(page.getByTestId('plugin-content-community-filter')).toBeVisible();
    await expect(page.getByTestId('plugin-content-status-filter')).toBeVisible();
    await page.getByTestId('plugin-content-back').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins/);
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
      expect(await globalEnable.text()).toContain('失败迁移');

      const communityEnable = await enableCommunityPlugin(request, 1, 'qa');
      expect(communityEnable.ok()).toBeFalsy();
      expect(await communityEnable.text()).toContain('失败迁移');

      await page.goto('/admin-next/plugins');
      await page.getByTestId('plugin-detail-qa').click();
      const drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: '迁移' }).click();
      const migrationsPanel = drawer.getByLabel('迁移');
      await expect(migrationsPanel.getByRole('cell', { name: 'failed' })).toBeVisible();
      await expect(migrationsPanel.getByText(errorText)).toBeVisible();
      await migrationsPanel.getByRole('button', { name: '重试' }).first().click();
      await expect(page.getByText('迁移重试完成')).toBeVisible();
      await expect(migrationsPanel.getByRole('cell', { name: 'success' }).first()).toBeVisible();

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

      await page.goto('/admin-next/plugins');
      await page.getByTestId('plugin-detail-qa').click();
      const drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: 'Hooks' }).click();
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

      await page.goto('/admin-next/plugins');
      await expect(page.getByText('文档插件')).toBeVisible();
      await page.getByTestId('plugin-detail-docs').click();
      let drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();
      await page.getByRole('tab', { name: '运行状态' }).click();
      await expect(drawer.getByText('healthy').first()).toBeVisible();
      await page.getByRole('button', { name: 'Close this dialog' }).click();

      await injectFailedPluginMigration(request, 'qa', 'qa_questions', migrationError);
      await page.reload();
      await page.getByTestId('plugin-detail-qa').click();
      drawer = page.getByTestId('plugin-detail-drawer');
      await page.getByRole('tab', { name: '运行状态' }).click();
      await expect(drawer.getByText('error').first()).toBeVisible();
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
      await expect(drawer.getByRole('row', { name: /hook_status.*warning|hook_status.*error/ }).first()).toBeVisible();
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
});
