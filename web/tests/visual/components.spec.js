// @ts-check
const { test, expect } = require('@playwright/test');

/**
 * Component-level visual regression tests.
 * Screenshots individual panels rather than the full page for more stable tests.
 */

test.describe('Projects Panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('empty state', async ({ page }) => {
    // Mock empty projects
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [] });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const panel = page.locator('#projects-panel');
    await expect(panel).toHaveScreenshot('projects-empty.png');
  });

  test('with items', async ({ page }) => {
    // Mock projects with data
    await page.route('**/api/projects', route => {
      route.fulfill({
        json: [
          { name: 'bearing', count: 5 },
          { name: 'sailkit', count: 3 },
          { name: 'fightingwithai.com', count: 2 },
        ],
      });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const panel = page.locator('#projects-panel');
    await expect(panel).toHaveScreenshot('projects-with-items.png');
  });

  test('with selection', async ({ page }) => {
    // Mock projects with data
    await page.route('**/api/projects', route => {
      route.fulfill({
        json: [
          { name: 'bearing', count: 5 },
          { name: 'sailkit', count: 3 },
          { name: 'fightingwithai.com', count: 2 },
        ],
      });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Navigate to second item
    await page.keyboard.press('j');

    const panel = page.locator('#projects-panel');
    await expect(panel).toHaveScreenshot('projects-selected.png');
  });
});

test.describe('Worktrees Table', () => {
  const mockWorktrees = [
    { repo: 'bearing', folder: 'bearing', branch: 'main', base: true, dirty: false, unpushed: 0, prState: null },
    { repo: 'bearing', folder: 'bearing-feature-1', branch: 'feature-1', base: false, dirty: true, unpushed: 2, prState: 'OPEN' },
    { repo: 'bearing', folder: 'bearing-bugfix', branch: 'bugfix', base: false, dirty: false, unpushed: 0, prState: 'MERGED' },
  ];

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('empty state', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'empty-project', count: 0 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: [] });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const panel = page.locator('#worktrees-panel');
    await expect(panel).toHaveScreenshot('worktrees-empty.png');
  });

  test('with data', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 3 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const panel = page.locator('#worktrees-panel');
    await expect(panel).toHaveScreenshot('worktrees-with-data.png');
  });

  test('sorted by folder ascending', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 3 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Click folder column to sort
    await page.click('[data-sort="folder"]');

    const panel = page.locator('#worktrees-panel');
    await expect(panel).toHaveScreenshot('worktrees-sorted-folder-asc.png');
  });

  test('sorted by folder descending', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 3 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Click folder column twice to sort descending
    await page.click('[data-sort="folder"]');
    await page.click('[data-sort="folder"]');

    const panel = page.locator('#worktrees-panel');
    await expect(panel).toHaveScreenshot('worktrees-sorted-folder-desc.png');
  });

  test('sorted by status', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 3 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Click status column to sort
    await page.click('[data-sort="status"]');

    const panel = page.locator('#worktrees-panel');
    await expect(panel).toHaveScreenshot('worktrees-sorted-status.png');
  });

  test('sorted by PR state', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 3 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Click PR column to sort
    await page.click('[data-sort="pr"]');

    const panel = page.locator('#worktrees-panel');
    await expect(panel).toHaveScreenshot('worktrees-sorted-pr.png');
  });
});

test.describe('Details Panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('empty state', async ({ page }) => {
    await page.route('**/api/projects', route => {
      route.fulfill({ json: [] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: [] });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const panel = page.locator('#details-panel');
    await expect(panel).toHaveScreenshot('details-empty.png');
  });

  test('with worktree selected', async ({ page }) => {
    const mockWorktrees = [
      {
        repo: 'bearing',
        folder: 'bearing-feature',
        branch: 'feature-branch',
        base: false,
        dirty: true,
        unpushed: 3,
        prState: 'OPEN',
        purpose: 'Adding new dashboard feature',
        status: 'In progress',
      },
    ];

    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 1 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Move to worktree table and select first item
    await page.keyboard.press('l');
    await page.keyboard.press('Enter');

    const panel = page.locator('#details-panel');
    await expect(panel).toHaveScreenshot('details-with-data.png');
  });

  test('with base worktree selected', async ({ page }) => {
    const mockWorktrees = [
      {
        repo: 'bearing',
        folder: 'bearing',
        branch: 'main',
        base: true,
        dirty: false,
        unpushed: 0,
        prState: null,
      },
    ];

    await page.route('**/api/projects', route => {
      route.fulfill({ json: [{ name: 'bearing', count: 1 }] });
    });
    await page.route('**/api/worktrees', route => {
      route.fulfill({ json: mockWorktrees });
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const panel = page.locator('#details-panel');
    await expect(panel).toHaveScreenshot('details-base-worktree.png');
  });
});

test.describe('Tab Bar', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('worktrees tab active (default)', async ({ page }) => {
    const tabBar = page.locator('.tab-bar');
    await expect(tabBar).toHaveScreenshot('tabs-worktrees-active.png');
  });

  test('issues tab active', async ({ page }) => {
    await page.keyboard.press('2');
    await page.waitForTimeout(100);

    const tabBar = page.locator('.tab-bar');
    await expect(tabBar).toHaveScreenshot('tabs-issues-active.png');
  });

  test('prs tab active', async ({ page }) => {
    await page.keyboard.press('3');
    await page.waitForTimeout(100);

    const tabBar = page.locator('.tab-bar');
    await expect(tabBar).toHaveScreenshot('tabs-prs-active.png');
  });
});

test.describe('Help Modal', () => {
  test('modal appearance', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Open help modal
    await page.keyboard.press('?');
    await page.waitForTimeout(100);

    const modal = page.locator('#help-modal');
    await expect(modal).toHaveScreenshot('help-modal.png');
  });
});

test.describe('Plans Modal', () => {
  test('empty state', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    await page.route('**/api/plans', route => {
      route.fulfill({ json: [] });
    });

    // Open plans modal
    await page.keyboard.press('p');
    await page.waitForTimeout(200);

    const modal = page.locator('#plans-modal');
    await expect(modal).toHaveScreenshot('plans-modal-empty.png');
  });

  test('with plans', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    await page.route('**/api/plans', route => {
      route.fulfill({
        json: [
          { title: 'TUI Test Foundation', project: 'bearing', status: 'active', issue: 17 },
          { title: 'Multi-view TUI', project: 'bearing', status: 'planned', issue: 13 },
          { title: 'Daemon Plans Indexing', project: 'bearing', status: 'completed', issue: 14 },
        ],
      });
    });

    // Open plans modal
    await page.keyboard.press('p');
    await page.waitForTimeout(200);

    const modal = page.locator('#plans-modal');
    await expect(modal).toHaveScreenshot('plans-modal-with-data.png');
  });
});

test.describe('Status Indicator', () => {
  test('connected state', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const indicator = page.locator('#status-indicator');
    await expect(indicator).toHaveScreenshot('status-connected.png');
  });
});

test.describe('Footer Bar', () => {
  test('default appearance', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const footer = page.locator('#footer-bar');
    await expect(footer).toHaveScreenshot('footer-bar.png');
  });
});
