import './index.css';
import { Home } from './pages/Home.js';
import { RecipeList } from './pages/RecipeList.js';
import { RecipeDetail } from './pages/RecipeDetail.js';
import { Login } from './pages/Login.js';
import { RecipeForm } from './pages/RecipeForm.js';
import { AdminUserList } from './pages/AdminUserList.js';
import { AdminUserForm } from './pages/AdminUserForm.js';
import { AdminRecipes } from './pages/AdminRecipes.js';
import { MyRecipes } from './pages/MyRecipes.js';
import { isLoggedIn, isAdmin, clearToken } from './api/auth.js';

const routes = [
  { pattern: /^\/admin\/users\/([^/]+)\/edit$/, factory: (m) => AdminUserForm({ id: m[1], onSave: () => navigate('#/admin/users') }) },
  { pattern: /^\/admin\/users\/new$/, factory: () => AdminUserForm({ onSave: () => navigate('#/admin/users') }) },
  { pattern: /^\/admin\/users$/, factory: () => AdminUserList() },
  { pattern: /^\/admin\/recipes$/, factory: () => AdminRecipes() },
  { pattern: /^\/recipes\/([^/]+)\/edit$/, factory: (m) => RecipeForm({ id: m[1], onSave: () => navigate(`#/recipes/${m[1]}`) }) },
  { pattern: /^\/recipes\/new$/, factory: () => RecipeForm({ onSave: (r) => navigate(`#/recipes/${r?.data?.id || ''}`) }) },
  { pattern: /^\/my-recipes$/, factory: () => MyRecipes() },
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

  // Auth guard for write routes and my-recipes
  const writeRoutes = /^\/(recipes\/new|recipes\/.+\/edit|my-recipes)/;
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
    el.className = 'text-center py-16 text-stone-400';
    el.textContent = 'Page not found.';
    root.appendChild(el);
  }
}

export function buildNav() {
  const nav = document.createElement('nav');
  nav.className = 'bg-stone-900 px-6 py-3 flex items-center gap-6';
  nav.innerHTML = `
    <a href="#/" class="text-stone-100 hover:text-amber-400 font-semibold text-lg">Cocktails</a>
    <a href="#/recipes" class="text-stone-100 hover:text-amber-400">All Recipes</a>
  `;
  if (isLoggedIn()) {
    if (isAdmin()) {
      const usersLink = document.createElement('a');
      usersLink.href = '#/admin/users';
      usersLink.className = 'text-stone-100 hover:text-amber-400';
      usersLink.textContent = 'Users';
      nav.appendChild(usersLink);

      const recipesLink = document.createElement('a');
      recipesLink.href = '#/admin/recipes';
      recipesLink.className = 'text-stone-100 hover:text-amber-400';
      recipesLink.textContent = 'Recipes';
      nav.appendChild(recipesLink);
    }

    const myRecipesLink = document.createElement('a');
    myRecipesLink.href = '#/my-recipes';
    myRecipesLink.className = 'text-stone-100 hover:text-amber-400';
    myRecipesLink.textContent = 'My Recipes';
    nav.appendChild(myRecipesLink);

    const createLink = document.createElement('a');
    createLink.href = '#/recipes/new';
    createLink.className = 'ml-auto bg-amber-500 text-stone-900 font-semibold px-3 py-1 rounded-xl hover:bg-amber-600 text-sm';
    createLink.textContent = '+ New Recipe';
    nav.appendChild(createLink);

    const logoutBtn = document.createElement('button');
    logoutBtn.textContent = 'Sign Out';
    logoutBtn.className = 'text-sm text-stone-400 hover:text-stone-200';
    logoutBtn.addEventListener('click', () => {
      clearToken();
      navigate('#/');
    });
    nav.appendChild(logoutBtn);
  } else {
    const loginLink = document.createElement('a');
    loginLink.href = '#/login';
    loginLink.className = 'ml-auto text-sm text-amber-400 hover:text-amber-300';
    loginLink.textContent = 'Sign In';
    nav.appendChild(loginLink);
  }
  return nav;
}

if (document.getElementById('app')) {
  window.addEventListener('hashchange', renderPage);
  renderPage();
}
