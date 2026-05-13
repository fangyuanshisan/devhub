import { expect, test } from '@playwright/test';
import { apiPost, adminToken, ensurePluginStatus } from './helpers/api.js';

test.describe.configure({ mode: 'serial' });

test.describe('frontend plugin navigation visibility', () => {
  test.afterEach(async ({ request }) => {
    await ensurePluginStatus(request, 'qa', 'enabled');
    await apiPost(request, '/api/v1/admin/communities/1/plugins/qa/enable', {}, adminToken).catch(() => {});
  });

  test('community navigation hides plugin items when community plugin disabled', async ({ page, request }) => {
    await ensurePluginStatus(request, 'qa', 'enabled');
    await apiPost(request, '/api/v1/admin/communities/1/plugins/qa/enable', {}, adminToken).catch(() => {});

    await page.goto('/');
    await expect(page.getByTestId('frontend-site-header')).toBeVisible();

    const scopeSelect = page.locator('[data-header-search] select[name="community_slug"]');
    await scopeSelect.selectOption('php');

    const questionNav = page.locator('[data-plugin-nav="question"]');
    await expect(questionNav).toBeVisible();
    await expect(questionNav).not.toHaveClass(/is-hidden/);

    await apiPost(request, '/api/v1/admin/communities/1/plugins/qa/disable', {}, adminToken);

    // Force refresh by toggling scope.
    await scopeSelect.selectOption('');
    await scopeSelect.selectOption('php');

    await expect(questionNav).toHaveClass(/is-hidden/);
  });
});

