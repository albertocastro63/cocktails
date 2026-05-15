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
  it('buildNav has bg-stone-900 class', () => {
    isLoggedIn.mockReturnValue(false);
    isAdmin.mockReturnValue(false);
    const nav = buildNav();
    expect(nav.className).toContain('bg-stone-900');
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
});
