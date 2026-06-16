import { test, expect } from '@playwright/test';
const BASE = 'http://localhost:5173';

test('未登录 / → /login', async ({ page }) => {
  await page.goto(`${BASE}/`);
  await page.waitForTimeout(800);
  expect(page.url()).toContain('/login');
});

test('未登录 /users → /login', async ({ page }) => {
  await page.goto(`${BASE}/users`);
  await page.waitForTimeout(800);
  expect(page.url()).toContain('/login');
});
