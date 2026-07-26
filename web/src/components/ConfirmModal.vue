<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useConfirm } from '@/composables/useConfirm'

const { state, accept, cancel } = useConfirm()

function onKeydown(e: KeyboardEvent) {
  if (!state.open) return
  if (e.key === 'Escape') {
    e.preventDefault()
    cancel()
  }
}

watch(
  () => state.open,
  (open) => {
    document.body.classList.toggle('modal-open', open)
    document.body.style.overflow = open ? 'hidden' : ''
  },
)

onMounted(() => {
  // Clear any leftover Bootstrap modal body state from prior broken attempts.
  document.body.classList.remove('modal-open')
  document.body.style.overflow = ''
  document.body.style.paddingRight = ''
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  document.body.classList.remove('modal-open')
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="state.open"
      id="siteConfirmModal"
      class="modal fade show d-block"
      tabindex="-1"
      role="dialog"
      aria-modal="true"
      aria-labelledby="siteConfirmModalLabel"
    >
      <div class="modal-dialog modal-dialog-centered" @click.stop>
        <div class="modal-content">
          <div class="modal-header">
            <h5 id="siteConfirmModalLabel" class="modal-title">{{ state.title }}</h5>
            <button type="button" class="btn-close" aria-label="Close" @click="cancel" />
          </div>
          <div class="modal-body">
            <p class="mb-0" style="white-space: pre-wrap">{{ state.message }}</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="cancel">
              {{ state.cancelLabel }}
            </button>
            <button
              type="button"
              class="btn"
              :class="state.danger ? 'btn-danger' : 'btn-primary'"
              @click="accept"
            >
              {{ state.confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </div>
    <div v-if="state.open" class="modal-backdrop fade show" />
  </Teleport>
</template>
