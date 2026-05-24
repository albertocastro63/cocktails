import { getRecipes } from '../api/client.js';
import { RecipeCard } from '../components/RecipeCard.js';
import { EmptyState } from '../components/EmptyState.js';
import { SearchBar } from '../components/SearchBar.js';
import { SortButtonGroup } from '../components/SortButtonGroup.js';
import { getUserID, isAdmin } from '../api/auth.js';

export function sortRecipes(recipes, direction) {
  return recipes.slice().sort((a, b) =>
    direction === 'asc'
      ? a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })
      : b.name.localeCompare(a.name, undefined, { sensitivity: 'base' })
  );
}

export function RecipeList() {
  const el = document.createElement('div');
  el.className = 'max-w-4xl mx-auto px-4 py-8';

  const heading = document.createElement('h1');
  heading.className = 'text-3xl font-bold text-stone-900 mb-6';
  heading.textContent = 'All Recipes';
  el.appendChild(heading);

  const searchWrap = document.createElement('div');
  searchWrap.className = 'mb-4';
  el.appendChild(searchWrap);

  const sortWrap = document.createElement('div');
  sortWrap.className = 'mb-4';
  el.appendChild(sortWrap);

  const content = document.createElement('div');
  content.textContent = 'Loading…';
  el.appendChild(content);

  let currentQ = '';
  let currentSortDir = null;
  let loadedData = [];

  function renderGrid(data, q) {
    const currentUser = getUserID() ? { id: getUserID(), isAdmin: isAdmin() } : null;
    content.innerHTML = '';
    if (!data || data.length === 0) {
      content.appendChild(EmptyState({ message: q ? 'No recipes match your search.' : 'No recipes yet. Add one!' }));
    } else {
      const displayData = currentSortDir ? sortRecipes(data, currentSortDir) : data;
      const grid = document.createElement('div');
      grid.className = 'grid gap-4 sm:grid-cols-2 lg:grid-cols-3';
      displayData.forEach((recipe) => grid.appendChild(RecipeCard({ recipe, currentUser })));
      content.appendChild(grid);
    }
  }

  function renderSortButtons() {
    sortWrap.innerHTML = '';
    sortWrap.appendChild(SortButtonGroup({ currentDir: currentSortDir, onSort }));
  }

  function onSort(dir) {
    currentSortDir = dir;
    renderSortButtons();
    renderGrid(loadedData, currentQ);
  }

  function loadRecipes(q = '') {
    content.innerHTML = '';
    const loadingP = document.createElement('p');
    loadingP.className = 'text-stone-500 animate-pulse py-4 text-center';
    loadingP.textContent = 'Loading…';
    content.appendChild(loadingP);
    const params = q ? { q } : {};
    getRecipes(params).then(({ data }) => {
      loadedData = data || [];
      renderGrid(loadedData, q);
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

  const hint = document.createElement('p');
  hint.className = 'text-stone-400 text-sm mt-1';
  hint.textContent = 'Tip: use "and" or "+" to search multiple ingredients — e.g. "gin and lemon"';
  searchWrap.appendChild(hint);

  renderSortButtons();
  loadRecipes();

  return el;
}
