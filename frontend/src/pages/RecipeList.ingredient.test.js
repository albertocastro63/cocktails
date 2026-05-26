import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipes: vi.fn(),
}));

import { getRecipes } from '../api/client.js';
import { RecipeList } from './RecipeList.js';

describe('RecipeList ingredient search hint', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    getRecipes.mockResolvedValue({ data: [], total: 0, page: 1, limit: 20 });
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders a hint text about multi-ingredient search syntax', () => {
    const el = RecipeList();
    document.body.appendChild(el);
    const paragraphs = Array.from(document.body.querySelectorAll('p'));
    const hint = paragraphs.find(p => /ingredient.*use.*and.*\+/i.test(p.textContent));
    expect(hint).toBeDefined();
  });
});
