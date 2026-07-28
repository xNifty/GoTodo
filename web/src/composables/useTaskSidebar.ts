import { ref } from 'vue'
import type { Task } from '@/api/types'

const open = ref(false)
const mode = ref<'add' | 'edit' | 'view'>('add')
const taskId = ref<number | null>(null)
const defaultDueDate = ref('')
const defaultProjectId = ref<number | string | null>(null)
const lastSavedTask = ref<Task | null>(null)

export function useTaskSidebar() {
  function openAdd(dueDate?: string, projectId?: number | string | null) {
    mode.value = 'add'
    taskId.value = null
    // Ignore PointerEvent when bound as @click="openAdd"
    defaultDueDate.value = typeof dueDate === 'string' ? dueDate.trim() : ''
    defaultProjectId.value = (typeof projectId === 'number' || typeof projectId === 'string') && projectId !== '0' ? projectId : null
    open.value = true
  }

  function openEdit(id: number) {
    mode.value = 'edit'
    taskId.value = id
    defaultDueDate.value = ''
    defaultProjectId.value = null
    open.value = true
  }

  function openView(id: number) {
    mode.value = 'view'
    taskId.value = id
    defaultDueDate.value = ''
    open.value = true
  }

  function close() {
    open.value = false
  }

  function notifySaved(task: Task, closeDrawer = true) {
    lastSavedTask.value = task
    if (closeDrawer) {
      close()
    }
  }

  return {
    open,
    mode,
    taskId,
    defaultDueDate,
    defaultProjectId,
    lastSavedTask,
    openAdd,
    openEdit,
    openView,
    close,
    notifySaved,
  }
}
