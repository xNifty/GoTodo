<script setup lang="ts">
import { nextTick, onUnmounted, ref, useAttrs, watch } from 'vue'
import RichBody from '@/components/RichBody.vue'
import { useImageUpload } from '@/composables/useImageUpload'
import { useToast } from '@/composables/useToast'
import { insertMarkdownAtCursor, mentionTokenAtCursor } from '@/utils/taskCommentBody'
import { IMAGE_ACCEPT, armTaskOverlayFileGuard, dropHasFiles, imageFileFromClipboard, imageFileFromDrop } from '@/utils/imageUpload'
import {
  continueListOnEnter,
  insertLinkMarkup,
  normalizeLinkHref,
  prefixLines,
  wrapInline,
  wouldExceedLimit,
  type ToolbarRange,
} from '@/utils/markdownToolbar'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    maxlength?: number
    placeholder?: string
    disabled?: boolean
    id?: string
    rows?: number
    taskTitle?: (id: number) => string | undefined
  }>(),
  {
    maxlength: 2000,
    placeholder: '',
    disabled: false,
    rows: 3,
  },
)

const model = defineModel<string>({ default: '' })
const emit = defineEmits<{
  'open-task': [id: number]
}>()

const attrs = useAttrs()
const toast = useToast()
const { enabled: imageEnabled, busy: imageBusy, uploadImageFile } = useImageUpload()

const textareaEl = ref<HTMLTextAreaElement | null>(null)
const fileEl = ref<HTMLInputElement | null>(null)
const mode = ref<'write' | 'preview'>('write')
const linkOpen = ref(false)
const linkTitle = ref('')
const linkHref = ref('')
const linkError = ref('')
const linkSel = ref({ start: 0, end: 0 })
const linkTitleEl = ref<HTMLInputElement | null>(null)
const linkHrefEl = ref<HTMLInputElement | null>(null)

function selection(): { start: number; end: number } {
  const el = textareaEl.value
  const fallback = model.value.length
  return {
    start: el?.selectionStart ?? fallback,
    end: el?.selectionEnd ?? fallback,
  }
}

async function applyRange(next: ToolbarRange) {
  if (wouldExceedLimit(next.body.length, props.maxlength)) {
    toast.push(`Comment would exceed ${props.maxlength} characters`, 'error')
    return
  }
  model.value = next.body
  mode.value = 'write'
  await nextTick()
  const el = textareaEl.value
  if (!el) return
  el.focus()
  el.setSelectionRange(next.start, next.end)
}

function wrap(marker: string) {
  if (props.disabled) return
  const { start, end } = selection()
  void applyRange(wrapInline(model.value, start, end, marker))
}

function list(kind: 'ul' | 'ol') {
  if (props.disabled) return
  const { start, end } = selection()
  void applyRange(prefixLines(model.value, start, end, kind))
}

function openLinkDialog() {
  if (props.disabled) return
  const { start, end } = selection()
  linkSel.value = { start, end }
  linkTitle.value = model.value.slice(start, end)
  linkHref.value = ''
  linkError.value = ''
  linkOpen.value = true
  mode.value = 'write'
  void nextTick(() => {
    if (linkTitle.value.trim()) linkHrefEl.value?.focus()
    else linkTitleEl.value?.focus()
  })
}

function closeLinkDialog() {
  if (!linkOpen.value) return
  linkOpen.value = false
  linkError.value = ''
  void nextTick(() => textareaEl.value?.focus())
}

function saveLink() {
  const href = normalizeLinkHref(linkHref.value)
  if (!href) {
    linkError.value = 'Enter a valid http(s) URL'
    linkHrefEl.value?.focus()
    return
  }
  const title = linkTitle.value.trim() || href
  const next = insertLinkMarkup(model.value, linkSel.value.start, linkSel.value.end, title, href)
  linkOpen.value = false
  linkError.value = ''
  void applyRange(next)
}

function onLinkDialogKey(e: KeyboardEvent) {
  if (!linkOpen.value) return
  if (e.key !== 'Escape') return
  e.preventDefault()
  e.stopPropagation()
  e.stopImmediatePropagation()
  closeLinkDialog()
}

watch(linkOpen, (open) => {
  if (open) window.addEventListener('keydown', onLinkDialogKey, true)
  else window.removeEventListener('keydown', onLinkDialogKey, true)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onLinkDialogKey, true)
})

function insertImageMarkdown(markdown: string) {
  const { start, end } = selection()
  const next = insertMarkdownAtCursor(model.value, markdown, start, end)
  void applyRange({ body: next.body, start: next.cursor, end: next.cursor })
}

function pickImage() {
  if (props.disabled || imageBusy.value || !imageEnabled.value) return
  armTaskOverlayFileGuard()
  fileEl.value?.click()
}

async function onFile(ev: Event) {
  armTaskOverlayFileGuard()
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  const markdown = await uploadImageFile(file)
  if (markdown) insertImageMarkdown(markdown)
}

async function ingestImageFile(file: File) {
  const markdown = await uploadImageFile(file)
  if (markdown) insertImageMarkdown(markdown)
}

function onPaste(e: ClipboardEvent) {
  const file = imageFileFromClipboard(e)
  if (!file || !imageEnabled.value) return
  e.preventDefault()
  void ingestImageFile(file)
}

function onDragOver(e: DragEvent) {
  if (!imageEnabled.value || !dropHasFiles(e)) return
  e.preventDefault()
}

function onDrop(e: DragEvent) {
  const file = imageFileFromDrop(e)
  if (!file || !imageEnabled.value) return
  e.preventDefault()
  void ingestImageFile(file)
}

function onKeydown(e: KeyboardEvent) {
  if (props.disabled) return
  const mod = e.metaKey || e.ctrlKey
  if (mod && !e.altKey) {
    const key = e.key.toLowerCase()
    if (key === 'b') {
      e.preventDefault()
      wrap('**')
      return
    }
    if (key === 'i') {
      e.preventDefault()
      wrap('*')
      return
    }
    if (key === 'u') {
      e.preventDefault()
      wrap('++')
    }
    return
  }
  if (e.key !== 'Enter' || e.shiftKey || e.altKey) return
  if (e.defaultPrevented) return
  const { start, end } = selection()
  if (start !== end) return
  if (mentionTokenAtCursor(model.value, start)) return
  const next = continueListOnEnter(model.value, start)
  if (!next) return
  e.preventDefault()
  void applyRange(next)
}

function onInput(e: Event) {
  model.value = (e.target as HTMLTextAreaElement).value
}

defineExpose({
  textarea: textareaEl,
  focus() {
    textareaEl.value?.focus()
  },
})
</script>

<template>
  <div class="md-editor">
    <div class="md-editor-bar">
      <div class="md-editor-tools" role="toolbar" aria-label="Markdown formatting">
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary md-editor-tool"
          title="Bold"
          :disabled="disabled"
          @mousedown.prevent
          @click="wrap('**')"
        >
          <i class="bi bi-type-bold" aria-hidden="true" />
          <span class="visually-hidden">Bold</span>
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary md-editor-tool"
          title="Italic"
          :disabled="disabled"
          @mousedown.prevent
          @click="wrap('*')"
        >
          <i class="bi bi-type-italic" aria-hidden="true" />
          <span class="visually-hidden">Italic</span>
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary md-editor-tool"
          title="Underline"
          :disabled="disabled"
          @mousedown.prevent
          @click="wrap('++')"
        >
          <i class="bi bi-type-underline" aria-hidden="true" />
          <span class="visually-hidden">Underline</span>
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary md-editor-tool"
          title="Numbered list"
          :disabled="disabled"
          @mousedown.prevent
          @click="list('ol')"
        >
          <i class="bi bi-list-ol" aria-hidden="true" />
          <span class="visually-hidden">Numbered list</span>
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary md-editor-tool"
          title="Bulleted list"
          :disabled="disabled"
          @mousedown.prevent
          @click="list('ul')"
        >
          <i class="bi bi-list-ul" aria-hidden="true" />
          <span class="visually-hidden">Bulleted list</span>
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary md-editor-tool"
          title="Insert link"
          :disabled="disabled"
          @mousedown.prevent
          @click="openLinkDialog"
        >
          <i class="bi bi-link-45deg" aria-hidden="true" />
          <span class="visually-hidden">Insert link</span>
        </button>
        <span v-if="imageEnabled" class="md-editor-image">
          <input
            ref="fileEl"
            type="file"
            class="d-none"
            :accept="IMAGE_ACCEPT"
            :disabled="disabled || imageBusy"
            @change="onFile"
          />
          <button
            type="button"
            class="btn btn-sm btn-outline-secondary md-editor-tool"
            title="Upload an image into this comment"
            :disabled="disabled || imageBusy"
            @mousedown.prevent
            @click="pickImage"
          >
            <i class="bi bi-image" aria-hidden="true" />
            <span class="visually-hidden">{{ imageBusy ? 'Uploading…' : 'Insert image' }}</span>
          </button>
        </span>
      </div>
      <div class="btn-group md-editor-modes" role="group" aria-label="Editor mode">
        <button
          type="button"
          class="btn btn-sm"
          :class="mode === 'write' ? 'btn-secondary' : 'btn-outline-secondary'"
          :aria-pressed="mode === 'write'"
          @click="mode = 'write'"
        >
          Write
        </button>
        <button
          type="button"
          class="btn btn-sm"
          :class="mode === 'preview' ? 'btn-secondary' : 'btn-outline-secondary'"
          :aria-pressed="mode === 'preview'"
          @click="mode = 'preview'"
        >
          Preview
        </button>
      </div>
    </div>
    <textarea
      v-show="mode === 'write'"
      :id="id"
      ref="textareaEl"
      class="form-control md-editor-input"
      :value="model"
      :maxlength="maxlength"
      :placeholder="placeholder"
      :disabled="disabled"
      :rows="rows"
      v-bind="attrs"
      @input="onInput"
      @keydown="onKeydown"
      @paste="onPaste"
      @dragover="onDragOver"
      @drop="onDrop"
    />
    <div v-if="mode === 'preview'" class="md-editor-preview" role="region" aria-label="Comment preview">
      <RichBody
        v-if="model.trim()"
        :body="model"
        :task-title="taskTitle"
        @open-task="emit('open-task', $event)"
      />
      <p v-else class="small text-muted mb-0">Nothing to preview</p>
    </div>
    <small class="text-muted md-editor-count">{{ model.length }}/{{ maxlength }}</small>
    <Teleport to="body">
      <div
        v-if="linkOpen"
        class="modal fade show d-block md-link-modal"
        tabindex="-1"
        role="dialog"
        aria-modal="true"
        aria-labelledby="md-link-modal-title"
        @click.self="closeLinkDialog"
      >
        <div class="modal-dialog modal-dialog-centered" @click.stop>
          <div class="modal-content border-0 shadow">
            <form @submit.prevent="saveLink">
              <div class="modal-header">
                <h5 id="md-link-modal-title" class="modal-title">Insert link</h5>
                <button type="button" class="btn-close" aria-label="Close" @click="closeLinkDialog" />
              </div>
              <div class="modal-body">
                <div class="mb-3">
                  <label class="form-label" for="md-link-title">Title</label>
                  <input
                    id="md-link-title"
                    ref="linkTitleEl"
                    v-model="linkTitle"
                    type="text"
                    class="form-control"
                    maxlength="200"
                    placeholder="Link text"
                    autocomplete="off"
                  />
                </div>
                <div class="mb-0">
                  <label class="form-label" for="md-link-href">URL</label>
                  <input
                    id="md-link-href"
                    ref="linkHrefEl"
                    v-model="linkHref"
                    type="text"
                    inputmode="url"
                    class="form-control"
                    :class="{ 'is-invalid': !!linkError }"
                    placeholder="https://example.com"
                    autocomplete="off"
                  />
                  <div v-if="linkError" class="invalid-feedback d-block">{{ linkError }}</div>
                </div>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary" @click="closeLinkDialog">Cancel</button>
                <button type="submit" class="btn btn-primary">Insert</button>
              </div>
            </form>
          </div>
        </div>
      </div>
      <div v-if="linkOpen" class="modal-backdrop fade show md-link-backdrop" />
    </Teleport>
  </div>
</template>

<style scoped>
.md-editor {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}
.md-editor-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem;
}
.md-editor-tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.2rem;
}
.md-editor-tool {
  width: 2rem;
  height: 2rem;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.md-editor-image {
  display: inline-flex;
}
.md-editor-input {
  min-height: 80px;
  resize: vertical;
}
.md-editor-preview {
  min-height: 80px;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  border-radius: 0.375rem;
  background: var(--ordryn-muted-bg, #f8f6ee);
  color: var(--ordryn-text);
}
.md-editor-count {
  align-self: flex-start;
}
.md-link-modal.modal {
  z-index: 1070;
}
.md-link-backdrop.modal-backdrop {
  z-index: 1065;
}
</style>
