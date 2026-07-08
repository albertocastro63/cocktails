// Single source of truth for navigation destinations. Both the desktop top nav
// and the mobile bottom bar decide *which* links to show from here, so the sets
// never drift. Pure and side-effect free for testability — the Sign Out action
// is marked with `action: true` and wired by the renderer, not imported here.

const svg = (paths, { fill = false } = {}) =>
  `<svg viewBox="0 0 24 24" ${fill ? 'fill="currentColor"' : 'fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"'} class="w-6 h-6" aria-hidden="true">${paths}</svg>`;

export const icons = {
  'all-recipes': svg('<path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/>'),
  'my-recipes': svg('<path d="M6 4h12v16l-6-4-6 4V4z"/>'),
  'new-recipe': svg('<path d="M12 5v14M5 12h14"/>'),
  'sign-in': svg('<path d="M15 12H3m0 0l4-4m-4 4l4 4M15 4h4a1 1 0 011 1v14a1 1 0 01-1 1h-4"/>'),
  users: svg('<path d="M15 20a6 6 0 00-12 0M9 11a4 4 0 100-8 4 4 0 000 8zM21 20a5 5 0 00-4-4.9M16 3.1a4 4 0 010 7.8"/>'),
  'admin-recipes': svg('<path d="M9 5a2 2 0 00-2 2v12a2 2 0 002 2h6a2 2 0 002-2V7a2 2 0 00-2-2M9 5a2 2 0 012-2h2a2 2 0 012 2M9 5h6M9 12h6M9 16h6"/>'),
  'sign-out': svg('<path d="M9 12h12m0 0l-4-4m4 4l-4 4M9 4H5a1 1 0 00-1 1v14a1 1 0 001 1h4"/>'),
  more: svg('<circle cx="5" cy="12" r="1.6"/><circle cx="12" cy="12" r="1.6"/><circle cx="19" cy="12" r="1.6"/>', { fill: true }),
};

// Catalog keyed by id. `match` is the exact hash path (getPath()) that marks the
// item active; action items have match=null and never appear active.
const CATALOG = {
  'all-recipes': { id: 'all-recipes', label: 'All', href: '#/recipes', priority: 1, match: '/recipes' },
  'my-recipes': { id: 'my-recipes', label: 'Mine', href: '#/my-recipes', priority: 2, match: '/my-recipes' },
  'new-recipe': { id: 'new-recipe', label: 'New', href: '#/recipes/new', priority: 3, match: '/recipes/new' },
  'sign-in': { id: 'sign-in', label: 'Sign in', href: '#/login', priority: 3, match: '/login' },
  users: { id: 'users', label: 'Users', href: '#/admin/users', priority: 4, match: '/admin/users' },
  'admin-recipes': { id: 'admin-recipes', label: 'Manage', href: '#/admin/recipes', priority: 5, match: '/admin/recipes' },
  'sign-out': { id: 'sign-out', label: 'Sign out', action: true, priority: 6, match: null },
};

const SETS = {
  visitor: ['all-recipes', 'sign-in'],
  user: ['all-recipes', 'my-recipes', 'new-recipe', 'sign-out'],
  admin: ['all-recipes', 'my-recipes', 'new-recipe', 'users', 'admin-recipes', 'sign-out'],
};

// Ordered destinations for an auth state ('visitor' | 'user' | 'admin').
export function navDestinations(state) {
  return (SETS[state] || []).map((id) => ({ ...CATALOG[id], icon: icons[id] }));
}

// Split into up to `maxSlots` bottom-bar slots. When the list fits, all are
// direct. When it overflows, the highest-priority (maxSlots-1) are direct and
// the rest go to `overflow` (rendered behind a synthesized "More" slot).
export function partitionNav(list, maxSlots = 5) {
  if (list.length <= maxSlots) return { direct: list, overflow: [] };
  const sorted = [...list].sort((a, b) => a.priority - b.priority);
  return { direct: sorted.slice(0, maxSlots - 1), overflow: sorted.slice(maxSlots - 1) };
}

// A destination is active only on an exact route match; actions are never active.
export function isActive(dest, path) {
  return dest.match != null && dest.match === path;
}
