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

  it('renders notes as markdown HTML when recipe has notes', async () => {
    getRecipe.mockResolvedValue({ ...fullRecipe, notes: '**bold notes**' });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('strong')).not.toBeNull();
    });
  });

  it('does not render a notes section when notes is empty', async () => {
    getRecipe.mockResolvedValue({ ...fullRecipe, notes: '' });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('h2')).not.toBeNull(); // some h2 exists
      const h2s = [...document.body.querySelectorAll('h2')].map(h => h.textContent);
      expect(h2s).not.toContain('Notes');
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

  it('recipe title h1 has text-stone-900 class', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const h1 = document.body.querySelector('h1');
      expect(h1).not.toBeNull();
      expect(h1.className).toContain('text-stone-900');
    });
  });

  it('notes container has prose prose-stone max-w-none overflow-x-auto classes for typography styling', async () => {
    getRecipe.mockResolvedValue({ ...fullRecipe, notes: '## Tips\n**Shake well.**' });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const strong = document.body.querySelector('strong');
      expect(strong).not.toBeNull();
      const notesContainer = strong.closest('div');
      expect(notesContainer.className).toContain('prose');
      expect(notesContainer.className).toContain('prose-stone');
      expect(notesContainer.className).toContain('max-w-none');
      expect(notesContainer.className).toContain('overflow-x-auto');
    });
  });
});
