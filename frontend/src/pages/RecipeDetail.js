import { getRecipe, deleteRecipe } from '../api/client.js';
import { getToken, isLoggedIn } from '../api/auth.js';
import { IngredientList } from '../components/IngredientList.js';
import { PropertyTable } from '../components/PropertyTable.js';

export function RecipeDetail({ id }) {
  const el = document.createElement('div');
  el.className = 'max-w-2xl mx-auto px-4 py-8';

  const content = document.createElement('div');
  content.textContent = 'Loading…';
  el.appendChild(content);

  getRecipe(id).then((recipe) => {
    content.innerHTML = '';

    const titleRow = document.createElement('div');
    titleRow.className = 'flex items-center justify-between mb-6';

    const title = document.createElement('h1');
    title.className = 'text-3xl font-bold text-gray-900';
    title.textContent = recipe.name;
    titleRow.appendChild(title);

    // Show Edit/Delete only to the creator (we check via JWT on server; client shows controls when logged in)
    if (isLoggedIn()) {
      const controls = document.createElement('div');
      controls.className = 'flex gap-2';
      const editBtn = document.createElement('a');
      editBtn.href = `#/recipes/${id}/edit`;
      editBtn.className = 'text-sm text-indigo-600 hover:text-indigo-800 border border-indigo-300 rounded px-3 py-1';
      editBtn.textContent = 'Edit';
      const deleteBtn = document.createElement('button');
      deleteBtn.type = 'button';
      deleteBtn.textContent = 'Delete';
      deleteBtn.className = 'text-sm text-red-600 hover:text-red-800 border border-red-300 rounded px-3 py-1';
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
      h2.className = 'text-xl font-semibold text-gray-700 mt-6 mb-2';
      h2.textContent = 'Ingredients';
      content.appendChild(h2);
      content.appendChild(IngredientList({ ingredients: recipe.ingredients }));
    }

    if (recipe.steps && recipe.steps.length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-xl font-semibold text-gray-700 mt-6 mb-2';
      h2.textContent = 'Steps';
      content.appendChild(h2);
      const ol = document.createElement('ol');
      ol.className = 'list-decimal list-inside space-y-1 text-gray-700';
      recipe.steps.forEach((step) => {
        const li = document.createElement('li');
        li.textContent = step;
        ol.appendChild(li);
      });
      content.appendChild(ol);
    }

    if (recipe.properties && Object.keys(recipe.properties).length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-xl font-semibold text-gray-700 mt-6 mb-2';
      h2.textContent = 'Properties';
      content.appendChild(h2);
      content.appendChild(PropertyTable({ properties: recipe.properties }));
    }
  }).catch((err) => {
    content.textContent = `Failed to load recipe: ${err.message}`;
  });

  return el;
}
