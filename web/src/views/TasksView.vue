<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '@/api/client'
import type { Project, SavedView, Tag, Task } from '@/api/types'
import { APIError } from '@/api/types'
import ModernSidebar from '@/components/modern/ModernSidebar.vue'
import ModernTaskFilterBar from '@/components/modern/ModernTaskFilterBar.vue'
import ModernTaskCard from '@/components/modern/ModernTaskCard.vue'
import AppFooter from '@/components/AppFooter.vue'
import { useAuth } from '@/composables/useAuth'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import { useTaskListFilters } from '@/composables/useTaskListFilters'
import { useTaskSidebar } from '@/composables/useTaskSidebar'
import { useTaskSortable } from '@/composables/useTaskSortable'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { useViewDensity } from '@/composables/useViewDensity'

defineProps<{
  mobileSidebarOpen?: boolean
}>()

const emit = defineEmits<{
  'close-mobile-sidebar': []
}>()

const { openAdd, openEdit, lastSavedTask } = useTaskSidebar()
const { density } = useViewDensity()

const sidebarCollapsed = ref(false)
const tasks = ref<Task[]>([])
const projects = ref<Project[]>([])
const tags = ref<Tag[]>([])
const savedViews = ref<SavedView[]>([])
const activeViewId = ref<string | null>(null)

const total = ref(0)
const loadedPage = ref(0)
const totalPages = ref(1)
const completedCount = ref(0)
const incompleteCount = ref(0)
const search = ref('')
const loading = ref(true)
const loadingMore = ref(false)

// Save View Modal state
const showSaveViewModal = ref(false)
const newViewName = ref('')

// Add Project Modal state
const showAddProjectModal = ref(false)
const newProjectName = ref('')

// Edit Project Modal state
const showEditProjectModal = ref(false)
const editingProject = ref<Project | null>(null)
const editedProjectName = ref('')

// Bulk Control Panel State
const bulkProject = ref('')
const bulkTag = ref('')
const bulkPriority = ref('')
const bulkDate = ref('')

const toast = useToast()
const { askConfirm } = useConfirm()
const { user } = useAuth()
const {
  filters,
  hasActiveFilters,
  toApiParams,
  toExportQuery,
  setFilter,
  clearFilters: resetFilters,
  applySavedView: applySavedViewFilters,
} = useTaskListFilters()
const undoToken = ref<string | null>(null)
const selected = ref<number[]>([])
const favoriteListEl = ref<HTMLElement | null>(null)
const taskListEl = ref<HTMLElement | null>(null)
const loadMoreSentinel = ref<HTMLElement | null>(null)

const favoriteTasks = computed(() => tasks.value.filter((t) => t.favorite))
const regularTasks = computed(() => tasks.value.filter((t) => !t.favorite))
const allSelected = computed(
  () => tasks.value.length > 0 && selected.value.length === tasks.value.length,
)
const isSearching = computed(() => filters.search !== '')
const showTaskTable = computed(
  () => total.value > 0 || favoriteTasks.value.length > 0 || hasActiveFilters.value,
)
const hasMore = computed(() => loadedPage.value < totalPages.value)
const sortableEnabled = computed(() => filters.sort !== 'priority' && !loading.value)
const showFavoriteList = computed(() => favoriteTasks.value.length > 0)

function getTodayStr() {
  return new Date().toISOString().slice(0, 10)
}

function getTomorrowStr() {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return d.toISOString().slice(0, 10)
}

function getNextWeekStr() {
  const d = new Date()
  d.setDate(d.getDate() + 7)
  return d.toISOString().slice(0, 10)
}

function taskMatchesCurrentFilters(task: Task): boolean {
  if (!taskMatchesStatusFilter(task)) return false
  if (filters.project === '0' && task.project_id != null) return false
  if (filters.project && filters.project !== '0') {
    const pid = parseInt(filters.project, 10)
    if (!Number.isNaN(pid) && task.project_id !== pid) return false
  }
  if (filters.priority && String(task.priority) !== filters.priority) return false
  if (filters.tag) {
    const tagId = parseInt(filters.tag, 10)
    if (!task.tags?.some((t) => t.id === tagId)) return false
  }
  if (filters.search) {
    const q = filters.search.toLowerCase()
    const hay = `${task.title} ${task.description}`.toLowerCase()
    if (!hay.includes(q)) return false
  }
  return true
}

function registerTaskAdded(task: Task) {
  total.value += 1
  if (task.completed) completedCount.value += 1
  else incompleteCount.value += 1
  if (!taskMatchesCurrentFilters(task) || tasks.value.some((t) => t.id === task.id)) return
  if (task.favorite) {
    tasks.value = [task, ...tasks.value]
  } else {
    tasks.value = [...tasks.value, task]
  }
}

function removeTaskLocally(task: Task) {
  if (!tasks.value.some((t) => t.id === task.id)) return
  tasks.value = tasks.value.filter((t) => t.id !== task.id)
  total.value = Math.max(0, total.value - 1)
  if (task.completed) completedCount.value = Math.max(0, completedCount.value - 1)
  else incompleteCount.value = Math.max(0, incompleteCount.value - 1)
  selected.value = selected.value.filter((id) => id !== task.id)
}

function taskMatchesStatusFilter(task: Task) {
  if (filters.status === 'complete') return task.completed
  if (filters.status === 'incomplete') return !task.completed
  return true
}

function adjustCompletionCounts(wasCompleted: boolean, isCompleted: boolean) {
  if (wasCompleted === isCompleted) return
  if (isCompleted) {
    completedCount.value += 1
    incompleteCount.value = Math.max(0, incompleteCount.value - 1)
  } else {
    completedCount.value = Math.max(0, completedCount.value - 1)
    incompleteCount.value += 1
  }
}

function applyTaskUpdate(updated: Task) {
  const idx = tasks.value.findIndex((t) => t.id === updated.id)
  const previous = idx >= 0 ? tasks.value[idx] : null

  if (!taskMatchesStatusFilter(updated)) {
    if (idx >= 0) {
      tasks.value = tasks.value.filter((t) => t.id !== updated.id)
      total.value = Math.max(0, total.value - 1)
    }
    if (previous) adjustCompletionCounts(previous.completed, updated.completed)
    return
  }

  if (idx >= 0) {
    adjustCompletionCounts(tasks.value[idx].completed, updated.completed)
    tasks.value[idx] = updated
    return
  }

  tasks.value = [...tasks.value, updated]
}

function reorderLocalTasks(orderedIds: number[], favorite: boolean) {
  const reordered = orderedIds
    .map((id) => tasks.value.find((t) => t.id === id))
    .filter((t): t is Task => !!t)
  if (favorite) {
    tasks.value = [...reordered, ...tasks.value.filter((t) => !t.favorite)]
  } else {
    tasks.value = [...tasks.value.filter((t) => t.favorite), ...reordered]
  }
}

async function saveReorder(orderedIds: number[], favorite: boolean) {
  reorderLocalTasks(orderedIds, favorite)
  try {
    await api.reorderTasks({
      task_ids: orderedIds,
      favorite,
      project: filters.project || undefined,
    })
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not save task order', 'error')
    await reloadInitial()
  }
}

const { refresh: refreshSortable } = useTaskSortable(
  favoriteListEl,
  taskListEl,
  sortableEnabled,
  showFavoriteList,
  saveReorder,
)

async function loadMeta() {
  try {
    const [projs, tagList, views] = await Promise.all([
      api.listProjects(),
      api.listTags(),
      api.listSavedViews(),
    ])
    projects.value = projs
    tags.value = tagList
    savedViews.value = views
  } catch {
    /* non-fatal */
  }
}

async function reloadInitial() {
  loading.value = true
  loadedPage.value = 0
  selected.value = []
  try {
    const perPage = user.value?.items_per_page || 50
    const list = await api.listTasks(toApiParams(1, perPage))
    tasks.value = list.tasks
    loadedPage.value = 1
    total.value = list.total
    totalPages.value = list.total_pages
    completedCount.value = list.completed_count
    incompleteCount.value = list.incomplete_count
    search.value = filters.search
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load tasks', 'error')
  } finally {
    loading.value = false
    await nextTick()
    refreshSortable()
  }
}

async function loadMore() {
  if (loadingMore.value || loading.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const perPage = user.value?.items_per_page || 50
    const nextPage = loadedPage.value + 1
    const list = await api.listTasks(toApiParams(nextPage, perPage))
    const existingIds = new Set(tasks.value.map((t) => t.id))
    const newTasks = list.tasks.filter((t) => !existingIds.has(t.id))
    tasks.value = [...tasks.value, ...newTasks]
    loadedPage.value = nextPage
    total.value = list.total
    totalPages.value = list.total_pages
    completedCount.value = list.completed_count
    incompleteCount.value = list.incomplete_count
    await nextTick()
    refreshSortable()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load tasks', 'error')
  } finally {
    loadingMore.value = false
  }
}

watch(lastSavedTask, async (task) => {
  if (!task) return
  const exists = tasks.value.some((t) => t.id === task.id)
  if (exists) {
    applyTaskUpdate(task)
  } else {
    registerTaskAdded(task)
  }
  lastSavedTask.value = null
  await nextTick()
  refreshSortable()
})

async function toggleComplete(task: Task) {
  try {
    const updated = await api.patchTask(task.id, { completed: !task.completed })
    applyTaskUpdate(updated)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Update failed', 'error')
  }
}

async function toggleFavorite(task: Task) {
  try {
    const updated = await api.patchTask(task.id, { favorite: !task.favorite })
    applyTaskUpdate(updated)
    await nextTick()
    refreshSortable()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Update failed', 'error')
  }
}

async function handleInlineTaskPatch(payload: { id: number; title?: string; description?: string }) {
  try {
    const updated = await api.patchTask(payload.id, {
      title: payload.title,
      description: payload.description,
    })
    applyTaskUpdate(updated)
    toast.push('Task updated', 'success')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Update failed', 'error')
  }
}

async function removeTask(task: Task) {
  const ok = await askConfirm({
    title: 'Delete task?',
    message: `Delete “${task.title}”?`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  try {
    const res = await api.deleteTask(task.id)
    undoToken.value = res.undo_token || null
    removeTaskLocally(task)
    toast.push(undoToken.value ? 'Task deleted — undo available' : 'Task deleted', 'info')
    await nextTick()
    refreshSortable()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Delete failed', 'error')
  }
}

async function undoDelete() {
  if (!undoToken.value) return
  try {
    await api.undo(undoToken.value)
    undoToken.value = null
    toast.push('Restored', 'success')
    await reloadInitial()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Undo failed', 'error')
  }
}

function toggleSelect(id: number, checked: boolean) {
  if (checked) {
    if (!selected.value.includes(id)) selected.value = [...selected.value, id]
  } else {
    selected.value = selected.value.filter((x) => x !== id)
  }
}

function toggleSelectAll(checked: boolean) {
  selected.value = checked ? tasks.value.map((t) => t.id) : []
}

async function bulk(action: string, extra: Record<string, unknown> = {}) {
  if (!selected.value.length) return
  if (action === 'delete') {
    const ok = await askConfirm({
      title: 'Delete tasks?',
      message: `Delete ${selected.value.length} selected task${selected.value.length === 1 ? '' : 's'}?`,
      confirmLabel: 'Delete',
      danger: true,
    })
    if (!ok) return
  }
  const affectedIds = [...selected.value]
  try {
    const res = await api.bulkTasks({ action, task_ids: affectedIds, ...extra })
    if (res.undo_token) undoToken.value = res.undo_token
    toast.push(`Bulk ${action}: ${res.affected ?? affectedIds.length}`, 'success')
    if (action === 'delete') {
      for (const id of affectedIds) {
        const task = tasks.value.find((t) => t.id === id)
        if (task) removeTaskLocally(task)
      }
      await nextTick()
      refreshSortable()
    } else if (action === 'complete' || action === 'incomplete') {
      for (const id of affectedIds) {
        const task = tasks.value.find((t) => t.id === id)
        if (!task) continue
        applyTaskUpdate({ ...task, completed: action === 'complete' })
      }
    } else {
      await reloadInitial()
    }
    selected.value = []
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Bulk action failed', 'error')
  }
}

function setFilterAndReload(key: Parameters<typeof setFilter>[0], value: string) {
  activeViewId.value = null
  setFilter(key, value)
  void reloadInitial()
}

function clearFilters() {
  activeViewId.value = null
  search.value = ''
  resetFilters()
  void reloadInitial()
}

// Toggle / Deselect project filter when clicking an already selected project
function selectProjectFilter(id: string) {
  activeViewId.value = null
  if (filters.project === id) {
    setFilterAndReload('project', '')
  } else {
    setFilterAndReload('project', id)
  }
}

function selectSavedViewFilter(id: string) {
  const view = savedViews.value.find((v) => String(v.id) === id)
  if (view) {
    activeViewId.value = String(view.id)
    applySavedViewFilters(view.filter || {})
    search.value = filters.search
    void reloadInitial()
  }
}

async function saveCurrentView() {
  if (!newViewName.value.trim()) return
  try {
    const created = await api.createSavedView({
      name: newViewName.value.trim(),
      filter: {
        status: filters.status || undefined,
        due: filters.due || undefined,
        project: filters.project || undefined,
        priority: filters.priority || undefined,
        tag: filters.tag || undefined,
        sort: filters.sort || undefined,
        search: filters.search || undefined,
      },
    })
    toast.push('View saved successfully!', 'success')
    activeViewId.value = String(created.id)
    newViewName.value = ''
    showSaveViewModal.value = false
    await loadMeta()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to save view', 'error')
  }
}

async function createProject() {
  if (!newProjectName.value.trim()) return
  try {
    await api.createProject(newProjectName.value.trim())
    toast.push('Project created!', 'success')
    newProjectName.value = ''
    showAddProjectModal.value = false
    await loadMeta()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to create project', 'error')
  }
}

function openEditProject(proj: Project) {
  editingProject.value = proj
  editedProjectName.value = proj.name
  showEditProjectModal.value = true
}

async function renameProject() {
  if (!editingProject.value || !editedProjectName.value.trim()) return
  try {
    await api.renameProject(editingProject.value.id, editedProjectName.value.trim())
    toast.push('Project renamed!', 'success')
    editingProject.value = null
    editedProjectName.value = ''
    showEditProjectModal.value = false
    await loadMeta()
    await reloadInitial()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to rename project', 'error')
  }
}

function selectHome() {
  activeViewId.value = null
  if (filters.project) {
    setFilterAndReload('project', '')
  }
}

async function exportTasks(format: 'json' | 'csv', filtered: boolean) {
  try {
    const suffix = filtered ? toExportQuery() : ''
    const path = `/api/v1/export?format=${format}${suffix}`
    const res = await fetch(path, { credentials: 'include' })
    if (!res.ok) throw new Error('Export failed')
    const blob = await res.blob()
    const disposition = res.headers.get('Content-Disposition') || ''
    const match = /filename="([^"]+)"/.exec(disposition)
    const filename = match?.[1] || `gotodo-export.${format}`
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    toast.push('Export failed', 'error')
  }
}

useInfiniteScroll(loadMoreSentinel, loadMore, hasMore)

onMounted(async () => {
  await loadMeta()
  await reloadInitial()
})
</script>

<template>
  <div class="ordryn-main-layout">
    <!-- Collapsible & Responsive Warm Sidebar -->
    <ModernSidebar
      :collapsed="sidebarCollapsed"
      :mobile-open="mobileSidebarOpen || false"
      :projects="projects"
      :saved-views="savedViews"
      :active-project="filters.project"
      :active-view="activeViewId || undefined"
      @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      @close-mobile="emit('close-mobile-sidebar')"
      @select-home="selectHome"
      @select-project="selectProjectFilter"
      @select-view="selectSavedViewFilter"
      @add-project="showAddProjectModal = true"
      @edit-project="openEditProject"
      @add-view="showSaveViewModal = true"
    />

    <!-- Main Content Area -->
    <div class="flex-grow-1 p-3 p-md-4 overflow-hidden d-flex flex-column justify-content-between">
      <div>
        <!-- Upper Header Bar: Action Buttons & Quick Stats -->
        <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
          <div class="d-flex align-items-center gap-2">
            <span class="badge bg-primary bg-opacity-10 text-primary border border-primary border-opacity-20 px-3 py-2 rounded-pill">
              Tasks: {{ total }}
            </span>
            <span class="badge bg-success bg-opacity-10 text-success border border-success border-opacity-20 px-3 py-2 rounded-pill">
              Completed: {{ completedCount }}
            </span>
            <span class="badge bg-warning bg-opacity-10 text-warning border border-warning border-opacity-20 px-3 py-2 rounded-pill">
              Incomplete: {{ incompleteCount }}
            </span>
          </div>

          <div class="d-flex align-items-center gap-2">
            <!-- Import/Export Dropdown -->
            <div class="dropdown">
              <button
                class="btn btn-sm btn-outline-secondary dropdown-toggle rounded-pill px-3"
                type="button"
                data-bs-toggle="dropdown"
              >
                <i class="bi bi-arrow-down-up me-1" /> Import / Export
              </button>
              <ul class="dropdown-menu dropdown-menu-end shadow-sm border-0">
                <li>
                  <RouterLink class="dropdown-item small" to="/import">
                    <i class="bi bi-upload me-2" />Import CSV
                  </RouterLink>
                </li>
                <li>
                  <RouterLink class="dropdown-item small" to="/settings#calendar-feed">
                    <i class="bi bi-calendar3 me-2" />Calendar Sync (ICS)
                  </RouterLink>
                </li>
                <li><hr class="dropdown-divider" /></li>
                <li>
                  <button type="button" class="dropdown-item small" @click="exportTasks('csv', true)">
                    <i class="bi bi-download me-2" />Export Filtered CSV
                  </button>
                </li>
                <li>
                  <button type="button" class="dropdown-item small" @click="exportTasks('json', true)">
                    <i class="bi bi-download me-2" />Export Filtered JSON
                  </button>
                </li>
              </ul>
            </div>

            <!-- Add Task Button -->
            <button
              type="button"
              class="btn btn-sm btn-success rounded-pill px-3 shadow-xs d-flex align-items-center gap-1"
              @click="() => openAdd()"
            >
              <i class="bi bi-plus-lg" />
              <span>Add Task</span>
            </button>

            <!-- Undo Delete Button -->
            <button
              v-if="undoToken"
              type="button"
              class="btn btn-sm btn-outline-warning rounded-pill px-3"
              @click="undoDelete"
            >
              <i class="bi bi-arrow-counterclockwise me-1" />Undo
            </button>
          </div>
        </div>

        <!-- Modern Filter Bar (Pills, Search, Density Toggle) -->
        <ModernTaskFilterBar
          :status="filters.status"
          :tag="filters.tag"
          :priority="filters.priority"
          :due-date-preset="filters.due"
          :sort="filters.sort"
          :search="search"
          :density="density"
          :tags="tags"
          @update:status="setFilterAndReload('status', $event)"
          @update:tag="setFilterAndReload('tag', $event)"
          @update:priority="setFilterAndReload('priority', $event)"
          @update:due-date-preset="setFilterAndReload('due', $event)"
          @update:sort="setFilterAndReload('sort', $event)"
          @update:search="search = $event; setFilter('search', $event); reloadInitial()"
          @update:density="density = $event"
          @clear-filters="clearFilters"
        />

        <!-- Bulk Actions Bar -->
        <div
          v-if="selected.length"
          class="bulk-action-bar alert alert-info py-2 px-3 rounded-3 shadow-sm d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3"
        >
          <span class="fw-semibold small">{{ selected.length }} tasks selected</span>
          <div class="d-flex align-items-center gap-2">
            <button type="button" class="btn btn-xs btn-success rounded-pill" @click="bulk('complete')">Complete</button>
            <button type="button" class="btn btn-xs btn-outline-secondary rounded-pill" @click="bulk('incomplete')">Incomplete</button>

            <!-- Compact Feature-Rich "More Actions" Popover Panel -->
            <div class="dropdown d-inline-block">
              <button class="btn btn-xs btn-outline-primary dropdown-toggle rounded-pill" type="button" data-bs-toggle="dropdown" data-bs-auto-close="outside">
                More Actions
              </button>
              <div class="dropdown-menu shadow-lg border p-3 rounded-3 mt-1" style="width: 280px; max-width: 90vw;">
                <!-- Move to Project -->
                <div class="mb-3">
                  <label class="form-label text-muted small fw-bold text-uppercase mb-1" style="font-size: 0.7rem;">Move to project...</label>
                  <select v-model="bulkProject" class="form-select form-select-sm mb-2">
                    <option value="">Select project...</option>
                    <option value="0">No Project</option>
                    <option v-for="p in projects" :key="p.id" :value="String(p.id)">{{ p.name }}</option>
                  </select>
                  <button
                    type="button"
                    class="btn btn-xs btn-outline-primary rounded-pill w-100"
                    :disabled="bulkProject === ''"
                    @click="bulk('move_project', { project_id: bulkProject })"
                  >
                    Move
                  </button>
                </div>

                <!-- Select Tag -->
                <div class="mb-3 border-top pt-2">
                  <label class="form-label text-muted small fw-bold text-uppercase mb-1" style="font-size: 0.7rem;">Select tag...</label>
                  <select v-model="bulkTag" class="form-select form-select-sm mb-2">
                    <option value="">Select tag...</option>
                    <option v-for="t in tags" :key="t.id" :value="String(t.id)">{{ t.name }}</option>
                  </select>
                  <div class="d-flex gap-2">
                    <button
                      type="button"
                      class="btn btn-xs btn-outline-primary rounded-pill flex-grow-1"
                      :disabled="!bulkTag"
                      @click="bulk('add_tag', { tag_id: parseInt(bulkTag, 10) })"
                    >
                      Add Tag
                    </button>
                    <button
                      type="button"
                      class="btn btn-xs btn-outline-secondary rounded-pill flex-grow-1"
                      :disabled="!bulkTag"
                      @click="bulk('remove_tag', { tag_id: parseInt(bulkTag, 10) })"
                    >
                      Remove Tag
                    </button>
                  </div>
                </div>

                <!-- Priority -->
                <div class="mb-3 border-top pt-2">
                  <label class="form-label text-muted small fw-bold text-uppercase mb-1" style="font-size: 0.7rem;">Priority</label>
                  <select v-model="bulkPriority" class="form-select form-select-sm mb-2">
                    <option value="">Select priority...</option>
                    <option value="3">High Priority</option>
                    <option value="2">Medium Priority</option>
                    <option value="1">Low Priority</option>
                    <option value="0">No Priority</option>
                  </select>
                  <button
                    type="button"
                    class="btn btn-xs btn-outline-primary rounded-pill w-100"
                    :disabled="bulkPriority === ''"
                    @click="bulk('set_priority', { priority: parseInt(bulkPriority, 10) })"
                  >
                    Set Priority
                  </button>
                </div>

                <!-- Due Date -->
                <div class="border-top pt-2">
                  <label class="form-label text-muted small fw-bold text-uppercase mb-1" style="font-size: 0.7rem;">Due Date</label>
                  <div class="btn-group btn-group-sm w-100 mb-2">
                    <button type="button" class="btn btn-xs btn-outline-secondary" @click="bulk('set_due_date', { due_date: getTodayStr() })">Today</button>
                    <button type="button" class="btn btn-xs btn-outline-secondary" @click="bulk('set_due_date', { due_date: getTomorrowStr() })">Tomorrow</button>
                    <button type="button" class="btn btn-xs btn-outline-secondary" @click="bulk('set_due_date', { due_date: getNextWeekStr() })">+1 Wk</button>
                    <button type="button" class="btn btn-xs btn-outline-secondary" @click="bulk('set_due_date', { due_date: '' })">Clear</button>
                  </div>
                  <input v-model="bulkDate" type="date" class="form-control form-control-sm mb-2" />
                  <div class="d-flex gap-2">
                    <button
                      type="button"
                      class="btn btn-xs btn-outline-primary rounded-pill flex-grow-1"
                      :disabled="!bulkDate"
                      @click="bulk('set_due_date', { due_date: bulkDate })"
                    >
                      Set Due
                    </button>
                    <button
                      type="button"
                      class="btn btn-xs btn-outline-secondary rounded-pill flex-grow-1"
                      @click="bulk('set_due_date', { due_date: '' })"
                    >
                      Clear Due
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <button type="button" class="btn btn-xs btn-danger rounded-pill" @click="bulk('delete')">Delete</button>
            <button type="button" class="btn btn-xs btn-link text-muted" @click="selected = []">Deselect</button>
          </div>
        </div>

        <!-- Task Lists Container -->
        <div id="task-container" aria-live="polite">
          <div v-if="loading && !tasks.length" class="text-center py-5 text-muted">
            <div class="spinner-border spinner-border-sm me-2" role="status" />Loading tasks…
          </div>

          <div v-else-if="showTaskTable">
            <!-- Select All Controls Bar -->
            <div class="d-flex align-items-center justify-content-between mb-2 px-1">
              <div class="form-check d-flex align-items-center gap-2">
                <input
                  id="select-all-tasks"
                  type="checkbox"
                  class="form-check-input"
                  :checked="allSelected"
                  @change="toggleSelectAll(($event.target as HTMLInputElement).checked)"
                />
                <label for="select-all-tasks" class="form-check-label small text-muted cursor-pointer">
                  Select all on page
                </label>
              </div>
              <button
                v-if="hasActiveFilters"
                type="button"
                class="btn btn-link btn-sm text-decoration-none p-0 text-muted small"
                @click="clearFilters"
              >
                <i class="bi bi-x-circle me-1" />Clear active filters
              </button>
            </div>

            <!-- Starred Tasks Section -->
            <div v-if="showFavoriteList" class="starred-tasks-section mb-3">
              <div class="small fw-bold text-muted mb-2 px-1 d-flex align-items-center gap-1">
                <i class="bi bi-star-fill text-warning" /> Starred Tasks
              </div>
              <div id="favorite-task-list" ref="favoriteListEl">
                <ModernTaskCard
                  v-for="task in favoriteTasks"
                  :key="task.id"
                  :task="task"
                  :selected="selected.includes(task.id)"
                  :density="density"
                  :show-project-pill="!filters.project"
                  @toggle-select="toggleSelect(task.id, $event)"
                  @toggle-complete="toggleComplete(task)"
                  @toggle-favorite="toggleFavorite(task)"
                  @patch-task="handleInlineTaskPatch"
                  @edit="openEdit(task.id)"
                  @remove="removeTask(task)"
                />
              </div>
            </div>

            <!-- Regular Task List -->
            <div id="task-list" ref="taskListEl">
              <ModernTaskCard
                v-for="task in regularTasks"
                :key="task.id"
                :task="task"
                :selected="selected.includes(task.id)"
                :density="density"
                :show-project-pill="!filters.project"
                @toggle-select="toggleSelect(task.id, $event)"
                @toggle-complete="toggleComplete(task)"
                @toggle-favorite="toggleFavorite(task)"
                @patch-task="handleInlineTaskPatch"
                @edit="openEdit(task.id)"
                @remove="removeTask(task)"
              />

              <!-- Empty Search Results -->
              <div v-if="!tasks.length && hasActiveFilters" class="text-center py-5 bg-white rounded-3 border">
                <p class="text-muted mb-2">No tasks match your active filters.</p>
                <button type="button" class="btn btn-sm btn-outline-primary rounded-pill" @click="clearFilters">
                  <i class="bi bi-x-circle me-1" />Clear filters
                </button>
              </div>

              <!-- Infinite Scroll Sentinel -->
              <div v-if="hasMore" ref="loadMoreSentinel" class="text-center py-3 text-muted small">
                <span v-if="loadingMore" class="spinner-border spinner-border-sm me-2" />
                {{ loadingMore ? 'Loading more tasks…' : 'Scroll for more tasks' }}
              </div>
            </div>
          </div>

          <!-- Zero State (No tasks match search) -->
          <div v-else-if="isSearching" class="text-center py-5 bg-white rounded-3 border">
            <p class="text-muted mb-2">No tasks match your search query.</p>
            <button type="button" class="btn btn-sm btn-outline-primary rounded-pill" @click="clearFilters">
              Clear Search
            </button>
          </div>

          <!-- Empty State (No tasks at all) -->
          <div v-else class="text-center py-5 bg-white rounded-3 border shadow-xs">
            <i class="bi bi-clipboard-check display-4 text-muted opacity-50" />
            <h4 class="mt-3 fw-bold">No tasks yet</h4>
            <p class="text-muted">Get started by creating your first task.</p>
            <button type="button" class="btn btn-success rounded-pill px-4" @click="() => openAdd()">
              <i class="bi bi-plus-lg me-1" /> Add Task
            </button>
          </div>
        </div>
      </div>

      <!-- Save Current View Modal -->
      <div v-if="showSaveViewModal" class="modal fade show d-block" style="background: rgba(0,0,0,0.5);" tabindex="-1">
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content border-0 shadow">
            <div class="modal-header border-0 pb-0">
              <h5 class="modal-title fw-bold">Save Current View</h5>
              <button type="button" class="btn-close" @click="showSaveViewModal = false" />
            </div>
            <div class="modal-body py-3">
              <p class="text-muted small mb-3">Save your current active filters into a custom named View in the sidebar.</p>
              <div class="mb-3">
                <label for="new-view-name" class="form-label small fw-bold">View Name</label>
                <input
                  id="new-view-name"
                  v-model="newViewName"
                  type="text"
                  class="form-control"
                  placeholder="e.g., High Priority Work"
                  @keyup.enter="saveCurrentView"
                />
              </div>
            </div>
            <div class="modal-footer border-0 pt-0 justify-content-end gap-2">
              <button type="button" class="btn btn-sm btn-outline-secondary" @click="showSaveViewModal = false">Cancel</button>
              <button type="button" class="btn btn-sm btn-primary px-3" :disabled="!newViewName.trim()" @click="saveCurrentView">Save View</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Add Project Modal -->
      <div v-if="showAddProjectModal" class="modal fade show d-block" style="background: rgba(0,0,0,0.5);" tabindex="-1">
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content border-0 shadow">
            <div class="modal-header border-0 pb-0">
              <h5 class="modal-title fw-bold">Create New Project</h5>
              <button type="button" class="btn-close" @click="showAddProjectModal = false" />
            </div>
            <div class="modal-body py-3">
              <div class="mb-3">
                <label for="new-project-name" class="form-label small fw-bold">Project Name</label>
                <input
                  id="new-project-name"
                  v-model="newProjectName"
                  type="text"
                  class="form-control"
                  placeholder="e.g., Marketing Campaign"
                  @keyup.enter="createProject"
                />
              </div>
            </div>
            <div class="modal-footer border-0 pt-0 justify-content-end gap-2">
              <button type="button" class="btn btn-sm btn-outline-secondary" @click="showAddProjectModal = false">Cancel</button>
              <button type="button" class="btn btn-sm btn-success px-3" :disabled="!newProjectName.trim()" @click="createProject">Create Project</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Edit Project Modal -->
      <div v-if="showEditProjectModal" class="modal fade show d-block" style="background: rgba(0,0,0,0.5);" tabindex="-1">
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content border-0 shadow">
            <div class="modal-header border-0 pb-0">
              <h5 class="modal-title fw-bold">Rename Project</h5>
              <button type="button" class="btn-close" @click="showEditProjectModal = false" />
            </div>
            <div class="modal-body py-3">
              <div class="mb-3">
                <label for="edit-project-name" class="form-label small fw-bold">Project Name</label>
                <input
                  id="edit-project-name"
                  v-model="editedProjectName"
                  type="text"
                  class="form-control"
                  placeholder="Project Name"
                  @keyup.enter="renameProject"
                />
              </div>
            </div>
            <div class="modal-footer border-0 pt-0 justify-content-end gap-2">
              <button type="button" class="btn btn-sm btn-outline-secondary" @click="showEditProjectModal = false">Cancel</button>
              <button type="button" class="btn btn-sm btn-primary px-3" :disabled="!editedProjectName.trim()" @click="renameProject">Save Name</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Clean Reusable Footer inside right content area -->
      <AppFooter />
    </div>
  </div>
</template>

<style scoped>
.btn-xs {
  padding: 0.15rem 0.5rem;
  font-size: 0.75rem;
}
</style>
