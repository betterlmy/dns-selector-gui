export const lightTheme: Record<string, string> = {
  '--bg-primary': '#ffffff',
  '--bg-secondary': '#f5f5f5',
  '--bg-tertiary': '#e8e8e8',
  '--text-primary': '#1a1a1a',
  '--text-secondary': '#666666',
  '--border-color': '#d9d9d9',
  '--accent-color': '#1890ff',
  '--success-color': '#52c41a',
  '--warning-color': '#faad14',
  '--error-color': '#ff4d4f',
  '--score-best': '#52c41a',
  '--score-zero': '#ff4d4f',
};

export const darkTheme: Record<string, string> = {
  '--bg-primary': '#1f1f1f',
  '--bg-secondary': '#2d2d2d',
  '--bg-tertiary': '#3a3a3a',
  '--text-primary': '#e8e8e8',
  '--text-secondary': '#a0a0a0',
  '--border-color': '#434343',
  '--accent-color': '#177ddc',
  '--success-color': '#49aa19',
  '--warning-color': '#d89614',
  '--error-color': '#d32029',
  '--score-best': '#49aa19',
  '--score-zero': '#d32029',
};

export function applyTheme(theme: 'light' | 'dark') {
  const vars = theme === 'dark' ? darkTheme : lightTheme;
  const root = document.documentElement;
  Object.entries(vars).forEach(([key, value]) => {
    root.style.setProperty(key, value);
  });
}
