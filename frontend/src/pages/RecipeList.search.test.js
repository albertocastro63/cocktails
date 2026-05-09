import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipes: vi.fn(),
}));

import { getRecipes } from '../api/client.js';
import { RecipeList } from './RecipeList.js';

describe('RecipeList search', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    getRecipes.mockResolvedValue({ data: [], total: 0, page: 1, limit: 20 });
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders a search input', async () => {
    const el = RecipeList();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('input')).not.toBeNull();
    });
  });

  it('calls getRecipes with q param when search changes', async () => {
    vi.useFakeTimers();
    const el = RecipeList();
    document.body.appendChild(el);

    await Promise.resolve();
    const input = document.body.querySelector('input');
    input.value = 'rum';
    input.dispatchEvent(new Event('input'));
    vi.advanceTimersByTime(350);

    expect(getRecipes).toHaveBeenCalledWith(expect.objectContaining({ q: 'rum' }));
    vi.useRealTimers();
  });
});
