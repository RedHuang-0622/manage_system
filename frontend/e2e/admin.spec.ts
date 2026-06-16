import { test, expect } from '@playwright/test';

test('全部菜单可见', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(500);
  await page.locator('.ant-menu-submenu-title').filter({ hasText: '借阅管理' }).click();
  await page.waitForTimeout(300);
  const s = await page.locator('.ant-menu-title-content').allTextContents();
  expect(s.some(t => t.includes('仪表盘'))).toBeTruthy();
  expect(s.some(t => t.includes('设备大厅'))).toBeTruthy();
  expect(s.some(t => t.includes('借阅管理'))).toBeTruthy();
  expect(s.some(t => t.includes('用户管理'))).toBeTruthy();
});

test('用户管理表格有数据', async ({ page }) => {
  await page.goto('/');
  await page.locator('.ant-menu-title-content').filter({ hasText: '用户管理' }).click();
  await page.waitForTimeout(1500);
  expect(await page.locator('.ant-table-row').count()).toBeGreaterThan(0);
});

test('设备大厅表格有数据', async ({ page }) => {
  await page.goto('/');
  await page.locator('.ant-menu-title-content').filter({ hasText: '设备大厅' }).click();
  await page.waitForTimeout(1000);
  expect(await page.locator('.ant-table-row').count()).toBeGreaterThan(0);
});
