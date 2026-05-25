import { describe, it, expect, vi, afterEach } from 'vitest';
import {
  getRecipes, getRandomRecipe, getMyRecipes, getRecipe,
  login, createRecipe, updateRecipe, deleteRecipe,
  listUsers, createUser, getUser, updateUser, deleteUser,
  downloadRecipeSchema, exportRecipes, importRecipes,
  getMyFavorites, getFavoriteStatus, favoriteRecipe, unfavoriteRecipe,
} from './client.js';

function mockFetch(status, body) {
  global.fetch = vi.fn().mockResolvedValue({
    status,
    ok: status >= 200 && status < 300,
    json: () => Promise.resolve(body),
    blob: () => Promise.resolve(new Blob([JSON.stringify(body)])),
  });
}

function mockFetchBlob(blob) {
  global.fetch = vi.fn().mockResolvedValue({ ok: true, blob: () => Promise.resolve(blob) });
}

function mockFetchError(status, errorBody) {
  global.fetch = vi.fn().mockResolvedValue({
    status,
    ok: false,
    json: () => Promise.resolve(errorBody),
  });
}

afterEach(() => vi.restoreAllMocks());

describe('request() — core behaviour', () => {
  it('returns parsed JSON for 200 responses', async () => {
    mockFetch(200, { data: [], total: 0 });
    const result = await getRecipes();
    expect(result).toEqual({ data: [], total: 0 });
  });

  it('returns null for 204 responses', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 204, ok: true });
    expect(await deleteRecipe('r1', 'tok')).toBeNull();
  });

  it('throws with status and code on error responses', async () => {
    mockFetchError(404, { error: { message: 'not found', code: 'NOT_FOUND' } });
    await expect(getRecipe('missing')).rejects.toMatchObject({
      message: 'not found',
      status: 404,
      code: 'NOT_FOUND',
    });
  });

  it('falls back to a generic message when error body cannot be parsed', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      status: 500,
      ok: false,
      json: () => Promise.reject(new Error('parse fail')),
    });
    await expect(getRecipe('x')).rejects.toMatchObject({ message: 'Request failed' });
  });

  it('attaches Authorization header when token is provided', async () => {
    mockFetch(200, { data: [] });
    await getMyRecipes('my-token');
    expect(global.fetch.mock.calls[0][1].headers['Authorization']).toBe('Bearer my-token');
  });

  it('omits Authorization header when no token', async () => {
    mockFetch(200, { data: [] });
    await getRecipes();
    expect(global.fetch.mock.calls[0][1].headers['Authorization']).toBeUndefined();
  });

  it('omits body when body argument is undefined', async () => {
    mockFetch(200, { data: [] });
    await listUsers('tok');
    expect(global.fetch.mock.calls[0][1].body).toBeUndefined();
  });

  it('serialises body to JSON when provided', async () => {
    mockFetch(200, { token: 'jwt' });
    await login('alice', 'secret');
    expect(global.fetch.mock.calls[0][1].body).toBe(JSON.stringify({ username: 'alice', password: 'secret' }));
  });
});

describe('recipe endpoints', () => {
  it('getRecipes calls GET /api/v1/recipes', async () => {
    mockFetch(200, { data: [] });
    await getRecipes();
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/recipes');
    expect(global.fetch.mock.calls[0][1].method).toBe('GET');
  });

  it('getRecipes appends query params', async () => {
    mockFetch(200, { data: [] });
    await getRecipes({ q: 'mojito', page: 2 });
    const url = global.fetch.mock.calls[0][0];
    expect(url).toContain('q=mojito');
    expect(url).toContain('page=2');
  });

  it('getRecipes with no params omits query string', async () => {
    mockFetch(200, { data: [] });
    await getRecipes({});
    expect(global.fetch.mock.calls[0][0]).not.toContain('?');
  });

  it('getRandomRecipe calls /api/v1/recipes/random', async () => {
    mockFetch(200, { data: {} });
    await getRandomRecipe();
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/recipes/random');
  });

  it('getMyRecipes appends query params', async () => {
    mockFetch(200, { data: [] });
    await getMyRecipes('tok', { page: 1 });
    expect(global.fetch.mock.calls[0][0]).toContain('page=1');
  });

  it('getRecipe calls GET /api/v1/recipes/:id', async () => {
    mockFetch(200, { data: {} });
    await getRecipe('r42');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/recipes/r42');
  });

  it('createRecipe posts with auth', async () => {
    mockFetch(200, { data: { id: 'r1' } });
    await createRecipe({ name: 'Mojito' }, 'tok');
    const opts = global.fetch.mock.calls[0][1];
    expect(opts.method).toBe('POST');
    expect(opts.headers['Authorization']).toBe('Bearer tok');
  });

  it('updateRecipe uses PUT', async () => {
    mockFetch(200, { data: {} });
    await updateRecipe('r1', { name: 'Mojito' }, 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('PUT');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/recipes/r1');
  });

  it('deleteRecipe uses DELETE', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 204, ok: true });
    await deleteRecipe('r1', 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('DELETE');
  });
});

describe('user admin endpoints', () => {
  it('listUsers calls /api/v1/admin/users', async () => {
    mockFetch(200, { data: [] });
    await listUsers('tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/admin/users');
  });

  it('createUser posts to admin users', async () => {
    mockFetch(200, { data: { id: 'u1' } });
    await createUser({ username: 'bob' }, 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('POST');
  });

  it('getUser calls GET /admin/users/:id', async () => {
    mockFetch(200, { data: {} });
    await getUser('u99', 'tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/admin/users/u99');
  });

  it('updateUser uses PUT', async () => {
    mockFetch(200, { data: {} });
    await updateUser('u1', { username: 'bob' }, 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('PUT');
  });

  it('deleteUser uses DELETE', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 204, ok: true });
    await deleteUser('u1', 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('DELETE');
  });
});

describe('fetchBlob() via admin export endpoints', () => {
  it('downloadRecipeSchema fetches from /api/v1/admin/schema', async () => {
    const blob = new Blob(['{}']);
    mockFetchBlob(blob);
    const result = await downloadRecipeSchema('tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/admin/schema');
    expect(result).toBe(blob);
  });

  it('downloadRecipeSchema includes Authorization header', async () => {
    mockFetchBlob(new Blob(['{}']));
    await downloadRecipeSchema('my-tok');
    expect(global.fetch.mock.calls[0][1].headers['Authorization']).toBe('Bearer my-tok');
  });

  it('exportRecipes fetches from /api/v1/admin/recipes/export', async () => {
    const blob = new Blob(['[]']);
    mockFetchBlob(blob);
    const result = await exportRecipes('tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/admin/recipes/export');
    expect(result).toBe(blob);
  });

  it('fetchBlob throws on non-ok response', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 403,
      json: () => Promise.resolve({ error: { message: 'Forbidden', code: 'FORBIDDEN' } }),
    });
    await expect(downloadRecipeSchema('tok')).rejects.toMatchObject({
      message: 'Forbidden',
      status: 403,
      code: 'FORBIDDEN',
    });
  });

  it('fetchBlob falls back to generic message when error body cannot be parsed', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.reject(new Error('parse fail')),
    });
    await expect(exportRecipes('tok')).rejects.toMatchObject({ message: 'Request failed' });
  });

  it('importRecipes posts array to /api/v1/admin/recipes/import', async () => {
    mockFetch(200, { imported: 2, skipped: 1 });
    const result = await importRecipes([{ name: 'Mojito' }], 'tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/admin/recipes/import');
    expect(result).toEqual({ imported: 2, skipped: 1 });
  });
});

describe('favorite endpoints', () => {
  it('getMyFavorites calls /api/v1/recipes/favorites', async () => {
    mockFetch(200, { data: [] });
    await getMyFavorites('tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/recipes/favorites');
  });

  it('getMyFavorites appends query params', async () => {
    mockFetch(200, { data: [] });
    await getMyFavorites('tok', { page: 2 });
    expect(global.fetch.mock.calls[0][0]).toContain('page=2');
  });

  it('getFavoriteStatus calls GET /recipes/:id/favorite', async () => {
    mockFetch(200, { is_favorite: true });
    await getFavoriteStatus('r1', 'tok');
    expect(global.fetch.mock.calls[0][0]).toContain('/api/v1/recipes/r1/favorite');
    expect(global.fetch.mock.calls[0][1].method).toBe('GET');
  });

  it('favoriteRecipe calls PUT /recipes/:id/favorite', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 204, ok: true });
    await favoriteRecipe('r1', 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('PUT');
  });

  it('unfavoriteRecipe calls DELETE /recipes/:id/favorite', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 204, ok: true });
    await unfavoriteRecipe('r1', 'tok');
    expect(global.fetch.mock.calls[0][1].method).toBe('DELETE');
  });
});
