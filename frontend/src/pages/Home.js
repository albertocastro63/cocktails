import { getRandomRecipe } from '../api/client.js';
import { EmptyState } from '../components/EmptyState.js';
import { IngredientList } from '../components/IngredientList.js';
import { PropertyTable } from '../components/PropertyTable.js';

export function Home() {
  const el = document.createElement('div');
  el.className = 'max-w-2xl mx-auto px-4 py-8';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-gray-900 mb-6';
  heading.textContent = 'Random Recipe';
  el.appendChild(heading);

  const content = document.createElement('div');
  content.textContent = 'Loading…';
  el.appendChild(content);

  getRandomRecipe().then((recipe) => {
    content.textContent = '';
    if (recipe === null) {
      content.appendChild(EmptyState({ message: 'No recipes yet. Add one!' }));
      return;
    }

    const title = document.createElement('h2');
    title.className = 'text-2xl font-bold text-gray-900 mb-4';
    title.textContent = recipe.name;
    content.appendChild(title);

    if (recipe.ingredients && recipe.ingredients.length > 0) {
      const h3 = document.createElement('h3');
      h3.className = 'text-lg font-semibold text-gray-700 mt-4 mb-2';
      h3.textContent = 'Ingredients';
      content.appendChild(h3);
      content.appendChild(IngredientList({ ingredients: recipe.ingredients }));
    }

    if (recipe.steps && recipe.steps.length > 0) {
      const h3 = document.createElement('h3');
      h3.className = 'text-lg font-semibold text-gray-700 mt-4 mb-2';
      h3.textContent = 'Steps';
      content.appendChild(h3);
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
      const h3 = document.createElement('h3');
      h3.className = 'text-lg font-semibold text-gray-700 mt-4 mb-2';
      h3.textContent = 'Properties';
      content.appendChild(h3);
      content.appendChild(PropertyTable({ properties: recipe.properties }));
    }

    if (recipe.notes) {
      const h3 = document.createElement('h3');
      h3.className = 'text-lg font-semibold text-gray-700 mt-4 mb-2';
      h3.textContent = 'Notes';
      content.appendChild(h3);
      const p = document.createElement('p');
      p.className = 'text-gray-700 whitespace-pre-wrap';
      p.textContent = recipe.notes;
      content.appendChild(p);
    }
  }).catch(() => {
    content.textContent = 'Failed to load recipe.';
  });

  return el;
}
