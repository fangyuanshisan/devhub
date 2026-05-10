import { expect, test } from '@playwright/test';

test.describe('frontend public navigation smoke', () => {
  test('homepage opens and hides admin entry for guests', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('frontend-site-header')).toBeVisible();
    await expect(page.getByTestId('frontend-site-nav')).toBeVisible();
    await expect(page.locator('[data-home-latest]')).toBeVisible();
    await expect(page.locator('a[href^="/admin-next"]')).toHaveCount(0);
  });

  test('community homepages open with SEO canonical links', async ({ page }) => {
    for (const slug of ['php', 'go']) {
      await page.goto(`/c/${slug}/`);
      await expect(page.locator('h1').first()).toBeVisible();
      await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', new RegExp(`/c/${slug}/?`));
    }
  });

  test('search page submits keyword and renders result region', async ({ page }) => {
    await page.goto('/search/');
    await expect(page.getByTestId('frontend-search-form')).toBeVisible();
    await page.getByTestId('frontend-search-input').fill('Go');
    await page.getByTestId('frontend-search-submit').click();
    await expect(page).toHaveURL(/keyword=Go/);
    await expect(page.getByTestId('frontend-search-results')).toBeVisible();
    await expect(page.locator('[data-result-title]')).toBeVisible();
  });

  test('topic detail opens with dynamic SEO content', async ({ page }) => {
    await page.goto('/topics/1/');
    await expect(page.locator('h1').first()).toBeVisible();
    await expect(page.locator('article').first()).toBeVisible();
    await expect(page.locator('script[type="application/ld+json"]').first()).toBeAttached();
  });

  test('guest publish page requires login before submit', async ({ page }) => {
    await page.goto('/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    await page.getByTestId('frontend-topic-publish-submit').click();
    await expect(page.getByTestId('frontend-publish-message')).toContainText('请先登录');
  });

  test('tag pages open with canonical metadata', async ({ page }) => {
    await page.goto('/tags/go/');
    await expect(page.locator('h1').first()).toBeVisible();
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', /\/tags\/go\/?/);

    await page.goto('/c/php/tags/laravel/');
    await expect(page.locator('h1').first()).toBeVisible();
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', /\/c\/php\/tags\/laravel\/?/);
  });
});
