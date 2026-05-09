export function IngredientList({ ingredients = [] } = {}) {
  const el = document.createElement('div');
  if (ingredients.length === 0) {
    el.className = 'text-gray-400 text-sm';
    el.textContent = 'No ingredients listed.';
    return el;
  }
  const ul = document.createElement('ul');
  ul.className = 'divide-y divide-gray-100';
  ingredients.forEach(({ name, quantity, unit }) => {
    const li = document.createElement('li');
    li.className = 'flex justify-between py-2 text-sm';
    li.innerHTML = `
      <span class="font-medium text-gray-800">${name}</span>
      <span class="text-gray-500">${quantity}${unit ? ' ' + unit : ''}</span>
    `;
    ul.appendChild(li);
  });
  el.appendChild(ul);
  return el;
}
