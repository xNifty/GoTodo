<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import type { ShareLinkTask } from '@/api/types'
import RichBody from '@/components/RichBody.vue'

const props = defineProps<{
  task: ShareLinkTask
}>()

const emit = defineEmits<{
  close: []
}>()

const statusLabel = computed(() => {
  if (props.task.status_name) return props.task.status_name
  return props.task.completed ? 'Done' : 'Open'
})

const description = computed(() => (props.task.description || '').trim())
const dueDate = computed(() => (props.task.due_date || '').trim())
const tags = computed(() => props.task.tags || [])

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div
    id="shareTaskModal"
    class="modal show d-block"
    style="background: rgba(0,0,0,0.5);"
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    aria-labelledby="shareTaskModalTitle"
    @click.self="emit('close')"
  >
    <div class="modal-dialog modal-dialog-centered modal-lg modal-dialog-scrollable">
      <div
        class="modal-content border-0 shadow"
        style="background: var(--ordryn-card-bg); color: var(--ordryn-text);"
      >
        <div class="modal-header border-0 pb-0">
          <h5 id="shareTaskModalTitle" class="modal-title fw-bold d-flex align-items-center gap-2 flex-wrap">
            View Task
            <span class="text-muted fw-normal">#{{ task.id }}</span>
          </h5>
          <button type="button" class="btn-close" aria-label="Close" @click="emit('close')" />
        </div>
        <div class="modal-body py-3">
          <div v-if="task.completed" class="alert alert-success py-2 mb-3">
            <i class="bi bi-check-circle" /> This task is completed
          </div>

          <div class="mb-3">
            <div class="form-label text-muted small mb-1">Title</div>
            <p class="mb-0 fw-semibold" :class="{ 'text-decoration-line-through text-muted': task.completed }">
              {{ task.title }}
            </p>
          </div>

          <div class="mb-3">
            <div class="form-label text-muted small mb-1">Description</div>
            <RichBody v-if="description" :body="description" />
            <p v-else class="text-muted small mb-0">No description</p>
          </div>

          <div class="mb-3">
            <div class="form-label text-muted small mb-1">Due date</div>
            <p v-if="dueDate" class="mb-0">{{ dueDate }}</p>
            <p v-else class="text-muted small mb-0">No due date</p>
          </div>

          <div class="mb-3">
            <div class="form-label text-muted small mb-1">Tags</div>
            <div v-if="tags.length" class="d-flex flex-wrap gap-1">
              <span
                v-for="tag in tags"
                :key="tag.id"
                class="tag-chip"
                :style="{ backgroundColor: tag.color || '#6c757d' }"
              >{{ tag.name }}</span>
            </div>
            <p v-else class="text-muted small mb-0">No tags</p>
          </div>

          <div>
            <div class="form-label text-muted small mb-1">Status</div>
            <p class="mb-0">{{ statusLabel }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
