import { expect, test } from '@playwright/test';
import { loginAsUser } from './helpers/auth.js';
import { uniqueTitle } from './helpers/api.js';

test.describe('frontend publish flow', () => {
  test('guest cannot publish from global or community publish pages', async ({ page }) => {
    for (const url of ['/topics/new/', '/c/php/topics/new/']) {
      await page.goto(url);
      await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
      await page.getByTestId('frontend-topic-publish-submit').click();
      await expect(page.getByTestId('frontend-publish-message')).toContainText('请先登录');
    }
  });

  test('logged-in publish page validates required title and body', async ({ page }) => {
    await loginAsUser(page);
    await page.goto('/c/php/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    await page.getByTestId('frontend-topic-publish-submit').click();
    await expect(page.getByTestId('frontend-publish-message')).toContainText(/标题长度|正文至少|请选择/);
  });

  test('logged-in user can publish an article and open the detail page', async ({ page }) => {
    await loginAsUser(page);
    const title = uniqueTitle('E2E Publish Topic');
    await page.goto('/c/php/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    await page.getByTestId('frontend-publish-community').selectOption('php');
    // In memory seed data, php community category ids are 101..107 (article is 101, question is 102, ...).
    // This case publishes an article, so we must pick the article category.
    await page.getByTestId('frontend-publish-category').selectOption('101');
    await page.getByTestId('frontend-publish-content-type').selectOption('article');
    await page.getByTestId('frontend-publish-title').fill(title);
    await page.getByTestId('frontend-publish-body').fill('这是一段由 E2E 发布流程创建的正文内容，用于验证发布成功后详情页可访问。');
    await page.getByTestId('frontend-topic-publish-submit').click();
    await expect(page).toHaveURL(/\/topics\/\d+\/?/);
    await expect(page.locator('h1').first()).toContainText(title);
    await expect(page.locator('article').first()).toBeVisible();
  });
});
