import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  login: vi.fn(),
}));
vi.mock('../api/auth.js', () => ({
  setToken: vi.fn(),
  getToken: vi.fn(() => null),
  isLoggedIn: vi.fn(() => false),
}));

import { login } from '../api/client.js';
import { setToken } from '../api/auth.js';
import { Login } from './Login.js';

describe('Login page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders username and password inputs', () => {
    const el = Login({ onSuccess: vi.fn() });
    document.body.appendChild(el);
    expect(document.body.querySelector('input[type="text"], input[name="username"]')).not.toBeNull();
    expect(document.body.querySelector('input[type="password"]')).not.toBeNull();
  });

  it('stores JWT and calls onSuccess on successful login', async () => {
    login.mockResolvedValue({ token: 'tok123', expires_at: new Date().toISOString() });
    const onSuccess = vi.fn();
    const el = Login({ onSuccess });
    document.body.appendChild(el);

    const form = document.body.querySelector('form');
    const usernameInput = form.querySelector('input[name="username"]') || form.querySelectorAll('input')[0];
    const passwordInput = form.querySelector('input[type="password"]');
    usernameInput.value = 'alice';
    passwordInput.value = 'secret';
    form.dispatchEvent(new Event('submit'));

    await vi.waitFor(() => {
      expect(setToken).toHaveBeenCalledWith('tok123');
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('shows error message on failed login', async () => {
    const err = new Error('invalid credentials');
    err.status = 401;
    login.mockRejectedValue(err);
    const el = Login({ onSuccess: vi.fn() });
    document.body.appendChild(el);

    const form = document.body.querySelector('form');
    form.dispatchEvent(new Event('submit'));

    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/invalid|error|failed/);
    });
  });
});
