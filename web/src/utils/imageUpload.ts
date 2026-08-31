export const IMAGE_ACCEPT =
  'image/jpeg,image/png,image/gif,image/webp,.jpg,.jpeg,.png,.gif,.webp'

const IMAGE_TYPES = new Set(['image/jpeg', 'image/png', 'image/gif', 'image/webp'])

export function isAllowedImageFile(file: File): boolean {
  const t = (file.type || '').toLowerCase()
  if (IMAGE_TYPES.has(t)) return true
  return /\.(jpe?g|png|gif|webp)$/i.test(file.name || '')
}

export function altFromFilename(name: string): string {
  const alt = (name || '')
    .replace(/\.[^.]+$/, '')
    .replace(/[[\]()]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  return alt || 'image'
}

export function toImageMarkdown(url: string, alt = 'image'): string {
  const safeAlt = alt.replace(/[[\]()]/g, ' ').replace(/\s+/g, ' ').trim() || 'image'
  return `![${safeAlt}](${url})`
}

let overlayGuardUntil = 0

/** Native file pickers steal focus; the following click often hits the task overlay. */
export function armTaskOverlayFileGuard() {
  overlayGuardUntil = Date.now() + 1500
}

export function shouldIgnoreTaskOverlayClose() {
  return Date.now() < overlayGuardUntil
}

export function imageFileFromClipboard(e: ClipboardEvent): File | null {
  const items = e.clipboardData?.items
  if (!items) return null
  for (const item of items) {
    if (item.kind === 'file' && item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (file) return file
    }
  }
  return null
}

export function imageFileFromDrop(e: DragEvent): File | null {
  const files = e.dataTransfer?.files
  if (!files?.length) return null
  for (const file of Array.from(files)) {
    if (isAllowedImageFile(file) || file.type.startsWith('image/')) return file
  }
  return null
}

export function dropHasFiles(e: DragEvent): boolean {
  return !!e.dataTransfer?.types && Array.from(e.dataTransfer.types).includes('Files')
}
