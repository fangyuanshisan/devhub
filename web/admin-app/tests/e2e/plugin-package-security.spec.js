import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package security (checksum + risk report)', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('shows checksum ok and low risk for safe package', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await page.getByTestId('plugin-package-path-input').fill('examples/plugins/demo_notice');
    await page.getByTestId('plugin-package-dry-run').click();

    const panel = page.getByTestId('plugin-package-result');
    await expect(panel).toBeVisible();
    await expect(panel.getByTestId('plugin-package-checksum-status')).toContainText('ok');
    await expect(panel.getByTestId('plugin-package-risk-level')).toContainText('low');
  });

  test('warns when checksums.json is missing', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await page.getByTestId('plugin-package-path-input').fill('examples/plugins/security-fixtures/no_checksums');
    await page.getByTestId('plugin-package-dry-run').click();

    const panel = page.getByTestId('plugin-package-result');
    await expect(panel).toBeVisible();
    await expect(panel.getByTestId('plugin-package-checksum-status')).toContainText('missing');
    await expect(panel.getByTestId('plugin-package-risk-level')).toContainText(/medium|high/);
  });

  test('blocks when checksum mismatched', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await page.getByTestId('plugin-package-path-input').fill('examples/plugins/security-fixtures/checksum_mismatch');
    await page.getByTestId('plugin-package-dry-run').click();

    const panel = page.getByTestId('plugin-package-result');
    await expect(panel).toBeVisible();
    await expect(panel.getByTestId('plugin-package-checksum-status')).toContainText('failed');
    await expect(panel.getByTestId('plugin-package-risk-level')).toContainText('blocked');
    await expect(panel.getByTestId('plugin-package-checksum-mismatched')).toBeVisible();
  });

  test('blocks when dangerous file exists', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await page.getByTestId('plugin-package-path-input').fill('examples/plugins/security-fixtures/dangerous_shell');
    await page.getByTestId('plugin-package-dry-run').click();

    const panel = page.getByTestId('plugin-package-result');
    await expect(panel).toBeVisible();
    await expect(panel.getByTestId('plugin-package-risk-level')).toContainText('blocked');
    await expect(panel.getByTestId('plugin-package-risk-items')).toContainText('plugin_package_dangerous_file');
  });
});

