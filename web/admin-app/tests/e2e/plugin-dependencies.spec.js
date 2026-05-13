import { expect, test } from '@playwright/test';
import { adminPost, archivePlugin, ensurePluginEnabled, installPluginManifest, restorePlugin } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin dependency and core compatibility governance', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ page, request }) => {
    await ensurePluginEnabled(request, 'qa');
    await seedAdminSession(page);
  });

  test('shows dependency matrix for satisfied, missing optional, missing required and incompatible core manifests', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();

    await openManifestValidate(page, buildManifest(`e2e_dep_ok_${Date.now()}`, { dependencies: [{ code: 'qa', version: '>=1.0.0', required: true, reason: 'E2E satisfied dependency' }] }));
    await expect(page.getByTestId('plugin-dependency-summary')).toContainText('已满足');
    await expect(page.getByTestId('plugin-dependency-summary')).toContainText('qa');
    await page.getByTestId('plugin-result-close').click();

    await openManifestValidate(page, buildManifest(`e2e_dep_optional_${Date.now()}`, { dependencies: [{ code: `missing_optional_${Date.now()}`, version: '>=1.0.0', required: false, reason: 'E2E optional dependency' }] }));
    await expect(page.getByTestId('plugin-dependency-summary')).toContainText('可选依赖缺失');
    await expect(page.getByTestId('plugin-result-panel')).toContainText('警告');
    await page.getByTestId('plugin-result-close').click();

    await openManifestValidate(page, buildManifest(`e2e_dep_missing_${Date.now()}`, { dependencies: [{ code: `missing_required_${Date.now()}`, version: '>=1.0.0', required: true, reason: 'E2E required dependency' }] }));
    await expect(page.getByTestId('plugin-dependency-summary')).toContainText('缺失');
    await expect(page.getByTestId('plugin-result-panel')).toContainText('错误');
    await page.getByTestId('plugin-result-close').click();

    await openManifestValidate(page, buildManifest(`e2e_dep_core_${Date.now()}`, { min_core_version: '99.0.0', compatible_core_version: '>=99.0.0' }));
    await expect(page.getByTestId('plugin-result-summary')).toContainText('不兼容');
    await expect(page.getByTestId('plugin-result-panel')).toContainText('plugin_core_version_incompatible');
    await page.getByTestId('plugin-result-close').click();
  });

  test('upgrade preview blocks new required dependency and detail drawer shows dependencies tab', async ({ page, request }) => {
    const code = `e2e_dep_upgrade_${Date.now()}`;
    let installed = false;
    try {
      await installPluginManifest(request, buildManifestObject(code, { dependencies: [{ code: 'qa', version: '>=1.0.0', required: true, reason: 'E2E base dependency' }] }));
      installed = true;
      await page.goto('/admin-next/plugins');
      await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
      await page.getByTestId('plugin-search').fill(code);
      await expect(page.getByText(code, { exact: true }).first()).toBeVisible();

      await page.getByTestId(`plugin-detail-${code}`).click();
      await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();
      await page.getByRole('tab', { name: '依赖' }).click();
      await expect(page.getByTestId('plugin-dependencies-table')).toContainText('qa');
      await expect(page.getByTestId('plugin-dependencies-table')).toContainText('已满足');
      await page.getByTestId('plugin-detail-drawer').getByRole('button', { name: '查看依赖插件' }).first().click();
      await expect(page.getByTestId('plugin-detail-drawer')).toContainText('问答插件');
      await page.keyboard.press('Escape');

      await page.getByTestId('plugin-search').fill(code);
      await page.getByRole('row', { name: new RegExp(code) }).getByRole('button', { name: '更多' }).click();
      await page.getByTestId(`plugin-upgrade-preview-${code}`).click();
      await expect(page.getByTestId('plugin-manifest-panel')).toBeVisible();
      const editor = page.getByTestId('plugin-manifest-input');
      const current = JSON.parse(await editor.inputValue());
      current.version = '2.0.0';
      current.dependencies = [
        { code: 'qa', version: '>=1.0.0', required: true, reason: 'E2E base dependency' },
        { code: `missing_upgrade_${Date.now()}`, version: '>=1.0.0', required: true, reason: 'E2E missing upgrade dependency' },
      ];
      await editor.fill(JSON.stringify(current, null, 2));
      await page.getByTestId('plugin-manifest-submit').click();
      await expect(page.getByTestId('plugin-upgrade-dependency-matrix')).toContainText('缺失');
      await expect(page.getByTestId('plugin-upgrade-dependency-diff')).toContainText('missing_upgrade_');

    } finally {
      if (installed) {
        await archivePlugin(request, code).catch(() => {});
        await restorePlugin(request, code).catch(() => {});
      }
    }
  });
});

async function openManifestValidate(page, manifestText) {
  await page.getByTestId('plugin-manifest-validate').click();
  await expect(page.getByTestId('plugin-manifest-panel')).toBeVisible();
  await page.getByTestId('plugin-manifest-input').fill(manifestText);
  await page.getByTestId('plugin-manifest-submit').click();
  await expect(page.getByTestId('plugin-result-panel')).toBeVisible();
}

function buildManifest(code, overrides = {}) {
  return JSON.stringify(buildManifestObject(code, overrides), null, 2);
}

function buildManifestObject(code, overrides = {}) {
  const contentType = overrides.content_type || `${code}_content`;
  return {
    code,
    name: `E2E Dependency ${code}`,
    version: overrides.version || '1.0.0',
    description: 'E2E dependency governance manifest',
    min_core_version: overrides.min_core_version || '1.4.0',
    compatible_core_version: overrides.compatible_core_version || '>=1.4.0',
    is_system: false,
    content_types: [contentType],
    content_type_definitions: [{
      type: contentType,
      name: 'E2E Dependency Content',
      plugin_code: code,
      create_permission: `${code}.content.create`,
      edit_permission: `${code}.content.edit`,
      delete_permission: `${code}.content.delete`,
      audit_permission: `${code}.content.audit`,
      default_status: 'draft',
      allow_comment: true,
      allow_like: true,
      allow_favorite: true,
      seo_type: 'Article',
    }],
    permissions: [
      { code: `${code}.content.create`, name: '创建内容', scope: 'community' },
      { code: `${code}.content.edit`, name: '编辑内容', scope: 'own' },
      { code: `${code}.content.delete`, name: '删除内容', scope: 'own' },
      { code: `${code}.content.audit`, name: '审核内容', scope: 'community' },
      { code: `${code}.manage`, name: '管理插件', scope: 'global' },
    ],
    menus: [{ code: `${code}.admin`, title: 'E2E Dependency', path: `/admin-next/${code}`, area: 'admin', location: 'admin', permission: `${code}.manage`, plugin_code: code, sort_order: 300 }],
    routes: [],
    config_schema: { type: 'object', properties: {} },
    dependencies: overrides.dependencies || [],
    hooks: [],
    migrations: [],
  };
}
