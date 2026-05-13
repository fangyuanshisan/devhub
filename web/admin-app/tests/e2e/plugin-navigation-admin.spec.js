import { expect, test } from '@playwright/test';
import { adminPost } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';
import { expectAdminPageReady } from './helpers/selectors.js';

test.describe.configure({ mode: 'serial' });

test.describe('admin plugin frontend menu preview', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedAdminSession(page);
    await adminPost(request, '/api/v1/admin/plugins/qa/enable').catch(() => {});
  });

  test('shows frontend menus preview with visibility and reasons', async ({ page, request }) => {
    try {
      await page.goto('/admin-next/plugins');
      await expectAdminPageReady(page, 'admin-plugins-page');

      await page.getByTestId('plugin-detail-qa').click();
      await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();

      await page.getByRole('tab', { name: '菜单' }).click();
      await expect(page.getByTestId('plugin-frontend-menus-preview')).toBeVisible();

      await page.getByPlaceholder('子站 slug（可选）').fill('php');
      await page.getByTestId('plugin-menu-preview-refresh').click();

      const preview = page.getByTestId('plugin-frontend-menus-preview');
      await expect(preview).toContainText('community_nav');
      await expect(preview).toContainText('问答');

      // Disable plugin and verify the preview shows blocked reason.
      await adminPost(request, '/api/v1/admin/plugins/qa/disable');
      await page.getByTestId('plugin-menu-preview-refresh').click();
      await expect(preview).toContainText('plugin_disabled');
    } finally {
      await adminPost(request, '/api/v1/admin/plugins/qa/enable').catch(() => {});
    }
  });
});
