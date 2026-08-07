<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

const SCROLL_THRESHOLD = 300
const visible = ref(false)

function updateVisibility() {
  visible.value = window.scrollY > SCROLL_THRESHOLD
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  updateVisibility()
  window.addEventListener('scroll', updateVisibility, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', updateVisibility)
})
</script>

<template>
  <button
    type="button"
    class="go-to-top-btn btn btn-sm rounded-pill"
    :class="{ 'is-visible': visible }"
    aria-label="Go to top"
    :tabindex="visible ? 0 : -1"
    @click="scrollToTop"
  >
    <i class="bi bi-arrow-up" aria-hidden="true" />
    <span class="go-to-top-btn__label">Top</span>
  </button>
</template>
