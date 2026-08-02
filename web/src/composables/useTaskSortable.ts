import { nextTick, onBeforeUnmount, type Ref, watch } from 'vue'
import Sortable from 'sortablejs'

type ReorderHandler = (
  taskIds: number[],
  favorite: boolean,
  parentId?: number | null,
) => void | Promise<void>

export function useTaskSortable(
  favoriteListEl: Ref<HTMLElement | null>,
  taskListEl: Ref<HTMLElement | null>,
  enabled: Ref<boolean>,
  showFavorites: Ref<boolean>,
  onReorder: ReorderHandler,
) {
  let favSortable: Sortable | null = null
  let regSortable: Sortable | null = null
  const childSortables: Sortable[] = []

  function destroy() {
    favSortable?.destroy()
    regSortable?.destroy()
    favSortable = null
    regSortable = null
    while (childSortables.length) {
      childSortables.pop()?.destroy()
    }
  }

  function createRootOptions(favorite: boolean): Sortable.Options {
    const coarse = window.matchMedia('(pointer: coarse)').matches
    return {
      handle: '.drag-handle',
      draggable: '.task-tree-root',
      animation: 150,
      delay: coarse ? 200 : 0,
      delayOnTouchOnly: true,
      touchStartThreshold: coarse ? 5 : 1,
      onEnd(evt) {
        const container = evt.to as HTMLElement
        const ids = Array.from(container.querySelectorAll(':scope > .task-tree-root'))
          .map((el) => parseInt((el as HTMLElement).dataset.taskId || '', 10))
          .filter((id) => !Number.isNaN(id))
        void onReorder(ids, favorite, null)
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
      onEnd(evt) {
        const container = evt.to as HTMLElement
        const ids = Array.from(container.querySelectorAll(':scope > .ordryn-task-card'))
          .map((el) => {
            const rawId = el.id.replace('task-card-', '').replace('task-', '')
            return parseInt(rawId, 10)
          })
          .filter((id) => !Number.isNaN(id))
        void onReorder(ids, false, parentId)
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

    if (showFavorites.value && favoriteListEl.value) {
      favSortable = Sortable.create(favoriteListEl.value, createRootOptions(true))
      initChildLists(favoriteListEl.value)
    }
    if (taskListEl.value) {
      regSortable = Sortable.create(taskListEl.value, createRootOptions(false))
      initChildLists(taskListEl.value)
    }
  }

  watch([enabled, showFavorites, favoriteListEl, taskListEl], () => {
    void nextTick(init)
  })

  onBeforeUnmount(destroy)

  return { refresh: () => void nextTick(init), destroy }
}
