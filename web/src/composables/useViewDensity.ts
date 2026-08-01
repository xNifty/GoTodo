import { onMounted, ref } from 'vue'

export type ViewDensity = 'comfortable' | 'dense'

const density = ref<ViewDensity>('comfortable')

function readSavedDensity(): ViewDensity {
  try {
    const saved = localStorage.getItem('ordryn_view_density')
    if (saved === 'comfortable' || saved === 'dense') {
      return saved
    }
  } catch {
    /* ignore */
  }
  return 'comfortable'
}

function applyDensity(next: ViewDensity) {
  density.value = next
  try {
    localStorage.setItem('ordryn_view_density', next)
  } catch {
    /* ignore */
  }
}

export function useViewDensity() {
  onMounted(() => {
    density.value = readSavedDensity()
  })

  function setDensity(next: ViewDensity) {
    applyDensity(next)
  }

  function toggleDensity() {
    applyDensity(density.value === 'comfortable' ? 'dense' : 'comfortable')
  }

  return {
    density,
    setDensity,
    toggleDensity,
  }
}
