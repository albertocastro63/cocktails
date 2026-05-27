import { createRecipe, updateRecipe, getRecipe } from '../api/client.js';
import { getToken } from '../api/auth.js';
import { MarkdownEditor } from '../components/MarkdownEditor.js';

export function RecipeForm({ id, onSave } = {}) {
  const el = document.createElement('div');
  el.className = 'max-w-2xl mx-auto px-4 py-8';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-6';
  heading.textContent = id ? 'Edit Recipe' : 'New Recipe';
  el.appendChild(heading);

  const form = document.createElement('form');
  form.className = 'space-y-6';
  el.appendChild(form);

  function field(label, name, placeholder = '') {
    const wrap = document.createElement('div');
    const lbl = document.createElement('label');
    lbl.className = 'block text-sm font-medium text-stone-700 mb-1';
    lbl.textContent = label;
    const input = document.createElement('input');
    input.name = name;
    input.type = 'text';
    input.placeholder = placeholder;
    input.className = 'w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent';
    wrap.appendChild(lbl);
    wrap.appendChild(input);
    return wrap;
  }

  form.appendChild(field('Name *', 'name', 'e.g. Mojito'));

  let editorEl = MarkdownEditor({ name: 'notes', placeholder: 'Personal notes, substitutions, tips…', value: '' });
  form.appendChild(editorEl);

  // Ingredients section
  const ingredientsSection = buildDynamicSection(
    'Ingredients',
    'Add Ingredient',
    () => {
      const row = document.createElement('div');
      row.className = 'flex gap-2 items-start';
      row.innerHTML = `
        <input name="ing_name" placeholder="Name" class="flex-1 border rounded px-2 py-1 text-sm" />
        <input name="ing_qty" placeholder="Qty" class="w-16 border rounded px-2 py-1 text-sm" />
        <input name="ing_unit" placeholder="Unit" class="w-16 border rounded px-2 py-1 text-sm" />
        <label class="flex items-center gap-1 text-xs text-amber-700 whitespace-nowrap cursor-pointer">
          <input type="checkbox" name="ing_base_spirit" class="accent-amber-500" />
          base spirit
        </label>
        <button type="button" class="text-red-400 hover:text-red-600 text-lg px-1">×</button>
      `;
      row.querySelector('[name="ing_base_spirit"]').addEventListener('change', (e) => {
        if (e.target.checked) {
          [...ingredientsSection._rows.children].forEach((r) => {
            const cb = r.querySelector('[name="ing_base_spirit"]');
            if (cb && cb !== e.target) cb.checked = false;
          });
        }
      });
      row.querySelector('button').addEventListener('click', () => row.remove());
      return row;
    }
  );
  form.appendChild(ingredientsSection);

  // Steps section
  const stepsSection = buildDynamicSection(
    'Steps',
    'Add Step',
    () => {
      const row = document.createElement('div');
      row.className = 'flex gap-2 items-start';
      row.innerHTML = `
        <input name="step" placeholder="Describe this step…" class="flex-1 border rounded px-2 py-1 text-sm" />
        <button type="button" class="text-red-400 hover:text-red-600 text-lg px-1">×</button>
      `;
      row.querySelector('button').addEventListener('click', () => row.remove());
      return row;
    }
  );
  form.appendChild(stepsSection);

  // Properties section
  const propertiesSection = buildDynamicSection(
    'Properties',
    'Add Property',
    () => {
      const row = document.createElement('div');
      row.className = 'flex gap-2 items-start';
      row.innerHTML = `
        <input name="prop_key" placeholder="Key" class="w-32 border rounded px-2 py-1 text-sm" />
        <input name="prop_val" placeholder="Value" class="flex-1 border rounded px-2 py-1 text-sm" />
        <button type="button" class="text-red-400 hover:text-red-600 text-lg px-1">×</button>
      `;
      row.querySelector('button').addEventListener('click', () => row.remove());
      return row;
    }
  );
  form.appendChild(propertiesSection);

  // Garnishes section (T011)
  const garnishesSection = buildDynamicSection(
    'Garnishes',
    'Add Garnish',
    () => {
      const row = document.createElement('div');
      row.className = 'flex gap-2 items-start';
      row.innerHTML = `
        <input name="garnish" placeholder="e.g. Express orange oil over the cocktail" class="flex-1 border rounded px-2 py-1 text-sm" />
        <button type="button" class="text-red-400 hover:text-red-600 text-lg px-1">×</button>
      `;
      row.querySelector('button').addEventListener('click', () => row.remove());
      return row;
    }
  );
  form.appendChild(garnishesSection);

  const errorMsg = document.createElement('p');
  errorMsg.className = 'text-red-600 text-sm hidden';
  form.appendChild(errorMsg);

  const submit = document.createElement('button');
  submit.type = 'submit';
  submit.className = 'bg-amber-500 text-stone-900 rounded-xl px-6 py-2 font-semibold hover:bg-amber-600 transition-colors';
  submit.textContent = id ? 'Save Changes' : 'Create Recipe';
  form.appendChild(submit);

  // Prefill if editing
  if (id) {
    getRecipe(id).then((recipe) => {
      form.querySelector('[name="name"]').value = recipe.name;
      (recipe.ingredients || []).forEach((ing) => {
        const row = ingredientsSection._addRow();
        row.querySelector('[name="ing_name"]').value = ing.name;
        row.querySelector('[name="ing_qty"]').value = ing.quantity;
        row.querySelector('[name="ing_unit"]').value = ing.unit || '';
        if (ing.is_base_spirit) {
          row.querySelector('[name="ing_base_spirit"]').checked = true;
        }
      });
      (recipe.steps || []).forEach((step) => {
        const row = stepsSection._addRow();
        row.querySelector('[name="step"]').value = step;
      });
      Object.entries(recipe.properties || {}).forEach(([k, v]) => {
        const row = propertiesSection._addRow();
        row.querySelector('[name="prop_key"]').value = k;
        row.querySelector('[name="prop_val"]').value = v;
      });
      (recipe.garnishes || []).forEach((g) => {
        const row = garnishesSection._addRow();
        row.querySelector('[name="garnish"]').value = g;
      });
      const newEditor = MarkdownEditor({ name: 'notes', placeholder: 'Personal notes, substitutions, tips…', value: recipe.notes || '' });
      form.replaceChild(newEditor, editorEl);
      editorEl = newEditor;
    }).catch(() => {});
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorMsg.classList.add('hidden');

    const name = form.querySelector('[name="name"]').value.trim();
    if (!name) {
      errorMsg.textContent = 'Name is required.';
      errorMsg.classList.remove('hidden');
      return;
    }

    const ingredients = [...ingredientsSection._rows.children].map((row) => {
      const isBase = row.querySelector('[name="ing_base_spirit"]')?.checked || false;
      return {
        name: (row.querySelector('[name="ing_name"]') || {}).value?.trim() || '',
        quantity: (row.querySelector('[name="ing_qty"]') || {}).value?.trim() || '',
        unit: (row.querySelector('[name="ing_unit"]') || {}).value?.trim() || '',
        ...(isBase && { is_base_spirit: true }),
      };
    }).filter((i) => i.name);

    const steps = [...stepsSection._rows.children].map((row) => {
      const input = row.querySelector('[name="step"]');
      return input ? input.value.trim() : '';
    }).filter(Boolean);

    const properties = {};
    [...propertiesSection._rows.children].forEach((row) => {
      const k = (row.querySelector('[name="prop_key"]') || {}).value?.trim() || '';
      const v = (row.querySelector('[name="prop_val"]') || {}).value?.trim() || '';
      if (k) properties[k] = v;
    });

    const notes = (form.querySelector('[name="notes"]') || {}).value?.trim() || '';

    const garnishes = [...garnishesSection._rows.children].map((row) => {
      const input = row.querySelector('[name="garnish"]');
      return input ? input.value.trim() : '';
    }).filter(Boolean);

    const payload = { name, ingredients, steps, properties, notes, garnishes };
    const token = getToken();

    try {
      const result = id
        ? await updateRecipe(id, payload, token)
        : await createRecipe(payload, token);
      if (onSave) onSave(result);
    } catch (err) {
      errorMsg.textContent = err.message || 'Failed to save recipe.';
      errorMsg.classList.remove('hidden');
    }
  });

  return el;
}

function buildDynamicSection(title, btnLabel, createRow) {
  const section = document.createElement('div');
  const lbl = document.createElement('div');
  lbl.className = 'flex items-center justify-between mb-2';
  lbl.innerHTML = `<span class="text-sm font-medium text-stone-700">${title}</span>`;

  const addBtn = document.createElement('button');
  addBtn.type = 'button';
  addBtn.textContent = `+ ${btnLabel}`;
  addBtn.className = 'text-sm text-amber-700 hover:text-amber-900';
  lbl.appendChild(addBtn);

  const rows = document.createElement('div');
  rows.className = 'space-y-2';

  section.appendChild(lbl);
  section.appendChild(rows);
  section._rows = rows;

  function addRow() {
    const row = createRow();
    rows.appendChild(row);
    return row;
  }

  addBtn.addEventListener('click', addRow);
  section._addRow = addRow;

  return section;
}
