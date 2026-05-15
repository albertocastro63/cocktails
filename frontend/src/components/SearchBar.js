export function SearchBar({ onSearch, value = '' } = {}) {
  const el = document.createElement('div');
  el.className = 'relative flex items-center';

  const input = document.createElement('input');
  input.type = 'search';
  input.value = value;
  input.placeholder = 'Search recipes…';
  input.className =
    'w-full rounded-xl border border-stone-300 px-4 py-2 pr-10 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent';
  el.appendChild(input);

  if (value) {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.textContent = '×';
    btn.className =
      'absolute right-2 text-stone-400 hover:text-amber-600 text-xl leading-none';
    btn.addEventListener('click', () => {
      input.value = '';
      onSearch('');
    });
    el.appendChild(btn);
  }

  let timer;
  input.addEventListener('input', () => {
    clearTimeout(timer);
    timer = setTimeout(() => onSearch(input.value), 300);
  });

  return el;
}
