import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  createUser: vi.fn(),
  getUser: vi.fn(),
  updateUser: vi.fn(),
}));

vi.mock('../api/auth.js', () => ({
  getToken: vi.fn(() => 'mock-token'),
  isLoggedIn: vi.fn(() => true),
  isAdmin: vi.fn(() => true),
}));

import { createUser, getUser, updateUser } from '../api/client.js';
import { AdminUserForm } from './AdminUserForm.js';

describe('AdminUserForm — create mode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders username, password, first_name, last_name, email inputs', () => {
    const el = AdminUserForm({});
    document.body.appendChild(el);
    expect(document.body.querySelector('input[name=username]')).toBeTruthy();
    expect(document.body.querySelector('input[name=password]')).toBeTruthy();
    expect(document.body.querySelector('input[name=first_name]')).toBeTruthy();
    expect(document.body.querySelector('input[name=last_name]')).toBeTruthy();
    expect(document.body.querySelector('input[name=email]')).toBeTruthy();
  });

  it('submitting without username shows validation error and does not call API', async () => {
    const el = AdminUserForm({});
    document.body.appendChild(el);
    document.body.querySelector('input[name=password]').value = 'pass123';
    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));
    expect(createUser).not.toHaveBeenCalled();
    expect(document.body.querySelector('p.text-red-600').textContent).toBeTruthy();
  });

  it('submitting without password shows validation error and does not call API', async () => {
    const el = AdminUserForm({});
    document.body.appendChild(el);
    document.body.querySelector('input[name=username]').value = 'alice';
    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));
    expect(createUser).not.toHaveBeenCalled();
    expect(document.body.querySelector('p.text-red-600').textContent).toBeTruthy();
  });

  it('successful create calls createUser with all provided fields and navigates', async () => {
    createUser.mockResolvedValue({ id: 'u1', username: 'alice' });
    const onSave = vi.fn();
    const el = AdminUserForm({ onSave });
    document.body.appendChild(el);

    document.body.querySelector('input[name=username]').value = 'alice';
    document.body.querySelector('input[name=password]').value = 'pass123';
    document.body.querySelector('input[name=first_name]').value = 'Alice';
    document.body.querySelector('input[name=last_name]').value = 'Smith';
    document.body.querySelector('input[name=email]').value = 'alice@example.com';
    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));

    await vi.waitFor(() => {
      expect(createUser).toHaveBeenCalledWith(
        { username: 'alice', password: 'pass123', first_name: 'Alice', last_name: 'Smith', email: 'alice@example.com' },
        'mock-token',
      );
    });
    expect(onSave).toHaveBeenCalled();
  });

  it('API error 409 USERNAME_CONFLICT is shown in error paragraph', async () => {
    const err = Object.assign(new Error('username already exists'), { code: 'CONFLICT' });
    createUser.mockRejectedValue(err);
    const el = AdminUserForm({});
    document.body.appendChild(el);
    document.body.querySelector('input[name=username]').value = 'alice';
    document.body.querySelector('input[name=password]').value = 'pass123';
    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));
    await vi.waitFor(() => {
      expect(document.body.querySelector('p.text-red-600').textContent).toBeTruthy();
    });
  });

  it('API error 409 EMAIL_CONFLICT is shown in error paragraph', async () => {
    const err = Object.assign(new Error('email already in use'), { code: 'EMAIL_CONFLICT' });
    createUser.mockRejectedValue(err);
    const el = AdminUserForm({});
    document.body.appendChild(el);
    document.body.querySelector('input[name=username]').value = 'alice';
    document.body.querySelector('input[name=password]').value = 'pass123';
    document.body.querySelector('input[name=email]').value = 'taken@example.com';
    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));
    await vi.waitFor(() => {
      expect(document.body.querySelector('p.text-red-600').textContent).toBeTruthy();
    });
  });
});

describe('AdminUserForm — edit mode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('edit mode pre-fills first_name, last_name, email from fetched user', async () => {
    getUser.mockResolvedValue({
      id: 'u1', username: 'alice', first_name: 'Alice', last_name: 'Smith', email: 'alice@example.com',
    });
    const el = AdminUserForm({ id: 'u1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('input[name=first_name]').value).toBe('Alice');
      expect(document.body.querySelector('input[name=last_name]').value).toBe('Smith');
      expect(document.body.querySelector('input[name=email]').value).toBe('alice@example.com');
    });
  });

  it('edit mode renders username as read-only (not an editable input)', async () => {
    getUser.mockResolvedValue({ id: 'u1', username: 'alice', first_name: '', last_name: '', email: '' });
    const el = AdminUserForm({ id: 'u1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('alice');
    });
    expect(document.body.querySelector('input[name=username]')).toBeNull();
  });

  it("password input has placeholder 'Leave blank to keep existing' in edit mode", async () => {
    getUser.mockResolvedValue({ id: 'u1', username: 'alice', first_name: '', last_name: '', email: '' });
    const el = AdminUserForm({ id: 'u1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const pwInput = document.body.querySelector('input[name=password]');
      expect(pwInput.placeholder.toLowerCase()).toMatch(/leave blank/);
    });
  });

  it('successful edit calls updateUser with correct payload and navigates', async () => {
    getUser.mockResolvedValue({ id: 'u1', username: 'alice', first_name: 'Alice', last_name: '', email: '' });
    updateUser.mockResolvedValue({ id: 'u1', username: 'alice', first_name: 'Alicia', last_name: '', email: '' });
    const onSave = vi.fn();
    const el = AdminUserForm({ id: 'u1', onSave });
    document.body.appendChild(el);

    await vi.waitFor(() => {
      expect(document.body.querySelector('input[name=first_name]').value).toBe('Alice');
    });

    document.body.querySelector('input[name=first_name]').value = 'Alicia';
    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));

    await vi.waitFor(() => {
      expect(updateUser).toHaveBeenCalledWith('u1', expect.objectContaining({ first_name: 'Alicia' }), 'mock-token');
    });
    expect(onSave).toHaveBeenCalled();
  });

  it('blank password field in edit mode does not include password in payload', async () => {
    getUser.mockResolvedValue({ id: 'u1', username: 'alice', first_name: '', last_name: '', email: '' });
    updateUser.mockResolvedValue({ id: 'u1', username: 'alice', first_name: '', last_name: '', email: '' });
    const el = AdminUserForm({ id: 'u1' });
    document.body.appendChild(el);

    await vi.waitFor(() => {
      expect(document.body.querySelector('input[name=first_name]')).toBeTruthy();
    });

    document.body.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true }));

    await vi.waitFor(() => {
      expect(updateUser).toHaveBeenCalled();
      const payload = updateUser.mock.calls[0][1];
      expect(payload.password).toBeUndefined();
    });
  });
});
