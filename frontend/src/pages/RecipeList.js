import { getRecipes } from '../api/client.js';
import { RecipeCard } from '../components/RecipeCard.js';
import { EmptyState } from '../components/EmptyState.js';
import { SearchBar } from '../components/SearchBar.js';

export function RecipeList() {
  const el = document.createElement('div');
  el.className = 'max-w-4xl mx-auto px-4 py-8';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-6';
  heading.textContent = 'All Recipes';
  el.appendChild(heading);

  const searchWrap = document.createElement('div');
  searchWrap.className = 'mb-6';
  el.appendChild(searchWrap);

  const content = document.createElement('div');
  content.textContent = 'Loading…';
  el.appendChild(content);

  let currentQ = '';

  function loadRecipes(q = '') {
    content.innerHTML = '';
    const loadingP = document.createElement('p');
    loadingP.className = 'text-stone-500 animate-pulse py-4 text-center';
    loadingP.textContent = 'Loading…';
    content.appendChild(loadingP);
    const params = q ? { q } : {};
    getRecipes(params).then(({ data }) => {
      content.textContent = '';
      if (!data || data.length === 0) {
        content.appendChild(EmptyState({ message: q ? 'No recipes match your search.' : 'No recipes yet. Add one!' }));
      } else {
        const grid = document.createElement('div');
        grid.className = 'grid gap-4 sm:grid-cols-2 lg:grid-cols-3';
        data.forEach((recipe) => grid.appendChild(RecipeCard({ recipe })));
        content.appendChild(grid);
      }
    }).catch(() => {
      content.textContent = 'Failed to load recipes.';
    });
  }

  const bar = SearchBar({
    onSearch(q) {
      currentQ = q;
      loadRecipes(q);
    },
    value: currentQ,
  });
  searchWrap.appendChild(bar);

  loadRecipes();

  return el;
}
