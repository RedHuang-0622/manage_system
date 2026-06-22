/**
 * Global setup: pre-login each role and save storageState.
 * Subsequent tests reuse these states — only 4 logins total.
 */
import { chromium, type Page } from '@playwright/test';

const BASE = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:5173';

async function loginAndSave(page: Page, user: string, pass: string, file: string) {
  await page.goto(`${BASE}/login`);
  await page.locator('#login_username').fill(user);
  await page.locator('#login_password').fill(pass);
  await page.locator('button[type="submit"]').click();
  await page.waitForURL('**/');
  await page.context().storageState({ path: file });
}

export default async function globalSetup() {
  const browser = await chromium.launch({ headless: true });

  console.log('  Setting up admin auth...');
  const adminCtx = await browser.newContext();
  await loginAndSave(await adminCtx.newPage(), 'admin', 'admin123', 'e2e/.auth/admin.json');
  await adminCtx.close();

  console.log('  Setting up member auth...');
  const memberCtx = await browser.newContext();
  await loginAndSave(await memberCtx.newPage(), 'lina', '123456', 'e2e/.auth/member.json');
  await memberCtx.close();

  console.log('  Setting up equip_manager auth...');
  const equipCtx = await browser.newContext();
  await loginAndSave(await equipCtx.newPage(), 'liulei', '123456', 'e2e/.auth/equip.json');
  await equipCtx.close();

  console.log('  Setting up viewer auth...');
  const viewerCtx = await browser.newContext();
  await loginAndSave(await viewerCtx.newPage(), 'sunyue', '123456', 'e2e/.auth/viewer.json');
  await viewerCtx.close();

  await browser.close();
  console.log('  Global setup done.');
}
