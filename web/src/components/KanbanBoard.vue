<template>
  <div class="kanban-board">
    <div v-if="loading" class="text-center py-4 text-muted">
      <div class="spinner-border spinner-border-sm me-2" role="status" />
      Loading board…
    </div>
    <div v-else-if="!statuses.length" class="text-muted small py-3">
      No status columns yet. Open project Board settings to add statuses.
    </div>
    <div v-else class="kanban-columns d-flex gap-3 overflow-auto pb-2">
      <div
        v-for="col in statuses"
        :key="col.id"
        class="kanban-column flex-shrink-0"
      >
        <div class="kanban-column-header d-flex align-items-center justify-content-between mb-2 px-1">
          <div class="d-flex align-items-center gap-1">
            <strong class="small">{{ col.name }}</strong>
            <span v-if="col.is_default" class="badge text-bg-info">default</span>
            <span v-if="col.is_done" class="badge text-bg-success">done</span>
          </div>
          <span class="badge text-bg-secondary">{{ tasksForStatus(col.id).length }}</span>
        </div>
        <div
          :ref="(el) => setColumnEl(col.id, el)"
          class="kanban-column-body"
          :data-status-id="col.id"
        >
          <div
            v-for="task in tasksForStatus(col.id)"
            :key="task.id"
            class="kanban-card card mb-2"
            :data-task-id="task.id"
            :class="{ 'kanban-card-readonly': !canDrag }"
          >
            <div class="card-body p-2">
              <div class="d-flex align-items-start gap-1">
                <span
                  v-if="canDrag"
                  class="drag-handle text-muted pe-1"
                  title="Drag"
                  aria-hidden="true"
                >
                  <i class="bi bi-grip-vertical" />
                </span>
                <div class="flex-grow-1 min-w-0">
                  <button
                    type="button"
                    class="btn btn-link text-start text-decoration-none p-0 fw-semibold text-body kanban-card-title"
                    @click="emit('open-task', task.id)"
                  >
                    {{ task.title }}
                  </button>
                  <div class="d-flex flex-wrap gap-1 mt-1">
                    <span
                      v-if="task.priority > 0"
                      class="badge"
                      :class="priorityClass(task.priority)"
                    >
                      {{ priorityLabel(task.priority) }}
                    </span>
                    <span v-if="task.due_date" class="badge text-bg-light text-muted border">
                      {{ task.due_date }}
                    </span>
                    <span
                      v-if="task.estimate_points != null"
                      class="badge text-bg-light text-muted border"
                      title="Estimate"
                    >
                      {{ task.estimate_points }} pts
                    </span>
                    <span
                      class="badge border"
                      :class="task.claimed_by ? 'text-bg-primary' : 'text-bg-light text-muted'"
                      title="Claimed by"
                    >
                      <i class="bi bi-person me-1" />{{ claimerLabel(task) }}
                    </span>
                  </div>
                  <div v-if="canDrag" class="mt-2">
                    <button
                      v-if="!task.claimed_by || task.claimed_by !== user?.id"
                      type="button"
                      class="btn btn-sm btn-outline-primary py-0 px-2"
                      :disabled="claimingId === task.id"
                      @click.stop="claimTask(task)"
                    >
                      {{ task.claimed_by ? 'Take over' : 'Claim' }}
                    </button>
                    <button
                      v-else
                      type="button"
                      class="btn btn-sm btn-outline-secondary py-0 px-2"
                      :disabled="claimingId === task.id"
                      @click.stop="unclaimTask(task)"
                    >
                      Release
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import Sortable from 'sortablejs'
import { api } from '@/api/client'
import type { ProjectStatus, Task } from '@/api/types'
import { APIError } from '@/api/types'
import { useAuth } from '@/composables/useAuth'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  projectId: number
  tasks: Task[]
  role?: 'owner' | 'editor' | 'viewer'
}>()

const emit = defineEmits<{
  'open-task': [id: number]
  changed: []
}>()

const toast = useToast()
const { user } = useAuth()
const statuses = ref<ProjectStatus[]>([])
const loading = ref(true)
const claimingId = ref<number | null>(null)
const columnEls = new Map<number, HTMLElement>()
const sortables: Sortable[] = []

const canDrag = computed(() => props.role !== 'viewer')

function claimerLabel(task: Task) {
  if (!task.claimed_by) return 'Unclaimed'
  if (user.value?.id && task.claimed_by === user.value.id) return 'You'
  return task.claimed_by_name || `User #${task.claimed_by}`
}

async function claimTask(task: Task) {
  claimingId.value = task.id
  try {
    await api.claimTask(task.id)
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not claim task', 'error')
  } finally {
    claimingId.value = null
  }
}

async function unclaimTask(task: Task) {
  claimingId.value = task.id
  try {
    await api.unclaimTask(task.id)
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not release task', 'error')
  } finally {
    claimingId.value = null
  }
}

function setColumnEl(statusId: number, el: unknown) {
  if (el instanceof HTMLElement) {
    columnEls.set(statusId, el)
  } else {
    columnEls.delete(statusId)
  }
}

const statusIdSet = computed(() => new Set(statuses.value.map((s) => s.id)))

const defaultStatusId = computed(() => {
  const def = statuses.value.find((s) => s.is_default)
  return def?.id ?? statuses.value[0]?.id ?? 0
})

function tasksForStatus(statusId: number): Task[] {
  const known = statusIdSet.value
  return props.tasks.filter((t) => {
    if (t.parent_id) return false
    // Place tasks with missing/unknown status into the default column so they
    // never disappear from the board after workflow changes.
    if (!t.status_id || !known.has(t.status_id)) {
      return statusId === defaultStatusId.value
    }
    return t.status_id === statusId
  })
}

function priorityLabel(p: number) {
  if (p === 3) return 'High'
  if (p === 2) return 'Medium'
  if (p === 1) return 'Low'
  return ''
}

function priorityClass(p: number) {
  if (p === 3) return 'text-bg-danger'
  if (p === 2) return 'text-bg-warning'
  if (p === 1) return 'text-bg-secondary'
  return 'text-bg-light'
}

async function loadStatuses() {
  loading.value = true
  try {
    statuses.value = await api.listProjectStatuses(props.projectId)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load board columns', 'error')
    statuses.value = []
  } finally {
    loading.value = false
    await nextTick()
    initSortables()
  }
}

function destroySortables() {
  while (sortables.length) {
    sortables.pop()?.destroy()
  }
}

function collectIds(container: HTMLElement): number[] {
  return Array.from(container.querySelectorAll(':scope > .kanban-card'))
    .map((el) => parseInt((el as HTMLElement).dataset.taskId || '', 10))
    .filter((id) => !Number.isNaN(id))
}

async function onCardDrop(evt: Sortable.SortableEvent) {
  const to = evt.to as HTMLElement
  const from = evt.from as HTMLElement
  const statusId = parseInt(to.dataset.statusId || '', 10)
  if (Number.isNaN(statusId)) return

  const taskId = parseInt((evt.item as HTMLElement).dataset.taskId || '', 10)
  const fromStatusId = parseInt(from.dataset.statusId || '', 10)
  const orderedIds = collectIds(to)

  try {
    if (!Number.isNaN(taskId) && fromStatusId !== statusId) {
      await api.patchTask(taskId, { status_id: statusId })
    }
    await api.reorderTasks({
      task_ids: orderedIds,
      favorite: false,
      status_id: statusId,
      project: String(props.projectId),
    })
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Could not update board', 'error')
    emit('changed')
  }
}

function initSortables() {
  destroySortables()
  if (!canDrag.value) return
  const coarse = window.matchMedia('(pointer: coarse)').matches
  for (const el of columnEls.values()) {
    sortables.push(
      Sortable.create(el, {
        group: 'kanban',
        handle: '.drag-handle',
        draggable: '.kanban-card',
        animation: 150,
        delay: coarse ? 200 : 0,
        delayOnTouchOnly: true,
        touchStartThreshold: coarse ? 5 : 1,
        emptyInsertThreshold: 24,
        onEnd(evt) {
          void onCardDrop(evt)
        },
      }),
    )
  }
}

watch(
  () => props.projectId,
  () => {
    void loadStatuses()
  },
)

watch(
  () => [props.tasks, canDrag.value] as const,
  async () => {
    await nextTick()
    initSortables()
  },
  { deep: true },
)

onMounted(() => {
  void loadStatuses()
})

onBeforeUnmount(destroySortables)
</script>

<style scoped>
.kanban-column {
  width: min(280px, 85vw);
  background: var(--bs-tertiary-bg, var(--ordryn-card-bg, #f8f9fa));
  border: 1px solid var(--bs-border-color, var(--ordryn-card-border, #dee2e6));
  border-radius: 0.5rem;
  padding: 0.5rem;
  min-height: 12rem;
}
.kanban-column-body {
  min-height: 8rem;
}
.kanban-card {
  cursor: default;
  border-color: var(--bs-border-color, var(--ordryn-card-border, #dee2e6));
  background: var(--bs-body-bg, var(--ordryn-card-bg, #fff));
}
.kanban-card .drag-handle {
  cursor: grab;
  line-height: 1.2;
}
.kanban-card-title {
  white-space: normal;
  word-break: break-word;
  font-size: 0.9rem;
  line-height: 1.3;
}
.kanban-card-readonly {
  opacity: 0.98;
}
</style>
