import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin config key rotation', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('shows key status and supports rotation dry-run + re-encrypt', async ({ page }) => {
    await page.goto('/admin-next/plugins/config-keys');
    await expect(page.getByTestId('plugin-config-keys-page')).toContainText('插件配置密钥');
    await expect(page.getByTestId('plugin-config-keys-page')).toContainText('不支持 KMS');
    await expect(page.getByTestId('plugin-config-keys-page')).not.toContainText('enc:v2:');

    await page.getByTestId('plugin-config-keys-rotation-dryrun').click();
    await expect(page.getByTestId('plugin-config-keys-rotation-status')).toBeVisible();

    // Should not expose ciphertext markers in UI.
    await expect(page.getByTestId('plugin-config-keys-page')).not.toContainText('enc:v1:');
    await expect(page.getByTestId('plugin-config-keys-page')).not.toContainText('enc:v2:');

    // Re-encrypt is allowed when not blocked; if blocked in env, just ensure the button exists.
    const btn = page.getByTestId('plugin-config-keys-reencrypt');
    await expect(btn).toBeVisible();
    if (!(await btn.isDisabled())) {
      await btn.click();
      await page.getByRole('button', { name: '确认' }).click();
      await expect(page.getByTestId('plugin-config-keys-page')).toContainText('re-encrypt');
    }
  });
});

