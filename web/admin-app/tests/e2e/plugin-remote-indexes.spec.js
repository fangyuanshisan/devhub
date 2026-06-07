import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin remote indexes readonly mirror', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
    await mockRemoteIndexApis(page);
  });

  test('manages readonly remote index metadata without download or install actions', async ({ page }) => {
    await page.goto('/admin-next/plugins/remote-indexes');
    await expect(page.getByTestId('plugin-remote-indexes-page')).toContainText('只读镜像');
    await expect(page.getByTestId('plugin-remote-indexes-page')).toContainText('不下载插件包');
    await expect(page.getByTestId('plugin-remote-indexes-page')).toContainText('不安装插件');
    await expect(page.getByTestId('plugin-remote-indexes-page')).not.toContainText('下载按钮');
    await expect(page.getByTestId('plugin-remote-indexes-page')).not.toContainText('自动更新按钮');
    await expect(page.getByTestId('plugin-remote-indexes-page')).not.toContainText('动态加载入口');

    await page.getByTestId('remote-index-create').click();
    await expect(page.getByTestId('remote-index-dialog')).toBeVisible();
    await page.getByTestId('remote-index-source-id').fill('e2e-index');
    await page.getByTestId('remote-index-name').fill('E2E Index');
    await page.getByTestId('remote-index-url').fill('https://example.com/devhub/plugins/index.json');
    await page.getByTestId('remote-index-submit').click();

    await expect(page.getByTestId('remote-index-list')).toContainText('e2e-index');
    await page.getByTestId('remote-index-fetch').click();
    await expect(page.getByTestId('remote-index-list')).toContainText('ok');
    await expect(page.getByTestId('remote-plugin-list')).toContainText('remote_demo');
    await expect(page.getByTestId('remote-plugin-list')).toContainText('trusted');
    await expect(page.getByTestId('remote-plugin-list')).toContainText('installed');
    await expect(page.getByTestId('remote-plugin-list')).toContainText('remote_legacy');
    await expect(page.getByTestId('remote-plugin-list')).toContainText('unknown');
    await expect(page.getByTestId('remote-plugin-list')).toContainText('incompatible');

    await page.getByRole('row', { name: /remote_demo/ }).getByTestId('remote-plugin-detail').click();
    await expect(page.getByTestId('remote-plugin-detail-drawer')).toContainText('package_url');
    await expect(page.getByTestId('remote-plugin-detail-drawer')).toContainText('package_sha256');
    await expect(page.getByTestId('remote-plugin-detail-drawer')).toContainText('publisher');
    await expect(page.getByTestId('remote-plugin-detail-drawer')).toContainText('metadata_only');

    await page.getByRole('button', { name: 'Close' }).click().catch(() => {});
    await page.getByTestId('remote-index-disable').click();
    await expect(page.getByTestId('remote-index-list')).toContainText('disabled');
  });
});

async function mockRemoteIndexApis(page) {
  const source = {
    id: 1,
    source_id: 'e2e-index',
    name: 'E2E Index',
    index_url: 'https://example.com/devhub/plugins/index.json',
    status: 'enabled',
    last_fetch_status: '',
  };
  let created = false;
  let disabled = false;

  await page.route('**/api/v1/admin/plugins/remote-indexes/1/fetch', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ source: { ...source, last_fetch_status: 'ok' }, index_hash: 'sha256:e2e', validation: { valid: true }, document: { plugins: [] } }) }));
  await page.route('**/api/v1/admin/plugins/remote-indexes/1/disable', (route) => {
    disabled = true;
    return route.fulfill({ contentType: 'application/json', body: JSON.stringify({ ...source, status: 'disabled' }) });
  });
  await page.route('**/api/v1/admin/plugins/remote-indexes/1/plugins/remote_demo', (route) => route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({
      source,
      plugin: { code: 'remote_demo', name: 'Remote Demo', latest_version: '1.0.0', description: '只读测试插件' },
      installed: true,
      local_version: '1.0.0',
      readonly: true,
      readonly_message: '当前仅展示远程索引元数据；不会下载、安装、执行代码或动态加载前端资产。',
      versions: [{
        version: '1.0.0',
        package_url: 'https://example.com/packages/remote_demo.zip',
        package_sha256: 'a'.repeat(64),
        signature_url: 'https://example.com/packages/remote_demo.signature.json',
        publisher_id: 'devhub-official',
        public_key_id: 'devhub-official-2026',
        publisher_trust_status: 'trusted',
        risk_level: 'low',
        risk_items: [],
      }],
    }),
  }));
  await page.route('**/api/v1/admin/plugins/remote-indexes/1/plugins**', (route) => {
    if (!new URL(route.request().url()).pathname.endsWith('/plugins')) return route.fallback();
    return route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          remotePlugin('remote_demo', 'trusted', 'installed', 'compatible', 'low'),
          remotePlugin('remote_legacy', 'unknown', 'not_installed', 'incompatible', 'blocked'),
        ],
        pagination: { page: 1, page_size: 20, total: 2 },
        summary: { total: 2, trusted: 1, unknown: 1, blocked: 0, installed: 1, update_available: 0, incompatible: 1 },
      }),
    });
  });
  await page.route('**/api/v1/admin/plugins/remote-indexes**', async (route) => {
    const req = route.request();
    if (!new URL(req.url()).pathname.endsWith('/remote-indexes')) return route.fallback();
    if (req.method() === 'GET') {
      return route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          items: created ? [{ ...source, status: disabled ? 'disabled' : 'enabled', last_fetch_status: 'ok' }] : [],
          pagination: { page: 1, page_size: 20, total: created ? 1 : 0 },
          summary: { total: created ? 1 : 0, enabled: created && !disabled ? 1 : 0, disabled: disabled ? 1 : 0, failed: 0 },
        }),
      });
    }
    if (req.method() === 'POST') {
      created = true;
      return route.fulfill({ contentType: 'application/json', body: JSON.stringify(source) });
    }
    return route.continue();
  });
}

function remotePlugin(code, trust, versionStatus, coreStatus, riskLevel) {
  return {
    code,
    name: code === 'remote_demo' ? 'Remote Demo' : 'Remote Legacy',
    latest_version: '1.0.0',
    publisher_id: trust === 'trusted' ? 'devhub-official' : 'unknown-publisher',
    public_key_id: trust === 'trusted' ? 'devhub-official-2026' : 'unknown-key',
    publisher_trust_status: trust,
    version_status: versionStatus,
    risk_level: riskLevel,
    risk_summary: riskLevel === 'low' ? '远程索引元数据未发现明显风险' : 'blocked：当前 Core 不兼容',
    core_compatibility: { status: coreStatus, messages: [coreStatus] },
    package_sha256: 'a'.repeat(64),
    signature_url: code === 'remote_demo' ? 'https://example.com/sig.json' : '',
  };
}
