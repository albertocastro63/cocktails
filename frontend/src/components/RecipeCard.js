export function RecipeCard({ recipe }) {
  const el = document.createElement('div');
  el.className = 'bg-white rounded-lg shadow p-4 hover:shadow-md transition-shadow';
  const count = recipe.ingredients ? recipe.ingredients.length : 0;
  el.innerHTML = `
    <a href="#/recipes/${recipe.id}" class="block">
      <h2 class="text-xl font-semibold text-gray-800">${recipe.name}</h2>
      <p class="text-sm text-gray-500 mt-1">${count} ingredient${count === 1 ? '' : 's'}</p>
    </a>
  `;
  return el;
}
