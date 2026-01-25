// Unit tests for pure JavaScript functions
const assert = require('assert');
const { escapeHtml, sortWorktrees, createState, saveState, loadState } = require('../lib.js');

// Simple test runner
let passed = 0;
let failed = 0;

function test(name, fn) {
  try {
    fn();
    passed++;
    console.log(`  ✓ ${name}`);
  } catch (err) {
    failed++;
    console.log(`  ✗ ${name}`);
    console.log(`    ${err.message}`);
  }
}

function describe(name, fn) {
  console.log(`\n${name}`);
  fn();
}

// Mock localStorage
function createMockStorage() {
  const store = {};
  return {
    getItem: (key) => store[key] || null,
    setItem: (key, value) => { store[key] = value; },
    clear: () => { for (const k in store) delete store[k]; },
  };
}

// Tests
describe('escapeHtml', () => {
  test('escapes ampersand', () => {
    assert.strictEqual(escapeHtml('foo & bar'), 'foo &amp; bar');
  });

  test('escapes less than', () => {
    assert.strictEqual(escapeHtml('<script>'), '&lt;script&gt;');
  });

  test('escapes greater than', () => {
    assert.strictEqual(escapeHtml('a > b'), 'a &gt; b');
  });

  test('escapes double quotes', () => {
    assert.strictEqual(escapeHtml('"quoted"'), '&quot;quoted&quot;');
  });

  test('escapes single quotes', () => {
    assert.strictEqual(escapeHtml("it's"), "it&#39;s");
  });

  test('handles null/undefined', () => {
    assert.strictEqual(escapeHtml(null), '');
    assert.strictEqual(escapeHtml(undefined), '');
    assert.strictEqual(escapeHtml(''), '');
  });

  test('prevents XSS injection', () => {
    const malicious = '<script>alert("xss")</script>';
    const escaped = escapeHtml(malicious);
    assert.ok(!escaped.includes('<script>'));
    assert.ok(escaped.includes('&lt;script&gt;'));
  });

  test('handles complex mixed content', () => {
    const input = '<div class="foo">Tom & Jerry\'s "Show"</div>';
    const expected = '&lt;div class=&quot;foo&quot;&gt;Tom &amp; Jerry&#39;s &quot;Show&quot;&lt;/div&gt;';
    assert.strictEqual(escapeHtml(input), expected);
  });
});

describe('sortWorktrees - default sort', () => {
  const worktrees = [
    { folder: 'feature-z', branch: 'z', dirty: false, unpushed: 0, prState: null },
    { folder: 'feature-a', branch: 'a', dirty: true, unpushed: 0, prState: null },
    { folder: 'feature-m', branch: 'm', dirty: false, unpushed: 0, prState: 'OPEN' },
    { folder: 'feature-b', branch: 'b', dirty: false, unpushed: 0, prState: 'CLOSED' },
  ];

  test('OPEN PRs come first', () => {
    const sorted = sortWorktrees(worktrees, 'default', 'asc');
    assert.strictEqual(sorted[0].prState, 'OPEN');
  });

  test('dirty worktrees come before clean (no PR)', () => {
    const sorted = sortWorktrees(worktrees, 'default', 'asc');
    // After OPEN and CLOSED PRs, dirty should come before clean
    const noPrWorktrees = sorted.filter(w => !w.prState);
    const dirtyIdx = noPrWorktrees.findIndex(w => w.dirty);
    const cleanIdx = noPrWorktrees.findIndex(w => !w.dirty);
    assert.ok(dirtyIdx < cleanIdx, 'Dirty worktrees should sort before clean');
  });

  test('respects PR state ordering: OPEN < DRAFT < MERGED < CLOSED', () => {
    const prWorktrees = [
      { folder: 'd', branch: 'd', dirty: false, unpushed: 0, prState: 'CLOSED' },
      { folder: 'a', branch: 'a', dirty: false, unpushed: 0, prState: 'OPEN' },
      { folder: 'c', branch: 'c', dirty: false, unpushed: 0, prState: 'MERGED' },
      { folder: 'b', branch: 'b', dirty: false, unpushed: 0, prState: 'DRAFT' },
    ];
    const sorted = sortWorktrees(prWorktrees, 'default', 'asc');
    assert.deepStrictEqual(
      sorted.map(w => w.prState),
      ['OPEN', 'DRAFT', 'MERGED', 'CLOSED']
    );
  });
});

describe('sortWorktrees - folder sort', () => {
  const worktrees = [
    { folder: 'zebra', branch: 'x', dirty: false, unpushed: 0 },
    { folder: 'alpha', branch: 'y', dirty: false, unpushed: 0 },
    { folder: 'middle', branch: 'z', dirty: false, unpushed: 0 },
  ];

  test('sorts alphabetically ascending', () => {
    const sorted = sortWorktrees(worktrees, 'folder', 'asc');
    assert.deepStrictEqual(sorted.map(w => w.folder), ['alpha', 'middle', 'zebra']);
  });

  test('sorts alphabetically descending', () => {
    const sorted = sortWorktrees(worktrees, 'folder', 'desc');
    assert.deepStrictEqual(sorted.map(w => w.folder), ['zebra', 'middle', 'alpha']);
  });
});

describe('sortWorktrees - branch sort', () => {
  const worktrees = [
    { folder: 'a', branch: 'feature/z', dirty: false, unpushed: 0 },
    { folder: 'b', branch: 'feature/a', dirty: false, unpushed: 0 },
    { folder: 'c', branch: 'main', dirty: false, unpushed: 0 },
  ];

  test('sorts by branch ascending', () => {
    const sorted = sortWorktrees(worktrees, 'branch', 'asc');
    assert.deepStrictEqual(sorted.map(w => w.branch), ['feature/a', 'feature/z', 'main']);
  });
});

describe('sortWorktrees - status sort', () => {
  const worktrees = [
    { folder: 'clean', branch: 'a', dirty: false, unpushed: 0 },
    { folder: 'unpushed', branch: 'b', dirty: false, unpushed: 2 },
    { folder: 'dirty', branch: 'c', dirty: true, unpushed: 0 },
  ];

  test('dirty first, then unpushed, then clean', () => {
    const sorted = sortWorktrees(worktrees, 'status', 'asc');
    assert.deepStrictEqual(sorted.map(w => w.folder), ['dirty', 'unpushed', 'clean']);
  });

  test('reverse order when descending', () => {
    const sorted = sortWorktrees(worktrees, 'status', 'desc');
    assert.deepStrictEqual(sorted.map(w => w.folder), ['clean', 'unpushed', 'dirty']);
  });
});

describe('sortWorktrees - pr sort', () => {
  const worktrees = [
    { folder: 'a', branch: 'a', dirty: false, unpushed: 0, prState: null },
    { folder: 'b', branch: 'b', dirty: false, unpushed: 0, prState: 'MERGED' },
    { folder: 'c', branch: 'c', dirty: false, unpushed: 0, prState: 'OPEN' },
  ];

  test('OPEN first, then other states, then null', () => {
    const sorted = sortWorktrees(worktrees, 'pr', 'asc');
    assert.strictEqual(sorted[0].prState, 'OPEN');
    assert.strictEqual(sorted[1].prState, 'MERGED');
    assert.strictEqual(sorted[2].prState, null);
  });
});

describe('sortWorktrees - immutability', () => {
  test('does not mutate original array', () => {
    const original = [
      { folder: 'z', branch: 'z', dirty: false, unpushed: 0 },
      { folder: 'a', branch: 'a', dirty: false, unpushed: 0 },
    ];
    const originalOrder = original.map(w => w.folder);
    sortWorktrees(original, 'folder', 'asc');
    assert.deepStrictEqual(original.map(w => w.folder), originalOrder);
  });
});

describe('state persistence - saveState', () => {
  test('saves state to storage', () => {
    const storage = createMockStorage();
    const state = {
      selectedProject: 'my-project',
      worktreeIndex: 5,
      sortColumn: 'folder',
      sortDirection: 'desc',
      currentView: 'prs',
    };
    saveState(state, storage);
    const saved = JSON.parse(storage.getItem('bearing-state'));
    assert.strictEqual(saved.selectedProject, 'my-project');
    assert.strictEqual(saved.worktreeIndex, 5);
    assert.strictEqual(saved.sortColumn, 'folder');
    assert.strictEqual(saved.sortDirection, 'desc');
    assert.strictEqual(saved.currentView, 'prs');
  });
});

describe('state persistence - loadState', () => {
  test('loads saved state', () => {
    const storage = createMockStorage();
    storage.setItem('bearing-state', JSON.stringify({
      selectedProject: 'loaded-project',
      worktreeIndex: 3,
      sortColumn: 'branch',
      sortDirection: 'asc',
      currentView: 'issues',
    }));
    const state = createState();
    loadState(state, storage);
    assert.strictEqual(state.selectedProject, 'loaded-project');
    assert.strictEqual(state.worktreeIndex, 3);
    assert.strictEqual(state.sortColumn, 'branch');
    assert.strictEqual(state.currentView, 'issues');
  });

  test('handles missing storage gracefully', () => {
    const storage = createMockStorage();
    const state = createState();
    loadState(state, storage);
    // Should use defaults
    assert.strictEqual(state.selectedProject, null);
    assert.strictEqual(state.sortColumn, 'default');
  });

  test('handles corrupted JSON gracefully', () => {
    const storage = createMockStorage();
    storage.setItem('bearing-state', 'not valid json {{{');
    const state = createState();
    // Should not throw
    loadState(state, storage);
    // Should retain defaults
    assert.strictEqual(state.selectedProject, null);
  });

  test('handles partial data gracefully', () => {
    const storage = createMockStorage();
    storage.setItem('bearing-state', JSON.stringify({
      selectedProject: 'partial-project',
      // Missing other fields
    }));
    const state = createState();
    loadState(state, storage);
    assert.strictEqual(state.selectedProject, 'partial-project');
    assert.strictEqual(state.worktreeIndex, 0); // Default
    assert.strictEqual(state.sortColumn, 'default'); // Default
  });
});

describe('state persistence - round trip', () => {
  test('save then load preserves state', () => {
    const storage = createMockStorage();
    const originalState = {
      selectedProject: 'roundtrip-test',
      worktreeIndex: 7,
      sortColumn: 'status',
      sortDirection: 'desc',
      currentView: 'worktrees',
    };
    saveState(originalState, storage);
    const loadedState = createState();
    loadState(loadedState, storage);
    assert.strictEqual(loadedState.selectedProject, originalState.selectedProject);
    assert.strictEqual(loadedState.worktreeIndex, originalState.worktreeIndex);
    assert.strictEqual(loadedState.sortColumn, originalState.sortColumn);
    assert.strictEqual(loadedState.sortDirection, originalState.sortDirection);
    assert.strictEqual(loadedState.currentView, originalState.currentView);
  });
});

// Summary
console.log(`\n${'='.repeat(40)}`);
console.log(`Tests: ${passed} passed, ${failed} failed`);
console.log(`${'='.repeat(40)}\n`);

process.exit(failed > 0 ? 1 : 0);
