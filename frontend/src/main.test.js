import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('./api/auth.js', () => ({
  isLoggedIn: vi.fn(),
  isAdmin: vi.fn(),
  getToken: vi.fn(() => null),
  clearToken: vi.fn(),
}));

vi.mock('./pages/AdminUserList.js', () => ({
  AdminUserList: vi.fn(() => {
    const el = document.createElement('div');
    el.textContent = 'AdminUserList';
    return el;
  }),
}));

vi.mock('./pages/AdminRecipes.js', () => ({
  AdminRecipes: vi.fn(() => {
    const el = document.createElement('div');
    el.textContent = 'AdminRecipes';
    return el;
  }),
}));

vi.mock('./pages/Login.js', () => ({
  Login: vi.fn(() => {
    const el = document.createElement('div');
    el.textContent = 'Login';
    return el;
  }),
}));

import { isLoggedIn, isAdmin } from './api/auth.js';
import { buildNav, renderAdminRoute } from './main.js';

describe('admin route guard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '<div id="app"></div>';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('navigating to #/admin/users when not logged in renders Login component', () => {
    isLoggedIn.mockReturnValue(false);
    isAdmin.mockReturnValue(false);
    renderAdminRoute(document.getElementById('app'));
    expect(document.getElementById('app').textContent).toContain('Login');
  });

  it('navigating to #/admin/users when logged in as non-admin renders access denied message', () => {
    isLoggedIn.mockReturnValue(true);
    isAdmin.mockReturnValue(false);
    renderAdminRoute(document.getElementById('app'));
    expect(document.getElementById('app').textContent.toLowerCase()).toMatch(/access denied/);
  });
});

describe('navigation redesign', () => {
  it('buildNav top nav has bg-stone-900 class', () => {
    isLoggedIn.mockReturnValue(false);
    isAdmin.mockReturnValue(false);
    const nav = buildNav();
    // buildNav now returns a container: desktop top nav (hidden below md) + a
    // slim mobile brand header + the bottom bar. The top nav keeps its styling.
    const topNav = nav.querySelector('nav.md\\:flex');
    expect(topNav).toBeTruthy();
    expect(topNav.className).toContain('bg-stone-900');
  });
});

describe('admin nav link', () => {
  it('Admin nav link is not rendered when isAdmin() returns false', () => {
    isLoggedIn.mockReturnValue(true);
    isAdmin.mockReturnValue(false);
    const nav = buildNav();
    expect(nav.textContent.toLowerCase()).not.toContain('admin');
  });

  it('Admin nav link is rendered when isAdmin() returns true', () => {
    isLoggedIn.mockReturnValue(true);
    isAdmin.mockReturnValue(true);
    const nav = buildNav();
    expect(nav.querySelector('[href="#/admin/users"]')).toBeTruthy();
  });

  it('admin nav renders both Users and Recipes links for admin user', () => {
    isLoggedIn.mockReturnValue(true);
    isAdmin.mockReturnValue(true);
    const nav = buildNav();
    expect(nav.querySelector('[href="#/admin/users"]')).toBeTruthy();
    expect(nav.querySelector('[href="#/admin/recipes"]')).toBeTruthy();
  });

  it('admin recipes top-nav link is labelled "Manage Recipes"', () => {
    isLoggedIn.mockReturnValue(true);
    isAdmin.mockReturnValue(true);
    const link = buildNav().querySelector('[href="#/admin/recipes"]');
    expect(link.textContent).toBe('Manage Recipes');
  });
});

describe('responsive nav contract (US3)', () => {
  it('desktop top nav is hidden below md; bottom bar + brand header are md:hidden (N1/N2, SC-004)', () => {
    isLoggedIn.mockReturnValue(false);
    isAdmin.mockReturnValue(false);
    window.location.hash = '#/recipes'; // non-home page: brand header shows on phones
    const nav = buildNav();

    const topNav = nav.querySelector('nav.md\\:flex');
    expect(topNav.className).toContain('hidden');
    expect(topNav.className).toContain('md:flex');

    const bottom = nav.querySelector('nav[aria-label="Primary"]');
    expect(bottom.className).toContain('md:hidden');

    const brand = nav.querySelector('header');
    expect(brand.className).toContain('md:hidden');
    expect(brand.querySelector('[href="#/"]')).toBeTruthy(); // Home reachable on mobile
  });

  it('hides the mobile brand header on the landing page (home route)', () => {
    isLoggedIn.mockReturnValue(false);
    isAdmin.mockReturnValue(false);
    window.location.hash = '#/'; // landing page
    const brand = buildNav().querySelector('header');
    expect(brand.className).toContain('hidden');
    expect(brand.className).not.toContain('flex'); // fully hidden, not phone-visible
  });

  it('visitor bottom bar shows only visitor destinations (N3)', () => {
    isLoggedIn.mockReturnValue(false);
    isAdmin.mockReturnValue(false);
    const bottom = buildNav().querySelector('nav[aria-label="Primary"]');
    const items = bottom.querySelectorAll('[data-nav-item]');
    expect(items).toHaveLength(2);
    expect(bottom.querySelector('[href="#/recipes"]')).toBeTruthy();
    expect(bottom.querySelector('[href="#/login"]')).toBeTruthy();
  });
});
