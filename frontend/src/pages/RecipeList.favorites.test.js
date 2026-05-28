import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipes: vi.fn(),
  getMyFavorites: vi.fn(),
}));

vi.mock('../api/auth.js', () => ({
  getUserID: vi.fn(),
  isAdmin: vi.fn(),
  getToken: vi.fn(),
}));

import { getRecipes, getMyFavorites } from '../api/client.js';
import { getUserID, isAdmin, getToken } from '../api/auth.js';
import { RecipeList } from './RecipeList.js';

const recipes = [
  { id: 'r1', name: 'Mojito', ingredients: [] },
  { id: 'r2', name: 'Negroni', ingredients: [] },
];

describe('RecipeList - favorite heart indicators', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    getRecipes.mockResolvedValue({ data: recipes, total: 2, page: 1, limit: 20 });
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('shows heart marker on favorited recipe when user is logged in', async () => {
    getUserID.mockReturnValue('user-1');
    isAdmin.mockReturnValue(false);
    getToken.mockReturnValue('tok');
    getMyFavorites.mockResolvedValue({ data: [recipes[0]] }); // r1 is favorited

    const el = RecipeList();
    document.body.appendChild(el);

    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });

    const hearts = document.body.querySelectorAll('[data-favorite="true"]');
    expect(hearts.length).toBe(1);
  });

  it('shows no heart markers when user has no favorites', async () => {
    getUserID.mockReturnValue('user-1');
    isAdmin.mockReturnValue(false);
    getToken.mockReturnValue('tok');
    getMyFavorites.mockResolvedValue({ data: [] });

    const el = RecipeList();
    document.body.appendChild(el);

    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });

    const hearts = document.body.querySelectorAll('[data-favorite="true"]');
    expect(hearts.length).toBe(0);
  });

  it('shows no heart markers when user is not logged in', async () => {
    getUserID.mockReturnValue(null);
    isAdmin.mockReturnValue(false);
    getToken.mockReturnValue(null);

    const el = RecipeList();
    document.body.appendChild(el);

    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });

    expect(getMyFavorites).not.toHaveBeenCalled();
    const hearts = document.body.querySelectorAll('[data-favorite="true"]');
    expect(hearts.length).toBe(0);
  });

  it('still loads all recipes when favorites fetch fails', async () => {
    getUserID.mockReturnValue('user-1');
    isAdmin.mockReturnValue(false);
    getToken.mockReturnValue('tok');
    getMyFavorites.mockRejectedValue(new Error('network error'));

    const el = RecipeList();
    document.body.appendChild(el);

    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });

    // recipes still show, just no hearts
    expect(document.body.textContent).toContain('Negroni');
    const hearts = document.body.querySelectorAll('[data-favorite="true"]');
    expect(hearts.length).toBe(0);
  });
});
