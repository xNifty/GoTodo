<template>
  <div class="workflow-panel">
    <h4 class="h6 mb-2">Board mode</h4>
    <p class="small text-muted mb-3">
      Kanban organizes tasks into status columns on a board. Turning board mode off is only allowed
      when the project has no tasks.
    </p>

    <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
      <div class="form-check form-switch m-0">
        <input
          :id="`kanban-toggle-${project.id}`"
          class="form-check-input"
          type="checkbox"
          :checked="isKanban"
          :disabled="savingMode"
          @change="onToggleMode(($event.target as HTMLInputElement).checked)"
        />
        <label class="form-check-label" :for="`kanban-toggle-${project.id}`">
          Enable kanban board
        </label>
      </div>
      <span v-if="isKanban" class="badge text-bg-primary">Kanban</span>
      <span v-else class="badge text-bg-secondary">Classic</span>
    </div>

    <template v-if="isKanban">
      <h4 class="h6">Statuses</h4>
      <p class="small text-muted mb-2">
        Up to {{ maxStatuses }} columns. One must be default (new tasks) and at least one must be done
        (completed).
      </p>

      <ul class="list-unstyled mb-3">
        <li
          v-for="s in statuses"
          :key="s.id"
          class="d-flex flex-wrap align-items-center gap-2 mb-2 pb-2 border-bottom"
        >
          <template v-if="renameId === s.id">
            <form class="d-flex align-items-center gap-1 flex-wrap flex-grow-1" @submit.prevent="saveRename(s)">
              <input
                v-model="renameValue"
                type="text"
                class="form-control form-control-sm"
                maxlength="40"
                required
                aria-label="Status name"
              />
              <button class="btn btn-sm btn-primary" type="submit" aria-label="Save status name">
                <i class="bi bi-check" />
              </button>
              <button
                class="btn btn-sm btn-secondary"
                type="button"
                aria-label="Cancel rename"
                @click="renameId = null"
              >
                <i class="bi bi-x" />
              </button>
            </form>
          </template>
          <template v-else>
            <strong class="me-1">{{ s.name }}</strong>
            <span v-if="s.is_default" class="badge text-bg-info">default</span>
            <span v-if="s.is_done" class="badge text-bg-success">done</span>
            <button
              class="btn btn-sm btn-link p-0"
              type="button"
              aria-label="Rename status"
              @click="beginRename(s)"
            >
              <i class="bi bi-pencil" />
            </button>
          </template>

          <div class="d-flex flex-wrap align-items-center gap-2 ms-auto">
            <div class="form-check form-check-inline m-0">
              <input
                :id="`default-${s.id}`"
                class="form-check-input"
                type="checkbox"
                :checked="s.is_default"
                :disabled="s.is_default"
                @change="setDefault(s)"
              />
              <label class="form-check-label small" :for="`default-${s.id}`">Default</label>
            </div>
            <div class="form-check form-check-inline m-0">
              <input
                :id="`done-${s.id}`"
                class="form-check-input"
                type="checkbox"
                :checked="s.is_done"
                @change="toggleDone(s, ($event.target as HTMLInputElement).checked)"
              />
              <label class="form-check-label small" :for="`done-${s.id}`">Done</label>
            </div>
            <button
              class="btn btn-sm btn-outline-danger"
              type="button"
              :disabled="statuses.length <= 1 || s.is_default"
              title="Delete status"
              @click="askDelete(s)"
            >
              <i class="bi bi-trash" />
            </button>
          </div>
        </li>
      </ul>

      <form v-if="statuses.length < maxStatuses" class="row g-2 align-items-end mb-3" @submit.prevent="addStatus">
        <div class="col-sm-6">
          <label class="form-label small mb-0" :for="`new-status-${project.id}`">Add status</label>
          <input
            :id="`new-status-${project.id}`"
            v-model="newStatusName"
            type="text"
            class="form-control form-control-sm"
            maxlength="40"
            required
            placeholder="e.g. In review"
          />
        </div>
        <div class="col-sm-3">
          <div class="form-check mt-3">
            <input id="new-status-done" v-model="newStatusDone" class="form-check-input" type="checkbox" />
            <label class="form-check-label small" for="new-status-done">Done column</label>
          </div>
        </div>
        <div class="col-sm-3">
          <button class="btn btn-sm btn-primary w-100" type="submit" :disabled="adding">Add</button>
        </div>
      </form>
      <p v-else class="small text-muted mb-3">Maximum of {{ maxStatuses }} statuses reached.</p>

      <div v-if="deleteTarget" class="border rounded p-2 bg-body mb-2">
        <p class="small mb-2">
          Delete <strong>{{ deleteTarget.name }}</strong>?
          <span v-if="moveOptions.length">
            Tasks in this column (if any) will move to the status you select.
          </span>
        </p>
        <div v-if="moveOptions.length" class="mb-2">
          <label class="form-label small mb-0" for="move-to-status">Move tasks to</label>
          <select id="move-to-status" v-model.number="moveToStatusId" class="form-select form-select-sm">
            <option :value="0" disabled>Select status…</option>
            <option v-for="opt in moveOptions" :key="opt.id" :value="opt.id">{{ opt.name }}</option>
          </select>
        </div>
        <div class="d-flex gap-2">
          <button class="btn btn-sm btn-danger" type="button" :disabled="deleting" @click="confirmDelete">
            Delete
          </button>
          <button class="btn btn-sm btn-outline-secondary" type="button" @click="cancelDelete">Cancel</button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { Project, ProjectStatus } from '@/api/types'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{ changed: [] }>()

const maxStatuses = 8
const statuses = ref<ProjectStatus[]>([])
const savingMode = ref(false)
const adding = ref(false)
const deleting = ref(false)
const newStatusName = ref('')
const newStatusDone = ref(false)
const renameId = ref<number | null>(null)
const renameValue = ref('')
const deleteTarget = ref<ProjectStatus | null>(null)
const moveToStatusId = ref(0)
const toast = useToast()

const isKanban = computed(() => (props.project.workflow_mode || 'classic') === 'kanban')
const moveOptions = computed(() =>
  statuses.value.filter((s) => s.id !== deleteTarget.value?.id),
)

async function loadStatuses() {
  if (!isKanban.value) {
    statuses.value = []
    return
  }
  try {
    statuses.value = await api.listProjectStatuses(props.project.id)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load statuses', 'error')
  }
}

async function onToggleMode(enabled: boolean) {
  savingMode.value = true
  try {
    await api.updateProject(props.project.id, {
      workflow_mode: enabled ? 'kanban' : 'classic',
    })
    toast.push(enabled ? 'Kanban board enabled' : 'Board mode disabled', 'success')
    emit('changed')
    if (enabled) await loadStatuses()
    else statuses.value = []
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not update workflow mode', 'error')
  } finally {
    savingMode.value = false
  }
}

function beginRename(s: ProjectStatus) {
  renameId.value = s.id
  renameValue.value = s.name
  deleteTarget.value = null
}

async function saveRename(s: ProjectStatus) {
  const name = renameValue.value.trim()
  if (!name) return
  try {
    await api.updateProjectStatus(props.project.id, s.id, { name })
    renameId.value = null
    toast.push('Status renamed', 'success')
    await loadStatuses()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Rename failed', 'error')
  }
}

async function setDefault(s: ProjectStatus) {
  if (s.is_default) return
  try {
    await api.updateProjectStatus(props.project.id, s.id, { is_default: true })
    toast.push('Default status updated', 'success')
    await loadStatuses()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Update failed', 'error')
    await loadStatuses()
  }
}

async function toggleDone(s: ProjectStatus, isDone: boolean) {
  try {
    await api.updateProjectStatus(props.project.id, s.id, { is_done: isDone })
    toast.push('Status updated', 'success')
    await loadStatuses()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Update failed', 'error')
    await loadStatuses()
  }
}

async function addStatus() {
  const name = newStatusName.value.trim()
  if (!name) return
  adding.value = true
  try {
    await api.createProjectStatus(props.project.id, {
      name,
      is_done: newStatusDone.value,
    })
    newStatusName.value = ''
    newStatusDone.value = false
    toast.push('Status added', 'success')
    await loadStatuses()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not add status', 'error')
  } finally {
    adding.value = false
  }
}

function askDelete(s: ProjectStatus) {
  deleteTarget.value = s
  renameId.value = null
  const first = statuses.value.find((x) => x.id !== s.id)
  moveToStatusId.value = first?.id ?? 0
}

function cancelDelete() {
  deleteTarget.value = null
  moveToStatusId.value = 0
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const moveTo = moveToStatusId.value > 0 ? moveToStatusId.value : undefined
    await api.deleteProjectStatus(props.project.id, deleteTarget.value.id, moveTo)
    toast.push('Status deleted', 'info')
    cancelDelete()
    await loadStatuses()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Delete failed', 'error')
  } finally {
    deleting.value = false
  }
}

watch(
  () => [props.project.id, props.project.workflow_mode] as const,
  () => {
    cancelDelete()
    renameId.value = null
    void loadStatuses()
  },
)

onMounted(loadStatuses)
</script>
