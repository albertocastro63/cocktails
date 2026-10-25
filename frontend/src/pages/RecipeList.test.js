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

describe('RecipeList sort controls', () => {
  const recipes = [
    { id: 'r1', name: 'Zombie', ingredients: [] },
    { id: 'r2', name: 'Aperol Spritz', ingredients: [] },
    { id: 'r3', name: 'Margarita', ingredients: [] },
  ];

  // Sorting is server-side: the API returns rows already ordered by the `sort`
  // param. This mock models that so the component only has to render what it
  // receives (and re-fetch when the sort changes).
  function mockSortedApi(rows = recipes) {
    getRecipes.mockImplementation((params = {}) => {
      let data = rows.slice();
      if (params.sort === 'name_asc') {
        data.sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }));
      } else if (params.sort === 'name_desc') {
        data.sort((a, b) => b.name.localeCompare(a.name, undefined, { sensitivity: 'base' }));
      }
      return Promise.resolve({ data, total: rows.length, page: params.page || 1, limit: 20 });
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders an "A→Z" and a "Z→A" button on load', async () => {
    mockSortedApi();
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      const buttons = document.body.querySelectorAll('button[data-dir]');
      expect(buttons).toHaveLength(2);
      expect(buttons[0].textContent).toBe('A→Z');
      expect(buttons[1].textContent).toBe('Z→A');
    });
  });

  it('neither sort button has aria-pressed="true" on initial render', async () => {
    mockSortedApi();
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.querySelectorAll('button[data-dir]')).toHaveLength(2);
    });
    document.body.querySelectorAll('button[data-dir]').forEach(btn => {
      expect(btn.getAttribute('aria-pressed')).toBe('false');
    });
  });

  it('clicking "A→Z" requests a name_asc sort and shows ascending order', async () => {
    mockSortedApi();
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.querySelectorAll('h2')).toHaveLength(3);
    });
    document.body.querySelector('button[data-dir="asc"]').click();
    await vi.waitFor(() => {
      const names = [...document.body.querySelectorAll('h2')].map(h => h.textContent.trim());
      expect(names).toEqual(['Aperol Spritz', 'Margarita', 'Zombie']);
    });
    expect(getRecipes).toHaveBeenCalledWith(expect.objectContaining({ sort: 'name_asc', page: 1 }));
  });

  it('clicking "Z→A" requests a name_desc sort and shows descending order', async () => {
    mockSortedApi();
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.querySelectorAll('h2')).toHaveLength(3);
    });
    document.body.querySelector('button[data-dir="desc"]').click();
    await vi.waitFor(() => {
      const names = [...document.body.querySelectorAll('h2')].map(h => h.textContent.trim());
      expect(names).toEqual(['Zombie', 'Margarita', 'Aperol Spritz']);
    });
    expect(getRecipes).toHaveBeenCalledWith(expect.objectContaining({ sort: 'name_desc' }));
  });

  it('switching from "A→Z" to "Z→A" updates the order', async () => {
    mockSortedApi();
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.querySelectorAll('h2')).toHaveLength(3);
    });
    document.body.querySelector('button[data-dir="asc"]').click();
    document.body.querySelector('button[data-dir="desc"]').click();
    await vi.waitFor(() => {
      const names = [...document.body.querySelectorAll('h2')].map(h => h.textContent.trim());
      expect(names).toEqual(['Zombie', 'Margarita', 'Aperol Spritz']);
    });
  });

  it('clicking "A→Z" on an empty list causes no error and keeps the empty state visible', async () => {
    getRecipes.mockResolvedValue({ data: [], total: 0, page: 1, limit: 20 });
    document.body.appendChild(RecipeList());
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/no recipe|empty|nothing|add one/);
    });
    expect(() => {
      document.body.querySelector('button[data-dir="asc"]').click();
    }).not.toThrow();
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/no recipe|empty|nothing|add one/);
    });
  });
});
