import { writable } from 'svelte/store';

function getInitial(): 'light' | 'dark' {
  if (typeof localStorage !== 'undefined') {
    const stored = localStorage.getItem('llmview-theme');
    if (stored === 'light' || stored === 'dark') return stored;
  }
  return 'light';
}

export const theme = writable<'light' | 'dark'>(getInitial());

theme.subscribe((v) => {
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', v);
    localStorage.setItem('llmview-theme', v);
  }
});

export function toggleTheme() {
  theme.update((v) => (v === 'light' ? 'dark' : 'light'));
}
