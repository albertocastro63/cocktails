import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipes: vi.fn(),
}));

import { getRecipes } from '../api/client.js';
import { RecipeList } from './RecipeList.js';

describe('RecipeList page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders loading state initially', () => {
    getRecipes.mockReturnValue(new Promise(() => {}));
    const el = RecipeList();
    document.body.appendChild(el);
    expect(document.body.textContent.toLowerCase()).toMatch(/loading/);
  });

  it('renders recipe cards when data loads', async () => {
    getRecipes.mockResolvedValue({
      data: [{ id: 'r1', name: 'Mojito', ingredients: [] }],
      total: 1, page: 1, limit: 20,
    });
    const el = RecipeList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });
  });

  it('renders EmptyState when data is empty', async () => {
    getRecipes.mockResolvedValue({ data: [], total: 0, page: 1, limit: 20 });
    const el = RecipeList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/no recipe|empty|nothing|add one/);
    });
  });
});
