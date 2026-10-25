import { getRecipes, getMyFavorites } from '../api/client.js';
import { RecipeCard } from '../components/RecipeCard.js';
import { EmptyState } from '../components/EmptyState.js';
import { SearchBar } from '../components/SearchBar.js';
import { SortButtonGroup } from '../components/SortButtonGroup.js';
import { Pagination } from '../components/Pagination.js';
import { getUserID, isAdmin, getToken } from '../api/auth.js';

const PAGE_SIZE = 20;

export function normaliseWhisky(q) {
  return q.replace(/\bwhisky\b/gi, (m) => m.slice(0, -1) + 'ey');
}

export function parseBaseSpirit(rawQ) {
  const match = rawQ.match(/base\s+spirit\s+(?:is|=)\s*(.*?)(?:\s+(?=base\s+spirit\s+)|$)/i);
  const baseSpirit = match ? match[1].trim() : '';
  const q = rawQ.replace(/base\s+spirit\s+(?:is|=)\s*[^\n]*/gi, '').trim();
  return { baseSpirit, q };
}

// sortParam maps the A→Z / Z→A control direction to the API sort value. Sorting
// is server-side so that order is consistent across page boundaries, not just
// within the recipes currently on screen.
function sortParam(dir) {
  if (dir === 'asc') return 'name_asc';
  if (dir === 'desc') return 'name_desc';
  return '';
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

  const totalLabel = document.createElement('p');
  totalLabel.className = 'text-stone-500 text-sm mb-3';
  totalLabel.dataset.testid = 'total-count';
  el.appendChild(totalLabel);

  const content = document.createElement('div');
  content.textContent = 'Loading…';
  el.appendChild(content);

  const pageWrap = document.createElement('div');
  el.appendChild(pageWrap);

  let currentQ = '';
  let currentSearchQ = '';
  let currentBaseSpirit = '';
  let currentSortDir = null;
  let currentPage = 1;
  let favoriteIds = new Set();

  function renderGrid(data) {
    const currentUser = getUserID() ? { id: getUserID(), isAdmin: isAdmin() } : null;
    content.innerHTML = '';
    if (!data || data.length === 0) {
      const searching = currentSearchQ || currentBaseSpirit;
      content.appendChild(EmptyState({ message: searching ? 'No recipes match your search.' : 'No recipes yet. Add one!' }));
      return;
    }
    const grid = document.createElement('div');
    grid.className = 'grid gap-4 sm:grid-cols-2 lg:grid-cols-3';
    data.forEach((recipe) => grid.appendChild(RecipeCard({ recipe, currentUser, isFavorite: favoriteIds.has(recipe.id) })));
    content.appendChild(grid);
  }

  function renderTotal(total) {
    if (!total) {
      totalLabel.textContent = '';
      return;
    }
    totalLabel.textContent = `${total} recipe${total === 1 ? '' : 's'}`;
  }

  function renderPagination(total) {
    pageWrap.innerHTML = '';
    const totalPages = Math.max(1, Math.ceil((total || 0) / PAGE_SIZE));
    pageWrap.appendChild(Pagination({ page: currentPage, totalPages, onPage: goToPage }));
  }

  function renderSortButtons() {
    sortWrap.innerHTML = '';
    sortWrap.appendChild(SortButtonGroup({ currentDir: currentSortDir, onSort }));
  }

  function goToPage(page) {
    currentPage = page;
    loadRecipes();
  }

  function onSort(dir) {
    currentSortDir = dir;
    currentPage = 1; // a new ordering starts from the first page
    renderSortButtons();
    loadRecipes();
  }

  // loadRecipes fetches the current page for the active query/filter/sort and
  // renders the grid, total, and page controls from the server response.
  function loadRecipes() {
    content.innerHTML = '';
    const loadingP = document.createElement('p');
    loadingP.className = 'text-stone-500 animate-pulse py-4 text-center';
    loadingP.textContent = 'Loading…';
    content.appendChild(loadingP);

    const params = { page: currentPage, limit: PAGE_SIZE };
    if (currentSearchQ) params.q = currentSearchQ;
    if (currentBaseSpirit) params.base_spirit = currentBaseSpirit;
    const sort = sortParam(currentSortDir);
    if (sort) params.sort = sort;

    const token = getToken();
    const favoritesPromise = token
      ? getMyFavorites(token).then(({ data }) => { favoriteIds = new Set((data || []).map((r) => r.id)); }).catch(() => {})
      : Promise.resolve();

    Promise.all([getRecipes(params), favoritesPromise]).then(([resp]) => {
      const data = resp.data || [];
      const total = resp.total || 0;
      renderGrid(data);
      renderTotal(total);
      renderPagination(total);
    }).catch(() => {
      content.textContent = 'Failed to load recipes.';
      pageWrap.innerHTML = '';
      totalLabel.textContent = '';
    });
  }

  const bar = SearchBar({
    onSearch(raw) {
      const normQ = normaliseWhisky(raw);
      const { baseSpirit, q } = parseBaseSpirit(normQ);
      currentQ = raw;
      currentSearchQ = q;
      currentBaseSpirit = baseSpirit;
      currentPage = 1; // a changed query resets to the first page of results
      loadRecipes();
    },
    value: currentQ,
  });
  searchWrap.appendChild(bar);

  const hint = document.createElement('p');
  hint.className = 'text-stone-400 text-sm mt-1';
  hint.textContent = 'Tip: search by ingredient (use "and" or "+" for multiple) — or try "base spirit is gin"';
  searchWrap.appendChild(hint);

  renderSortButtons();
  loadRecipes();

  return el;
}
