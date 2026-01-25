// @ts-check
const { test, expect } = require('@playwright/test');

const mockProjects = [
  { name: 'bearing', count: 3 },
  { name: 'sailkit', count: 2 },
  { name: 'surfdeeper', count: 1 },
];

const mockWorktrees = [
  { repo: 'bearing', folder: 'bearing', branch: 'main', base: true, dirty: false, unpushed: 0, prState: null },
  { repo: 'bearing', folder: 'bearing-feature-x', branch: 'feature-x', base: false, dirty: true, unpushed: 2, prState: 'OPEN' },
  { repo: 'bearing', folder: 'bearing-bugfix', branch: 'bugfix', base: false, dirty: false, unpushed: 0, prState: 'MERGED' },
  { repo: 'sailkit', folder: 'sailkit', branch: 'main', base: true, dirty: false, unpushed: 0, prState: null },
  { repo: 'sailkit', folder: 'sailkit-docs', branch: 'docs', base: false, dirty: true, unpushed: 1, prState: 'DRAFT' },
];

const mockPlans = [
  { title: 'TUI improvements', project: 'bearing', status: 'in_progress', issue: 42 },
  { title: 'Add compass component', project: 'sailkit', status: 'pending', issue: 15 },
];

test.describe('Keyboard - j/k Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
    await page.waitForSelector('#project-list .list-item');
  });

  test('j key moves selection down in project list', async ({ page }) => {
    // Focus project list first (h key)
    await page.keyboard.press('h');

    // First project should be selected initially
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'bearing');

    // Press j to move down
    await page.keyboard.press('j');

    // Second project should now be selected
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'sailkit');
  });

  test('k key moves selection up in project list', async ({ page }) => {
    await page.keyboard.press('h');

    // Move down twice
    await page.keyboard.press('j');
    await page.keyboard.press('j');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'surfdeeper');

    // Move back up
    await page.keyboard.press('k');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'sailkit');
  });

  test('j/k navigates worktree list when focused', async ({ page }) => {
    // Focus worktree table (l key moves right)
    await page.keyboard.press('l');

    // First worktree should be selected
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '0');

    // Press j to move down
    await page.keyboard.press('j');
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '1');

    // Press k to move back up
    await page.keyboard.press('k');
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '0');
  });

  test('j does not go past last item', async ({ page }) => {
    await page.keyboard.press('h');

    // Move to last project
    await page.keyboard.press('j');
    await page.keyboard.press('j');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'surfdeeper');

    // Try to go further - should stay on last
    await page.keyboard.press('j');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'surfdeeper');
  });

  test('k does not go past first item', async ({ page }) => {
    await page.keyboard.press('h');

    // Already at first, try to go up
    await page.keyboard.press('k');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'bearing');
  });

  test('Arrow keys work same as j/k', async ({ page }) => {
    await page.keyboard.press('h');

    await page.keyboard.press('ArrowDown');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'sailkit');

    await page.keyboard.press('ArrowUp');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'bearing');
  });
});

test.describe('Keyboard - h/l Panel Focus', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
    await page.waitForSelector('#project-list .list-item');
  });

  test('h key focuses project list', async ({ page }) => {
    await page.keyboard.press('h');

    // j/k should now navigate projects, not worktrees
    await page.keyboard.press('j');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'sailkit');
  });

  test('l key focuses worktree table', async ({ page }) => {
    await page.keyboard.press('l');

    // j/k should navigate worktrees
    await page.keyboard.press('j');
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '1');
  });

  test('Enter in project list moves focus to worktrees', async ({ page }) => {
    await page.keyboard.press('h');

    // Select a project with Enter
    await page.keyboard.press('Enter');

    // Focus should move to worktree table - test by navigating
    await page.keyboard.press('j');
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '1');
  });

  test('Arrow keys work for panel focus', async ({ page }) => {
    await page.keyboard.press('ArrowLeft');

    await page.keyboard.press('ArrowDown');
    await expect(page.locator('#project-list .list-item.selected')).toHaveAttribute('data-project', 'sailkit');

    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowDown');
    await expect(page.locator('#worktree-rows .table-row.selected')).toHaveAttribute('data-index', '1');
  });
});

test.describe('Keyboard - Number Keys for Views', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
  });

  test('1 key switches to worktrees view', async ({ page }) => {
    // First switch away
    await page.keyboard.press('2');
    await expect(page.locator('.tab[data-view="issues"]')).toHaveClass(/active/);

    // Press 1 to go back
    await page.keyboard.press('1');
    await expect(page.locator('.tab[data-view="worktrees"]')).toHaveClass(/active/);
    await expect(page.locator('#main-container')).toHaveCSS('display', 'flex');
  });

  test('2 key switches to issues view', async ({ page }) => {
    await page.keyboard.press('2');

    await expect(page.locator('.tab[data-view="issues"]')).toHaveClass(/active/);
    await expect(page.locator('#main-container')).toHaveCSS('display', 'none');
  });

  test('3 key switches to PRs view', async ({ page }) => {
    await page.keyboard.press('3');

    await expect(page.locator('.tab[data-view="prs"]')).toHaveClass(/active/);
    await expect(page.locator('#placeholder-view')).toContainText('Pull Requests');
  });

  test('number keys update localStorage', async ({ page }) => {
    await page.keyboard.press('3');

    const savedState = await page.evaluate(() => {
      return JSON.parse(localStorage.getItem('bearing-state') || '{}');
    });

    expect(savedState.currentView).toBe('prs');
  });
});

test.describe('Keyboard - Tab Cycling', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
  });

  test('Tab key cycles through views', async ({ page }) => {
    // Start on worktrees
    await expect(page.locator('.tab[data-view="worktrees"]')).toHaveClass(/active/);

    // Tab -> issues
    await page.keyboard.press('Tab');
    await expect(page.locator('.tab[data-view="issues"]')).toHaveClass(/active/);

    // Tab -> prs
    await page.keyboard.press('Tab');
    await expect(page.locator('.tab[data-view="prs"]')).toHaveClass(/active/);

    // Tab -> back to worktrees (cycle)
    await page.keyboard.press('Tab');
    await expect(page.locator('.tab[data-view="worktrees"]')).toHaveClass(/active/);
  });

  test('Tab cycling persists state', async ({ page }) => {
    await page.keyboard.press('Tab');

    const savedState = await page.evaluate(() => {
      return JSON.parse(localStorage.getItem('bearing-state') || '{}');
    });

    expect(savedState.currentView).toBe('issues');
  });
});

test.describe('Keyboard - Help Modal', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
  });

  test('? key opens help modal', async ({ page }) => {
    await page.keyboard.press('?');

    await expect(page.locator('#help-modal')).not.toHaveClass(/hidden/);
    await expect(page.locator('#help-modal')).toContainText('Keybindings');
  });

  test('Escape closes help modal', async ({ page }) => {
    await page.keyboard.press('?');
    await expect(page.locator('#help-modal')).not.toHaveClass(/hidden/);

    await page.keyboard.press('Escape');
    await expect(page.locator('#help-modal')).toHaveClass(/hidden/);
  });

  test('? key also closes help modal', async ({ page }) => {
    await page.keyboard.press('?');
    await expect(page.locator('#help-modal')).not.toHaveClass(/hidden/);

    await page.keyboard.press('?');
    await expect(page.locator('#help-modal')).toHaveClass(/hidden/);
  });

  test('other keys blocked when help modal open', async ({ page }) => {
    await page.keyboard.press('?');

    // Try to change view - should not work
    await page.keyboard.press('2');
    await expect(page.locator('.tab[data-view="worktrees"]')).toHaveClass(/active/);
  });
});

test.describe('Keyboard - Plans Modal', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/projects', route => route.fulfill({ json: mockProjects }));
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/plans', route => route.fulfill({ json: mockPlans }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
  });

  test('p key opens plans modal', async ({ page }) => {
    await page.keyboard.press('p');

    await expect(page.locator('#plans-modal')).not.toHaveClass(/hidden/);
  });

  test('j/k navigates plans in modal', async ({ page }) => {
    await page.keyboard.press('p');
    await page.waitForSelector('#plans-list .list-item');

    // First plan selected
    await expect(page.locator('#plans-list .list-item.selected')).toContainText('TUI improvements');

    // Navigate down
    await page.keyboard.press('j');
    await expect(page.locator('#plans-list .list-item.selected')).toContainText('Add compass');

    // Navigate back up
    await page.keyboard.press('k');
    await expect(page.locator('#plans-list .list-item.selected')).toContainText('TUI improvements');
  });

  test('Escape closes plans modal', async ({ page }) => {
    await page.keyboard.press('p');
    await expect(page.locator('#plans-modal')).not.toHaveClass(/hidden/);

    await page.keyboard.press('Escape');
    await expect(page.locator('#plans-modal')).toHaveClass(/hidden/);
  });

  test('p key also closes plans modal', async ({ page }) => {
    await page.keyboard.press('p');
    await expect(page.locator('#plans-modal')).not.toHaveClass(/hidden/);

    await page.keyboard.press('p');
    await expect(page.locator('#plans-modal')).toHaveClass(/hidden/);
  });
});

test.describe('Keyboard - Refresh', () => {
  test('r key triggers data refresh', async ({ page }) => {
    let refreshCount = 0;
    await page.route('**/api/projects', route => {
      refreshCount++;
      route.fulfill({ json: mockProjects });
    });
    await page.route('**/api/worktrees', route => route.fulfill({ json: mockWorktrees }));
    await page.route('**/api/events', route => route.abort());

    await page.goto('/');
    await page.waitForSelector('#project-list .list-item');

    const initialCount = refreshCount;

    // Press r to refresh
    await page.keyboard.press('r');

    // Wait for API call
    await page.waitForTimeout(100);

    expect(refreshCount).toBeGreaterThan(initialCount);
  });
});
