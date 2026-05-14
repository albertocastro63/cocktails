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
  ingredients.slice(0, MAX_VISIBLE).forEach(({ name }) => {
    const li = document.createElement('li');
    li.textContent = name;
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

export function RecipeCard({ recipe }) {
  const el = document.createElement('div');
  el.className = 'relative bg-white rounded-2xl border border-stone-200 shadow-sm border-l-4 border-l-amber-400 p-4 hover:shadow-lg hover:border-l-amber-500 transition-shadow';
  const count = recipe.ingredients ? recipe.ingredients.length : 0;
  el.innerHTML = `
    <a href="#/recipes/${recipe.id}" class="block">
      <h2 class="text-xl font-semibold text-stone-900">${recipe.name}</h2>
      <p class="text-sm text-stone-500 mt-1">${count} ingredient${count === 1 ? '' : 's'}</p>
    </a>
  `;

  const ingredients = recipe.ingredients || [];
  el.addEventListener('mouseenter', () => el.appendChild(buildIngredientPopover(ingredients)));
  el.addEventListener('mouseleave', () => {
    const popover = el.querySelector('[data-popover]');
    if (popover) popover.remove();
  });

  return el;
}
