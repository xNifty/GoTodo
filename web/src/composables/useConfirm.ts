import { reactive } from 'vue'

export type ConfirmOptions = {
  title?: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

type ConfirmState = {
  open: boolean
  title: string
  message: string
  confirmLabel: string
  cancelLabel: string
  danger: boolean
  resolve: ((value: boolean) => void) | null
}

const state = reactive<ConfirmState>({
  open: false,
  title: 'Confirm',
  message: '',
  confirmLabel: 'Confirm',
  cancelLabel: 'Cancel',
  danger: false,
  resolve: null,
})

export function useConfirm() {
  function askConfirm(options: ConfirmOptions): Promise<boolean> {
    if (state.resolve) {
      state.resolve(false)
      state.resolve = null
    }
    state.title = options.title || 'Confirm'
    state.message = options.message
    state.confirmLabel = options.confirmLabel || 'Confirm'
    state.cancelLabel = options.cancelLabel || 'Cancel'
    state.danger = !!options.danger
    state.open = true
    return new Promise((resolve) => {
      state.resolve = resolve
    })
  }

  function accept() {
    state.open = false
    state.resolve?.(true)
    state.resolve = null
  }

  function cancel() {
    state.open = false
    state.resolve?.(false)
    state.resolve = null
  }

  return { state, askConfirm, accept, cancel }
}
