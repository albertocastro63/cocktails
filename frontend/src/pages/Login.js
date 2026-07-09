import { login } from '../api/client.js';
import { setToken } from '../api/auth.js';

export function Login({ onSuccess } = {}) {
  const el = document.createElement('div');
  el.className = 'max-w-sm mx-auto px-4 py-16';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-8 text-center';
  heading.textContent = 'Sign In';
  el.appendChild(heading);

  const card = document.createElement('div');
  card.className = 'bg-white rounded-2xl shadow-sm border border-stone-200 p-8';

  const form = document.createElement('form');
  form.className = 'space-y-4';
  form.innerHTML = `
    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Username</label>
      <input name="username" type="text" autocomplete="username"
        class="w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent" />
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Password</label>
      <input name="password" type="password" autocomplete="current-password"
        class="w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent" />
    </div>
    <button type="submit"
      class="w-full bg-amber-500 text-stone-900 rounded-xl px-4 py-2 font-semibold hover:bg-amber-600 transition-colors">
      Sign In
    </button>
    <p class="text-center text-sm"><a href="#/forgot" class="text-amber-700 hover:text-amber-800">Forgot password?</a></p>
  `;
  card.appendChild(form);
  el.appendChild(card);

  const errorMsg = document.createElement('p');
  errorMsg.className = 'mt-4 text-red-600 text-sm text-center hidden';
  el.appendChild(errorMsg);

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorMsg.classList.add('hidden');
    const username = form.querySelector('[name="username"]').value;
    const password = form.querySelector('[name="password"]').value;
    try {
      const { token } = await login(username, password);
      setToken(token);
      if (onSuccess) onSuccess();
    } catch (err) {
      errorMsg.textContent = err.message || 'Login failed. Check your credentials.';
      errorMsg.classList.remove('hidden');
    }
  });

  return el;
}
