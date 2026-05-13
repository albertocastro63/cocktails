import { createUser, getUser, updateUser } from '../api/client.js';
import { getToken } from '../api/auth.js';

export function AdminUserForm({ id, onSave } = {}) {
  const isEdit = Boolean(id);
  const root = document.createElement('div');
  root.className = 'max-w-2xl mx-auto px-4 py-8';

  const heading = document.createElement('h1');
  heading.className = 'text-2xl font-bold text-gray-900 mb-6';
  heading.textContent = isEdit ? 'Edit User' : 'New User';
  root.appendChild(heading);

  const errorP = document.createElement('p');
  errorP.className = 'text-red-600 mb-4 hidden';
  root.appendChild(errorP);

  const form = document.createElement('form');
  form.className = 'space-y-4';

  if (isEdit) {
    const usernameLabel = document.createElement('label');
    usernameLabel.className = 'block text-sm font-medium text-gray-700';
    usernameLabel.textContent = 'Username';
    const usernameVal = document.createElement('p');
    usernameVal.id = 'edit-username';
    usernameVal.className = 'mt-1 text-gray-900';
    root.insertBefore(usernameLabel, errorP);
    root.insertBefore(usernameVal, errorP);
  } else {
    form.appendChild(field('username', 'Username', 'text', '', false));
  }

  const pwPlaceholder = isEdit ? 'Leave blank to keep existing' : '';
  form.appendChild(field('first_name', 'First Name', 'text', '', false));
  form.appendChild(field('last_name', 'Last Name', 'text', '', false));
  form.appendChild(field('email', 'Email', 'email', '', false));
  form.appendChild(field('password', 'Password', 'password', pwPlaceholder, false));

  const submitBtn = document.createElement('button');
  submitBtn.type = 'submit';
  submitBtn.className = 'bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 text-sm';
  submitBtn.textContent = isEdit ? 'Save Changes' : 'Create User';
  form.appendChild(submitBtn);

  form.addEventListener('submit', (e) => {
    e.preventDefault();
    handleSubmit(form, errorP, id, onSave);
  });

  root.appendChild(form);

  if (isEdit) {
    loadUser(id, form, root);
  }

  return root;
}

function field(name, label, type, placeholder, required) {
  const div = document.createElement('div');
  const lbl = document.createElement('label');
  lbl.className = 'block text-sm font-medium text-gray-700';
  lbl.textContent = label;
  const input = document.createElement('input');
  input.name = name;
  input.type = type;
  input.placeholder = placeholder;
  input.required = required;
  input.className = 'mt-1 block w-full border border-gray-300 rounded px-3 py-2 text-sm';
  div.appendChild(lbl);
  div.appendChild(input);
  return div;
}

function showError(errorP, msg) {
  errorP.textContent = msg;
  errorP.classList.remove('hidden');
}

function hideError(errorP) {
  errorP.classList.add('hidden');
}

function handleSubmit(form, errorP, id, onSave) {
  hideError(errorP);
  const isEdit = Boolean(id);
  const token = getToken();

  const username = form.querySelector('input[name=username]')?.value.trim();
  const password = form.querySelector('input[name=password]').value;
  const first_name = form.querySelector('input[name=first_name]').value.trim();
  const last_name = form.querySelector('input[name=last_name]').value.trim();
  const email = form.querySelector('input[name=email]').value.trim();

  if (!isEdit) {
    if (!username) {
      showError(errorP, 'Username is required.');
      return;
    }
    if (!password) {
      showError(errorP, 'Password is required.');
      return;
    }
    createUser({ username, password, first_name, last_name, email }, token)
      .then((user) => (onSave ? onSave(user) : (location.hash = '#/admin/users')))
      .catch((err) => showError(errorP, err.message || 'Failed to create user.'));
    return;
  }

  const payload = { first_name, last_name, email };
  if (password) payload.password = password;
  updateUser(id, payload, token)
    .then((user) => (onSave ? onSave(user) : (location.hash = '#/admin/users')))
    .catch((err) => showError(errorP, err.message || 'Failed to update user.'));
}

function loadUser(id, form, root) {
  const token = getToken();
  getUser(id, token).then((user) => {
    const usernameEl = root.querySelector('#edit-username');
    if (usernameEl) usernameEl.textContent = user.username;
    form.querySelector('input[name=first_name]').value = user.first_name || '';
    form.querySelector('input[name=last_name]').value = user.last_name || '';
    form.querySelector('input[name=email]').value = user.email || '';
  });
}
