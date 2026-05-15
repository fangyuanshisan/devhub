import { expect } from '@playwright/test';

export async function openPluginPage(page, navKey, expectedURL, pageTestId) {
  await page.goto('/admin-next/plugins/overview');
  await page.getByTestId(`admin-sub-nav-${navKey}`).click();
  await expect(page).toHaveURL(expectedURL);
  await expect(page.getByTestId(pageTestId)).toBeVisible();
}

export async function expectPluginBoundaryNotice(page) {
  const notice = page.getByTestId('plugin-package-boundary-notice');
  await expect(notice).toBeVisible();
  await expect(notice).toContainText(/不执行第三方代码|不会执行插件代码/);
  await expect(notice).toContainText(/不执行外部 SQL|不会执行 SQL/);
  await expect(notice).toContainText(/不动态加载前端资产|不会加载前端资产/);
}
