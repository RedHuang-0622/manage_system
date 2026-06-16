import { test, expect } from '@playwright/test';

test('有设备大厅+审批, 无用户管理', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(500);
  await page.locator('.ant-menu-submenu-title').filter({ hasText: '借阅管理' }).click();
  await page.waitForTimeout(300);
  const s = await page.locator('.ant-menu-title-content').allTextContents();
  expect(s.some(t => t.includes('用户管理'))).toBeFalsy();
  expect(s.some(t => t.includes('设备大厅'))).toBeTruthy();
  expect(s.some(t => t.includes('待审批'))).toBeTruthy();
});
