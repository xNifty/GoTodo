<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { Project, Tag, Task, TaskEvent } from '@/api/types'
import { APIError } from '@/api/types'
import { useTaskSidebar } from '@/composables/useTaskSidebar'
import { useToast } from '@/composables/useToast'
import { projectOptionLabel } from '@/utils/projectLabel'

const {
  open,
  mode,
  taskId,
  defaultDueDate,
  defaultProjectId,
  defaultParentId,
  defaultParentTitle,
  close,
  notifySaved,
} = useTaskSidebar()
const toast = useToast()

const loading = ref(false)
const saving = ref(false)
const projects = ref<Project[]>([])
const allTags = ref<Tag[]>([])
const rootTasks = ref<Task[]>([])
const events = ref<TaskEvent[]>([])
const eventsLoaded = ref(false)
const eventsLoading = ref(false)
const descriptionError = ref('')

const titleInput = ref<HTMLInputElement | null>(null)
const title = ref('')
const description = ref('')
const projectId = ref<number | ''>('')
const parentId = ref<number | ''>('')
const parentTitle = ref('')
const priority = ref(0)
const dueDate = ref('')
const selectedTagIds = ref<number[]>([])
const newTags = ref('')
const completed = ref(false)
const isSubtask = computed(() => parentId.value !== '' && Number(parentId.value) > 0)

const readOnly = computed(() => mode.value === 'view')
const sidebarTitle = computed(() => {
  if (mode.value === 'view') return 'View Task'
  if (mode.value === 'edit') return 'Edit Task'
  return 'Add Task'
})
const submitText = computed(() => (mode.value === 'edit' ? 'Save Task' : 'Add Task'))
const charCount = computed(() => description.value.length)

function resetForm() {
  title.value = ''
  description.value = ''
  projectId.value = ''
  parentId.value = ''
  parentTitle.value = ''
  priority.value = 0
  dueDate.value = ''
  selectedTagIds.value = []
  newTags.value = ''
  completed.value = false
  descriptionError.value = ''
  events.value = []
  eventsLoaded.value = false
}

async function loadMeta() {
  const [projs, tags, list] = await Promise.all([
    api.listProjects(),
    api.listTags(),
    api.listTasks({ page: 1, per_page: 100 }),
  ])
  projects.value = projs
  allTags.value = tags
  rootTasks.value = list.tasks.filter((t) => !t.parent_id)
}

async function loadTask(id: number) {
  loading.value = true
  try {
    const task = await api.getTask(id)
    title.value = task.title
    description.value = task.description || ''
    projectId.value = task.project_id ?? ''
    parentId.value = task.parent_id ?? ''
    parentTitle.value = ''
    if (task.parent_id) {
      const parent = rootTasks.value.find((t) => t.id === task.parent_id)
      parentTitle.value = parent?.title || `Task #${task.parent_id}`
    }
    priority.value = task.priority
    dueDate.value = task.due_date || ''
    selectedTagIds.value = task.tags?.map((t) => t.id) ?? []
    newTags.value = ''
    completed.value = task.completed
    descriptionError.value = ''
    events.value = []
    eventsLoaded.value = false
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load task', 'error')
    close()
  } finally {
    loading.value = false
  }
}

async function resolveTagIds(): Promise<number[]> {
  const ids = new Set(selectedTagIds.value)
  const parts = newTags.value
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
  for (const name of parts) {
    const existing = allTags.value.find((t) => t.name.toLowerCase() === name.toLowerCase())
    if (existing) {
      ids.add(existing.id)
      continue
    }
    const created = await api.createTag(name)
    allTags.value = [...allTags.value, created]
    ids.add(created.id)
  }
  const out = Array.from(ids)
  if (out.length > 5) {
    throw new Error('Maximum 5 tags per task')
  }
  return out
}

function validateDescription() {
  if (description.value.length > 1000) {
    descriptionError.value = 'Description must be 1000 characters or fewer.'
    return false
  }
  descriptionError.value = ''
  return true
}

async function save(keepOpen = false) {
  if (readOnly.value) return
  if (!title.value.trim()) return
  if (!validateDescription()) return
  saving.value = true
  try {
    const tagIds = await resolveTagIds()
    if (mode.value === 'add') {
      const created = await api.createTask({
        title: title.value.trim(),
        description: description.value,
        project_id: isSubtask.value
          ? undefined
          : projectId.value === ''
            ? null
            : Number(projectId.value),
        parent_id: isSubtask.value ? Number(parentId.value) : null,
        priority: priority.value,
        due_date: dueDate.value || undefined,
        tag_ids: tagIds,
      })
      notifySaved(created, !keepOpen)
      toast.push(isSubtask.value ? 'Subtask created' : 'Task created', 'success')

      if (keepOpen) {
        // Clear title & description for next task while keeping project/due/parent pre-selected
        title.value = ''
        description.value = ''
        newTags.value = ''
        descriptionError.value = ''
        await nextTick()
        titleInput.value?.focus()
      } else {
        resetForm()
      }
      return
    }
    if (!taskId.value) return
    const payload: Parameters<typeof api.patchTask>[1] = {
      title: title.value.trim(),
      description: description.value,
      priority: priority.value,
      tag_ids: tagIds,
      // 0 clears parent (promotes to root); JSON null is indistinguishable from omitted for **int.
      parent_id: isSubtask.value ? Number(parentId.value) : 0,
    }
    if (!isSubtask.value) {
      payload.project_id = projectId.value === '' ? null : Number(projectId.value)
    }
    if (dueDate.value) payload.due_date = dueDate.value
    else payload.clear_due_date = true
    const updated = await api.patchTask(taskId.value, payload)
    notifySaved(updated, true)
    toast.push('Task saved', 'success')
  } catch (err) {
    const msg = err instanceof APIError ? err.message : err instanceof Error ? err.message : 'Save failed'
    toast.push(msg, 'error')
  } finally {
    saving.value = false
  }
}

function applyDuePreset(preset: string) {
  const today = new Date()
  if (preset === 'today') {
    dueDate.value = today.toISOString().slice(0, 10)
  } else if (preset === 'tomorrow') {
    today.setDate(today.getDate() + 1)
    dueDate.value = today.toISOString().slice(0, 10)
  } else if (preset === 'week') {
    today.setDate(today.getDate() + 7)
    dueDate.value = today.toISOString().slice(0, 10)
  } else if (preset === 'clear') {
    dueDate.value = ''
  }
}

function toggleTag(id: number, checked: boolean) {
  if (checked) {
    if (!selectedTagIds.value.includes(id)) selectedTagIds.value = [...selectedTagIds.value, id]
  } else {
    selectedTagIds.value = selectedTagIds.value.filter((x) => x !== id)
  }
}

async function loadEvents() {
  if (!taskId.value || eventsLoaded.value || eventsLoading.value) return
  eventsLoading.value = true
  try {
    events.value = await api.listTaskEvents(taskId.value)
    eventsLoaded.value = true
  } catch {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

watch(description, validateDescription)

watch(
  () => ({
    isOpen: open.value,
    m: mode.value,
    id: taskId.value,
    due: defaultDueDate.value,
    proj: defaultProjectId.value,
    parent: defaultParentId.value,
    parentLabel: defaultParentTitle.value,
  }),
  async ({ isOpen, m, id, due, proj, parent, parentLabel }) => {
    if (!isOpen) return
    await loadMeta()
    if ((m === 'edit' || m === 'view') && id) {
      await loadTask(id)
    } else {
      resetForm()
      if (due) dueDate.value = due
      if (proj) projectId.value = Number(proj)
      if (parent) {
        parentId.value = parent
        parentTitle.value = parentLabel || rootTasks.value.find((t) => t.id === parent)?.title || ''
        const p = rootTasks.value.find((t) => t.id === parent)
        if (p?.project_id) projectId.value = p.project_id
      }
      await nextTick()
      titleInput.value?.focus()
    }
  },
  { immediate: true },
)

function onParentChange() {
  if (!isSubtask.value) {
    parentTitle.value = ''
    return
  }
  const p = rootTasks.value.find((t) => t.id === Number(parentId.value))
  parentTitle.value = p?.title || ''
  if (p?.project_id) projectId.value = p.project_id
  else projectId.value = ''
}
</script>

<template>
  <div
    id="sidebar-backdrop"
    class="sidebar-backdrop"
    :class="{ active: open }"
    :aria-hidden="!open"
    @click="close"
  />
  <div id="sidebar" :class="{ active: open }">
    <div class="sidebar-header">
      <button type="button" class="btn-close float-end" id="closeSidebar" aria-label="Close sidebar" @click="close" />
      <h5>{{ sidebarTitle }}</h5>
    </div>
    <div class="sidebar-body">
      <div v-if="loading" class="sidebar-loading" aria-busy="true">
        <div class="spinner-border text-primary" role="status">
          <span class="visually-hidden">Loading task…</span>
        </div>
        <p class="mb-0">Loading task…</p>
      </div>
      <form v-else id="newTaskForm" @submit.prevent="save(false)">
        <div class="form-group">
          <label for="title">Title:</label>
          <input
            id="title"
            ref="titleInput"
            v-model="title"
            type="text"
            class="form-control"
            :required="!readOnly"
            :readonly="readOnly"
            :disabled="readOnly"
            placeholder="Your Task Title"
          />
        </div>
        <div class="form-group">
          <div v-if="completed" class="alert alert-success py-2 mb-2">
            <i class="bi bi-check-circle" /> This task is completed
          </div>
          <label for="description">Description:</label>
          <textarea
            id="description"
            v-model="description"
            class="form-control"
            maxlength="1000"
            rows="4"
            :readonly="readOnly"
            :disabled="readOnly"
          />
          <div v-if="!readOnly" class="d-flex justify-content-between align-items-center mt-1">
            <small class="form-hint">Max 1000 Characters</small>
            <small class="text-muted"><span id="char-count">{{ charCount }}</span>/1000</small>
          </div>
          <div v-if="descriptionError" id="description-error" class="invalid-feedback d-block">
            {{ descriptionError }}
          </div>
        </div>
        <div class="form-group mt-2">
          <label for="parent_id">Parent task (optional):</label>
          <select
            id="parent_id"
            v-model="parentId"
            class="form-select"
            :disabled="readOnly || (mode === 'edit' && (rootTasks.find((t) => t.id === taskId)?.child_count || 0) > 0)"
            @change="onParentChange"
          >
            <option value="">No parent — top level</option>
            <option
              v-for="t in rootTasks.filter((r) => r.id !== taskId)"
              :key="t.id"
              :value="t.id"
            >
              {{ t.title }}
            </option>
          </select>
          <small v-if="isSubtask" class="form-hint d-block mt-1">
            Subtask of {{ parentTitle || 'selected parent' }}. Project is inherited.
          </small>
        </div>
        <div class="form-group mt-2">
          <label for="project_id">Project (optional):</label>
          <select id="project_id" v-model="projectId" class="form-select" :disabled="readOnly || isSubtask">
            <option value="">No project</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ projectOptionLabel(p) }}</option>
          </select>
        </div>
        <div class="form-group mt-2">
          <label for="priority">Priority:</label>
          <select id="priority" v-model.number="priority" class="form-select" :disabled="readOnly">
            <option :value="0">None</option>
            <option :value="1">Low</option>
            <option :value="2">Medium</option>
            <option :value="3">High</option>
          </select>
        </div>
        <div class="form-group mt-2">
          <label for="due_date">Due Date (optional):</label>
          <input id="due_date" v-model="dueDate" type="date" class="form-control" :disabled="readOnly" :readonly="readOnly" />
          <div v-if="!readOnly" class="btn-group btn-group-sm mt-1" role="group" aria-label="Due date presets">
            <button type="button" class="btn btn-outline-secondary" @click="applyDuePreset('today')">Today</button>
            <button type="button" class="btn btn-outline-secondary" @click="applyDuePreset('tomorrow')">Tomorrow</button>
            <button type="button" class="btn btn-outline-secondary" @click="applyDuePreset('week')">+1 week</button>
            <button type="button" class="btn btn-outline-secondary" @click="applyDuePreset('clear')">Clear</button>
          </div>
        </div>
        <div v-if="allTags.length" class="form-group mt-2">
          <label>Tags (max 5)</label>
          <div v-for="tag in allTags" :key="tag.id" class="form-check">
            <input
              :id="`tag-${tag.id}`"
              type="checkbox"
              class="form-check-input"
              :checked="selectedTagIds.includes(tag.id)"
              :disabled="readOnly"
              @change="toggleTag(tag.id, ($event.target as HTMLInputElement).checked)"
            />
            <label class="form-check-label" :for="`tag-${tag.id}`">
              <span class="tag-chip" :style="{ backgroundColor: tag.color || '#6c757d' }">{{ tag.name }}</span>
            </label>
          </div>
        </div>
        <div v-if="!readOnly" class="form-group mt-2">
          <label for="new_tags">Add tags (comma-separated)</label>
          <input
            id="new_tags"
            v-model="newTags"
            type="text"
            class="form-control"
            placeholder="e.g. work, urgent"
            maxlength="200"
          />
          <small class="form-hint">New tag names are created on save (max 5 tags per task).</small>
        </div>

        <!-- Form Submit Action Buttons -->
        <div v-if="!readOnly" class="d-flex gap-2 mt-3">
          <button type="submit" class="btn btn-primary flex-grow-1" :disabled="saving">
            {{ saving ? 'Saving…' : submitText }}
          </button>
          <button
            v-if="mode === 'add'"
            type="button"
            class="btn btn-outline-primary flex-grow-1"
            :disabled="saving || !title.trim()"
            @click="save(true)"
          >
            Save &amp; Add Another
          </button>
        </div>

        <details
          v-if="mode === 'edit' || mode === 'view'"
          class="task-timeline mt-3"
          @toggle="(e) => { if ((e.target as HTMLDetailsElement).open) void loadEvents() }"
        >
          <summary class="mb-2">Activity</summary>
          <div class="task-timeline-body">
            <p v-if="eventsLoading" class="text-muted small mb-0">Loading activity…</p>
            <ul v-else-if="events.length" class="list-unstyled small mb-0">
              <li v-for="ev in events" :key="ev.id" class="mb-1">
                <span class="text-muted">{{ ev.created_at }}</span> — {{ ev.label }}
              </li>
            </ul>
            <p v-else class="text-muted small mb-0">No activity recorded.</p>
          </div>
        </details>
      </form>
    </div>
  </div>
</template>
