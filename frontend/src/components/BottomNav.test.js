import { describe, it, expect, beforeEach, vi } from 'vitest';
import { buildBottomNav } from './BottomNav.js';
import { navDestinations } from '../nav/destinations.js';

vi.mock('../api/auth.js', () => ({ clearToken: vi.fn() }));

beforeEach(() => {
  window.location.hash = '';
  document.body.innerHTML = '';
});

describe('BottomNav — direct items (US1)', () => {
  it('renders one item per user destination with icon + label, and no More (N7, N4)', () => {
    const bar = buildBottomNav(navDestinations('user'));
    const items = bar.querySelectorAll('[data-nav-item]');
    expect(items).toHaveLength(4);
    for (const item of items) {
      expect(item.querySelector('svg'), 'icon').toBeTruthy();
      expect(item.textContent.trim(), 'label').not.toBe('');
    }
    expect(bar.querySelector('[data-nav-more]')).toBeNull();
  });

  it('marks the item matching the current route active (B2)', () => {
    window.location.hash = '#/recipes';
    const bar = buildBottomNav(navDestinations('user'));
    const active = bar.querySelector('[aria-current="page"]');
    expect(active).toBeTruthy();
    expect(active.textContent).toContain('All');
  });

  it('does not mark All Recipes active on the New Recipe route', () => {
    window.location.hash = '#/recipes/new';
    const bar = buildBottomNav(navDestinations('user'));
    const active = bar.querySelector('[aria-current="page"]');
    expect(active.textContent).toContain('New');
  });

  it('every item carries a min tap-target size class (N6/SC-003)', () => {
    const bar = buildBottomNav(navDestinations('user'));
    for (const item of bar.querySelectorAll('[data-nav-item]')) {
      expect(item.className).toMatch(/min-w-\[44px\]/);
      expect(item.className).toMatch(/min-h-\[44px\]/);
    }
  });

  it('is a labelled nav fixed to the bottom and hidden at md+ (N1/N2)', () => {
    const bar = buildBottomNav(navDestinations('user'));
    expect(bar.tagName).toBe('NAV');
    expect(bar.getAttribute('aria-label')).toBeTruthy();
    expect(bar.className).toContain('bottom-nav');
    expect(bar.className).toContain('fixed');
    expect(bar.className).toContain('bottom-0');
    expect(bar.className).toContain('md:hidden');
  });

  it('visitor set renders All + Sign in (N3)', () => {
    const bar = buildBottomNav(navDestinations('visitor'));
    const labels = [...bar.querySelectorAll('[data-nav-item]')].map((i) => i.textContent.trim());
    expect(labels).toHaveLength(2);
    expect(bar.querySelector('[href="#/recipes"]')).toBeTruthy();
    expect(bar.querySelector('[href="#/login"]')).toBeTruthy();
  });
});

describe('BottomNav — admin overflow "More" menu (US2)', () => {
  function mountAdmin() {
    const bar = buildBottomNav(navDestinations('admin'));
    document.body.appendChild(bar);
    return bar;
  }

  it('renders 4 direct items + a More button (N5)', () => {
    const bar = mountAdmin();
    expect(bar.querySelectorAll('nav > [data-nav-item]')).toHaveLength(4);
    const more = bar.querySelector('[data-nav-more]');
    expect(more).toBeTruthy();
    expect(more.getAttribute('aria-expanded')).toBe('false');
    expect(more.getAttribute('aria-controls')).toBe('nav-more');
  });

  it('More menu holds the overflow destinations (Manage, Sign out)', () => {
    const bar = mountAdmin();
    const menu = bar.querySelector('#nav-more');
    const labels = [...menu.querySelectorAll('[data-nav-item]')].map((i) => i.textContent.trim());
    expect(labels.join(' ')).toContain('Manage');
    expect(labels.join(' ')).toContain('Sign out');
  });

  it('opening More sets aria-expanded and moves focus to the first item (B4)', () => {
    const bar = mountAdmin();
    const more = bar.querySelector('[data-nav-more]');
    more.click();
    expect(more.getAttribute('aria-expanded')).toBe('true');
    const firstItem = bar.querySelector('#nav-more [data-nav-item]');
    expect(document.activeElement).toBe(firstItem);
  });

  it('Escape closes the menu and returns focus to More (B5)', () => {
    const bar = mountAdmin();
    const more = bar.querySelector('[data-nav-more]');
    more.click();
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    expect(more.getAttribute('aria-expanded')).toBe('false');
    expect(document.activeElement).toBe(more);
  });

  it('selecting an overflow item closes the menu (B5)', () => {
    const bar = mountAdmin();
    bar.querySelector('[data-nav-more]').click();
    bar.querySelector('#nav-more [data-nav-item]').click();
    expect(bar.querySelector('[data-nav-more]').getAttribute('aria-expanded')).toBe('false');
  });

  it('an outside pointer-down closes the menu (B5)', () => {
    const bar = mountAdmin();
    bar.querySelector('[data-nav-more]').click();
    document.body.dispatchEvent(new MouseEvent('pointerdown', { bubbles: true }));
    expect(bar.querySelector('[data-nav-more]').getAttribute('aria-expanded')).toBe('false');
  });

  it('More carries the active token when the route is in the overflow set (B3)', () => {
    window.location.hash = '#/admin/recipes';
    const bar = mountAdmin();
    const more = bar.querySelector('[data-nav-more]');
    expect(more.className).toContain('text-amber-400');
  });

  it('overflow items meet the min tap-target size (N6)', () => {
    const bar = mountAdmin();
    for (const item of bar.querySelectorAll('#nav-more [data-nav-item]')) {
      expect(item.className).toMatch(/min-h-\[44px\]/);
    }
  });
});
