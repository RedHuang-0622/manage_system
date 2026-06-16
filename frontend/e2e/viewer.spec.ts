import { test, expect } from '@playwright/test';
const BASE = 'http://localhost:5173';

test('只读菜单, 不能发起申请', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(500);
  const s = await page.locator('.ant-menu-title-content').allTextContents();
  expect(s.some(t => t.includes('发起申请'))).toBeFalsy();
  expect(s.some(t => t.includes('待审批'))).toBeFalsy();
  expect(s.some(t => t.includes('设备大厅'))).toBeTruthy();
});

test('访问 /equipments/new 被拦截', async ({ page }) => {
  await page.goto(`${BASE}/equipments/new`);
  await page.waitForTimeout(800);
  expect(page.url()).toBe(`${BASE}/`);
});
