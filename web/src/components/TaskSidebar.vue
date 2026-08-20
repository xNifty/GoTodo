<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { Project, ProjectStatus, Tag, Task, TaskEvent, TaskGitHubIssue, TaskTimeEntry } from '@/api/types'
import { APIError } from '@/api/types'
import ParentTaskCombobox from '@/components/ParentTaskCombobox.vue'
import { useAuth } from '@/composables/useAuth'
import { useTaskSidebar } from '@/composables/useTaskSidebar'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
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
  openEdit,
  openView,
} = useTaskSidebar()
const toast = useToast()
const { askConfirm } = useConfirm()
const { user } = useAuth()
const router = useRouter()
const showTaskNumber = computed(() => (mode.value === 'edit' || mode.value === 'view') && !!taskId.value)

async function copyPermalink() {
  if (!taskId.value) return
  const href = router.resolve({ name: 'task', params: { id: String(taskId.value) } }).href
  const url = new URL(href, window.location.origin).href
  try {
    await navigator.clipboard.writeText(url)
    toast.push('Copied', 'success')
  } catch {
    toast.push(url, 'info')
  }
}

const loading = ref(false)
const saving = ref(false)
const projects = ref<Project[]>([])
const allTags = ref<Tag[]>([])
const rootTasks = ref<Task[]>([])
const currentTask = ref<Task | null>(null)
const events = ref<TaskEvent[]>([])
const eventsLoaded = ref(false)
const eventsLoading = ref(false)
const descriptionError = ref('')
const statuses = ref<ProjectStatus[]>([])
const timeEntries = ref<TaskTimeEntry[]>([])
const timeSpentMinutes = ref(0)
const newEntryMinutes = ref<number | ''>('')
const newEntryNote = ref('')
const addingTime = ref(false)
const taskWorkflow = ref('')

const titleInput = ref<HTMLInputElement | null>(null)
const descriptionInput = ref<HTMLTextAreaElement | null>(null)
const title = ref('')
const description = ref('')
const projectId = ref<number | ''>('')
const parentId = ref<number | ''>('')
const parentTitle = ref('')
const priority = ref(0)
const dueDate = ref('')
const selectedTagIds = ref<number[]>([])
const taskTags = ref<Tag[]>([])
const newTags = ref('')
const completed = ref(false)
const statusId = ref<number | ''>('')
const estimatePoints = ref<number | ''>('')
const claimedBy = ref<number | null>(null)
const claimedByName = ref('')
const claiming = ref(false)
const githubIssue = ref<TaskGitHubIssue | null>(null)
const projectHasGitHub = ref(false)
const githubIssueRef = ref('')
const githubBusy = ref(false)
const isSubtask = computed(() => parentId.value !== '' && Number(parentId.value) > 0)
const selectedProject = computed(() => {
  if (projectId.value === '') return null
  return projects.value.find((p) => p.id === Number(projectId.value)) ?? null
})
/** Viewers may open via edit entry points; treat their project role as read-only. */
const readOnly = computed(() => {
  if (mode.value === 'view') return true
  if (mode.value === 'edit' && selectedProject.value?.role === 'viewer') return true
  return false
})
const canManageTags = computed(() => {
  if (readOnly.value) return false
  const role = selectedProject.value?.role
  if (!role) return true
  return role === 'owner' || role === 'editor'
})
const parentOptions = computed(() =>
  rootTasks.value.filter((r) => r.id !== taskId.value && taskInSelectedProject(r)),
)
const parentPickerDisabled = computed(
  () =>
    readOnly.value ||
    (mode.value === 'edit' && (currentTask.value?.child_count || currentTask.value?.children?.length || 0) > 0),
)
const relatedParent = computed(() => {
  if (!isSubtask.value) return null
  const id = Number(parentId.value)
  return { id, title: parentTitle.value || rootTasks.value.find((t) => t.id === id)?.title || `Task #${id}` }
})
const relatedChildren = computed(() => currentTask.value?.children ?? [])
const isKanbanTask = computed(() => {
  if (taskWorkflow.value === 'kanban') return true
  return (selectedProject.value?.workflow_mode || 'classic') === 'kanban'
})
const sidebarTitle = computed(() => {
  if (readOnly.value) return 'View Task'
  if (mode.value === 'edit') return 'Edit Task'
  return 'Add Task'
})
const submitText = computed(() => (mode.value === 'edit' ? 'Save Task' : 'Add Task'))
const charCount = computed(() => description.value.length)
const timeSpentLabel = computed(() => formatMinutes(timeSpentMinutes.value))
const claimerLabel = computed(() => {
  if (!claimedBy.value) return 'Unclaimed'
  if (user.value?.id && claimedBy.value === user.value.id) return 'You'
  return claimedByName.value || `User #${claimedBy.value}`
})

function formatMinutes(total: number) {
  if (total <= 0) return '0m'
  const h = Math.floor(total / 60)
  const m = total % 60
  if (h <= 0) return `${m}m`
  if (m <= 0) return `${h}h`
  return `${h}h ${m}m`
}

function taskInSelectedProject(t: Task): boolean {
  if (projectId.value === '') return t.project_id == null || t.project_id === 0
  return t.project_id === Number(projectId.value)
}

function stubParentTask(id: number, titleText: string, pid: number | ''): Task {
  return {
    id,
    title: titleText,
    description: '',
    completed: false,
    due_date: '',
    project_id: pid === '' ? null : Number(pid),
    priority: 0,
    favorite: false,
    position: 0,
    tags: [],
    created_at: '',
    modified_at: '',
  }
}

const DESCRIPTION_MIN_HEIGHT = 80

function autosizeDescription() {
  const el = descriptionInput.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.max(el.scrollHeight, DESCRIPTION_MIN_HEIGHT)}px`
}

function resetForm() {
  title.value = ''
  description.value = ''
  projectId.value = ''
  parentId.value = ''
  parentTitle.value = ''
  currentTask.value = null
  priority.value = 0
  dueDate.value = ''
  selectedTagIds.value = []
  taskTags.value = []
  newTags.value = ''
  completed.value = false
  descriptionError.value = ''
  events.value = []
  eventsLoaded.value = false
  statusId.value = ''
  estimatePoints.value = ''
  claimedBy.value = null
  claimedByName.value = ''
  statuses.value = []
  timeEntries.value = []
  timeSpentMinutes.value = 0
  newEntryMinutes.value = ''
  newEntryNote.value = ''
  taskWorkflow.value = ''
  githubIssue.value = null
  projectHasGitHub.value = false
  githubIssueRef.value = ''
}

async function loadMeta() {
  projects.value = await api.listProjects()
  await loadTagsForProject(projectId.value)
}

async function loadTagsForProject(pid: number | '') {
  const project_id = pid === '' || pid == null ? 0 : Number(pid)
  try {
    allTags.value = await api.listTags({ project_id })
  } catch {
    allTags.value = []
  }
}

async function loadParentCandidates(pid: number | '', keepParentId: number | null = null) {
  const list = await api.listTasks({
    page: 1,
    per_page: 100,
    workflow_claim_scope: 'all',
    project: pid === '' || pid == null ? 0 : pid,
  })
  let loaded = list.tasks.filter((t) => !t.parent_id)
  if (keepParentId && !loaded.some((t) => t.id === keepParentId)) {
    const cached = rootTasks.value.find((t) => t.id === keepParentId && !t.parent_id)
    if (cached) {
      loaded = [cached, ...loaded]
    } else {
      try {
        const parent = await api.getTask(keepParentId)
        if (!parent.parent_id) {
          loaded = [parent, ...loaded]
        }
      } catch {
        if (parentTitle.value) {
          loaded = [stubParentTask(keepParentId, parentTitle.value, pid), ...loaded]
        }
      }
    }
  }
  rootTasks.value = loaded
}

async function loadStatusesForProject(pid: number | '') {
  if (pid === '' || pid == null) {
    statuses.value = []
    return
  }
  const proj = projects.value.find((p) => p.id === Number(pid))
  const kanban =
    taskWorkflow.value === 'kanban' || (proj?.workflow_mode || 'classic') === 'kanban'
  if (!kanban) {
    statuses.value = []
    return
  }
  try {
    statuses.value = await api.listProjectStatuses(Number(pid))
    if (statusId.value === '' && mode.value === 'add') {
      const def = statuses.value.find((s) => s.is_default)
      if (def) statusId.value = def.id
    }
  } catch {
    statuses.value = []
  }
}

async function loadTimeEntries(id: number) {
  try {
    const entries = await api.listTimeEntries(id)
    timeEntries.value = entries
    timeSpentMinutes.value = entries.reduce((sum, e) => sum + e.minutes, 0)
  } catch {
    timeEntries.value = []
  }
}

async function refreshProjectGitHub(pid: number | '') {
  projectHasGitHub.value = false
  if (pid === '' || pid == null) return
  try {
    const link = await api.getProjectGitHub(Number(pid))
    projectHasGitHub.value = !!link.linked
  } catch {
    projectHasGitHub.value = false
  }
}

async function resolveParentTitle(parentTaskId: number) {
  const cached = rootTasks.value.find((t) => t.id === parentTaskId)
  if (cached) {
    parentTitle.value = cached.title
    return
  }
  try {
    const parent = await api.getTask(parentTaskId)
    parentTitle.value = parent.title
    if (!parent.parent_id && !rootTasks.value.some((t) => t.id === parent.id)) {
      rootTasks.value = [parent, ...rootTasks.value]
    }
  } catch {
    parentTitle.value = `Task #${parentTaskId}`
  }
}

async function loadTask(id: number) {
  const task = await api.getTask(id)
  currentTask.value = task
  title.value = task.title
  description.value = task.description || ''
  projectId.value = task.project_id ?? ''
  parentId.value = task.parent_id ?? ''
  parentTitle.value = ''
  await loadParentCandidates(projectId.value, task.parent_id ?? null)
  if (task.parent_id) {
    await resolveParentTitle(task.parent_id)
  }
  priority.value = task.priority
  dueDate.value = task.due_date || ''
  taskTags.value = task.tags ? [...task.tags] : []
  selectedTagIds.value = taskTags.value.map((t) => t.id)
  newTags.value = ''
  completed.value = task.completed
  descriptionError.value = ''
  events.value = []
  eventsLoaded.value = false
  statusId.value = task.status_id ?? ''
  estimatePoints.value = task.estimate_points ?? ''
  claimedBy.value = task.claimed_by ?? null
  claimedByName.value = task.claimed_by_name || ''
  timeSpentMinutes.value = task.time_spent_minutes ?? 0
  taskWorkflow.value = task.project_workflow || ''
  githubIssue.value = task.github ?? null
  githubIssueRef.value = ''
  await loadStatusesForProject(projectId.value)
  await loadTagsForProject(projectId.value)
  await refreshProjectGitHub(projectId.value)
  if (taskWorkflow.value === 'kanban' || isKanbanTask.value) {
    await loadTimeEntries(id)
  } else {
    timeEntries.value = []
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
    const created = await api.createTag(name, projectId.value === '' ? null : Number(projectId.value))
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
        ...(isKanbanTask.value && statusId.value !== ''
          ? { status_id: Number(statusId.value) }
          : {}),
        ...(isKanbanTask.value && estimatePoints.value !== ''
          ? { estimate_points: Number(estimatePoints.value) }
          : {}),
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
        autosizeDescription()
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
    if (isKanbanTask.value) {
      if (statusId.value !== '') payload.status_id = Number(statusId.value)
      payload.estimate_points =
        estimatePoints.value === '' ? null : Number(estimatePoints.value)
    }
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

async function onProjectChange() {
  const names = new Set(
    [...taskTags.value, ...allTags.value]
      .filter((t) => selectedTagIds.value.includes(t.id))
      .map((t) => t.name.toLowerCase()),
  )
  taskWorkflow.value = ''
  if (statusId.value !== '' && !statuses.value.some((s) => s.id === statusId.value)) {
    statusId.value = ''
  }
  await loadStatusesForProject(projectId.value)
  await refreshProjectGitHub(projectId.value)
  await loadTagsForProject(projectId.value)
  selectedTagIds.value = allTags.value.filter((t) => names.has(t.name.toLowerCase())).map((t) => t.id)
  const previousParent = isSubtask.value ? Number(parentId.value) : null
  await loadParentCandidates(projectId.value)
  if (previousParent && !parentOptions.value.some((t) => t.id === previousParent)) {
    parentId.value = ''
    parentTitle.value = ''
  }
}

let loadSeq = 0

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
    if (!isOpen) {
      loading.value = false
      return
    }
    // Show spinner immediately so the modal never flashes empty/stale form content.
    const seq = ++loadSeq
    loading.value = true
    try {
      await loadMeta()
      if (seq !== loadSeq) return
      if ((m === 'edit' || m === 'view') && id) {
        await loadTask(id)
      } else {
        resetForm()
        if (due) dueDate.value = due
        if (proj) projectId.value = Number(proj)
        if (parent) {
          parentId.value = parent
          parentTitle.value = parentLabel || ''
          await loadParentCandidates(projectId.value, parent)
          const p = rootTasks.value.find((t) => t.id === parent)
          parentTitle.value = parentLabel || p?.title || ''
          const inheritedProject = p?.project_id ?? ''
          if (inheritedProject !== projectId.value) {
            projectId.value = inheritedProject
            await loadParentCandidates(projectId.value, parent)
          }
        } else {
          await loadParentCandidates(projectId.value)
        }
        await loadStatusesForProject(projectId.value)
        await loadTagsForProject(projectId.value)
        await refreshProjectGitHub(projectId.value)
      }
    } catch (err) {
      if (seq !== loadSeq) return
      toast.push(err instanceof APIError ? err.message : 'Failed to load task', 'error')
      close()
      return
    } finally {
      if (seq === loadSeq) loading.value = false
    }
    if (seq !== loadSeq || !open.value) return
    await nextTick()
    autosizeDescription()
    if (m === 'add') {
      titleInput.value?.focus()
    }
  },
  { immediate: true },
)

async function onParentChange() {
  if (!isSubtask.value) {
    parentTitle.value = ''
    return
  }
  const p = rootTasks.value.find((t) => t.id === Number(parentId.value))
  parentTitle.value = p?.title || ''
  if (p?.project_id) projectId.value = p.project_id
  else projectId.value = ''
  taskWorkflow.value = p?.project_workflow || ''
  await loadStatusesForProject(projectId.value)
  await refreshProjectGitHub(projectId.value)
}

function openRelated(id: number) {
  if (mode.value === 'view') openView(id)
  else openEdit(id)
}

async function claimCurrentTask() {
  if (!taskId.value || readOnly.value) return
  claiming.value = true
  try {
    const task = await api.claimTask(taskId.value)
    claimedBy.value = task.claimed_by ?? null
    claimedByName.value = task.claimed_by_name || ''
    toast.push('Task claimed', 'success')
    notifySaved(task, false)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not claim task', 'error')
  } finally {
    claiming.value = false
  }
}

async function unclaimCurrentTask() {
  if (!taskId.value || readOnly.value) return
  claiming.value = true
  try {
    const task = await api.unclaimTask(taskId.value)
    claimedBy.value = task.claimed_by ?? null
    claimedByName.value = task.claimed_by_name || ''
    toast.push('Task released', 'success')
    notifySaved(task, false)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not release task', 'error')
  } finally {
    claiming.value = false
  }
}

async function createGitHubIssue() {
  if (!taskId.value || readOnly.value) return
  githubBusy.value = true
  try {
    githubIssue.value = await api.createTaskGitHubIssue(taskId.value)
    toast.push(`Created GitHub issue #${githubIssue.value.issue_number}`, 'success')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not create GitHub issue', 'error')
  } finally {
    githubBusy.value = false
  }
}

async function linkGitHubIssue() {
  if (!taskId.value || readOnly.value || !githubIssueRef.value.trim()) return
  githubBusy.value = true
  try {
    githubIssue.value = await api.linkTaskGitHubIssue(taskId.value, githubIssueRef.value.trim())
    githubIssueRef.value = ''
    toast.push(`Linked GitHub issue #${githubIssue.value.issue_number}`, 'success')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not link GitHub issue', 'error')
  } finally {
    githubBusy.value = false
  }
}

async function unlinkGitHubIssue() {
  if (!taskId.value || readOnly.value) return
  const ok = await askConfirm({
    title: 'Unlink GitHub issue?',
    message: 'The GitHub issue itself is not deleted.',
    confirmLabel: 'Unlink',
    danger: true,
  })
  if (!ok) return
  githubBusy.value = true
  try {
    await api.unlinkTaskGitHubIssue(taskId.value)
    githubIssue.value = null
    toast.push('GitHub issue unlinked', 'info')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not unlink GitHub issue', 'error')
  } finally {
    githubBusy.value = false
  }
}

async function addTimeEntry() {
  if (!taskId.value || readOnly.value) return
  const minutes = Number(newEntryMinutes.value)
  if (!Number.isFinite(minutes) || minutes <= 0) {
    toast.push('Enter minutes greater than 0', 'error')
    return
  }
  addingTime.value = true
  try {
    await api.addTimeEntry(taskId.value, minutes, newEntryNote.value.trim())
    newEntryMinutes.value = ''
    newEntryNote.value = ''
    toast.push('Time logged', 'success')
    await loadTimeEntries(taskId.value)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not add time entry', 'error')
  } finally {
    addingTime.value = false
  }
}

async function removeTimeEntry(entryId: number) {
  if (!taskId.value || readOnly.value) return
  const ok = await askConfirm({
    title: 'Delete time entry?',
    message: 'Remove this time entry?',
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  try {
    await api.deleteTimeEntry(taskId.value, entryId)
    toast.push('Time entry deleted', 'info')
    await loadTimeEntries(taskId.value)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not delete time entry', 'error')
  }
}
</script>

<template>
  <div
    v-if="open"
    id="taskModal"
    class="modal show d-block"
    style="background: rgba(0,0,0,0.5);"
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    @click.self="close"
  >
    <div class="modal-dialog modal-dialog-centered modal-lg modal-dialog-scrollable">
      <div
        class="modal-content border-0 shadow"
        style="background: var(--ordryn-card-bg); color: var(--ordryn-text);"
      >
        <div class="modal-header border-0 pb-0">
          <h5 class="modal-title fw-bold d-flex align-items-center gap-2 flex-wrap">
            {{ sidebarTitle }}
            <span v-if="showTaskNumber" class="text-muted fw-normal">#{{ taskId }}</span>
            <button
              v-if="showTaskNumber"
              type="button"
              class="btn btn-sm btn-outline-secondary"
              title="Copy link"
              aria-label="Copy task link"
              @click="copyPermalink"
            >
              <i class="bi bi-link-45deg" /> Copy link
            </button>
          </h5>
          <button type="button" class="btn-close" id="closeSidebar" aria-label="Close" @click="close" />
        </div>
        <div class="modal-body py-3">
      <div v-if="loading" class="d-flex flex-column align-items-center justify-content-center gap-2 py-5" aria-busy="true">
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
            ref="descriptionInput"
            v-model="description"
            class="form-control task-description-input"
            maxlength="1000"
            rows="4"
            :readonly="readOnly"
            :disabled="readOnly"
            @input="autosizeDescription"
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
          <ParentTaskCombobox
            v-model="parentId"
            input-id="parent_id"
            :options="parentOptions"
            :disabled="parentPickerDisabled"
            placeholder="Type to search parent tasks…"
            @change="onParentChange"
          />
          <small v-if="isSubtask" class="form-hint d-block mt-1">
            Subtask of {{ parentTitle || 'selected parent' }}. Project is inherited.
          </small>
        </div>
        <div
          v-if="(mode === 'edit' || mode === 'view') && (relatedParent || relatedChildren.length)"
          class="form-group mt-2"
        >
          <label class="d-block">Related tasks</label>
          <div v-if="relatedParent" class="mb-1">
            <span class="form-hint d-block">Parent</span>
            <button
              type="button"
              class="btn btn-link text-start text-decoration-none p-0"
              @click="openRelated(relatedParent.id)"
            >
              {{ relatedParent.title }}
            </button>
          </div>
          <div v-if="relatedChildren.length">
            <span class="form-hint d-block">Subtasks</span>
            <ul class="list-unstyled mb-0">
              <li v-for="child in relatedChildren" :key="child.id">
                <button
                  type="button"
                  class="btn btn-link text-start text-decoration-none p-0"
                  @click="openRelated(child.id)"
                >
                  {{ child.title }}
                </button>
              </li>
            </ul>
          </div>
        </div>
        <div class="form-group mt-2">
          <label for="project_id">Project (optional):</label>
          <select
            id="project_id"
            v-model="projectId"
            class="form-select"
            :disabled="readOnly || isSubtask"
            @change="onProjectChange"
          >
            <option value="">No project</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ projectOptionLabel(p) }}</option>
          </select>
        </div>
        <div v-if="isKanbanTask" class="form-group mt-2">
          <label for="status_id">Status:</label>
          <select id="status_id" v-model="statusId" class="form-select" :disabled="readOnly">
            <option v-if="!statuses.length" value="">No statuses</option>
            <option v-for="s in statuses" :key="s.id" :value="s.id">
              {{ s.name }}{{ s.is_done ? ' (done)' : '' }}{{ s.is_default ? ' (default)' : '' }}
            </option>
          </select>
        </div>
        <div v-if="isKanbanTask && (mode === 'edit' || mode === 'view')" class="form-group mt-2">
          <label class="d-block">Claimed by</label>
          <div class="d-flex align-items-center gap-2 flex-wrap">
            <span class="badge border" :class="claimedBy ? 'text-bg-primary' : 'text-bg-light text-muted'">
              <i class="bi bi-person me-1" />{{ claimerLabel }}
            </span>
            <template v-if="!readOnly && taskId">
              <button
                v-if="!claimedBy || claimedBy !== user?.id"
                type="button"
                class="btn btn-sm btn-outline-primary"
                :disabled="claiming"
                @click="claimCurrentTask"
              >
                {{ claimedBy ? 'Take over' : 'Claim' }}
              </button>
              <button
                v-else
                type="button"
                class="btn btn-sm btn-outline-secondary"
                :disabled="claiming"
                @click="unclaimCurrentTask"
              >
                Release
              </button>
            </template>
          </div>
        </div>
        <div v-if="isKanbanTask" class="form-group mt-2">
          <label for="estimate_points">Estimate (points):</label>
          <input
            id="estimate_points"
            v-model="estimatePoints"
            type="number"
            class="form-control"
            min="0"
            max="100"
            step="1"
            :disabled="readOnly"
            :readonly="readOnly"
            placeholder="0–100"
          />
        </div>
        <div
          v-if="(mode === 'edit' || mode === 'view') && projectHasGitHub"
          class="form-group mt-3"
        >
          <label class="d-block">GitHub issue</label>
          <div v-if="githubIssue" class="d-flex flex-column gap-2">
            <div class="d-flex flex-wrap align-items-center gap-2">
              <a
                :href="githubIssue.issue_url"
                target="_blank"
                rel="noopener noreferrer"
                class="fw-semibold"
              >
                #{{ githubIssue.issue_number }}
                <span v-if="githubIssue.issue_title"> — {{ githubIssue.issue_title }}</span>
              </a>
              <span
                class="badge"
                :class="githubIssue.issue_state === 'closed' ? 'text-bg-secondary' : 'text-bg-success'"
              >
                {{ githubIssue.issue_state }}
              </span>
            </div>
            <p v-if="githubIssue.last_sync_error" class="text-danger small mb-0">
              Sync error: {{ githubIssue.last_sync_error }}
            </p>
            <button
              v-if="!readOnly"
              type="button"
              class="btn btn-sm btn-outline-secondary align-self-start"
              :disabled="githubBusy"
              @click="unlinkGitHubIssue"
            >
              Unlink
            </button>
          </div>
          <div v-else-if="!readOnly" class="d-flex flex-column gap-2">
            <button
              type="button"
              class="btn btn-sm btn-outline-primary align-self-start"
              :disabled="githubBusy"
              @click="createGitHubIssue"
            >
              {{ githubBusy ? 'Working…' : 'Create GitHub issue' }}
            </button>
            <div class="input-group input-group-sm">
              <input
                v-model="githubIssueRef"
                type="text"
                class="form-control"
                placeholder="Issue # or URL"
                :disabled="githubBusy"
              />
              <button
                type="button"
                class="btn btn-outline-secondary"
                :disabled="githubBusy || !githubIssueRef.trim()"
                @click="linkGitHubIssue"
              >
                Link existing
              </button>
            </div>
            <div class="form-hint mb-0">
              Creates or links an issue in the project’s GitHub repository. Does not import issues as tasks.
            </div>
          </div>
          <p v-else class="text-muted small mb-0">No linked GitHub issue.</p>
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
        <div v-if="readOnly" class="form-group mt-2">
          <label>Tags</label>
          <div v-if="taskTags.length" class="d-flex flex-wrap gap-1 mt-1">
            <span
              v-for="tag in taskTags"
              :key="tag.id"
              class="tag-chip"
              :style="{ backgroundColor: tag.color || '#6c757d' }"
            >{{ tag.name }}</span>
          </div>
          <p v-else class="text-muted small mb-0 mt-1">No tags</p>
        </div>
        <template v-else>
          <div v-if="allTags.length" class="form-group mt-2">
            <label>Tags (max 5)</label>
            <div v-for="tag in allTags" :key="tag.id" class="form-check">
              <input
                :id="`tag-${tag.id}`"
                type="checkbox"
                class="form-check-input"
                :checked="selectedTagIds.includes(tag.id)"
                @change="toggleTag(tag.id, ($event.target as HTMLInputElement).checked)"
              />
              <label class="form-check-label" :for="`tag-${tag.id}`">
                <span class="tag-chip" :style="{ backgroundColor: tag.color || '#6c757d' }">{{ tag.name }}</span>
              </label>
            </div>
          </div>
          <div class="form-group mt-2" v-if="canManageTags">
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
        </template>

        <div
          v-if="isKanbanTask && (mode === 'edit' || mode === 'view')"
          class="form-group mt-3 border-top pt-3"
        >
          <label class="d-block">Time tracking</label>
          <p class="small text-muted mb-2">
            Total logged: <strong>{{ timeSpentLabel }}</strong>
            <span v-if="timeSpentMinutes">({{ timeSpentMinutes }} min)</span>
          </p>
          <form v-if="!readOnly" class="row g-2 align-items-end mb-2" @submit.prevent="addTimeEntry">
            <div class="col-4">
              <label class="form-label small mb-0" for="time_minutes">Minutes</label>
              <input
                id="time_minutes"
                v-model="newEntryMinutes"
                type="number"
                class="form-control form-control-sm"
                min="1"
                max="1440"
                required
              />
            </div>
            <div class="col-5">
              <label class="form-label small mb-0" for="time_note">Note</label>
              <input
                id="time_note"
                v-model="newEntryNote"
                type="text"
                class="form-control form-control-sm"
                maxlength="200"
                placeholder="Optional"
              />
            </div>
            <div class="col-3">
              <button class="btn btn-sm btn-outline-primary w-100" type="submit" :disabled="addingTime">
                Add
              </button>
            </div>
          </form>
          <ul v-if="timeEntries.length" class="list-unstyled small mb-0">
            <li
              v-for="entry in timeEntries"
              :key="entry.id"
              class="d-flex justify-content-between align-items-start gap-2 mb-1"
            >
              <span>
                <strong>{{ entry.minutes }}m</strong>
                <span v-if="entry.note"> — {{ entry.note }}</span>
                <span class="text-muted"> · {{ entry.user_name || entry.user_email || 'user' }}</span>
              </span>
              <button
                v-if="!readOnly"
                class="btn btn-sm btn-link text-danger p-0"
                type="button"
                @click="removeTimeEntry(entry.id)"
              >
                Delete
              </button>
            </li>
          </ul>
          <p v-else class="small text-muted mb-0">No time entries yet.</p>
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
                <span v-if="ev.actor_user_name || ev.actor_email">
                  · {{ ev.actor_user_name || ev.actor_email }}
                </span>
              </li>
            </ul>
            <p v-else class="text-muted small mb-0">No activity recorded.</p>
          </div>
        </details>
      </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
textarea.task-description-input {
  height: auto;
  min-height: 80px;
  resize: vertical;
  overflow-y: hidden;
}
</style>
