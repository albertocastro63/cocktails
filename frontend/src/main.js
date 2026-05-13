import './index.css';
import { Home } from './pages/Home.js';
import { RecipeList } from './pages/RecipeList.js';
import { RecipeDetail } from './pages/RecipeDetail.js';
import { Login } from './pages/Login.js';
import { RecipeForm } from './pages/RecipeForm.js';
import { AdminUserList } from './pages/AdminUserList.js';
import { AdminUserForm } from './pages/AdminUserForm.js';
import { isLoggedIn, isAdmin, clearToken } from './api/auth.js';

const routes = [
  { pattern: /^\/admin\/users\/([^/]+)\/edit$/, factory: (m) => AdminUserForm({ id: m[1], onSave: () => navigate('#/admin/users') }) },
  { pattern: /^\/admin\/users\/new$/, factory: () => AdminUserForm({ onSave: () => navigate('#/admin/users') }) },
  { pattern: /^\/admin\/users$/, factory: () => AdminUserList() },
  { pattern: /^\/recipes\/([^/]+)\/edit$/, factory: (m) => RecipeForm({ id: m[1], onSave: () => navigate(`#/recipes/${m[1]}`) }) },
  { pattern: /^\/recipes\/new$/, factory: () => RecipeForm({ onSave: (r) => navigate(`#/recipes/${r?.data?.id || ''}`) }) },
  { pattern: /^\/recipes\/([^/]+)$/, factory: (m) => RecipeDetail({ id: m[1] }) },
  { pattern: /^\/recipes$/, factory: () => RecipeList() },
  { pattern: /^\/login$/, factory: () => Login({ onSuccess: () => navigate('#/') }) },
  { pattern: /^\/$/, factory: () => Home() },
];

function navigate(hash) {
  location.hash = hash || '/';
}

function getPath() {
  return location.hash.replace(/^#/, '') || '/';
}

export function renderAdminRoute(root) {
  if (!isLoggedIn()) {
    root.appendChild(Login({ onSuccess: () => { root.innerHTML = ''; root.appendChild(AdminUserList()); } }));
    return;
  }
  if (!isAdmin()) {
    const p = document.createElement('p');
    p.className = 'text-center py-16 text-red-600';
    p.textContent = 'Access denied. Admin only.';
    root.appendChild(p);
    return;
  }
  root.appendChild(AdminUserList());
}

function renderPage() {
  const path = getPath();
  const root = document.getElementById('app');
  root.innerHTML = '';
  root.appendChild(buildNav());

  // Admin route guard
  if (/^\/admin/.test(path)) {
    if (!isLoggedIn()) {
      root.appendChild(Login({ onSuccess: () => renderPage() }));
      return;
    }
    if (!isAdmin()) {
      const p = document.createElement('p');
      p.className = 'text-center py-16 text-red-600';
      p.textContent = 'Access denied. Admin only.';
      root.appendChild(p);
      return;
    }
    // Fall through to route matching for the specific admin page
  }

  // Auth guard for write routes
  const writeRoutes = /^\/(recipes\/new|recipes\/.+\/edit)/;
  if (writeRoutes.test(path) && !isLoggedIn()) {
    root.appendChild(Login({ onSuccess: () => renderPage() }));
    return;
  }

  let matched = false;
  for (const { pattern, factory } of routes) {
    const m = path.match(pattern);
    if (m) {
      root.appendChild(factory(m));
      matched = true;
      break;
    }
  }
  if (!matched) {
    const el = document.createElement('p');
    el.className = 'text-center py-16 text-gray-400';
    el.textContent = 'Page not found.';
    root.appendChild(el);
  }
}

export function buildNav() {
  const nav = document.createElement('nav');
  nav.className = 'bg-white border-b border-gray-200 px-6 py-3 flex items-center gap-6';
  nav.innerHTML = `
    <a href="#/" class="text-gray-700 hover:text-indigo-600 font-semibold text-lg">Cocktails</a>
    <a href="#/recipes" class="text-gray-600 hover:text-indigo-600">All Recipes</a>
  `;
  if (isLoggedIn()) {
    if (isAdmin()) {
      const adminLink = document.createElement('a');
      adminLink.href = '#/admin/users';
      adminLink.className = 'text-gray-600 hover:text-indigo-600';
      adminLink.textContent = 'Admin';
      nav.appendChild(adminLink);
    }

    const createLink = document.createElement('a');
    createLink.href = '#/recipes/new';
    createLink.className = 'ml-auto bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700 text-sm';
    createLink.textContent = '+ New Recipe';
    nav.appendChild(createLink);

    const logoutBtn = document.createElement('button');
    logoutBtn.textContent = 'Sign Out';
    logoutBtn.className = 'text-sm text-gray-500 hover:text-gray-700';
    logoutBtn.addEventListener('click', () => {
      clearToken();
      navigate('#/');
    });
    nav.appendChild(logoutBtn);
  } else {
    const loginLink = document.createElement('a');
    loginLink.href = '#/login';
    loginLink.className = 'ml-auto text-sm text-indigo-600 hover:text-indigo-800';
    loginLink.textContent = 'Sign In';
    nav.appendChild(loginLink);
  }
  return nav;
}

if (document.getElementById('app')) {
  window.addEventListener('hashchange', renderPage);
  renderPage();
}
