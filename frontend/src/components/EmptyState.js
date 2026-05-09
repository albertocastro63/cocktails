export function EmptyState({ message = 'Nothing here yet.' } = {}) {
  const el = document.createElement('div');
  el.className = 'flex flex-col items-center justify-center py-16 text-gray-400';
  el.innerHTML = `
    <p class="text-lg">${message}</p>
  `;
  return el;
}
