import { partitionNav, isActive, icons } from '../nav/destinations.js';
import { clearToken } from '../api/auth.js';

const FOCUS = 'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-amber-400';
const BAR_ITEM =
  `flex-1 flex flex-col items-center justify-center gap-0.5 min-w-[44px] min-h-[44px] py-2 text-xs ${FOCUS}`;
const MENU_ITEM =
  `flex items-center gap-3 w-full min-h-[44px] px-4 py-2 text-sm text-left ${FOCUS}`;
const ACTIVE = 'text-amber-400';
const INACTIVE = 'text-stone-300 hover:text-amber-400';

function currentPath() {
  return window.location.hash.replace(/^#/, '') || '/';
}

// Wires the Sign Out action onto an element (clears the token, returns home).
function wireAction(el, dest) {
  if (dest.action === true && dest.id === 'sign-out') {
    el.type = 'button';
    el.addEventListener('click', () => {
      clearToken();
      window.location.hash = '#/';
    });
  }
}

// Builds a tappable destination as either a bar tab or an overflow-menu row.
function buildItem(dest, path, variant) {
  const active = isActive(dest, path);
  const el = document.createElement(dest.href ? 'a' : 'button');
  el.className = `${variant === 'menu' ? MENU_ITEM : BAR_ITEM} ${active ? ACTIVE : INACTIVE}`;
  el.setAttribute('data-nav-item', dest.id);
  if (dest.href) el.setAttribute('href', dest.href);
  if (active) el.setAttribute('aria-current', 'page');
  el.innerHTML = `${dest.icon}<span>${dest.label}</span>`;
  wireAction(el, dest);
  return el;
}

// Builds the "More" disclosure (button + popup menu) and returns { more, menu }
// with open/close behavior: Escape, outside pointer-down, selection, and route
// change all close it; opening moves focus into the menu (WCAG 2.1 AA).
function buildOverflow(overflow, path) {
  const active = overflow.some((d) => isActive(d, path));
  const more = document.createElement('button');
  more.type = 'button';
  more.className = `${BAR_ITEM} ${active ? ACTIVE : INACTIVE}`;
  more.setAttribute('data-nav-more', '');
  more.setAttribute('aria-haspopup', 'true');
  more.setAttribute('aria-expanded', 'false');
  more.setAttribute('aria-controls', 'nav-more');
  more.innerHTML = `${icons.more}<span>More</span>`;

  const menu = document.createElement('div');
  menu.id = 'nav-more';
  menu.setAttribute('data-nav-menu', '');
  menu.setAttribute('role', 'menu');
  menu.className =
    'hidden absolute bottom-full right-2 mb-2 min-w-[12rem] rounded-xl bg-stone-800 py-2 shadow-lg';
  for (const dest of overflow) menu.appendChild(buildItem(dest, path, 'menu'));

  const onKeydown = (e) => { if (e.key === 'Escape') { close(); more.focus(); } };
  const onOutside = (e) => { if (!more.contains(e.target) && !menu.contains(e.target)) close(); };

  function open() {
    menu.classList.remove('hidden');
    more.setAttribute('aria-expanded', 'true');
    menu.querySelector('[data-nav-item]')?.focus();
    document.addEventListener('keydown', onKeydown);
    document.addEventListener('pointerdown', onOutside);
    window.addEventListener('hashchange', close);
  }
  function close() {
    menu.classList.add('hidden');
    more.setAttribute('aria-expanded', 'false');
    document.removeEventListener('keydown', onKeydown);
    document.removeEventListener('pointerdown', onOutside);
    window.removeEventListener('hashchange', close);
  }

  more.addEventListener('click', () =>
    more.getAttribute('aria-expanded') === 'true' ? close() : open());
  menu.querySelectorAll('[data-nav-item]').forEach((item) =>
    item.addEventListener('click', close));

  return { more, menu };
}

// Fixed bottom tab bar for phones (< 768px): up to `maxSlots` slots — direct
// destinations plus a "More" overflow when they exceed the bound. Hidden at
// md+ where the top nav takes over.
export function buildBottomNav(destinations, maxSlots = 5) {
  const { direct, overflow } = partitionNav(destinations, maxSlots);
  const path = currentPath();

  const nav = document.createElement('nav');
  nav.setAttribute('aria-label', 'Primary');
  nav.className =
    'bottom-nav fixed bottom-0 inset-x-0 z-40 flex md:hidden bg-stone-900 border-t border-stone-800 pb-[env(safe-area-inset-bottom)]';

  for (const dest of direct) nav.appendChild(buildItem(dest, path, 'bar'));

  if (overflow.length) {
    const { more, menu } = buildOverflow(overflow, path);
    nav.appendChild(more);
    nav.appendChild(menu);
  }

  return nav;
}
