import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    // Clear cookies before each test to ensure a clean auth state
    await page.context().clearCookies();
  });

  test('redirects unauthenticated users to login page', async ({ page }) => {
    // Attempt to access a protected route without authentication
    await page.goto('/dashboard');

    // Should be redirected to login
    await expect(page).toHaveURL(/\/auth\/login/);

    // The redirect parameter should preserve the intended destination
    const url = new URL(page.url());
    expect(url.searchParams.get('redirect')).toBe('/dashboard');
  });

  test('login page renders correctly', async ({ page }) => {
    await page.goto('/auth/login');

    // Login form should be visible
    await expect(page.locator('input[type="email"], input[name="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();

    // Submit button should be present
    await expect(
      page.locator('button[type="submit"], button:has-text("Sign in"), button:has-text("Log in")')
    ).toBeVisible();
  });

  test('shows error message with invalid credentials', async ({ page }) => {
    await page.goto('/auth/login');

    // Fill in invalid credentials
    await page.fill('input[type="email"], input[name="email"]', 'invalid@example.com');
    await page.fill('input[type="password"]', 'wrong-password');

    // Submit the form
    await page.click('button[type="submit"], button:has-text("Sign in"), button:has-text("Log in")');

    // Should show an error message (wait for the API response)
    await expect(
      page.locator('[role="alert"], .text-red-500, .text-red-600, .error-message')
    ).toBeVisible({ timeout: 10_000 });
  });

  test('successful login redirects to dashboard', async ({ page }) => {
    await page.goto('/auth/login');

    // Fill in valid test credentials
    await page.fill('input[type="email"], input[name="email"]', 'admin@complianceforge.io');
    await page.fill('input[type="password"]', 'admin123');

    // Submit the form
    await page.click('button[type="submit"], button:has-text("Sign in"), button:has-text("Log in")');

    // Should redirect to dashboard after successful login
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 });
  });

  test('login page is accessible without authentication', async ({ page }) => {
    // The login page should load without being redirected
    const response = await page.goto('/auth/login');
    expect(response?.status()).toBeLessThan(400);
    await expect(page).toHaveURL(/\/auth\/login/);
  });
});
