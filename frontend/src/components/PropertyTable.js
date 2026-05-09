export function PropertyTable({ properties = {} } = {}) {
  const el = document.createElement('div');
  const entries = Object.entries(properties);
  if (entries.length === 0) {
    el.className = 'text-gray-400 text-sm';
    el.textContent = 'No additional properties.';
    return el;
  }
  const table = document.createElement('table');
  table.className = 'w-full text-sm';
  const tbody = document.createElement('tbody');
  table.appendChild(tbody);
  entries.forEach(([key, value]) => {
    const tr = document.createElement('tr');
    tr.className = 'border-b border-gray-100';
    tr.innerHTML = `
      <td class="py-2 pr-4 text-gray-500 capitalize">${key}</td>
      <td class="py-2 text-gray-800">${value}</td>
    `;
    tbody.appendChild(tr);
  });
  el.appendChild(table);
  return el;
}
