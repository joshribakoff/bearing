import type { NavItem } from 'sailkit/packages/compass';

export const navigation: NavItem[] = [
  'index',
  'workspace-layout',
  {
    slug: 'concepts',
    children: [
      'base-folders',
      'worktrees',
      'state-files',
      'workflow-health',
    ],
  },
  {
    slug: 'commands',
    children: [
      'worktree-new',
      'worktree-cleanup',
      'worktree-recover',
      'worktree-status',
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
      'plan-sync',
      'ai-features',
    ],
  },
];

export const titles: Record<string, string> = {
  'index': 'Introduction',
  'workspace-layout': 'Workspace Layout',
  'commands': 'Commands',
  'worktree-new': 'worktree-new',
  'worktree-cleanup': 'worktree-cleanup',
  'worktree-recover': 'worktree-recover',
  'worktree-sync': 'worktree-sync',
  'worktree-list': 'worktree-list',
  'worktree-register': 'worktree-register',
  'worktree-check': 'worktree-check',
  'concepts': 'Concepts',
  'base-folders': 'Base Folders',
  'worktrees': 'Worktrees',
  'state-files': 'State Files',
  'workflow-health': 'Workflow Health',
  'worktree-status': 'worktree-status',
  'integration': 'Integration',
  'claude-code-hooks': 'Claude Code Hooks',
  'slash-commands': 'Slash Commands',
  'plan-sync': 'Plan Sync',
  'ai-features': 'AI Features',
};

export function getTitle(slug: string): string {
  return titles[slug] || slug.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}
