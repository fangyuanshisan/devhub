import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package real signature verification', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('shows verified, unknown and blocked signature states', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();

    await page.getByTestId('plugin-package-repo-root').fill('plugins-local/repository-fixtures-signature');
    await page.getByTestId('plugin-package-repo-scan').click();
    const table = page.getByTestId('plugin-package-repo-table');
    await expect(table).toBeVisible();

    const signedRow = table.getByRole('row', { name: /repository-fixtures-signature\/demo_signed_notice/ }).first();
    await expect(signedRow).toContainText('trusted');
    await expect(signedRow).toContainText('verified');
    await expect(signedRow).not.toContainText('structural_only');

    const unknownRow = table.getByRole('row', { name: /repository-fixtures-signature\/signature_unknown_publisher/ }).first();
    await expect(unknownRow).toContainText('unknown');
    await expect(unknownRow).toContainText('verified');

    const invalidRow = table.getByRole('row', { name: /repository-fixtures-signature\/signature_invalid_json/ }).first();
    await expect(invalidRow).toContainText('blocked');

    await signedRow.getByTestId('plugin-package-repo-detail-btn').click();
    await expect(page.getByTestId('plugin-package-repo-detail-signature')).toContainText('verified');
    await expect(page.getByTestId('plugin-package-repo-detail-signature')).toContainText('devhub-official');
    await expect(page.getByTestId('plugin-package-repo-detail-signed-files')).toContainText('checksums.json');

    await expect(page.getByTestId('plugin-install-page')).not.toContainText('远程市场入口');
    await expect(page.getByTestId('plugin-install-page')).not.toContainText('自动下载入口');
    await expect(page.getByTestId('plugin-install-page')).not.toContainText('动态加载入口');
  });
});
