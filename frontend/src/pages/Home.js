import { getRandomRecipe } from '../api/client.js';
import { RecipeCard } from '../components/RecipeCard.js';
import { EmptyState } from '../components/EmptyState.js';

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
    } else {
      content.appendChild(RecipeCard({ recipe }));
    }
  }).catch(() => {
    content.textContent = 'Failed to load recipe.';
  });

  return el;
}
