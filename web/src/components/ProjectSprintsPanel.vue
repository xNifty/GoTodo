<template>
  <div v-if="isKanban" class="sprints-panel">
    <h4 class="h6 mb-2">Sprints</h4>
    <p class="small text-muted mb-3">
      Name a sprint and give it a date range. The board can switch between sprints; tasks with no
      sprint stay in the backlog.
    </p>

    <ul class="list-unstyled mb-3">
      <li
        v-for="s in sprints"
        :key="s.id"
        class="d-flex flex-wrap align-items-start gap-2 mb-2 pb-2 border-bottom"
      >
        <template v-if="editId === s.id">
          <form class="d-flex flex-column gap-2 flex-grow-1" @submit.prevent="saveEdit(s)">
            <input
              v-model="editName"
              type="text"
              class="form-control form-control-sm"
              maxlength="60"
              required
              aria-label="Sprint name"
            />
            <div class="d-flex flex-wrap gap-2">
              <input
                v-model="editStart"
                type="date"
                class="form-control form-control-sm"
                required
                aria-label="Sprint start date"
              />
              <input
                v-model="editEnd"
                type="date"
                class="form-control form-control-sm"
                required
                aria-label="Sprint end date"
              />
            </div>
            <div class="d-flex gap-1">
              <button class="btn btn-sm btn-primary" type="submit" :disabled="saving">Save</button>
              <button class="btn btn-sm btn-secondary" type="button" @click="editId = null">Cancel</button>
            </div>
          </form>
        </template>
        <template v-else>
          <div class="min-w-0 flex-grow-1">
            <div class="d-flex align-items-center gap-1 flex-wrap">
              <strong>{{ s.name }}</strong>
              <span v-if="s.is_active" class="badge text-bg-success">active</span>
              <button
                v-if="isOwner"
                class="btn btn-sm btn-link p-0"
                type="button"
                aria-label="Edit sprint"
                @click="beginEdit(s)"
              >
                <i class="bi bi-pencil" />
              </button>
            </div>
            <div class="small text-muted">
              {{ formatRange(s.start_date, s.end_date) }}
              · {{ s.task_count }} task{{ s.task_count === 1 ? '' : 's' }}
            </div>
          </div>
          <button
            v-if="isOwner"
            class="btn btn-sm btn-outline-danger"
            type="button"
            :disabled="deletingId === s.id"
            aria-label="Delete sprint"
            @click="removeSprint(s)"
          >
            <i class="bi bi-trash" />
          </button>
        </template>
      </li>
    </ul>

    <p v-if="!sprints.length" class="small text-muted mb-3">No sprints yet.</p>

    <form v-if="isOwner" class="row g-2 align-items-end" @submit.prevent="addSprint">
      <div class="col-12">
        <label class="form-label small mb-0" :for="`new-sprint-name-${project.id}`">New sprint</label>
        <input
          :id="`new-sprint-name-${project.id}`"
          v-model="newName"
          type="text"
          class="form-control form-control-sm"
          maxlength="60"
          required
          placeholder="Sprint name"
        />
      </div>
      <div class="col-sm-6">
        <label class="form-label small mb-0" :for="`new-sprint-start-${project.id}`">Starts</label>
        <input
          :id="`new-sprint-start-${project.id}`"
          v-model="newStart"
          type="date"
          class="form-control form-control-sm"
          required
        />
      </div>
      <div class="col-sm-6">
        <label class="form-label small mb-0" :for="`new-sprint-end-${project.id}`">Ends</label>
        <input
          :id="`new-sprint-end-${project.id}`"
          v-model="newEnd"
          type="date"
          class="form-control form-control-sm"
          required
        />
      </div>
      <div class="col-12">
        <button class="btn btn-sm btn-primary" type="submit" :disabled="adding">Add sprint</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { Project, ProjectSprint } from '@/api/types'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  project: Project
}>()

const emit = defineEmits<{
  changed: []
}>()

const toast = useToast()
const sprints = ref<ProjectSprint[]>([])
const adding = ref(false)
const saving = ref(false)
const deletingId = ref<number | null>(null)
const newName = ref('')
const newStart = ref('')
const newEnd = ref('')
const editId = ref<number | null>(null)
const editName = ref('')
const editStart = ref('')
const editEnd = ref('')

const isKanban = computed(() => (props.project.workflow_mode || 'classic') === 'kanban')
const isOwner = computed(() => (props.project.role || 'owner') === 'owner')

function formatRange(start: string, end: string) {
  return `${start} – ${end}`
}

async function loadSprints() {
  if (!isKanban.value) {
    sprints.value = []
    return
  }
  try {
    sprints.value = await api.listProjectSprints(props.project.id)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load sprints', 'error')
    sprints.value = []
  }
}

function beginEdit(s: ProjectSprint) {
  editId.value = s.id
  editName.value = s.name
  editStart.value = s.start_date
  editEnd.value = s.end_date
}

async function addSprint() {
  if (!newName.value.trim() || !newStart.value || !newEnd.value) return
  adding.value = true
  try {
    await api.createProjectSprint(props.project.id, {
      name: newName.value.trim(),
      start_date: newStart.value,
      end_date: newEnd.value,
    })
    newName.value = ''
    newStart.value = ''
    newEnd.value = ''
    toast.push('Sprint created', 'success')
    await loadSprints()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not create sprint', 'error')
  } finally {
    adding.value = false
  }
}

async function saveEdit(s: ProjectSprint) {
  if (!editName.value.trim() || !editStart.value || !editEnd.value) return
  saving.value = true
  try {
    await api.updateProjectSprint(props.project.id, s.id, {
      name: editName.value.trim(),
      start_date: editStart.value,
      end_date: editEnd.value,
    })
    editId.value = null
    toast.push('Sprint updated', 'success')
    await loadSprints()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not update sprint', 'error')
  } finally {
    saving.value = false
  }
}

async function removeSprint(s: ProjectSprint) {
  deletingId.value = s.id
  try {
    await api.deleteProjectSprint(props.project.id, s.id)
    toast.push('Sprint deleted; tasks moved to backlog', 'success')
    await loadSprints()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not delete sprint', 'error')
  } finally {
    deletingId.value = null
  }
}

watch(
  () => [props.project.id, props.project.workflow_mode] as const,
  () => {
    void loadSprints()
  },
  { immediate: true },
)
</script>
