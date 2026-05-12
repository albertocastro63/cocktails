import { listUsers, deleteUser } from '../api/client.js';
import { getToken } from '../api/auth.js';

export function AdminUserList() {
  const root = document.createElement('div');
  root.className = 'max-w-4xl mx-auto px-4 py-8';

  const header = document.createElement('div');
  header.className = 'flex items-center justify-between mb-6';
  header.innerHTML = '<h1 class="text-2xl font-bold text-gray-900">Users</h1>';

  const addBtn = document.createElement('button');
  addBtn.setAttribute('data-add-user', '');
  addBtn.className = 'bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 text-sm';
  addBtn.textContent = '+ Add User';
  addBtn.addEventListener('click', () => { location.hash = '#/admin/users/new'; });
  header.appendChild(addBtn);

  const content = document.createElement('div');
  content.setAttribute('data-content', '');
  content.innerHTML = '<p class="text-gray-500">Loading…</p>';

  root.appendChild(header);
  root.appendChild(content);

  loadUsers(content);

  return root;
}

function loadUsers(content) {
  const token = getToken();
  listUsers(token)
    .then((users) => renderUsers(content, users))
    .catch(() => renderError(content));
}

function renderUsers(content, users) {
  content.innerHTML = '';
  if (users.length === 0) {
    const p = document.createElement('p');
    p.className = 'text-gray-500';
    p.textContent = 'No users yet.';
    content.appendChild(p);
    return;
  }

  const table = document.createElement('table');
  table.className = 'w-full text-left border-collapse';
  table.innerHTML = `
    <thead>
      <tr class="border-b">
        <th class="py-2 pr-4 text-gray-600 font-medium">Username</th>
        <th class="py-2 pr-4 text-gray-600 font-medium">Name</th>
        <th class="py-2 pr-4 text-gray-600 font-medium">Email</th>
        <th class="py-2 text-gray-600 font-medium">Actions</th>
      </tr>
    </thead>
  `;

  const tbody = document.createElement('tbody');
  users.forEach((user) => {
    const tr = document.createElement('tr');
    tr.className = 'border-b hover:bg-gray-50';

    const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ') || '—';
    const email = user.email || '—';

    tr.innerHTML = `
      <td class="py-2 pr-4">${escapeHtml(user.username)}</td>
      <td class="py-2 pr-4">${escapeHtml(fullName)}</td>
      <td class="py-2 pr-4">${escapeHtml(email)}</td>
      <td class="py-2 flex gap-3">
        <a href="#/admin/users/${user.id}/edit" class="text-indigo-600 hover:underline text-sm">Edit</a>
      </td>
    `;

    const deleteBtn = document.createElement('button');
    deleteBtn.setAttribute('data-delete-user', user.id);
    deleteBtn.className = 'text-red-600 hover:underline text-sm';
    deleteBtn.textContent = 'Delete';
    deleteBtn.addEventListener('click', () => handleDelete(user, content));
    tr.querySelector('td:last-child').appendChild(deleteBtn);

    tbody.appendChild(tr);
  });

  table.appendChild(tbody);
  content.appendChild(table);
}

function handleDelete(user, content) {
  if (!confirm(`Delete user ${user.username}?`)) return;
  const token = getToken();
  deleteUser(user.id, token)
    .then(() => loadUsers(content))
    .catch(() => {
      const err = document.createElement('p');
      err.className = 'text-red-600 text-sm mt-2';
      err.textContent = `Failed to delete ${user.username}.`;
      content.appendChild(err);
    });
}

function renderError(content) {
  content.innerHTML = '';
  const p = document.createElement('p');
  p.className = 'text-red-600';
  p.textContent = 'Failed to load users.';
  content.appendChild(p);
}

function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
