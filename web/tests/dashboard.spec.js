// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Dashboard', () => {
  test('page loads and renders header', async ({ page }) => {
    await page.goto('/');

    // Check page title
    await expect(page).toHaveTitle(/Bearing/);

    // Check header is present
    const header = page.locator('header, .header, h1').first();
    await expect(header).toBeVisible();
  });

  test('displays worktrees view by default', async ({ page }) => {
    await page.goto('/');

    // Wait for data to load
    await page.waitForLoadState('networkidle');

    // Check that worktrees section exists
    const worktreesSection = page.locator('[data-view="worktrees"], .worktrees, .view-worktrees').first();
    await expect(worktreesSection).toBeVisible({ timeout: 5000 });
  });

  test('renders project list', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Check for project items or empty state
    const projects = page.locator('[data-project], .project-item, .project');
    const count = await projects.count();

    // Either has projects or shows empty state
    if (count === 0) {
      const emptyState = page.locator('.empty-state, .no-data');
      await expect(emptyState).toBeVisible();
    } else {
      await expect(projects.first()).toBeVisible();
    }
  });
});

test.describe('Keyboard Navigation', () => {
  test('w key switches to worktrees view', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Press 'w' key
    await page.keyboard.press('w');

    // Check worktrees view is active
    const worktreesView = page.locator('[data-view="worktrees"].active, .view-worktrees.active, .worktrees-view');
    await expect(worktreesView).toBeVisible({ timeout: 2000 });
  });

  test('p key switches to projects view', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Press 'p' key
    await page.keyboard.press('p');

    // Check projects view is active
    const projectsView = page.locator('[data-view="projects"].active, .view-projects.active, .projects-view');
    await expect(projectsView).toBeVisible({ timeout: 2000 });
  });

  test('j/k keys navigate list items', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Navigate down with 'j'
    await page.keyboard.press('j');

    // Check if selection moved (look for selected/focused class)
    const selectedItem = page.locator('.selected, .focused, [aria-selected="true"]').first();
    // May not have items if empty
    const count = await selectedItem.count();
    if (count > 0) {
      await expect(selectedItem).toBeVisible();
    }
  });

  test('? key shows help modal', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Press '?' key
    await page.keyboard.press('?');

    // Check help modal appears
    const helpModal = page.locator('.modal, .help-modal, [role="dialog"]');
    await expect(helpModal).toBeVisible({ timeout: 2000 });
  });

  test('Escape closes modal', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Open help modal
    await page.keyboard.press('?');

    // Wait for modal
    const helpModal = page.locator('.modal, .help-modal, [role="dialog"]');
    await expect(helpModal).toBeVisible({ timeout: 2000 });

    // Close with Escape
    await page.keyboard.press('Escape');

    // Modal should be hidden
    await expect(helpModal).toBeHidden({ timeout: 2000 });
  });
});

test.describe('View Switching', () => {
  test('can switch between all views', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const views = ['w', 'p', 'l', 'r'];

    for (const key of views) {
      await page.keyboard.press(key);
      // Small delay to let view switch
      await page.waitForTimeout(100);
    }

    // Back to worktrees
    await page.keyboard.press('w');
    const worktreesView = page.locator('[data-view="worktrees"], .worktrees-view, .view-worktrees');
    await expect(worktreesView.first()).toBeVisible({ timeout: 2000 });
  });
});

test.describe('API Integration', () => {
  test('fetches worktrees from API', async ({ page }) => {
    // Intercept API call
    const responsePromise = page.waitForResponse('**/api/worktrees**');

    await page.goto('/');

    const response = await responsePromise;
    expect(response.status()).toBe(200);

    const data = await response.json();
    expect(Array.isArray(data)).toBe(true);
  });

  test('fetches projects from API', async ({ page }) => {
    const responsePromise = page.waitForResponse('**/api/projects**');

    await page.goto('/');
    await page.keyboard.press('p'); // Switch to projects view

    const response = await responsePromise;
    expect(response.status()).toBe(200);

    const data = await response.json();
    expect(Array.isArray(data)).toBe(true);
  });

  test('status endpoint returns running', async ({ page }) => {
    await page.goto('/');

    const response = await page.request.get('/api/status');
    expect(response.status()).toBe(200);

    const data = await response.json();
    expect(data.running).toBe(true);
  });
});
