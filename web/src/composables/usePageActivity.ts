import { onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'

/**
 * Run `handler` on SPA route changes and when the tab becomes visible.
 * Use this instead of polling or refetching on every live SSE event.
 */
export function usePageActivity(handler: () => void): void {
  const route = useRoute()
  watch(
    () => route.path,
    () => {
      handler()
    },
  )

  function onVisibility() {
    if (document.visibilityState === 'visible') handler()
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', onVisibility)
  })
  onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibility)
  })
}
