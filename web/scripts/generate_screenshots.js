#!/usr/bin/env node

/**
 * Screenshot generation script for the Bearing web dashboard.
 * Captures screenshots for documentation, matching TUI screenshot scenarios.
 */

const { chromium } = require('playwright');
const path = require('path');
const fs = require('fs');

const OUTPUT_DIR = path.join(__dirname, '../../docs/public/images');
const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

// Screenshot scenarios matching TUI
const scenarios = [
  {
    name: 'web-dashboard-worktrees',
    description: 'Worktrees view showing all worktrees',
    setup: async (page) => {
      await page.keyboard.press('w');
    },
  },
  {
    name: 'web-dashboard-projects',
    description: 'Projects view showing project list',
    setup: async (page) => {
      await page.keyboard.press('p');
    },
  },
  {
    name: 'web-dashboard-plans',
    description: 'Plans view showing all plans',
    setup: async (page) => {
      await page.keyboard.press('l');
    },
  },
  {
    name: 'web-dashboard-prs',
    description: 'PRs view showing pull requests',
    setup: async (page) => {
      await page.keyboard.press('r');
    },
  },
  {
    name: 'web-dashboard-help',
    description: 'Help modal showing keyboard shortcuts',
    setup: async (page) => {
      await page.keyboard.press('?');
    },
  },
];

async function generateScreenshots() {
  console.log('Starting screenshot generation...');
  console.log(`Output directory: ${OUTPUT_DIR}`);
  console.log(`Base URL: ${BASE_URL}`);

  // Ensure output directory exists
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }

  const browser = await chromium.launch();
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    deviceScaleFactor: 2, // Retina
  });
  const page = await context.newPage();

  try {
    // Navigate to dashboard
    console.log('Loading dashboard...');
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    for (const scenario of scenarios) {
      console.log(`Capturing: ${scenario.name}`);

      // Reset to default state
      await page.goto(BASE_URL);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // Run setup function
      if (scenario.setup) {
        await scenario.setup(page);
        await page.waitForTimeout(300);
      }

      // Capture screenshot
      const filename = `${scenario.name}.png`;
      const filepath = path.join(OUTPUT_DIR, filename);

      await page.screenshot({
        path: filepath,
        fullPage: false,
      });

      console.log(`  Saved: ${filename}`);
    }

    console.log('\nAll screenshots generated successfully!');
    console.log(`Files saved to: ${OUTPUT_DIR}`);

  } catch (error) {
    console.error('Error generating screenshots:', error);
    process.exit(1);
  } finally {
    await browser.close();
  }
}

// Check if server is available
async function checkServer() {
  const { chromium } = require('playwright');
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    const response = await page.goto(BASE_URL, { timeout: 5000 });
    await browser.close();
    return response && response.ok();
  } catch {
    await browser.close();
    return false;
  }
}

async function main() {
  console.log('Bearing Web Dashboard Screenshot Generator\n');

  // Check if server is running
  console.log('Checking if server is running...');
  const serverAvailable = await checkServer();

  if (!serverAvailable) {
    console.error(`Error: Server not available at ${BASE_URL}`);
    console.error('Please start the daemon with: go run ./cmd/bearing daemon start --port 8080');
    process.exit(1);
  }

  console.log('Server is available.\n');
  await generateScreenshots();
}

main();
