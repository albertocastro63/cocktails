import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipe: vi.fn(),
}));

import { getRecipe } from '../api/client.js';
import { RecipeDetail } from './RecipeDetail.js';

const fullRecipe = {
  id: 'r1',
  name: 'Mojito',
  ingredients: [{ name: 'rum', quantity: '50', unit: 'ml' }],
  steps: ['Muddle mint', 'Add rum', 'Top with soda'],
  properties: { style: 'refreshing', base_spirit: 'rum' },
  creator_id: 'u1',
};

describe('RecipeDetail page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders loading state initially', () => {
    getRecipe.mockReturnValue(new Promise(() => {}));
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    expect(document.body.textContent.toLowerCase()).toMatch(/loading/);
  });

  it('renders recipe name and ingredients', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
      expect(document.body.textContent).toContain('rum');
    });
  });

  it('renders steps', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Muddle mint');
    });
  });

  it('renders all property pairs', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('style');
      expect(document.body.textContent).toContain('refreshing');
    });
  });

  it('renders error state on failure', async () => {
    getRecipe.mockRejectedValue(new Error('not found'));
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/error|not found|failed/);
    });
  });
});
