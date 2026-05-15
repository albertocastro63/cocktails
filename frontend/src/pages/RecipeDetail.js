import { getRecipe, deleteRecipe } from '../api/client.js';
import { getToken, isLoggedIn } from '../api/auth.js';
import { IngredientList } from '../components/IngredientList.js';
import { PropertyTable } from '../components/PropertyTable.js';
import { renderMarkdown } from '../utils/markdown.js';

export function RecipeDetail({ id }) {
  const el = document.createElement('div');
  el.className = 'max-w-2xl mx-auto px-4 py-8';

  const content = document.createElement('div');
  const loadingP = document.createElement('p');
  loadingP.className = 'text-stone-500 animate-pulse py-4 text-center';
  loadingP.textContent = 'Loading…';
  content.appendChild(loadingP);
  el.appendChild(content);

  getRecipe(id).then((recipe) => {
    content.innerHTML = '';

    const titleRow = document.createElement('div');
    titleRow.className = 'flex items-center justify-between mb-6';

    const title = document.createElement('h1');
    title.className = 'text-3xl font-bold text-stone-900';
    title.textContent = recipe.name;
    titleRow.appendChild(title);

    if (isLoggedIn()) {
      const controls = document.createElement('div');
      controls.className = 'flex gap-2';
      const editBtn = document.createElement('a');
      editBtn.href = `#/recipes/${id}/edit`;
      editBtn.className = 'text-sm text-stone-700 hover:text-amber-700 border border-stone-300 hover:border-amber-500 rounded-xl px-3 py-1';
      editBtn.textContent = 'Edit';
      const deleteBtn = document.createElement('button');
      deleteBtn.type = 'button';
      deleteBtn.textContent = 'Delete';
      deleteBtn.className = 'text-sm text-red-600 hover:text-red-800 border border-red-300 hover:bg-red-50 rounded-xl px-3 py-1';
      deleteBtn.addEventListener('click', async () => {
        if (!confirm('Delete this recipe?')) return;
        try {
          await deleteRecipe(id, getToken());
          location.hash = '#/recipes';
        } catch (err) {
          alert(err.message || 'Failed to delete.');
        }
      });
      controls.appendChild(editBtn);
      controls.appendChild(deleteBtn);
      titleRow.appendChild(controls);
    }

    content.appendChild(titleRow);

    if (recipe.ingredients && recipe.ingredients.length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Ingredients';
      content.appendChild(h2);
      content.appendChild(IngredientList({ ingredients: recipe.ingredients }));
    }

    if (recipe.steps && recipe.steps.length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Steps';
      content.appendChild(h2);
      const ol = document.createElement('ol');
      ol.className = 'list-decimal list-inside space-y-1 text-stone-700';
      recipe.steps.forEach((step) => {
        const li = document.createElement('li');
        li.textContent = step;
        ol.appendChild(li);
      });
      content.appendChild(ol);
    }

    if (recipe.properties && Object.keys(recipe.properties).length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Properties';
      content.appendChild(h2);
      content.appendChild(PropertyTable({ properties: recipe.properties }));
    }

    if (recipe.notes) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Notes';
      content.appendChild(h2);
      const div = document.createElement('div');
      div.className = 'text-stone-700';
      div.innerHTML = renderMarkdown(recipe.notes);
      content.appendChild(div);
    }
  }).catch((err) => {
    content.innerHTML = '';
    const errP = document.createElement('p');
    errP.className = 'text-stone-500 py-4 text-center';
    errP.textContent = `Failed to load recipe: ${err.message}`;
    content.appendChild(errP);
  });

  return el;
}
