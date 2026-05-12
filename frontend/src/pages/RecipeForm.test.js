import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  createRecipe: vi.fn(),
  updateRecipe: vi.fn(),
  getRecipe: vi.fn(),
}));
vi.mock('../api/auth.js', () => ({
  getToken: vi.fn(() => 'tok'),
  isLoggedIn: vi.fn(() => true),
}));

import { createRecipe } from '../api/client.js';
import { RecipeForm } from './RecipeForm.js';

describe('RecipeForm page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders a name input', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    expect(document.body.querySelector('input[name="name"]')).not.toBeNull();
  });

  it('renders a button to add ingredients', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const btns = [...document.body.querySelectorAll('button')];
    const found = btns.some(b => b.textContent.toLowerCase().includes('ingredient'));
    expect(found).toBe(true);
  });

  it('renders a button to add steps', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const btns = [...document.body.querySelectorAll('button')];
    const found = btns.some(b => b.textContent.toLowerCase().includes('step'));
    expect(found).toBe(true);
  });

  it('renders a button to add properties', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const btns = [...document.body.querySelectorAll('button')];
    const found = btns.some(b => b.textContent.toLowerCase().includes('propert'));
    expect(found).toBe(true);
  });

  it('renders a textarea for notes with a preview toggle button', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    expect(document.body.querySelector('textarea[name="notes"]')).not.toBeNull();
    const btns = [...document.body.querySelectorAll('button')];
    const previewBtn = btns.find(b => b.textContent === 'Preview');
    expect(previewBtn).not.toBeUndefined();
  });

  it('submitting the form while notes editor is in preview mode sends raw markdown in the payload', async () => {
    createRecipe.mockResolvedValue({ data: { id: 'r1', name: 'New' }, warnings: [] });
    const el = RecipeForm({});
    document.body.appendChild(el);

    document.body.querySelector('input[name="name"]').value = 'My Cocktail';
    const textarea = document.body.querySelector('textarea[name="notes"]');
    textarea.value = '**raw markdown**';

    const previewBtn = [...document.body.querySelectorAll('button')].find(b => b.textContent === 'Preview');
    previewBtn.click();

    document.body.querySelector('form').dispatchEvent(new Event('submit'));

    await vi.waitFor(() => {
      expect(createRecipe).toHaveBeenCalledWith(
        expect.objectContaining({ notes: '**raw markdown**' }),
        expect.anything(),
      );
    });
  });

  it('calls createRecipe on submit for new recipe', async () => {
    createRecipe.mockResolvedValue({ data: { id: 'r1', name: 'New' }, warnings: [] });
    const onSave = vi.fn();
    const el = RecipeForm({ onSave });
    document.body.appendChild(el);

    const nameInput = document.body.querySelector('input[name="name"]');
    nameInput.value = 'My Cocktail';
    const form = document.body.querySelector('form');
    form.dispatchEvent(new Event('submit'));

    await vi.waitFor(() => {
      expect(createRecipe).toHaveBeenCalled();
    });
  });
});
