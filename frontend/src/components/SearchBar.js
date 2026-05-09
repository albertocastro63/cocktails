export function SearchBar({ onSearch, value = '' } = {}) {
  const el = document.createElement('div');
  el.className = 'relative flex items-center';

  const input = document.createElement('input');
  input.type = 'search';
  input.value = value;
  input.placeholder = 'Search recipes…';
  input.className =
    'w-full rounded-lg border border-gray-300 px-4 py-2 pr-10 focus:outline-none focus:ring-2 focus:ring-indigo-500';
  el.appendChild(input);

  if (value) {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.textContent = '×';
    btn.className =
      'absolute right-2 text-gray-400 hover:text-gray-600 text-xl leading-none';
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
