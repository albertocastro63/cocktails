import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MyRecipes } from './MyRecipes.js';

vi.mock('../api/client.js', () => ({
  getMyRecipes: vi.fn(() => Promise.resolve({ data: [], total: 0, page: 1, limit: 20 })),
}));

vi.mock('../api/auth.js', () => ({
  getToken: vi.fn(() => 'test-token'),
  getUserID: vi.fn(() => 'u1'),
  isAdmin: vi.fn(() => false),
}));

beforeEach(() => {
  document.body.innerHTML = '';
  vi.clearAllMocks();
});

describe('MyRecipes', () => {
  it('renders a heading', () => {
    const el = MyRecipes();
    document.body.appendChild(el);
    const h1 = document.body.querySelector('h1');
    expect(h1).not.toBeNull();
    expect(h1.textContent).toMatch(/my recipes/i);
  });

  it('shows empty state message when no recipes', async () => {
    const { getMyRecipes } = await import('../api/client.js');
    getMyRecipes.mockResolvedValueOnce({ data: [], total: 0, page: 1, limit: 20 });
    const el = MyRecipes();
    document.body.appendChild(el);
    await Promise.resolve();
    await Promise.resolve();
    expect(document.body.textContent).toMatch(/no recipes|haven't added/i);
  });

  it('renders recipe cards when recipes exist', async () => {
    const { getMyRecipes } = await import('../api/client.js');
    getMyRecipes.mockResolvedValueOnce({
      data: [
        { id: 'r1', name: 'Mojito', creator_id: 'u1', ingredients: [] },
        { id: 'r2', name: 'Daiquiri', creator_id: 'u1', ingredients: [] },
      ],
      total: 2, page: 1, limit: 20,
    });
    const el = MyRecipes();
    document.body.appendChild(el);
    await Promise.resolve();
    await Promise.resolve();
    const links = document.body.querySelectorAll('a[href*="r1"], a[href*="r2"]');
    expect(links.length).toBeGreaterThan(0);
  });

  it('always passes currentUser to each RecipeCard (edit/delete always visible)', async () => {
    const { getMyRecipes } = await import('../api/client.js');
    getMyRecipes.mockResolvedValueOnce({
      data: [{ id: 'r1', name: 'Mojito', creator_id: 'u1', ingredients: [] }],
      total: 1, page: 1, limit: 20,
    });
    const el = MyRecipes();
    document.body.appendChild(el);
    await Promise.resolve();
    await Promise.resolve();
    const editLink = document.body.querySelector('a[href*="edit"]');
    expect(editLink).not.toBeNull();
  });

  it('shows error state when API fails', async () => {
    const { getMyRecipes } = await import('../api/client.js');
    getMyRecipes.mockRejectedValueOnce(new Error('Network error'));
    const el = MyRecipes();
    document.body.appendChild(el);
    await Promise.resolve();
    await Promise.resolve();
    expect(document.body.textContent).toMatch(/failed|error/i);
  });
});
