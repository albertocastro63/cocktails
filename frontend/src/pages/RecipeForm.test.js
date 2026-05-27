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

import { createRecipe, updateRecipe, getRecipe } from '../api/client.js';
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

  it('submit button has bg-amber-500 class', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const submitBtn = document.body.querySelector('button[type="submit"]');
    expect(submitBtn).not.toBeNull();
    expect(submitBtn.className).toContain('bg-amber-500');
  });

  it('shows error message when name is empty on submit', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    const err = document.body.querySelector('p.text-red-600');
    expect(err.classList.contains('hidden')).toBe(false);
    expect(err.textContent).toBe('Name is required.');
  });

  it('shows error message when createRecipe throws', async () => {
    createRecipe.mockRejectedValue(new Error('Server error'));
    const el = RecipeForm({});
    document.body.appendChild(el);
    document.body.querySelector('[name="name"]').value = 'Mojito';
    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    await vi.waitFor(() => {
      const err = document.body.querySelector('p.text-red-600');
      expect(err.classList.contains('hidden')).toBe(false);
      expect(err.textContent).toBe('Server error');
    });
  });

  it('calls updateRecipe when an id is provided', async () => {
    getRecipe.mockResolvedValue({ name: 'Existing', ingredients: [], steps: [], properties: {}, notes: '' });
    updateRecipe.mockResolvedValue({ data: { id: 'r1' } });
    const el = RecipeForm({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('[name="name"]').value).toBe('Existing');
    });
    document.body.querySelector('[name="name"]').value = 'Updated';
    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    await vi.waitFor(() => expect(updateRecipe).toHaveBeenCalledWith('r1', expect.any(Object), 'tok'));
  });

  it('includes steps and properties in the submit payload', async () => {
    createRecipe.mockResolvedValue({ data: { id: 'r1' } });
    const el = RecipeForm({});
    document.body.appendChild(el);
    document.body.querySelector('[name="name"]').value = 'Mojito';

    const addStep = [...document.body.querySelectorAll('button')].find(b => b.textContent.includes('Add Step'));
    addStep.click();
    document.body.querySelector('[name="step"]').value = 'Shake it';

    const addProp = [...document.body.querySelectorAll('button')].find(b => b.textContent.includes('Add Property'));
    addProp.click();
    document.body.querySelector('[name="prop_key"]').value = 'style';
    document.body.querySelector('[name="prop_val"]').value = 'tropical';

    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    await vi.waitFor(() => expect(createRecipe).toHaveBeenCalled());

    const payload = createRecipe.mock.calls[0][0];
    expect(payload.steps).toEqual(['Shake it']);
    expect(payload.properties).toEqual({ style: 'tropical' });
  });

  it('removing a step row removes it from the submit payload', async () => {
    createRecipe.mockResolvedValue({ data: { id: 'r1' } });
    const el = RecipeForm({});
    document.body.appendChild(el);
    document.body.querySelector('[name="name"]').value = 'Mojito';

    const addStep = [...document.body.querySelectorAll('button')].find(b => b.textContent.includes('Add Step'));
    addStep.click();
    const stepRow = document.body.querySelector('[name="step"]').closest('div');
    stepRow.querySelector('button').click();

    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    await vi.waitFor(() => expect(createRecipe).toHaveBeenCalled());
    expect(createRecipe.mock.calls[0][0].steps).toEqual([]);
  });
});

describe('RecipeForm — garnishes section (T010)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders a button to add garnishes', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const btns = [...document.body.querySelectorAll('button')];
    const found = btns.some(b => b.textContent.toLowerCase().includes('garnish'));
    expect(found).toBe(true);
  });

  it('clicking Add Garnish creates a garnish input row', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const addBtn = [...document.body.querySelectorAll('button')].find(b =>
      b.textContent.toLowerCase().includes('add garnish')
    );
    addBtn.click();
    expect(document.body.querySelector('[name="garnish"]')).not.toBeNull();
  });

  it('clicking remove on a garnish row removes the input', () => {
    const el = RecipeForm({});
    document.body.appendChild(el);
    const addBtn = [...document.body.querySelectorAll('button')].find(b =>
      b.textContent.toLowerCase().includes('add garnish')
    );
    addBtn.click();
    const row = document.body.querySelector('[name="garnish"]').closest('div');
    row.querySelector('button').click();
    expect(document.body.querySelector('[name="garnish"]')).toBeNull();
  });

  it('blank garnish entries are excluded from the submit payload', async () => {
    createRecipe.mockResolvedValue({ data: { id: 'r1', name: 'Sour' }, warnings: [] });
    const el = RecipeForm({});
    document.body.appendChild(el);
    document.body.querySelector('[name="name"]').value = 'Sour';

    const addBtn = [...document.body.querySelectorAll('button')].find(b =>
      b.textContent.toLowerCase().includes('add garnish')
    );
    addBtn.click();
    addBtn.click();
    const inputs = [...document.body.querySelectorAll('[name="garnish"]')];
    inputs[0].value = 'Express orange oil';
    inputs[1].value = '   ';

    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    await vi.waitFor(() => expect(createRecipe).toHaveBeenCalled());
    const payload = createRecipe.mock.calls[0][0];
    expect(payload.garnishes).toEqual(['Express orange oil']);
  });

  it('saved garnishes are pre-populated in edit mode', async () => {
    getRecipe.mockResolvedValue({
      name: 'Old Fashioned',
      ingredients: [],
      steps: [],
      properties: {},
      notes: '',
      garnishes: ['Express orange oil over the cocktail', 'Use orange peel to garnish'],
    });
    const el = RecipeForm({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const inputs = [...document.body.querySelectorAll('[name="garnish"]')];
      expect(inputs.length).toBe(2);
      expect(inputs[0].value).toBe('Express orange oil over the cocktail');
      expect(inputs[1].value).toBe('Use orange peel to garnish');
    });
  });
});

describe('RecipeForm — base spirit toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });
  afterEach(() => { document.body.innerHTML = ''; });

  function clickAddIngredient() {
    const btn = [...document.body.querySelectorAll('button')].find(b =>
      b.textContent.toLowerCase().includes('add ingredient')
    );
    btn.click();
  }

  it('newly added ingredient row contains a base spirit checkbox', () => {
    document.body.appendChild(RecipeForm({}));
    clickAddIngredient();
    expect(document.body.querySelector('[name="ing_base_spirit"]')).not.toBeNull();
  });

  it('checking base spirit on row B clears row A', () => {
    document.body.appendChild(RecipeForm({}));
    clickAddIngredient();
    clickAddIngredient();
    const cbs = [...document.body.querySelectorAll('[name="ing_base_spirit"]')];
    expect(cbs.length).toBe(2);
    cbs[0].click();
    expect(cbs[0].checked).toBe(true);
    cbs[1].click();
    expect(cbs[1].checked).toBe(true);
    expect(cbs[0].checked).toBe(false);
  });

  it('unchecking the active checkbox leaves all rows unchecked', () => {
    document.body.appendChild(RecipeForm({}));
    clickAddIngredient();
    const cb = document.body.querySelector('[name="ing_base_spirit"]');
    cb.click();
    expect(cb.checked).toBe(true);
    cb.click();
    expect(cb.checked).toBe(false);
  });

  it('deleting the base spirit row does not activate any remaining row (FR-009)', () => {
    document.body.appendChild(RecipeForm({}));
    clickAddIngredient();
    clickAddIngredient();
    const cbs = () => [...document.body.querySelectorAll('[name="ing_base_spirit"]')];
    cbs()[0].click();
    expect(cbs()[0].checked).toBe(true);
    const removeBtn = cbs()[0].closest('div').querySelector('button');
    removeBtn.click();
    expect(cbs().length).toBe(1);
    expect(cbs()[0].checked).toBe(false);
  });

  it('edit prefill restores is_base_spirit on the correct row only', async () => {
    getRecipe.mockResolvedValue({
      name: 'Manhattan',
      ingredients: [
        { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
        { name: 'Vermouth', quantity: '30', unit: 'ml' },
      ],
      steps: [],
      properties: {},
      notes: '',
    });
    document.body.appendChild(RecipeForm({ id: 'r1' }));
    await vi.waitFor(() => {
      const cbs = [...document.body.querySelectorAll('[name="ing_base_spirit"]')];
      expect(cbs.length).toBe(2);
      expect(cbs[0].checked).toBe(true);
      expect(cbs[1].checked).toBe(false);
    });
  });

  it('submit payload includes is_base_spirit:true only on the checked ingredient', async () => {
    createRecipe.mockResolvedValue({ data: { id: 'r1', name: 'New' }, warnings: [] });
    document.body.appendChild(RecipeForm({}));
    document.body.querySelector('[name="name"]').value = 'Manhattan';
    clickAddIngredient();
    clickAddIngredient();
    const nameInputs = [...document.body.querySelectorAll('[name="ing_name"]')];
    nameInputs[0].value = 'Rye';
    nameInputs[1].value = 'Vermouth';
    document.body.querySelectorAll('[name="ing_base_spirit"]')[0].click();
    document.body.querySelector('form').dispatchEvent(new Event('submit'));
    await vi.waitFor(() => expect(createRecipe).toHaveBeenCalled());
    const payload = createRecipe.mock.calls[0][0];
    const rye = payload.ingredients.find(i => i.name === 'Rye');
    const vermouth = payload.ingredients.find(i => i.name === 'Vermouth');
    expect(rye.is_base_spirit).toBe(true);
    expect(vermouth.is_base_spirit).toBeFalsy();
  });
});
