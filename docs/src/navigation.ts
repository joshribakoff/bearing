import type { NavItem } from 'sailkit/packages/compass';

export const navigation: NavItem[] = [
  'index',
  'workspace-layout',
  'tui',
  {
    slug: 'concepts',
    children: [
      'base-folders',
      'worktrees',
      'state-files',
    ],
  },
  {
    slug: 'commands',
    children: [
      'worktree-new',
      'worktree-cleanup',
      'worktree-sync',
      'worktree-list',
      'worktree-register',
      'worktree-check',
    ],
  },
  {
    slug: 'integration',
    children: [
      'claude-code-hooks',
      'slash-commands',
    ],
  },
];

export const titles: Record<string, string> = {
  'index': 'Introduction',
  'workspace-layout': 'Workspace Layout',
  'tui': 'Terminal UI',
  'commands': 'Commands',
  'worktree-new': 'worktree-new',
  'worktree-cleanup': 'worktree-cleanup',
  'worktree-sync': 'worktree-sync',
  'worktree-list': 'worktree-list',
  'worktree-register': 'worktree-register',
  'worktree-check': 'worktree-check',
  'concepts': 'Concepts',
  'base-folders': 'Base Folders',
  'worktrees': 'Worktrees',
  'state-files': 'State Files',
  'integration': 'Integration',
  'claude-code-hooks': 'Claude Code Hooks',
  'slash-commands': 'Slash Commands',
};

export function getTitle(slug: string): string {
  return titles[slug] || slug.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}
