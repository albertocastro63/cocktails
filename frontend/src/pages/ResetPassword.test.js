import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../api/client.js', () => ({ resetPassword: vi.fn() }));
import { resetPassword } from '../api/client.js';
import { ResetPassword } from './ResetPassword.js';

function mount(hash = '#/reset?uid=u1&token=tok') {
  window.location.hash = hash;
  const el = ResetPassword();
  document.body.appendChild(el);
  return el;
}

describe('ResetPassword page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    resetPassword.mockResolvedValue({});
  });

  it('rejects mismatched passwords without calling the API', async () => {
    const el = mount();
    el.querySelector('#rp-pw').value = 'NewStrong1!aa';
    el.querySelector('#rp-pw2').value = 'Different1!aa';
    el.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await Promise.resolve();
    expect(document.body.textContent.toLowerCase()).toContain('do not match');
    expect(resetPassword).not.toHaveBeenCalled();
  });

  it('rejects a weak password without calling the API', async () => {
    const el = mount();
    el.querySelector('#rp-pw').value = 'weak';
    el.querySelector('#rp-pw2').value = 'weak';
    el.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await Promise.resolve();
    expect(resetPassword).not.toHaveBeenCalled();
  });

  it('submits a matching strong password with uid+token and shows success', async () => {
    const el = mount();
    el.querySelector('#rp-pw').value = 'NewStrong1!aa';
    el.querySelector('#rp-pw2').value = 'NewStrong1!aa';
    el.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toContain('has been reset');
    });
    expect(resetPassword).toHaveBeenCalledWith('u1', 'tok', 'NewStrong1!aa');
  });

  it('shows a generic invalid-link message with a request-new link on INVALID_RESET', async () => {
    const err = new Error('this reset link is invalid or has expired');
    err.code = 'INVALID_RESET';
    resetPassword.mockRejectedValue(err);
    const el = mount();
    el.querySelector('#rp-pw').value = 'NewStrong1!aa';
    el.querySelector('#rp-pw2').value = 'NewStrong1!aa';
    el.querySelector('form').dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toContain('invalid or has expired');
    });
    expect(document.body.querySelector('[href="#/forgot"]')).toBeTruthy();
  });
});
