import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin versions repository and upgrade diff', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
    await mockVersionApis(page);
  });

  test('shows version sources and upgrade diff before approval', async ({ page }) => {
    await page.goto('/admin-next/plugins/versions');
    await expect(page.getByTestId('plugin-versions-page')).toContainText('插件版本仓库');
    await expect(page.getByTestId('plugin-versions-page')).toContainText('不会自动升级');
    await expect(page.getByTestId('plugin-versions-page')).toContainText('不执行代码');

    await expect(page.getByTestId('plugin-version-list')).toContainText('demo_notice');
    await expect(page.getByTestId('plugin-version-list')).toContainText('1.0.0');
    await expect(page.getByTestId('plugin-version-list')).toContainText('1.1.0');
    await expect(page.getByTestId('plugin-version-list')).toContainText('remote_index');

    await page.getByRole('row', { name: /demo_notice/ }).getByTestId('plugin-version-detail-btn').click();
    await expect(page.getByTestId('plugin-version-detail-drawer')).toContainText('local_package');
    await expect(page.getByTestId('plugin-version-detail-drawer')).toContainText('remote_index');
    await expect(page.getByTestId('plugin-version-detail-drawer')).toContainText('只读');

    await page.getByRole('row', { name: /1.1.0.*local_package/ }).getByTestId('plugin-upgrade-diff-btn').click();
    await expect(page.getByTestId('plugin-upgrade-diff-drawer')).toContainText('1.0.0 → 1.1.0');
    await expect(page.getByTestId('plugin-upgrade-diff')).toContainText('权限变化');
    await expect(page.getByTestId('plugin-upgrade-diff')).toContainText('配置 schema');
    await expect(page.getByTestId('plugin-upgrade-diff')).toContainText('依赖变化');
    await expect(page.getByTestId('plugin-upgrade-diff')).toContainText('新增高危插件管理权限');
    await expect(page.getByTestId('plugin-upgrade-diff')).toContainText('[REDACTED]');
    await expect(page.getByTestId('plugin-upgrade-submit-approval')).toBeEnabled();

    await page.getByTestId('plugin-upgrade-submit-approval').click();
    await expect(page.getByText('已提交升级审批')).toBeVisible();
  });
});

async function mockVersionApis(page) {
  await page.route('**/*', (route) => {
    const req = route.request();
    const url = new URL(req.url());
    if (url.pathname.endsWith('/admin/plugins/versions')) {
      return route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          items: [{
            plugin_code: 'demo_notice',
            plugin_name: 'Demo Notice',
            installed_version: '1.0.0',
            latest_local_version: '1.1.0',
            latest_remote_version: '1.2.0',
            update_available: true,
            sources: ['installed', 'local_package', 'remote_index'],
          }],
          pagination: { page: 1, page_size: 20, total: 1 },
          summary: { total: 1, installed: 1, local_packages: 1, remote_index: 1, update_available: 1, readonly: 1 },
        }),
      });
    }
    if (url.pathname.endsWith('/admin/plugins/demo_notice/versions')) {
      return route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          plugin_code: 'demo_notice',
          plugin_name: 'Demo Notice',
          installed_version: '1.0.0',
          versions: [
            { plugin_code: 'demo_notice', plugin_name: 'Demo Notice', version: '1.0.0', source: 'installed', status: 'installed', is_installed: true, risk_level: 'low' },
            { plugin_code: 'demo_notice', plugin_name: 'Demo Notice', version: '1.1.0', source: 'local_package', status: 'available', package_path: 'storage/plugins/packages/demo_notice-1.1.0', risk_level: 'low', signature_status: 'verified', trust_status: 'trusted', installed_version: '1.0.0', is_upgrade_candidate: true },
            { plugin_code: 'demo_notice', plugin_name: 'Demo Notice', version: '1.2.0', source: 'remote_index', status: 'readonly', remote_source_id: 'e2e-index', risk_level: 'warning', readonly: true, readonly_message: '远程索引版本仅展示元数据；不会下载、安装或执行代码。' },
          ],
        }),
      });
    }
    if (url.pathname.endsWith('/admin/plugins/demo_notice/versions/1.1.0/upgrade-diff')) {
      return route.fulfill({ contentType: 'application/json', body: JSON.stringify(upgradeDiffPayload()) });
    }
    if (url.pathname.endsWith('/admin/plugins/approvals') && req.method() === 'POST') {
      return route.fulfill({ contentType: 'application/json', body: JSON.stringify({ id: 88, status: 'pending', action: 'upgrade', plugin_code: 'demo_notice' }) });
    }
    return route.fallback();
  });
  await page.route(/.*\/api\/v1\/admin\/plugins\/versions(?:\?.*)?$/, (route) => {
    return route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        items: [{
          plugin_code: 'demo_notice',
          plugin_name: 'Demo Notice',
          installed_version: '1.0.0',
          latest_local_version: '1.1.0',
          latest_remote_version: '1.2.0',
          update_available: true,
          sources: ['installed', 'local_package', 'remote_index'],
        }],
        pagination: { page: 1, page_size: 20, total: 1 },
        summary: { total: 1, installed: 1, local_packages: 1, remote_index: 1, update_available: 1, readonly: 1 },
      }),
    });
  });
  await page.route(/.*\/api\/v1\/admin\/plugins\/demo_notice\/versions(?:\?.*)?$/, (route) => route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({
      plugin_code: 'demo_notice',
      plugin_name: 'Demo Notice',
      installed_version: '1.0.0',
      versions: [
        { plugin_code: 'demo_notice', plugin_name: 'Demo Notice', version: '1.0.0', source: 'installed', status: 'installed', is_installed: true, risk_level: 'low' },
        { plugin_code: 'demo_notice', plugin_name: 'Demo Notice', version: '1.1.0', source: 'local_package', status: 'available', package_path: 'storage/plugins/packages/demo_notice-1.1.0', risk_level: 'low', signature_status: 'verified', trust_status: 'trusted', installed_version: '1.0.0', is_upgrade_candidate: true },
        { plugin_code: 'demo_notice', plugin_name: 'Demo Notice', version: '1.2.0', source: 'remote_index', status: 'readonly', remote_source_id: 'e2e-index', risk_level: 'warning', readonly: true, readonly_message: '远程索引版本仅展示元数据；不会下载、安装或执行代码。' },
      ],
    }),
  }));
  await page.route('**/api/v1/admin/plugins/demo_notice/versions/1.1.0/upgrade-diff', (route) => route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify(upgradeDiffPayload()),
  }));
  await page.route('**/api/v1/admin/plugins/approvals', (route) => {
    if (route.request().method() !== 'POST') return route.fallback();
    return route.fulfill({ contentType: 'application/json', body: JSON.stringify({ id: 88, status: 'pending', action: 'upgrade', plugin_code: 'demo_notice' }) });
  });
}

function upgradeDiffPayload() {
  return {
    plugin_code: 'demo_notice',
    current_version: '1.0.0',
    target_version: '1.1.0',
    source: 'local_package',
    status: 'warning',
    summary: { added: 3, removed: 0, changed: 2, high_risk: 2, blocked: 0 },
    risk_report: { level: 'high', score: 70, summary: '升级差异包含 2 个高风险变更' },
    diff_sections: [
      { section: 'permissions', title: '权限变化', risk_level: 'high', items: [{ section: 'permissions', path: 'permissions.demo_notice.manage', type: 'added', risk_level: 'high', before: null, after: { code: 'demo_notice.manage', risk: 'high' }, message: '新增高危插件管理权限' }] },
      { section: 'config_schema', title: '配置 schema', risk_level: 'high', items: [{ section: 'config_schema', path: 'config_schema.oauth.app_secret', type: 'changed', risk_level: 'high', before: '[REDACTED]', after: '[REDACTED]', message: '配置字段类型变化可能导致现有配置不兼容' }] },
      { section: 'dependencies', title: '依赖变化', risk_level: 'high', items: [{ section: 'dependencies', path: 'dependencies.search', type: 'added', risk_level: 'high', before: null, after: { code: 'search', required: true }, message: '新增 required 依赖会阻断缺失环境升级' }] },
    ],
  };
}
