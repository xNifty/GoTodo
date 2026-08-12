import { computed, reactive } from 'vue'
import type { SavedViewFilter } from '@/api/types'

/** Default list filter: incomplete only, so the queue stays uncluttered. */
export const DEFAULT_TASK_LIST_STATUS = 'incomplete'

const filterKeys = ['status', 'due', 'completed', 'priority', 'tag', 'sort', 'project', 'search'] as const
type FilterKey = (typeof filterKeys)[number]
type TaskListFilterState = Record<FilterKey, string>

const defaultFilters: TaskListFilterState = {
  status: DEFAULT_TASK_LIST_STATUS,
  due: '',
  completed: '',
  priority: '',
  tag: '',
  sort: '',
  project: '',
  search: '',
}

export const taskListFilters = reactive<TaskListFilterState>({ ...defaultFilters })

export function useTaskListFilters() {
  const hasActiveFilters = computed(() =>
    filterKeys.some((k) => taskListFilters[k] !== defaultFilters[k]),
  )

  function toApiParams(page: number, perPage: number) {
    const params: Record<string, string | number> = { page, per_page: perPage }
    for (const key of filterKeys) {
      const v = taskListFilters[key]
      if (v) params[key] = v
    }
    return params
  }

  function toExportQuery() {
    const params = new URLSearchParams()
    for (const key of filterKeys) {
      const v = taskListFilters[key]
      if (v) params.set(key, v)
    }
    const s = params.toString()
    return s ? `&${s}` : ''
  }

  function setFilter(key: FilterKey, value: string) {
    taskListFilters[key] = value
  }

  function clearFilters() {
    for (const key of filterKeys) taskListFilters[key] = defaultFilters[key]
  }

  function applySavedView(filter: SavedViewFilter) {
    clearFilters()
    for (const key of filterKeys) {
      const v = filter[key]
      if (typeof v === 'string' && v) taskListFilters[key] = v
    }
  }

  return {
    filters: taskListFilters,
    filterKeys,
    hasActiveFilters,
    toApiParams,
    toExportQuery,
    setFilter,
    clearFilters,
    applySavedView,
  }
}
