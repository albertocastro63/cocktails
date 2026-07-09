import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../api/client.js', () => ({ requestPasswordReset: vi.fn().mockResolvedValue({}) }));
import { requestPasswordReset } from '../api/client.js';
import { ForgotPassword } from './ForgotPassword.js';

describe('ForgotPassword page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  it('submits the email and shows a neutral confirmation', async () => {
    const el = ForgotPassword();
    document.body.appendChild(el);
    el.querySelector('#fp-email').value = 'a@b.com';
    el.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));

    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toContain('if an account exists');
    });
    expect(requestPasswordReset).toHaveBeenCalledWith('a@b.com');
  });

  it('has a link back to sign in', () => {
    const el = ForgotPassword();
    expect(el.querySelector('[href="#/login"]')).toBeTruthy();
  });
});
