import { expect, test } from '@playwright/test';
import { archivePlugin, disablePlugin, ensurePluginEnabled, installPluginManifest, uniqueTitle } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

function buildManifest(code, deps) {
  return JSON.stringify(
    {
      code,
      name: 'E2E Manifest Plugin',
      version: '1.0.0',
      description: uniqueTitle('e2e readiness'),
      is_system: false,
      compatible_core_version: '>=1.3.4',
      content_types: [
        {
          type: `${code}_item`,
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
      dependencies: deps || [],
    },
    null,
    2,
  );
}

test.describe('plugin readiness errors', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('shows blocked readiness when required dependency missing', async ({ page, request }) => {
    const readinessProbe = await request.get('/api/v1/admin/plugins/qa/readiness?action=enable', {
      headers: { Authorization: 'Bearer devhub-admin-1' },
    });
    expect(readinessProbe.ok(), await readinessProbe.text()).toBeTruthy();

    const code = `e2e_readiness_${Date.now()}`;
    const manifest = buildManifest(code, [{ code: 'qa', version: '>=1.0.0', required: true, reason: 'E2E depends on qa' }]);
    let installed = false;
    try {
      await ensurePluginEnabled(request, 'qa');
      await installPluginManifest(request, manifest);
      installed = true;
      await disablePlugin(request, 'qa');

      await page.goto('/admin-next/plugins');
      await page.getByTestId(`plugin-detail-${code}`).click();
      const drawer = page.getByTestId('plugin-detail-drawer');
      await expect(drawer).toBeVisible();

      await page.getByRole('tab', { name: '操作诊断' }).click();
      await expect(page.getByTestId('plugin-readiness-table')).toBeVisible();
      await expect(drawer).toContainText('阻断');
      await expect(drawer).toContainText('plugin_dependency_disabled');
    } finally {
      await ensurePluginEnabled(request, 'qa').catch(() => {});
      if (installed) await archivePlugin(request, code).catch(() => {});
    }
  });
});
