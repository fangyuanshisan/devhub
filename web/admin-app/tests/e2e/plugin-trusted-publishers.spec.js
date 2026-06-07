import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

const PUBLIC_KEY = 'Jf8LxG7EtiK11AHz1Ce3zxPdzIoIk0cUWmPFsNyJAi0=';

test.describe('plugin trusted publishers', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('creates, edits, blocks and restores trusted publisher', async ({ page }) => {
    const id = `e2e-publisher-${Date.now()}`;
    await page.goto('/admin-next/plugins/trusted-publishers');
    await expect(page.getByTestId('plugin-trusted-publishers-page')).toContainText('可信发布者');
    await expect(page.getByTestId('plugin-trusted-publishers-page')).toContainText('不支持远程可信源同步');
    await expect(page.getByTestId('plugin-trusted-publishers-page')).not.toContainText('远程市场入口');
    await expect(page.getByTestId('plugin-trusted-publishers-page')).not.toContainText('自动下载入口');
    await expect(page.getByTestId('plugin-trusted-publishers-page')).not.toContainText('动态加载入口');

    await page.getByTestId('trusted-publisher-create').click();
    await expect(page.getByTestId('trusted-publisher-dialog')).toBeVisible();
    await page.getByTestId('trusted-publisher-publisher-id').fill(id);
    await page.getByTestId('trusted-publisher-name').fill('E2E Publisher');
    await page.getByTestId('trusted-publisher-homepage').fill('https://example.com/e2e');
    await page.getByTestId('trusted-publisher-key-id').fill(`${id}-key`);
    await page.getByTestId('trusted-publisher-public-key').fill(PUBLIC_KEY);
    await page.getByTestId('trusted-publisher-notes').fill('created by e2e');
    await page.getByTestId('trusted-publisher-submit').click();

    await expect(page.getByTestId('trusted-publisher-list')).toContainText(id);
    await expect(page.getByTestId('trusted-publisher-list')).toContainText('trusted');
    await expect(page.getByTestId('trusted-publisher-list')).toContainText('sha256:');

    const row = page.getByRole('row', { name: new RegExp(id) }).first();
    await row.getByTestId('trusted-publisher-edit').click();
    await page.getByTestId('trusted-publisher-notes').fill('updated by e2e');
    await page.getByTestId('trusted-publisher-submit').click();
    await expect(page.getByTestId('trusted-publisher-list')).toContainText(id);

    await row.getByTestId('trusted-publisher-block').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByRole('row', { name: new RegExp(id) }).first()).toContainText('blocked');

    await page.getByRole('row', { name: new RegExp(id) }).first().getByTestId('trusted-publisher-restore').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByRole('row', { name: new RegExp(id) }).first()).toContainText('trusted');

    await page.getByRole('row', { name: new RegExp(id) }).first().getByTestId('trusted-publisher-detail').click();
    await expect(page.getByTestId('trusted-publisher-detail-drawer')).toContainText(PUBLIC_KEY);
  });
});
