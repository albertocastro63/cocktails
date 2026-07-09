import { resetPassword } from '../api/client.js';
import { PASSWORD_RULES, isPasswordValid } from '../utils/password.js';

function parseParams() {
  const q = (window.location.hash.split('?')[1]) || '';
  const p = new URLSearchParams(q);
  return { uid: p.get('uid') || '', token: p.get('token') || '' };
}

export function ResetPassword() {
  const el = document.createElement('div');
  el.className = 'max-w-sm mx-auto px-4 py-16';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-8 text-center';
  heading.textContent = 'Choose a new password';
  el.appendChild(heading);

  const card = document.createElement('div');
  card.className = 'bg-white rounded-2xl shadow-sm border border-stone-200 p-8';
  el.appendChild(card);

  const { uid, token } = parseParams();

  const form = document.createElement('form');
  form.className = 'space-y-4';
  form.innerHTML = `
    <div>
      <label for="rp-pw" class="block text-sm font-medium text-stone-700 mb-1">New password</label>
      <input id="rp-pw" name="password" type="password" autocomplete="new-password"
        class="w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent" />
    </div>
    <div>
      <label for="rp-pw2" class="block text-sm font-medium text-stone-700 mb-1">Confirm new password</label>
      <input id="rp-pw2" name="confirm" type="password" autocomplete="new-password"
        class="w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent" />
    </div>
    <ul data-rules class="text-sm space-y-1"></ul>
    <button type="submit"
      class="w-full bg-amber-500 text-stone-900 rounded-xl px-4 py-2 font-semibold hover:bg-amber-600 transition-colors">
      Reset password
    </button>
  `;
  card.appendChild(form);

  const pwInput = form.querySelector('#rp-pw');
  const confirmInput = form.querySelector('#rp-pw2');
  const rulesEl = form.querySelector('[data-rules]');
  const renderRules = (pw) => {
    rulesEl.innerHTML = PASSWORD_RULES.map((r) => {
      const ok = r.test(pw);
      return `<li class="${ok ? 'text-green-700' : 'text-stone-500'}">${ok ? '✓' : '•'} ${r.label}</li>`;
    }).join('');
  };
  renderRules('');
  pwInput.addEventListener('input', () => renderRules(pwInput.value));

  const error = document.createElement('p');
  error.className = 'mt-4 text-red-600 text-sm text-center hidden';
  error.setAttribute('role', 'alert');
  el.appendChild(error);
  const showError = (msg) => { error.textContent = msg; error.classList.remove('hidden'); };

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    error.classList.add('hidden');
    const pw = pwInput.value;
    if (pw !== confirmInput.value) {
      showError('Passwords do not match.');
      return;
    }
    if (!isPasswordValid(pw)) {
      showError('Password does not meet the requirements above.');
      return;
    }
    const btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    try {
      await resetPassword(uid, token, pw);
      card.innerHTML = `
        <p role="status" class="text-stone-700">Your password has been reset. You can now sign in with your new password.</p>
        <p class="mt-6 text-center"><a href="#/login" class="inline-block bg-amber-500 text-stone-900 rounded-xl px-4 py-2 font-semibold hover:bg-amber-600">Go to sign in</a></p>
      `;
    } catch (err) {
      btn.disabled = false;
      if (err.code === 'INVALID_RESET') {
        card.innerHTML = `
          <p role="alert" class="text-stone-700">This reset link is invalid or has expired.</p>
          <p class="mt-6 text-center"><a href="#/forgot" class="text-amber-700 hover:text-amber-800">Request a new reset link</a></p>
        `;
      } else {
        showError(err.message || 'Something went wrong. Please try again.');
      }
    }
  });

  return el;
}
