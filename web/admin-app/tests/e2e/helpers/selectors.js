import { expect } from '@playwright/test';

export async function expectAdminPageReady(page, testId) {
  await expect(page.getByTestId(testId)).toBeVisible();
  await expect(page.locator('.el-table').first()).toBeVisible();
}

export async function waitForToast(page, text) {
  await expect(page.locator('.el-message').filter({ hasText: text }).first()).toBeVisible();
}

export function findRowByText(page, text) {
  return page.locator('.el-table__row').filter({ hasText: text }).first();
}
