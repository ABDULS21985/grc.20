import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set an authentication cookie so protected routes are accessible.
    // In a real environment, this token would come from the auth API.
    await context.addCookies([
      {
        name: 'cf_token',
        value: 'e2e-test-token',
        domain: 'localhost',
        path: '/',
      },
    ]);
  });

  test('dashboard page loads successfully', async ({ page }) => {
    await page.goto('/dashboard');

    // Page should not redirect away
    await expect(page).toHaveURL(/\/dashboard/);

    // Should contain dashboard content (heading or title)
    await expect(
      page.locator('h1, h2, [data-testid="dashboard-title"]')
    ).toBeVisible({ timeout: 10_000 });
  });

  test('dashboard displays KPI cards', async ({ page }) => {
    await page.goto('/dashboard');

    // KPI/metric cards should be visible on the dashboard
    // These are typically rendered as card components with statistics
    const cards = page.locator(
      '[data-testid="kpi-card"], .rounded-lg.border, [class*="card"]'
    );

    // Expect at least one card to be present
    await expect(cards.first()).toBeVisible({ timeout: 10_000 });
  });

  test('sidebar navigation is visible on desktop', async ({ page }) => {
    // Set desktop viewport
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto('/dashboard');

    // Sidebar should contain navigation links
    const sidebar = page.locator('aside, nav, [data-testid="sidebar"]');
    await expect(sidebar.first()).toBeVisible();

    // Key navigation items should be present
    await expect(page.locator('a[href="/dashboard"]')).toBeVisible();
    await expect(page.locator('a[href="/frameworks"]')).toBeVisible();
    await expect(page.locator('a[href="/risks"]')).toBeVisible();
  });

  test('navigation from dashboard to frameworks works', async ({ page }) => {
    await page.goto('/dashboard');

    // Click on the Frameworks navigation link
    await page.click('a[href="/frameworks"]');

    // Should navigate to frameworks page
    await expect(page).toHaveURL(/\/frameworks/);
  });

  test('navigation from dashboard to risks works', async ({ page }) => {
    await page.goto('/dashboard');

    // Click on the Risks navigation link
    await page.click('a[href="/risks"]');

    // Should navigate to risk register page
    await expect(page).toHaveURL(/\/risks/);
  });
});
