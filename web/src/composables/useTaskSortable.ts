import { nextTick, onBeforeUnmount, type Ref, watch } from 'vue'
import Sortable from 'sortablejs'
import { pauseLiveReload, resumeLiveReload } from '@/composables/useLiveUpdates'

type ReorderHandler = (
  taskIds: number[],
  parentId?: number | null,
) => void | Promise<void>

export function useTaskSortable(
  taskListEl: Ref<HTMLElement | null>,
  enabled: Ref<boolean>,
  onReorder: ReorderHandler,
) {
  let rootSortable: Sortable | null = null
  const childSortables: Sortable[] = []
  let dragPaused = 0

  function beginDrag() {
    if (dragPaused === 0) pauseLiveReload()
    dragPaused += 1
  }

  function endDrag() {
    if (dragPaused === 0) return
    dragPaused -= 1
    if (dragPaused === 0) resumeLiveReload()
  }

  function destroy() {
    while (dragPaused > 0) endDrag()
    rootSortable?.destroy()
    rootSortable = null
    while (childSortables.length) {
      childSortables.pop()?.destroy()
    }
  }

  function createRootOptions(): Sortable.Options {
    const coarse = window.matchMedia('(pointer: coarse)').matches
    return {
      handle: '.drag-handle',
      draggable: '.task-tree-root',
      animation: 150,
      delay: coarse ? 200 : 0,
      delayOnTouchOnly: true,
      touchStartThreshold: coarse ? 5 : 1,
      onStart() {
        beginDrag()
      },
      onEnd(evt) {
        endDrag()
        const container = evt.to as HTMLElement
        const ids = Array.from(container.querySelectorAll(':scope > .task-tree-root'))
          .map((el) => parseInt((el as HTMLElement).dataset.taskId || '', 10))
          .filter((id) => !Number.isNaN(id))
        void onReorder(ids, null)
      },
    }
  }

  function createChildOptions(parentId: number): Sortable.Options {
    const coarse = window.matchMedia('(pointer: coarse)').matches
    return {
      handle: '.drag-handle',
      draggable: '.ordryn-task-card',
      animation: 150,
      delay: coarse ? 200 : 0,
      delayOnTouchOnly: true,
      touchStartThreshold: coarse ? 5 : 1,
      onStart() {
        beginDrag()
      },
      onEnd(evt) {
        endDrag()
        const container = evt.to as HTMLElement
        const ids = Array.from(container.querySelectorAll(':scope > .ordryn-task-card'))
          .map((el) => {
            const rawId = el.id.replace('task-card-', '').replace('task-', '')
            return parseInt(rawId, 10)
          })
          .filter((id) => !Number.isNaN(id))
        void onReorder(ids, parentId)
      },
    }
  }

  function initChildLists(root: HTMLElement | null) {
    if (!root) return
    root.querySelectorAll<HTMLElement>('.task-children[data-parent-id]').forEach((el) => {
      const parentId = parseInt(el.dataset.parentId || '', 10)
      if (Number.isNaN(parentId)) return
      childSortables.push(Sortable.create(el, createChildOptions(parentId)))
    })
  }

  function init() {
    destroy()
    if (!enabled.value) return

    if (taskListEl.value) {
      rootSortable = Sortable.create(taskListEl.value, createRootOptions())
      initChildLists(taskListEl.value)
    }
  }

  watch([enabled, taskListEl], () => {
    void nextTick(init)
  })

  onBeforeUnmount(destroy)

  return { refresh: () => void nextTick(init), destroy }
}
