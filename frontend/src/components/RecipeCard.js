function buildIngredientPopover(ingredients) {
  const popover = document.createElement('div');
  popover.className = 'absolute top-full left-0 z-10 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-xl p-3';
  popover.setAttribute('data-popover', '');

  if (ingredients.length === 0) {
    const msg = document.createElement('p');
    msg.className = 'text-sm text-gray-400';
    msg.textContent = 'No ingredients listed.';
    popover.appendChild(msg);
    return popover;
  }

  const ul = document.createElement('ul');
  ul.className = 'text-sm text-gray-700 space-y-1';
  const MAX_VISIBLE = 5;
  ingredients.slice(0, MAX_VISIBLE).forEach(({ name, is_base_spirit }) => {
    const li = document.createElement('li');
    if (is_base_spirit) {
      const nameSpan = document.createElement('span');
      nameSpan.className = 'font-semibold text-stone-900';
      nameSpan.textContent = name;
      const label = document.createElement('span');
      label.className = 'text-amber-600 text-xs ml-1';
      label.textContent = '(base spirit)';
      li.appendChild(nameSpan);
      li.appendChild(label);
    } else {
      li.textContent = name;
    }
    ul.appendChild(li);
  });
  if (ingredients.length > MAX_VISIBLE) {
    const li = document.createElement('li');
    li.className = 'text-gray-400';
    li.textContent = '…';
    ul.appendChild(li);
  }
  popover.appendChild(ul);
  return popover;
}

export function RecipeCard({ recipe, currentUser = null, isFavorite = false }) {
  const el = document.createElement('div');
  el.className = 'relative bg-white rounded-2xl border border-stone-200 shadow-sm border-l-4 border-l-amber-400 p-4 hover:shadow-lg hover:border-l-amber-500 transition-shadow';

  const count = recipe.ingredients ? recipe.ingredients.length : 0;
  el.innerHTML = `
    <a href="#/recipes/${recipe.id}" class="block">
      <h2 class="text-xl font-semibold text-stone-900">${recipe.name}</h2>
      <p class="text-sm text-stone-500 mt-1">${count} ingredient${count === 1 ? '' : 's'}</p>
    </a>
  `;

  if (isFavorite) {
    const badge = document.createElement('span');
    badge.setAttribute('data-favorite', 'true');
    badge.setAttribute('aria-label', 'Favorited');
    badge.className = 'absolute top-2 right-2';
    badge.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-4 h-4 text-red-500"><path d="M11.645 20.91l-.007-.003-.022-.012a15.247 15.247 0 01-.383-.218 25.18 25.18 0 01-4.244-3.17C4.688 15.36 2.25 12.174 2.25 8.25 2.25 5.322 4.714 3 7.688 3A5.5 5.5 0 0112 5.052 5.5 5.5 0 0116.313 3c2.973 0 5.437 2.322 5.437 5.25 0 3.925-2.438 7.111-4.739 9.256a25.175 25.175 0 01-4.244 3.17 15.247 15.247 0 01-.383.218l-.022.012-.007.003-.003.001a.752.752 0 01-.704 0l-.003-.001z"/></svg>`;
    el.appendChild(badge);
  }

  const canEdit = currentUser && (currentUser.isAdmin || (recipe.creator_id && currentUser.id === recipe.creator_id));
  if (canEdit) {
    const controls = document.createElement('div');
    controls.className = 'flex gap-2 mt-2';
    const editLink = document.createElement('a');
    editLink.href = `#/recipes/${recipe.id}/edit`;
    editLink.className = 'text-sm text-amber-600 hover:underline';
    editLink.textContent = 'Edit';
    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'text-sm text-red-500 hover:underline';
    deleteBtn.textContent = 'Delete';
    controls.appendChild(editLink);
    controls.appendChild(deleteBtn);
    el.appendChild(controls);
  }

  const ingredients = recipe.ingredients || [];
  el.addEventListener('mouseenter', () => el.appendChild(buildIngredientPopover(ingredients)));
  el.addEventListener('mouseleave', () => {
    const popover = el.querySelector('[data-popover]');
    if (popover) popover.remove();
  });

  return el;
}
