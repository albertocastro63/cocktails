export function IngredientList({ ingredients = [] } = {}) {
  const el = document.createElement('div');
  if (ingredients.length === 0) {
    el.className = 'text-stone-400 text-sm';
    el.textContent = 'No ingredients listed.';
    return el;
  }
  const ul = document.createElement('ul');
  ul.className = 'divide-y divide-stone-100';
  ingredients.forEach(({ name, quantity, unit, is_base_spirit }) => {
    const li = document.createElement('li');
    li.className = 'flex justify-between py-2 text-sm';
    const nameSpan = document.createElement('span');
    if (is_base_spirit) {
      nameSpan.className = 'font-semibold text-stone-900';
      nameSpan.textContent = name;
      const label = document.createElement('span');
      label.className = 'text-amber-600 text-xs ml-1';
      label.textContent = '(base spirit)';
      li.appendChild(nameSpan);
      li.appendChild(label);
    } else {
      nameSpan.className = 'font-medium text-stone-800';
      nameSpan.textContent = name;
      li.appendChild(nameSpan);
    }
    const qtySpan = document.createElement('span');
    qtySpan.className = 'text-stone-500';
    qtySpan.textContent = `${quantity}${unit ? ' ' + unit : ''}`;
    li.appendChild(qtySpan);
    ul.appendChild(li);
  });
  el.appendChild(ul);
  return el;
}
