import { getRecipe, deleteRecipe, getFavoriteStatus, favoriteRecipe, unfavoriteRecipe } from '../api/client.js';
import { getToken, getUserID, isAdmin } from '../api/auth.js';
import { FavoriteButton } from '../components/FavoriteButton.js';
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

    const uid = getUserID();
    const canEdit = uid && (isAdmin() || (recipe.creator_id && uid === recipe.creator_id));
    if (canEdit) {
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

    if (uid && recipe.creator_id !== uid) {
      getFavoriteStatus(id, getToken()).then(({ is_favorite }) => {
        let isFavorited = is_favorite;
        const renderHeart = () => {
          const existing = titleRow.querySelector('button[data-heart]');
          if (existing) existing.remove();
          const btn = FavoriteButton({
            isFavorited,
            onToggle: () => {
              const action = isFavorited ? unfavoriteRecipe : favoriteRecipe;
              isFavorited = !isFavorited;
              renderHeart();
              action(id, getToken()).catch(() => {
                isFavorited = !isFavorited;
                renderHeart();
                const err = document.createElement('p');
                err.className = 'text-red-600 text-sm mt-1';
                err.textContent = 'Failed to save. Try again.';
                titleRow.appendChild(err);
                setTimeout(() => err.remove(), 4000);
              });
            },
          });
          btn.dataset.heart = '';
          btn.className += ' ml-3';
          titleRow.appendChild(btn);
        };
        renderHeart();
      }).catch(() => {});
    }

    content.appendChild(titleRow);

    if (recipe.ingredients && recipe.ingredients.length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Ingredients';
      content.appendChild(h2);
      content.appendChild(IngredientList({ ingredients: recipe.ingredients }));
    }

    if (recipe.garnishes && recipe.garnishes.length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Garnishes';
      content.appendChild(h2);
      const ul = document.createElement('ul');
      ul.className = 'space-y-1 text-stone-700';
      recipe.garnishes.forEach((g) => {
        const li = document.createElement('li');
        const em = document.createElement('em');
        em.textContent = g;
        li.appendChild(em);
        ul.appendChild(li);
      });
      content.appendChild(ul);
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

    if (recipe.notes) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Notes';
      content.appendChild(h2);
      const div = document.createElement('div');
      div.className = 'prose prose-stone max-w-none overflow-x-auto';
      div.innerHTML = renderMarkdown(recipe.notes);
      content.appendChild(div);
    }

    if (recipe.properties && Object.keys(recipe.properties).length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h2.textContent = 'Properties';
      content.appendChild(h2);
      content.appendChild(PropertyTable({ properties: recipe.properties }));
    }

    // Related cocktails (feature 028) — shown only here, at the bottom, when any
    // exist. Server returns them alphabetically as {id, name}. Never on Home.
    if (recipe.related && recipe.related.length > 0) {
      const h2 = document.createElement('h2');
      h2.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-8 mb-2';
      h2.textContent = 'Related cocktails';
      content.appendChild(h2);
      const ul = document.createElement('ul');
      ul.className = 'flex flex-wrap gap-x-4 gap-y-1';
      recipe.related.forEach((r) => {
        const li = document.createElement('li');
        const a = document.createElement('a');
        a.href = `#/recipes/${r.id}`;
        a.textContent = r.name;
        a.className = 'text-amber-700 hover:text-amber-800 underline';
        li.appendChild(a);
        ul.appendChild(li);
      });
      content.appendChild(ul);
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
