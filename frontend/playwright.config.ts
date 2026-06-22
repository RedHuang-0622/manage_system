import { defineConfig } from '@playwright/test';

const CI = !!process.env.CI;

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  expect: { timeout: 10000 },
  fullyParallel: true,
  retries: CI ? 2 : 0,
  workers: CI ? 1 : undefined,
  reporter: [
    ['list'],
    ['junit', { outputFile: 'test-results/junit.xml' }],
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
  ],
  globalSetup: './e2e/global.setup.ts',
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:5173',
    headless: true,
    screenshot: 'only-on-failure',
    trace: CI ? 'on-first-retry' : 'off',
  },
  // Auto-start dev servers in CI (webServer is not used locally — dev.ps1 handles that)
  webServer: CI
    ? [
        {
          // Pre-built binary (see CI build step) — much faster than go run
          command: 'cd ../backend && ./server',
          port: 8080,
          reuseExistingServer: false,
          timeout: 60000,
        },
        {
          command: 'npx vite --port 5173 --strictPort',
          port: 5173,
          reuseExistingServer: false,
          timeout: 30000,
        },
      ]
    : undefined,
  projects: [
    {
      name: 'admin',
      use: { storageState: 'e2e/.auth/admin.json' },
      testMatch: /admin\.spec\.ts/,
    },
    {
      name: 'member',
      use: { storageState: 'e2e/.auth/member.json' },
      testMatch: /member\.spec\.ts/,
    },
    {
      name: 'equip',
      use: { storageState: 'e2e/.auth/equip.json' },
      testMatch: /equip\.spec\.ts/,
    },
    {
      name: 'viewer',
      use: { storageState: 'e2e/.auth/viewer.json' },
      testMatch: /viewer\.spec\.ts/,
    },
    {
      name: 'noauth',
      testMatch: /noauth\.spec\.ts/,
    },
  ],
});
