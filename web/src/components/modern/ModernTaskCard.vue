<script setup lang="ts">
import { ref } from 'vue'
import type { Task } from '@/api/types'
import type { ViewDensity } from '@/composables/useViewDensity'

const props = withDefaults(
  defineProps<{
    task: Task
    selected: boolean
    density?: ViewDensity
    depth?: number
    showProjectPill?: boolean
    canWrite?: boolean
  }>(),
  {
    density: 'comfortable',
    depth: 0,
    showProjectPill: true,
    canWrite: true,
  },
)

const emit = defineEmits<{
  'toggle-select': [checked: boolean]
  'toggle-complete': []
  'toggle-favorite': []
  'patch-task': [payload: { id: number; title?: string; description?: string }]
  edit: []
  remove: []
}>()

// Inline Editing State (Desktop)
const isEditingTitle = ref(false)
const titleVal = ref('')
const isEditingDesc = ref(false)
const descVal = ref('')

function startEditTitle() {
  if (!props.canWrite) return
  titleVal.value = props.task.title
  isEditingTitle.value = true
}

function saveTitle() {
  if (!isEditingTitle.value) return
  isEditingTitle.value = false
  const trimmed = titleVal.value.trim()
  if (trimmed && trimmed !== props.task.title) {
    emit('patch-task', { id: props.task.id, title: trimmed })
  }
}

function cancelEditTitle() {
  isEditingTitle.value = false
}

function startEditDesc() {
  if (!props.canWrite) return
  descVal.value = props.task.description || ''
  isEditingDesc.value = true
}

function saveDesc() {
  if (!isEditingDesc.value) return
  isEditingDesc.value = false
  const trimmed = descVal.value.trim()
  if (trimmed !== (props.task.description || '')) {
    emit('patch-task', { id: props.task.id, description: trimmed })
  }
}

function cancelEditDesc() {
  isEditingDesc.value = false
}

function priorityLabel(priority: number) {
  if (priority === 1) return 'Low'
  if (priority === 2) return 'Med'
  if (priority === 3) return 'High'
  return ''
}

function formatDueDate(dateStr?: string): string {
  if (!dateStr) return ''
  const todayStr = new Date().toISOString().slice(0, 10)
  const currentYear = new Date().getFullYear().toString()
  const isOverdue = dateStr < todayStr
  const year = dateStr.slice(0, 4)
  if (isOverdue || year !== currentYear) {
    return dateStr
  }
  const dateObj = new Date(dateStr + 'T00:00:00')
  return dateObj.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>

<template>
  <div
    :id="`task-card-${task.id}`"
    class="ordryn-task-card"
    :class="{
      'is-completed': task.completed,
      'density-comfortable': density === 'comfortable',
      'density-dense': density === 'dense',
    }"
    :style="{
      paddingLeft: depth > 0 ? `calc(${depth} * 28px + ${density === 'dense' ? 0.65 : 1.25}rem)` : undefined
    }"
  >
    <!-- Visual Tree Connector Line for Nested Tasks -->
    <div
      v-if="depth > 0"
      class="nested-tree-line"
      :style="{ '--nest-depth': depth }"
    />

    <div class="d-flex align-items-center justify-content-between gap-2 flex-wrap flex-md-nowrap">
      <!-- Left Controls Container: Tight 5px gap moving everything slightly left -->
      <div class="d-flex align-items-center flex-grow-1 min-w-0" style="gap: 5px;">
        <!-- Drag Handle -->
        <span
          class="drag-handle text-muted flex-shrink-0 d-inline-flex align-items-center justify-content-center m-0 p-0"
          style="cursor: grab; opacity: 0.5;"
          title="Drag to reorder"
        >
          <i class="bi bi-grip-vertical" style="font-size: 1.15rem; line-height: 1;" />
        </span>

        <!-- Multi-Select Checkbox -->
        <div
          v-if="canWrite"
          class="hover-reveal flex-shrink-0 d-inline-flex align-items-center justify-content-center m-0 p-0"
          :class="{ 'is-visible': selected }"
          title="Select task for bulk actions"
        >
          <input
            type="checkbox"
            class="cursor-pointer m-0 p-0"
            :checked="selected"
            style="width: 1.05rem; height: 1.05rem; cursor: pointer; accent-color: var(--ordryn-accent);"
            :aria-label="`Select task ${task.title}`"
            @change="emit('toggle-select', ($event.target as HTMLInputElement).checked)"
          />
        </div>

        <!-- Favorite Star Button -->
        <button
          type="button"
          class="btn btn-link p-0 m-0 border-0 text-decoration-none hover-reveal flex-shrink-0 d-inline-flex align-items-center justify-content-center"
          :class="{ 'is-visible': task.favorite }"
          :aria-label="task.favorite ? 'Unstar task' : 'Star task'"
          @click="emit('toggle-favorite')"
        >
          <i
            :class="task.favorite ? 'bi bi-star-fill text-warning' : 'bi bi-star text-muted opacity-60'"
            style="font-size: 1.05rem; line-height: 1;"
          />
        </button>

        <!-- Completion Checkmark Button -->
        <button
          type="button"
          class="completion-check-btn flex-shrink-0 d-inline-flex align-items-center justify-content-center m-0 p-0 border-0 bg-transparent"
          :title="task.completed ? 'Mark incomplete' : 'Mark complete'"
          :disabled="!canWrite"
          @click="emit('toggle-complete')"
        >
          <i v-if="task.completed" class="bi bi-check-circle-fill text-success" style="font-size: 1.15rem; line-height: 1;" />
          <span v-else class="position-relative d-inline-flex align-items-center justify-content-center">
            <i class="bi bi-circle" style="font-size: 1.15rem; line-height: 1;" />
            <i class="bi bi-check-lg text-success position-absolute top-50 start-50 translate-middle opacity-0 hover-check" style="font-size: 0.85rem;" />
          </span>
        </button>

        <!-- Title & Badges Container -->
        <div class="d-flex flex-column gap-1 min-w-0 flex-grow-1 ms-1">
          <div class="d-flex align-items-center gap-2 flex-wrap">
            <!-- Inline Title Editor (Desktop) -->
            <input
              v-if="isEditingTitle && canWrite"
              v-model="titleVal"
              type="text"
              class="inline-edit-input"
              @blur="saveTitle"
              @keyup.enter="saveTitle"
              @keyup.escape="cancelEditTitle"
            />
            <!-- Title Display (Single click inline edit) -->
            <span
              v-else
              class="task-title fw-semibold text-truncate"
              style="color: var(--ordryn-text);"
              :style="{ cursor: canWrite ? 'pointer' : 'default' }"
              :title="canWrite ? 'Click to rename task' : undefined"
              @click="startEditTitle"
            >
              {{ task.title }}
            </span>

            <!-- Project Badge Pill -->
            <span
              v-if="showProjectPill && task.project"
              class="badge rounded-pill bg-secondary bg-opacity-10 text-secondary border border-secondary border-opacity-20 px-2 py-1 small"
            >
              <i class="bi bi-folder2 me-1" />{{ task.project }}
            </span>

            <!-- Tag Badges -->
            <template v-if="task.tags && task.tags.length > 0">
              <span
                v-for="tag in task.tags"
                :key="tag.id"
                class="ordryn-badge ordryn-badge-tag"
              >
                {{ tag.name }}
              </span>
            </template>
          </div>

          <!-- Inline Description Editor or Preview (Comfortable mode only) -->
          <div v-if="density === 'comfortable'">
            <textarea
              v-if="isEditingDesc && canWrite"
              v-model="descVal"
              class="inline-edit-textarea mt-1"
              rows="2"
              @blur="saveDesc"
              @keyup.escape="cancelEditDesc"
            />
            <div
              v-else-if="task.description"
              class="task-description text-muted small text-break"
              :class="{ 'cursor-pointer': canWrite }"
              :title="canWrite ? 'Click to edit description' : undefined"
              @click="startEditDesc"
            >
              {{ task.description }}
            </div>
          </div>
        </div>
      </div>

      <!-- Right Group (Desktop): Due Date, Priority Badge, Edit/Delete Action Buttons -->
      <div class="d-none d-md-flex align-items-center gap-3 ms-auto">
        <!-- Smart Formatted Due Date (Icon hidden in Dense mode) -->
        <div v-if="task.due_date" class="d-flex align-items-center gap-1 text-muted small whitespace-nowrap">
          <i v-if="density !== 'dense'" class="bi bi-calendar-event opacity-75" />
          <span>{{ formatDueDate(task.due_date) }}</span>
        </div>

        <!-- Priority Pill -->
        <span
          v-if="priorityLabel(task.priority)"
          class="ordryn-badge"
          :class="{
            'ordryn-badge-priority-low': task.priority === 1,
            'ordryn-badge-priority-med': task.priority === 2,
            'ordryn-badge-priority-high': task.priority === 3,
          }"
        >
          Priority {{ priorityLabel(task.priority) }}
        </span>

        <!-- Edit / Delete Actions (Hover-revealed on Desktop, hidden if viewer/read-only) -->
        <div v-if="canWrite" class="d-flex align-items-center gap-1 action-buttons-group hover-reveal">
          <button
            type="button"
            class="btn btn-sm btn-icon text-muted hover-accent border-0 p-1"
            title="Open task details editor"
            @click="emit('edit')"
          >
            <i class="bi bi-pencil" />
          </button>

          <button
            type="button"
            class="btn btn-sm btn-icon text-danger hover-danger border-0 p-1"
            title="Delete task"
            @click="emit('remove')"
          >
            <i class="bi bi-trash" />
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile Action Row (Placed below task on mobile screens) -->
    <div class="d-md-none mobile-actions-wrapper">
      <div class="d-flex align-items-center gap-2 justify-content-between w-100">
        <div class="d-flex align-items-center gap-2">
          <div v-if="task.due_date" class="text-muted small">
            <i v-if="density !== 'dense'" class="bi bi-calendar-event me-1" />{{ formatDueDate(task.due_date) }}
          </div>
          <span
            v-if="priorityLabel(task.priority)"
            class="ordryn-badge"
            :class="{
              'ordryn-badge-priority-low': task.priority === 1,
              'ordryn-badge-priority-med': task.priority === 2,
              'ordryn-badge-priority-high': task.priority === 3,
            }"
          >
            {{ priorityLabel(task.priority) }}
          </span>
        </div>
        <div v-if="canWrite" class="d-flex align-items-center gap-2">
          <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-2" @click="emit('edit')">
            <i class="bi bi-pencil me-1" />Edit
          </button>
          <button type="button" class="btn btn-sm btn-outline-danger py-0 px-2" @click="emit('remove')">
            <i class="bi bi-trash me-1" />Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.completion-check-btn:hover .hover-check {
  opacity: 0.6 !important;
}
.btn-icon {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: background-color 0.15s ease;
}
.btn-icon:hover {
  background-color: var(--ordryn-muted-bg);
}
.hover-danger:hover {
  background-color: rgba(239, 68, 68, 0.1) !important;
}
</style>
