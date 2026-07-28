import { ref, watch } from 'vue'

const saved = typeof localStorage !== 'undefined' && localStorage.getItem('ordryn_sidebar_collapsed') === 'true'
const sidebarCollapsed = ref<boolean>(saved)

watch(sidebarCollapsed, (val) => {
  try {
    localStorage.setItem('ordryn_sidebar_collapsed', String(val))
  } catch {
    /* ignore */
  }
})

export function useSidebarState() {
  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setCollapsed(val: boolean) {
    sidebarCollapsed.value = val
  }

  return {
    sidebarCollapsed,
    toggleSidebar,
    setCollapsed,
  }
}
