import { requestPasswordReset } from '../api/client.js';

const NEUTRAL = 'If an account exists for that email, a password reset link has been sent. Please check your inbox.';

export function ForgotPassword() {
  const el = document.createElement('div');
  el.className = 'max-w-sm mx-auto px-4 py-16';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-8 text-center';
  heading.textContent = 'Forgot password';
  el.appendChild(heading);

  const card = document.createElement('div');
  card.className = 'bg-white rounded-2xl shadow-sm border border-stone-200 p-8';
  el.appendChild(card);

  const form = document.createElement('form');
  form.className = 'space-y-4';
  form.innerHTML = `
    <p class="text-sm text-stone-600">Enter your account email and we'll send you a link to reset your password.</p>
    <div>
      <label for="fp-email" class="block text-sm font-medium text-stone-700 mb-1">Email</label>
      <input id="fp-email" name="email" type="email" autocomplete="email" required
        class="w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent" />
    </div>
    <button type="submit"
      class="w-full bg-amber-500 text-stone-900 rounded-xl px-4 py-2 font-semibold hover:bg-amber-600 transition-colors">
      Send reset link
    </button>
    <p class="text-center text-sm"><a href="#/login" class="text-amber-700 hover:text-amber-800">Back to sign in</a></p>
  `;
  card.appendChild(form);

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = form.querySelector('#fp-email').value.trim();
    const btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    // Always show the same neutral confirmation regardless of the outcome
    // (no account enumeration); network errors are swallowed intentionally.
    try {
      await requestPasswordReset(email);
    } catch {
      /* neutral regardless */
    }
    card.innerHTML = `
      <p role="status" class="text-stone-700">${NEUTRAL}</p>
      <p class="mt-6 text-center text-sm"><a href="#/login" class="text-amber-700 hover:text-amber-800">Back to sign in</a></p>
    `;
  });

  return el;
}
