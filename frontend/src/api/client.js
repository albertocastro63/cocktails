import { clearToken } from './auth.js';

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

function apiPrefix() {
  return import.meta.env.VITE_API_PATH_PREFIX || '/api';
}

// handleSessionExpired drops the stale token and sends the user to sign in.
// Called when an authenticated request comes back 401 — the token has expired
// or been invalidated, so the UI must not keep pretending the session is live.
function handleSessionExpired() {
  clearToken();
  if (typeof window !== 'undefined' && !window.location.hash.startsWith('#/login')) {
    window.location.hash = '#/login';
  }
}

async function request(method, path, body, token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(`${BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (res.status === 204) return null;
  if (!res.ok) {
    // A 401 on a request that carried a token means the session is no longer
    // valid (expired or invalidated). Log out so a stale token can't linger and
    // surface confusing per-page errors. Unauthenticated 401s (e.g. bad login
    // credentials) carry no token and are left for the caller to handle.
    if (res.status === 401 && token) {
      handleSessionExpired();
    }
    const err = await res.json().catch(() => ({ error: { message: 'Request failed' } }));
    const e = new Error(err.error?.message || 'Request failed');
    e.status = res.status;
    e.code = err.error?.code;
    throw e;
  }
  return res.json();
}

export function getRecipes(params = {}) {
  const qs = new URLSearchParams(params).toString();
  return request('GET', `${apiPrefix()}/v1/recipes${qs ? '?' + qs : ''}`);
}

export function getRandomRecipe() {
  return request('GET', `${apiPrefix()}/v1/recipes/random`);
}

export function getMyRecipes(token, params = {}) {
  const qs = new URLSearchParams(params).toString();
  return request('GET', `${apiPrefix()}/v1/recipes/mine${qs ? '?' + qs : ''}`, undefined, token);
}

export function getRecipe(id) {
  return request('GET', `${apiPrefix()}/v1/recipes/${id}`);
}

// getRecipeNames returns a lightweight [{id, name}] list for the related-cocktails
// type-ahead (filtered client-side).
export function getRecipeNames() {
  return request('GET', `${apiPrefix()}/v1/recipes/names`);
}

export function login(username, password) {
  return request('POST', `${apiPrefix()}/v1/auth/login`, { username, password });
}

export function requestPasswordReset(email) {
  return request('POST', `${apiPrefix()}/v1/auth/forgot-password`, { email });
}

export function resetPassword(uid, token, password) {
  return request('POST', `${apiPrefix()}/v1/auth/reset-password`, { uid, token, password });
}

export function createRecipe(recipe, token) {
  return request('POST', `${apiPrefix()}/v1/recipes`, recipe, token);
}

export function updateRecipe(id, recipe, token) {
  return request('PUT', `${apiPrefix()}/v1/recipes/${id}`, recipe, token);
}

export function deleteRecipe(id, token) {
  return request('DELETE', `${apiPrefix()}/v1/recipes/${id}`, null, token);
}

export function listUsers(token) {
  return request('GET', `${apiPrefix()}/v1/admin/users`, undefined, token);
}

export function createUser(data, token) {
  return request('POST', `${apiPrefix()}/v1/admin/users`, data, token);
}

export function getUser(id, token) {
  return request('GET', `${apiPrefix()}/v1/admin/users/${id}`, undefined, token);
}

export function updateUser(id, data, token) {
  return request('PUT', `${apiPrefix()}/v1/admin/users/${id}`, data, token);
}

export function deleteUser(id, token) {
  return request('DELETE', `${apiPrefix()}/v1/admin/users/${id}`, undefined, token);
}

async function fetchBlob(path, token) {
  const headers = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(`${BASE_URL}${path}`, { headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: { message: 'Request failed' } }));
    const e = new Error(err.error?.message || 'Request failed');
    e.status = res.status;
    e.code = err.error?.code;
    throw e;
  }
  return res.blob();
}

export function downloadRecipeSchema(token) {
  return fetchBlob(`${apiPrefix()}/v1/admin/schema`, token);
}

export function exportRecipes(token) {
  return fetchBlob(`${apiPrefix()}/v1/admin/recipes/export`, token);
}

export function importRecipes(recipes, token) {
  return request('POST', `${apiPrefix()}/v1/admin/recipes/import`, recipes, token);
}

export function getMyFavorites(token, params = {}) {
  const qs = new URLSearchParams(params).toString();
  return request('GET', `${apiPrefix()}/v1/recipes/favorites${qs ? '?' + qs : ''}`, undefined, token);
}

export function getFavoriteStatus(id, token) {
  return request('GET', `${apiPrefix()}/v1/recipes/${id}/favorite`, undefined, token);
}

export function favoriteRecipe(id, token) {
  return request('PUT', `${apiPrefix()}/v1/recipes/${id}/favorite`, null, token);
}

export function unfavoriteRecipe(id, token) {
  return request('DELETE', `${apiPrefix()}/v1/recipes/${id}/favorite`, null, token);
}
