import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin pages navigation', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('redirects /plugins to overview and renders grouped sub navigation', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await expect(page).toHaveURL(/\/admin-next\/plugins\/overview/);
    await expect(page.getByTestId('plugin-overview-page')).toBeVisible();

    const subNav = page.getByTestId('admin-sub-nav');
    await expect(subNav).toBeVisible();
    await expect(subNav.getByTestId('admin-sub-nav-group-ops')).toBeVisible();
    await expect(subNav.getByTestId('admin-sub-nav-group-govern')).toBeVisible();
    await expect(subNav.getByTestId('admin-sub-nav-group-runtime')).toBeVisible();
    await expect(subNav.getByTestId('admin-sub-nav-group-dev')).toBeVisible();
  });

  test('navigates between plugin function pages and keeps legacy routes working', async ({ page }) => {
    await page.goto('/admin-next/plugins/overview');
    await expect(page.getByTestId('plugin-overview-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-list').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/list/);
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();

    // plugin detail drawer still works on list page
    await page.getByTestId('plugin-detail-qa').click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();
    await page.getByRole('button', { name: 'Close this dialog' }).click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeHidden();

    await page.getByTestId('admin-sub-nav-content').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/content/);
    await expect(page.getByTestId('plugin-content-hub-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-install').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/install/);
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();
    await page.getByTestId('plugin-manifest-validate').click();
    await expect(page.getByTestId('plugin-manifest-panel')).toBeVisible();
    await page.getByTestId('plugin-manifest-cancel').click();

    await page.getByTestId('admin-sub-nav-config').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/config/);
    await expect(page.getByTestId('plugin-config-hub-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-dependencies').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/dependencies/);
    await expect(page.getByTestId('plugin-dependencies-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-hooks').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/hooks/);
    await expect(page.getByTestId('plugin-hooks-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-events').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/events/);
    await expect(page.getByTestId('plugin-events-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-searchIndex').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/search-index/);
    await expect(page.getByTestId('plugin-search-index-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-navigation').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/navigation/);
    await expect(page.getByTestId('plugin-navigation-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-permissions').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/permissions/);
    await expect(page.getByTestId('plugin-permissions-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-audit').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/audit/);
    await expect(page.getByTestId('plugin-audit-page')).toBeVisible();

    await page.getByTestId('admin-sub-nav-developer').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/developer/);
    await expect(page.getByTestId('plugin-developer-page')).toBeVisible();

    // content hub can jump to legacy PluginContent pages
    await page.goto('/admin-next/plugins/content');
    await page.getByTestId('plugin-content-hub-open-qa').click();
    await expect(page).toHaveURL(/\/admin-next\/qa/);
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.getByTestId('plugin-content-back').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins\/content/);

    // legacy routes keep working
    await page.goto('/admin-next/qa');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.goto('/admin-next/plugins/governance');
    await expect(page).toHaveURL(/\/admin-next\/plugins\/overview/);
  });
});
