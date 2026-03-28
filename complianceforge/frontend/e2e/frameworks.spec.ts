import { test, expect } from '@playwright/test';

test.describe('Frameworks', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set an authentication cookie so protected routes are accessible.
    await context.addCookies([
      {
        name: 'cf_token',
        value: 'e2e-test-token',
        domain: 'localhost',
        path: '/',
      },
    ]);
  });

  test('frameworks list page loads', async ({ page }) => {
    await page.goto('/frameworks');

    // Should stay on the frameworks page
    await expect(page).toHaveURL(/\/frameworks/);

    // Page should have a heading or title related to frameworks
    await expect(
      page.locator('h1:has-text("Framework"), h2:has-text("Framework"), [data-testid="frameworks-title"]')
    ).toBeVisible({ timeout: 10_000 });
  });

  test('frameworks page displays framework items', async ({ page }) => {
    await page.goto('/frameworks');

    // Wait for framework list to load (table rows, cards, or list items)
    const frameworkItems = page.locator(
      'table tbody tr, [data-testid="framework-item"], [data-testid="framework-card"], .framework-card'
    );

    // Wait for at least one framework to appear (or an empty state message)
    await expect(
      frameworkItems.first().or(page.locator('text=No frameworks, text=Get started, text=empty'))
    ).toBeVisible({ timeout: 10_000 });
  });

  test('clicking a framework navigates to detail page', async ({ page }) => {
    await page.goto('/frameworks');

    // Wait for framework items to load
    const firstFramework = page.locator(
      'table tbody tr a, [data-testid="framework-item"] a, [data-testid="framework-card"], a[href*="/frameworks/"]'
    ).first();

    // Only proceed if frameworks exist on the page
    const count = await firstFramework.count();
    if (count > 0) {
      await firstFramework.click();

      // Should navigate to a framework detail page
      await expect(page).toHaveURL(/\/frameworks\/[a-zA-Z0-9-]+/);
    }
  });

  test('frameworks page has proper breadcrumb or back navigation', async ({ page }) => {
    await page.goto('/frameworks');

    // The sidebar should still show the Frameworks link as active or present
    await expect(page.locator('a[href="/frameworks"]')).toBeVisible();

    // Dashboard link should be accessible for navigation back
    await expect(page.locator('a[href="/dashboard"]')).toBeVisible();
  });
});
