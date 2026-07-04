const BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

function apiPrefix() {
  return import.meta.env.VITE_API_PATH_PREFIX || '/api';
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

export function login(username, password) {
  return request('POST', `${apiPrefix()}/v1/auth/login`, { username, password });
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
