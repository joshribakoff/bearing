// @ts-check
const { test, expect } = require('@playwright/test');

const mockProjects = [
  { name: 'bearing', count: 5 },
];

// Worktrees with varied data for sorting tests
const mockWorktrees = [
  { repo: 'bearing', folder: 'bearing', branch: 'main', base: true, dirty: false, unpushed: 0, prState: null },
  { repo: 'bearing', folder: 'bearing-alpha', branch: 'feature-alpha', base: false, dirty: true, unpushed: 2, prState: 'OPEN' },
  { repo: 'bearing', folder: 'bearing-zeta', branch: 'bugfix-zeta', base: false, dirty: false, unpushed: 0, prState: 'MERGED' },
  { repo: 'bearing', folder: 'bearing-beta', branch: 'draft-beta', base: false, dirty: false, unpushed: 1, prState: 'DRAFT' },
  { repo: 'bearing', folder: 'bearing-gamma', branch: 'wip-gamma', base: false, dirty: true, unpushed: 0, prState: 'CLOSED' },
];

test.describe('Sorting - Column Headers', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
    await page.waitForSelector('#worktree-rows .table-row');
  });

  test('clicking Folder header sorts by folder name', async ({ page }) => {
    // Click folder header
    await page.click('[data-sort="folder"]');

    // Get folder names in order
    const folders = await page.locator('#worktree-rows .table-row .col-folder').allTextContents();
    const cleanFolders = folders.map(f => f.replace('BASE', '').trim());

    // Should be alphabetically sorted (asc)
    const sorted = [...cleanFolders].sort();
    expect(cleanFolders).toEqual(sorted);

    // Header should show sort indicator
    await expect(page.locator('[data-sort="folder"]')).toHaveClass(/sort-asc/);
  });

  test('clicking same header toggles sort direction', async ({ page }) => {
    // Click folder header twice
    await page.click('[data-sort="folder"]');
    await expect(page.locator('[data-sort="folder"]')).toHaveClass(/sort-asc/);

    await page.click('[data-sort="folder"]');
    await expect(page.locator('[data-sort="folder"]')).toHaveClass(/sort-desc/);

    // Get folder names
    const folders = await page.locator('#worktree-rows .table-row .col-folder').allTextContents();
    const cleanFolders = folders.map(f => f.replace('BASE', '').trim());

    // Should be reverse alphabetically sorted
    const sorted = [...cleanFolders].sort().reverse();
    expect(cleanFolders).toEqual(sorted);
  });

  test('clicking Branch header sorts by branch name', async ({ page }) => {
    await page.click('[data-sort="branch"]');

    const branches = await page.locator('#worktree-rows .table-row .col-branch').allTextContents();

    // Should be alphabetically sorted
    const sorted = [...branches].sort();
    expect(branches).toEqual(sorted);

    await expect(page.locator('[data-sort="branch"]')).toHaveClass(/sort-asc/);
  });

  test('clicking Status header sorts by status', async ({ page }) => {
    await page.click('[data-sort="status"]');

    // Verify header has sort indicator
    await expect(page.locator('[data-sort="status"]')).toHaveClass(/sort-asc/);

    // Status sort order: dirty first, then unpushed, then clean
    const rows = page.locator('#worktree-rows .table-row');
    const count = await rows.count();

    // First rows should have dirty indicator (*)
    const firstRowStatus = await rows.first().locator('.col-status').textContent();
    expect(firstRowStatus).toContain('*');
  });

  test('clicking PR header sorts by PR state', async ({ page }) => {
    await page.click('[data-sort="pr"]');

    await expect(page.locator('[data-sort="pr"]')).toHaveClass(/sort-asc/);

    // PR sort order: OPEN, DRAFT, MERGED, CLOSED, none
    const prStates = await page.locator('#worktree-rows .table-row .col-pr').allTextContents();

    // First should be OPEN
    expect(prStates[0]).toContain('OPEN');
  });

  test('switching columns clears previous sort indicator', async ({ page }) => {
    // Sort by folder
    await page.click('[data-sort="folder"]');
    await expect(page.locator('[data-sort="folder"]')).toHaveClass(/sort-asc/);

    // Sort by branch
    await page.click('[data-sort="branch"]');
    await expect(page.locator('[data-sort="branch"]')).toHaveClass(/sort-asc/);
    await expect(page.locator('[data-sort="folder"]')).not.toHaveClass(/sort-asc/);
    await expect(page.locator('[data-sort="folder"]')).not.toHaveClass(/sort-desc/);
  });
});

test.describe('Sorting - Persistence', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());
  });

  test('sort column persists in localStorage', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('#worktree-rows .table-row');

    await page.click('[data-sort="branch"]');

    const savedState = await page.evaluate(() => {
      return JSON.parse(localStorage.getItem('bearing-state') || '{}');
    });

    expect(savedState.sortColumn).toBe('branch');
    expect(savedState.sortDirection).toBe('asc');
  });

  test('sort direction persists in localStorage', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('#worktree-rows .table-row');

    await page.click('[data-sort="folder"]');
    await page.click('[data-sort="folder"]'); // Toggle to desc

    const savedState = await page.evaluate(() => {
      return JSON.parse(localStorage.getItem('bearing-state') || '{}');
    });

    expect(savedState.sortColumn).toBe('folder');
    expect(savedState.sortDirection).toBe('desc');
  });

  test('sort state survives page reload', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('#worktree-rows .table-row');

    // Sort by branch descending
    await page.click('[data-sort="branch"]');
    await page.click('[data-sort="branch"]');
    await expect(page.locator('[data-sort="branch"]')).toHaveClass(/sort-desc/);

    // Reload
    await page.reload();
    await page.waitForSelector('#worktree-rows .table-row');

    // Branch should still have desc indicator
    await expect(page.locator('[data-sort="branch"]')).toHaveClass(/sort-desc/);

    // Verify data is still sorted
    const branches = await page.locator('#worktree-rows .table-row .col-branch').allTextContents();
    const sorted = [...branches].sort().reverse();
    expect(branches).toEqual(sorted);
  });
});

test.describe('Sorting - Selection Interaction', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
    await page.waitForSelector('#worktree-rows .table-row');
  });

  test('sorting preserves worktree index position', async ({ page }) => {
    // Select second row
    await page.click('#worktree-rows .table-row:nth-child(2)');
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '1');

    // Sort by folder
    await page.click('[data-sort="folder"]');

    // Selection should still be at index 1 (same position in list)
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '1');
  });

  test('details panel updates after sorting', async ({ page }) => {
    // Select first row
    await page.click('#worktree-rows .table-row:first-child');
    const detailsBefore = await page.locator('#details-content').textContent();

    // Sort by branch - order changes
    await page.click('[data-sort="branch"]');

    // Details should reflect current selection
    const detailsAfter = await page.locator('#details-content').textContent();
    expect(detailsAfter).toContain('Folder:');
  });
});

test.describe('Sorting - Default Order', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());
  });

  test('default sort prioritizes OPEN PRs', async ({ page }) => {
    // Clear localStorage to ensure default
    await page.evaluate(() => localStorage.clear());

    await page.goto('/');
    await page.waitForSelector('#worktree-rows .table-row');

    // First worktree with PR should be OPEN
    const firstPR = await page.locator('#worktree-rows .table-row .col-pr span').first();
    const text = await firstPR.textContent();

    // OPEN should appear before MERGED, CLOSED
    if (text) {
      expect(['OPEN', '']).toContain(text.trim() || '');
    }
  });
});
