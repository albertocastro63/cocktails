import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipes: vi.fn(),
  getMyFavorites: vi.fn(),
}));

import { getRecipes } from '../api/client.js';
import { RecipeList } from './RecipeList.js';

// makeRecipes builds n recipes named "Recipe 00".."Recipe nn".
function makeRecipes(n) {
  return Array.from({ length: n }, (_, i) => ({
    id: `r${String(i).padStart(2, '0')}`,
    name: `Recipe ${String(i).padStart(2, '0')}`,
    ingredients: [],
  }));
}

// mockPaginatedApi models the server: it slices the full catalog by page/limit
// and reports the true total, so the component drives paging from the response.
function mockPaginatedApi(all) {
  getRecipes.mockImplementation((params = {}) => {
    const page = params.page || 1;
    const limit = params.limit || 20;
    const start = (page - 1) * limit;
    return Promise.resolve({
      data: all.slice(start, start + limit),
      total: all.length,
      page,
      limit,
    });
  });
}

function cardNames() {
  return [...document.body.querySelectorAll('h2')].map((h) => h.textContent.trim());
}

describe('RecipeList pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('requests page 1 on initial load', async () => {
    mockPaginatedApi(makeRecipes(30));
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.querySelectorAll('h2').length).toBe(20);
    });
    expect(getRecipes).toHaveBeenCalledWith(expect.objectContaining({ page: 1, limit: 20 }));
  });

  it('shows the total recipe count from the response', async () => {
    mockPaginatedApi(makeRecipes(30));
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.querySelector('[data-testid="total-count"]').textContent).toBe('30 recipes');
    });
  });

  it('clicking Next fetches and renders the next page', async () => {
    mockPaginatedApi(makeRecipes(30));
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => expect(document.body.querySelectorAll('h2').length).toBe(20));

    document.body.querySelector('[data-dir="next"]').click();
    await vi.waitFor(() => {
      expect(cardNames()).toContain('Recipe 20');
      expect(document.body.querySelectorAll('h2').length).toBe(10);
    });
    expect(getRecipes).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }));
    expect(document.body.querySelector('[data-testid="page-indicator"]').textContent).toBe('Page 2 of 2');
  });

  it('clicking Previous goes back a page', async () => {
    mockPaginatedApi(makeRecipes(30));
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => expect(document.body.querySelectorAll('h2').length).toBe(20));
    document.body.querySelector('[data-dir="next"]').click();
    await vi.waitFor(() => expect(document.body.querySelector('[data-dir="prev"]').disabled).toBe(false));
    document.body.querySelector('[data-dir="prev"]').click();
    await vi.waitFor(() => {
      expect(cardNames()).toContain('Recipe 00');
      expect(document.body.querySelector('[data-testid="page-indicator"]').textContent).toBe('Page 1 of 2');
    });
  });

  it('hides pagination controls when everything fits on one page', async () => {
    mockPaginatedApi(makeRecipes(5));
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => expect(document.body.querySelectorAll('h2').length).toBe(5));
    expect(document.body.querySelector('[data-dir="next"]')).toBeNull();
    expect(document.body.querySelector('[data-dir="prev"]')).toBeNull();
  });

  it('shows the empty state and no controls when the catalog is empty', async () => {
    mockPaginatedApi([]);
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/no recipe|empty|nothing|add one/);
    });
    expect(document.body.querySelector('[data-dir="next"]')).toBeNull();
    expect(document.body.querySelector('[data-testid="total-count"]').textContent).toBe('');
  });

  it('resets to page 1 when the search query changes', async () => {
    vi.useFakeTimers();
    const all = makeRecipes(30);
    mockPaginatedApi(all);
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => expect(document.body.querySelectorAll('h2').length).toBe(20));

    // Go to page 2 first.
    document.body.querySelector('[data-dir="next"]').click();
    await vi.waitFor(() => expect(getRecipes).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 })));

    // Now search — must reset to page 1.
    const input = document.body.querySelector('input');
    input.value = 'Recipe 0';
    input.dispatchEvent(new Event('input'));
    vi.advanceTimersByTime(350);
    vi.useRealTimers();

    await vi.waitFor(() => {
      expect(getRecipes).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, q: 'Recipe 0' }));
    });
  });

  it('resets to page 1 when the sort changes', async () => {
    mockPaginatedApi(makeRecipes(30));
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => expect(document.body.querySelectorAll('h2').length).toBe(20));
    document.body.querySelector('[data-dir="next"]').click();
    await vi.waitFor(() => expect(getRecipes).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 })));

    document.body.querySelector('button[data-dir="asc"]').click();
    await vi.waitFor(() => {
      expect(getRecipes).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, sort: 'name_asc' }));
    });
  });
});
