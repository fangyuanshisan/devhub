import { expect, test } from '@playwright/test';
import { adminGet } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

test.describe('system effective config troubleshooting', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      window.__devhubCopiedText = '';
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: async (text) => {
            window.__devhubCopiedText = String(text || '');
          },
        },
        configurable: true,
      });
    });
    await seedAdminSession(page);
  });

  test('renders redacted troubleshooting view and copies redacted diagnostics', async ({ page, request }) => {
    const response = await adminGet(request, '/api/v1/admin/system/effective-config');
    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    const diagnosticText = String(body.diagnostic_text || '');
    expect(diagnosticText).toContain('external_services');
    expect(diagnosticText).not.toContain('Authorization');
    expect(diagnosticText).not.toContain('encrypted_value');
    expect(diagnosticText).not.toContain('DEVHUB_PLUGIN_CONFIG_KEYS');

    await page.goto('/admin-next/');
    await page.evaluate(() => {
      window.history.pushState({}, '', '/admin-next/system?tab=effective');
      window.dispatchEvent(new PopStateEvent('popstate'));
    });
    await expect(page.getByTestId('system-effective-metrics')).toBeVisible();
    await expect(page.getByTestId('system-effective-runtime')).toBeVisible();
    await expect(page.getByTestId('system-effective-secretcenter')).toBeVisible();
    await expect(page.getByTestId('system-effective-webhook-callback')).toBeVisible();
    await expect(page.getByText('插件 external_service 运行配置')).toBeVisible();
    await expect(page.getByText('external_service HTTP Allowlist', { exact: true })).toBeVisible();
    await expect(page.locator('body')).not.toContainText('undefined');
    await expect(page.locator('body')).not.toContainText('encrypted_value');

    await page.getByTestId('system-copy-diagnostics').click();
    const copied = await page.evaluate(() => window.__devhubCopiedText || '');
    expect(copied).toContain('external_services');
    expect(copied).not.toContain('Authorization');
    expect(copied).not.toContain('encrypted_value');
    expect(copied).not.toContain('DEVHUB_PLUGIN_CONFIG_KEYS');
  });
});
