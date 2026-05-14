import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package signature + trusted publishers', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('shows signature trust/verification status in repository list + detail', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();

    await page.getByTestId('plugin-package-repo-root').fill('plugins-local/repository-fixtures-signature');
    await page.getByTestId('plugin-package-repo-scan').click();

    const table = page.getByTestId('plugin-package-repo-table');
    await expect(table).toBeVisible();

    // trusted + verified
    const signedRow = table.getByRole('row', { name: /repository-fixtures-signature\/demo_signed_notice/ }).first();
    await expect(signedRow).toContainText('trusted');
    await expect(signedRow).toContainText('verified');

    // unsigned (signature missing)
    const unsignedRow = table.getByRole('row', { name: /repository-fixtures-signature\/demo_notice/ }).first();
    await expect(unsignedRow).toContainText('unsigned');
    await expect(unsignedRow).toContainText('missing');

    // revoked publisher => blocked
    const revokedRow = table.getByRole('row', { name: /repository-fixtures-signature\/publisher_revoked/ }).first();
    await expect(revokedRow).toContainText('blocked');

    // open detail and inspect signed_files/unsigned_files
    await signedRow.getByTestId('plugin-package-repo-detail-btn').click();
    await expect(page.getByTestId('plugin-package-repo-detail-content')).toBeVisible();
    await expect(page.getByTestId('plugin-package-repo-detail-signature')).toBeVisible();
    await expect(page.getByTestId('plugin-package-repo-detail-signed-files')).toBeVisible();
  });
});
