export function Footer() {
  const footer = document.createElement('footer');

  const wrapper = document.createElement('div');
  wrapper.className = 'max-w-4xl mx-auto px-4';

  const hr = document.createElement('hr');
  hr.className = 'border-stone-300 mt-4 mb-0';

  const copy = document.createElement('p');
  copy.className = 'text-stone-500 text-sm text-center py-4';
  copy.textContent = `© ${new Date().getFullYear()} Cocktails`;

  wrapper.appendChild(hr);
  wrapper.appendChild(copy);
  footer.appendChild(wrapper);

  return footer;
}
