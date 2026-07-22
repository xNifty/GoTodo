import { onMounted, ref } from 'vue'

export type ThemeName = 'warm' | 'light' | 'dark'

export interface ThemeOption {
  id: ThemeName
  label: string
  icon: string
}

export const AVAILABLE_THEMES: ThemeOption[] = [
  { id: 'warm', label: 'Warm Pastel', icon: 'bi-sun-fill' },
  { id: 'light', label: 'Classic Light', icon: 'bi-brightness-high' },
  { id: 'dark', label: 'Classic Dark', icon: 'bi-moon-stars-fill' },
]

const currentTheme = ref<ThemeName>('warm')

function readSavedTheme(): ThemeName {
  try {
    const saved = localStorage.getItem('ordryn_theme') || localStorage.getItem('theme')
    if (saved === 'warm' || saved === 'light' || saved === 'dark') {
      return saved as ThemeName
    }
  } catch {
    /* ignore */
  }
  return 'warm'
}

function applyTheme(next: ThemeName) {
  currentTheme.value = next
  document.documentElement.setAttribute('data-theme', next)
  document.documentElement.setAttribute('data-bs-theme', next === 'dark' ? 'dark' : 'light')
  try {
    localStorage.setItem('ordryn_theme', next)
    localStorage.setItem('theme', next)
    document.cookie = `theme=${next}; path=/; max-age=31536000; SameSite=Lax`
  } catch {
    /* ignore */
  }
}

export function useTheme() {
  onMounted(() => {
    applyTheme(readSavedTheme())
  })

  function setTheme(name: ThemeName) {
    applyTheme(name)
  }

  function cycleTheme() {
    const currentIndex = AVAILABLE_THEMES.findIndex((t) => t.id === currentTheme.value)
    const nextIndex = (currentIndex + 1) % AVAILABLE_THEMES.length
    applyTheme(AVAILABLE_THEMES[nextIndex].id)
  }

  const isDark = () => currentTheme.value === 'dark'

  return {
    theme: currentTheme,
    currentTheme,
    availableThemes: AVAILABLE_THEMES,
    setTheme,
    cycleTheme,
    toggleTheme: cycleTheme,
    isDark,
  }
}
