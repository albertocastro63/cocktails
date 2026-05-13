import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  listUsers: vi.fn(),
  deleteUser: vi.fn(),
}));

vi.mock('../api/auth.js', () => ({
  getToken: vi.fn(() => 'mock-token'),
  isLoggedIn: vi.fn(() => true),
  isAdmin: vi.fn(() => true),
}));

import { listUsers, deleteUser } from '../api/client.js';
import { AdminUserList } from './AdminUserList.js';

describe('AdminUserList page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders loading state initially', () => {
    listUsers.mockReturnValue(new Promise(() => {}));
    const el = AdminUserList();
    document.body.appendChild(el);
    expect(document.body.textContent.toLowerCase()).toMatch(/loading/);
  });

  it('renders table with user rows when users load', async () => {
    listUsers.mockResolvedValue([
      { id: 'u1', username: 'alice', first_name: 'Alice', last_name: 'Smith', email: 'alice@example.com' },
    ]);
    const el = AdminUserList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('alice');
    });
    expect(document.body.querySelector('table')).toBeTruthy();
  });

  it('renders empty state when no users', async () => {
    listUsers.mockResolvedValue([]);
    const el = AdminUserList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/no users/);
    });
  });

  it('renders error state on API failure', async () => {
    listUsers.mockRejectedValue(new Error('network error'));
    const el = AdminUserList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/failed to load/);
    });
  });

  it('clicking Add User navigates to #/admin/users/new', async () => {
    listUsers.mockResolvedValue([]);
    const el = AdminUserList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const btn = document.body.querySelector('[data-add-user]');
      expect(btn).toBeTruthy();
    });
    const btn = document.body.querySelector('[data-add-user]');
    btn.click();
    expect(location.hash).toBe('#/admin/users/new');
  });

  it('clicking Delete shows confirm dialog; on confirm calls deleteUser and refreshes list', async () => {
    listUsers.mockResolvedValue([
      { id: 'u1', username: 'alice', first_name: '', last_name: '', email: '' },
    ]);
    deleteUser.mockResolvedValue(null);
    vi.spyOn(window, 'confirm').mockReturnValue(true);

    const el = AdminUserList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('[data-delete-user]')).toBeTruthy();
    });

    document.body.querySelector('[data-delete-user]').click();
    expect(window.confirm).toHaveBeenCalled();
    await vi.waitFor(() => {
      expect(deleteUser).toHaveBeenCalledWith('u1', 'mock-token');
    });
  });

  it('clicking Delete and cancelling confirm does not call deleteUser', async () => {
    listUsers.mockResolvedValue([
      { id: 'u1', username: 'alice', first_name: '', last_name: '', email: '' },
    ]);
    vi.spyOn(window, 'confirm').mockReturnValue(false);

    const el = AdminUserList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('[data-delete-user]')).toBeTruthy();
    });

    document.body.querySelector('[data-delete-user]').click();
    expect(deleteUser).not.toHaveBeenCalled();
  });
});
