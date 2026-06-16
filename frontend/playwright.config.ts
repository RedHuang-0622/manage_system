import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  expect: { timeout: 10000 },
  fullyParallel: false,
  retries: 0,
  reporter: 'list',
  globalSetup: './e2e/global.setup.ts',
  use: {
    baseURL: 'http://localhost:5173',
    headless: true,
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'admin',
      use: {
        storageState: 'e2e/.auth/admin.json',
        launchOptions: { executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe' },
      },
      testMatch: /admin\.spec\.ts/,
    },
    {
      name: 'member',
      use: {
        storageState: 'e2e/.auth/member.json',
        launchOptions: { executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe' },
      },
      testMatch: /member\.spec\.ts/,
    },
    {
      name: 'equip',
      use: {
        storageState: 'e2e/.auth/equip.json',
        launchOptions: { executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe' },
      },
      testMatch: /equip\.spec\.ts/,
    },
    {
      name: 'viewer',
      use: {
        storageState: 'e2e/.auth/viewer.json',
        launchOptions: { executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe' },
      },
      testMatch: /viewer\.spec\.ts/,
    },
    {
      name: 'noauth',
      use: {
        launchOptions: { executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe' },
      },
      testMatch: /noauth\.spec\.ts/,
      dependencies: [], // no global setup needed
    },
  ],
});
