<script setup lang="ts">
import { inject, onMounted, ref, type Ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { DashboardStats, Project, SavedView, Task } from '@/api/types'
import { APIError } from '@/api/types'
import ModernSidebar from '@/components/modern/ModernSidebar.vue'
import AppFooter from '@/components/AppFooter.vue'
import { useToast } from '@/composables/useToast'
import { useSidebarState } from '@/composables/useSidebarState'
import { useTaskSidebar } from '@/composables/useTaskSidebar'

defineProps<{
  mobileSidebarOpen?: boolean
}>()

const emit = defineEmits<{
  'close-mobile-sidebar': []
}>()

const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const projects = ref<Project[]>([])
const savedViews = ref<SavedView[]>([])
const overdueTasks = ref<Task[]>([])
const completing = ref<Record<number, boolean>>({})
const { sidebarCollapsed, toggleSidebar } = useSidebarState()
const { openEdit, lastSavedTask } = useTaskSidebar()
const overdueCount = inject<Ref<number>>('overdueCount')
const loading = ref(true)
const toast = useToast()

function selectHome() {
  void router.push('/')
}

function selectProject(id: string) {
  void router.push({ path: '/', query: { project: id } })
}

function selectView(id: string) {
  void router.push({ path: '/', query: { view: id } })
}

function formatDueDate(dateStr?: string): string {
  if (!dateStr) return ''
  const dateObj = new Date(dateStr + 'T00:00:00')
  return dateObj.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

async function loadOverdueTasks() {
  const list = await api.listTasks({ due: 'overdue', page: 1, per_page: 50 })
  overdueTasks.value = [...list.tasks].sort((a, b) => (a.due_date || '').localeCompare(b.due_date || ''))
}

async function refreshDashboard() {
  const [dashStats] = await Promise.all([api.dashboard(), loadOverdueTasks()])
  stats.value = dashStats
  if (overdueCount) overdueCount.value = dashStats.overdue_count
}

async function toggleComplete(task: Task) {
  if (completing.value[task.id]) return
  completing.value = { ...completing.value, [task.id]: true }
  try {
    await api.patchTask(task.id, { completed: true })
    overdueTasks.value = overdueTasks.value.filter((t) => t.id !== task.id)
    if (stats.value) {
      stats.value = {
        ...stats.value,
        overdue_count: Math.max(0, stats.value.overdue_count - 1),
        completed_this_week: stats.value.completed_this_week + 1,
      }
    }
    if (overdueCount) overdueCount.value = Math.max(0, overdueCount.value - 1)
    toast.push('Task completed', 'success')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not update task', 'error')
  } finally {
    const next = { ...completing.value }
    delete next[task.id]
    completing.value = next
  }
}

onMounted(async () => {
  try {
    const [dashStats, projList, viewList] = await Promise.all([
      api.dashboard(),
      api.listProjects(),
      api.listSavedViews(),
      loadOverdueTasks(),
    ])
    stats.value = dashStats
    projects.value = projList
    savedViews.value = viewList
    if (overdueCount) overdueCount.value = dashStats.overdue_count
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load dashboard', 'error')
  } finally {
    loading.value = false
  }
})

watch(lastSavedTask, () => {
  void refreshDashboard().catch(() => {
    /* keep current view if refresh fails */
  })
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
      @toggle-collapse="toggleSidebar"
      @close-mobile="emit('close-mobile-sidebar')"
      @select-home="selectHome"
      @select-project="selectProject"
      @select-view="selectView"
      @add-project="router.push('/projects')"
      @edit-project="() => router.push('/projects')"
      @add-view="router.push('/views')"
    />

    <!-- Main Content Area -->
    <div class="flex-grow-1 p-3 p-md-4 overflow-auto d-flex flex-column justify-content-between">
      <div>
        <h1 class="h3 fw-bold mb-3">Dashboard</h1>
        <p v-if="loading" class="text-muted">Loading…</p>
        <template v-else-if="stats">
          <div class="row g-3 mb-4">
            <div class="col-sm-6 col-lg-3">
              <div class="card text-center border-0 shadow-xs">
                <div class="card-body py-3">
                  <div class="text-muted small fw-medium">Overdue</div>
                  <div class="display-6 fw-bold text-danger">{{ stats.overdue_count }}</div>
                </div>
              </div>
            </div>
            <div class="col-sm-6 col-lg-3">
              <div class="card text-center border-0 shadow-xs">
                <div class="card-body py-3">
                  <div class="text-muted small fw-medium">Due today</div>
                  <div class="display-6 fw-bold text-primary">{{ stats.due_today_count }}</div>
                </div>
              </div>
            </div>
            <div class="col-sm-6 col-lg-3">
              <div class="card text-center border-0 shadow-xs">
                <div class="card-body py-3">
                  <div class="text-muted small fw-medium">Done this week</div>
                  <div class="display-6 fw-bold text-success">{{ stats.completed_this_week }}</div>
                </div>
              </div>
            </div>
            <div class="col-sm-6 col-lg-3">
              <div class="card text-center border-0 shadow-xs">
                <div class="card-body py-3">
                  <div class="text-muted small fw-medium">Streak</div>
                  <div class="display-6 fw-bold text-warning">{{ stats.streak_days }}d</div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="overdueTasks.length" class="card border-0 shadow-xs mb-4">
            <div class="card-header bg-transparent border-0 pt-3 pb-0 d-flex justify-content-between align-items-center">
              <h2 class="h5 fw-bold mb-0">Overdue</h2>
              <span class="badge rounded-pill bg-danger bg-opacity-10 text-danger border border-danger border-opacity-20 px-2.5 py-1">
                {{ stats.overdue_count }}
              </span>
            </div>
            <ul class="list-group list-group-flush">
              <li
                v-for="task in overdueTasks"
                :key="task.id"
                class="list-group-item d-flex align-items-center gap-3"
              >
                <button
                  type="button"
                  class="btn btn-link p-0 text-danger flex-shrink-0"
                  :title="`Mark ${task.title} complete`"
                  :aria-label="`Mark ${task.title} complete`"
                  :disabled="!!completing[task.id]"
                  @click="toggleComplete(task)"
                >
                  <i class="bi bi-circle fs-5" aria-hidden="true" />
                </button>
                <button
                  type="button"
                  class="btn btn-link text-start text-decoration-none flex-grow-1 p-0 min-w-0"
                  @click="openEdit(task.id)"
                >
                  <span class="fw-medium text-body d-block text-truncate">{{ task.title }}</span>
                  <span v-if="task.project" class="text-muted small">{{ task.project }}</span>
                </button>
                <span class="text-danger small fw-medium flex-shrink-0 whitespace-nowrap">
                  {{ formatDueDate(task.due_date) }}
                </span>
              </li>
            </ul>
          </div>

          <div class="row g-4">
            <div class="col-lg-6">
              <div class="card border-0 shadow-xs">
                <div class="card-header bg-transparent border-0 pt-3 pb-0"><h2 class="h5 fw-bold mb-0">By project</h2></div>
                <ul class="list-group list-group-flush">
                  <li v-for="row in (stats.by_project ?? [])" :key="row.name" class="list-group-item d-flex justify-content-between align-items-center">
                    <span class="fw-medium">{{ row.name || 'No project' }}</span>
                    <span class="badge rounded-pill bg-secondary bg-opacity-10 text-secondary border border-secondary border-opacity-20 px-2.5 py-1">{{ row.count }}</span>
                  </li>
                  <li v-if="!(stats.by_project?.length)" class="list-group-item text-muted">No project breakdown.</li>
                </ul>
              </div>
            </div>
            <div class="col-lg-6">
              <div class="card border-0 shadow-xs">
                <div class="card-header bg-transparent border-0 pt-3 pb-0"><h2 class="h5 fw-bold mb-0">Last 7 days</h2></div>
                <ul class="list-group list-group-flush">
                  <li v-for="day in stats.completions_last_7_days" :key="day.date" class="list-group-item d-flex justify-content-between align-items-center">
                    <span>{{ day.date }}</span>
                    <span class="badge rounded-pill bg-success bg-opacity-10 text-success border border-success border-opacity-20 px-2.5 py-1">{{ day.count }}</span>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Clean Reusable Footer inside right content area -->
      <AppFooter />
    </div>
  </div>
</template>
