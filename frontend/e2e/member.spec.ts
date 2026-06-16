import { test, expect } from '@playwright/test';
const BASE = 'http://localhost:5173';

test('无用户管理 + 无审批菜单', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(500);
  const s = await page.locator('.ant-menu-title-content').allTextContents();
  expect(s.some(t => t.includes('用户管理'))).toBeFalsy();
  expect(s.some(t => t.includes('待审批'))).toBeFalsy();
  expect(s.some(t => t.includes('设备大厅'))).toBeTruthy();
});

test('访问 /users 被踢回首页', async ({ page }) => {
  await page.goto(`${BASE}/users`);
  await page.waitForTimeout(800);
  expect(page.url()).toBe(`${BASE}/`);
});

test('我的借阅可导航', async ({ page }) => {
  await page.goto('/');
  await page.locator('.ant-menu-submenu-title').filter({ hasText: '借阅管理' }).click();
  await page.waitForTimeout(300);
  await page.locator('.ant-menu-title-content').filter({ hasText: '我的借阅' }).click();
  await page.waitForTimeout(1000);
  expect(page.url()).toContain('/borrows/my');
});
