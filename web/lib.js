// Bearing Web Dashboard - Pure functions (testable)
// These are extracted from app.js for unit testing

function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/[&<>"']/g, (c) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  }[c]));
}

function sortWorktrees(worktrees, sortColumn, sortDirection) {
  const sorted = [...worktrees];
  const dir = sortDirection === 'asc' ? 1 : -1;

  if (sortColumn === 'default') {
    // Default: PR state (OPEN first), then dirty, then folder
    sorted.sort((a, b) => {
      const prOrder = { OPEN: 0, DRAFT: 1, MERGED: 2, CLOSED: 3 };
      const aPr = prOrder[a.prState] ?? 4;
      const bPr = prOrder[b.prState] ?? 4;
      if (aPr !== bPr) return (aPr - bPr) * dir;
      if (a.dirty !== b.dirty) return (b.dirty - a.dirty) * dir;
      return a.folder.localeCompare(b.folder) * dir;
    });
  } else if (sortColumn === 'folder') {
    sorted.sort((a, b) => a.folder.localeCompare(b.folder) * dir);
  } else if (sortColumn === 'branch') {
    sorted.sort((a, b) => a.branch.localeCompare(b.branch) * dir);
  } else if (sortColumn === 'status') {
    sorted.sort((a, b) => {
      // dirty first, then unpushed, then clean
      const aScore = a.dirty ? 0 : (a.unpushed > 0 ? 1 : 2);
      const bScore = b.dirty ? 0 : (b.unpushed > 0 ? 1 : 2);
      return (aScore - bScore) * dir;
    });
  } else if (sortColumn === 'pr') {
    const prOrder = { OPEN: 0, DRAFT: 1, MERGED: 2, CLOSED: 3 };
    sorted.sort((a, b) => {
      const aPr = prOrder[a.prState] ?? 4;
      const bPr = prOrder[b.prState] ?? 4;
      return (aPr - bPr) * dir;
    });
  }

  return sorted;
}

function createState() {
  return {
    selectedProject: null,
    worktreeIndex: 0,
    sortColumn: 'default',
    sortDirection: 'asc',
    currentView: 'worktrees',
  };
}

function saveState(state, storage) {
  const persisted = {
    selectedProject: state.selectedProject,
    worktreeIndex: state.worktreeIndex,
    sortColumn: state.sortColumn,
    sortDirection: state.sortDirection,
    currentView: state.currentView,
  };
  storage.setItem('bearing-state', JSON.stringify(persisted));
}

function loadState(state, storage) {
  try {
    const saved = storage.getItem('bearing-state');
    if (saved) {
      const persisted = JSON.parse(saved);
      state.selectedProject = persisted.selectedProject || null;
      state.worktreeIndex = persisted.worktreeIndex || 0;
      state.sortColumn = persisted.sortColumn || 'default';
      state.sortDirection = persisted.sortDirection || 'asc';
      state.currentView = persisted.currentView || 'worktrees';
    }
  } catch (e) {
    // Silently fail on parse errors
  }
  return state;
}

// Export for Node.js testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { escapeHtml, sortWorktrees, createState, saveState, loadState };
}
