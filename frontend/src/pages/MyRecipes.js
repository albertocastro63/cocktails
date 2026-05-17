import { getMyRecipes } from '../api/client.js';
import { getToken, getUserID, isAdmin } from '../api/auth.js';
import { RecipeCard } from '../components/RecipeCard.js';
import { EmptyState } from '../components/EmptyState.js';

export function MyRecipes() {
  const el = document.createElement('div');
  el.className = 'max-w-4xl mx-auto px-4 py-8';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-6';
  heading.textContent = 'My Recipes';
  el.appendChild(heading);

  const content = document.createElement('div');
  const loadingP = document.createElement('p');
  loadingP.className = 'text-stone-500 animate-pulse py-4 text-center';
  loadingP.textContent = 'Loading…';
  content.appendChild(loadingP);
  el.appendChild(content);

  const currentUser = getUserID() ? { id: getUserID(), isAdmin: isAdmin() } : null;

  getMyRecipes(getToken()).then(({ data }) => {
    content.innerHTML = '';
    if (!data || data.length === 0) {
      content.appendChild(EmptyState({ message: "You haven't added any recipes yet." }));
    } else {
      const grid = document.createElement('div');
      grid.className = 'grid gap-4 sm:grid-cols-2 lg:grid-cols-3';
      data.forEach((recipe) => grid.appendChild(RecipeCard({ recipe, currentUser })));
      content.appendChild(grid);
    }
  }).catch(() => {
    content.innerHTML = '';
    const errP = document.createElement('p');
    errP.className = 'text-stone-500 py-4 text-center';
    errP.textContent = 'Failed to load recipes.';
    content.appendChild(errP);
  });

  return el;
}
