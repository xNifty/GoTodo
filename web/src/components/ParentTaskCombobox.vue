<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { Task } from '@/api/types'

const props = withDefaults(
  defineProps<{
    modelValue: number | ''
    options: Task[]
    inputId?: string
    disabled?: boolean
    placeholder?: string
    noneLabel?: string
  }>(),
  {
    inputId: 'parent_id',
    disabled: false,
    placeholder: 'Search parent tasks…',
    noneLabel: 'No parent — top level',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | '']
  change: []
}>()

const query = ref('')
const open = ref(false)
const highlight = ref(0)
const inputEl = ref<HTMLInputElement | null>(null)
const listEl = ref<HTMLElement | null>(null)

const selected = computed(() =>
  props.modelValue === '' ? null : props.options.find((t) => t.id === props.modelValue) ?? null,
)

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.options
  return props.options.filter((t) => {
    const title = t.title.toLowerCase()
    const project = (t.project || '').toLowerCase()
    return title.includes(q) || project.includes(q)
  })
})

const menuItems = computed(() => {
  // Always offer "no parent" first, then matches.
  return [{ id: 0 as const, kind: 'none' as const }, ...filtered.value.map((t) => ({ id: t.id, kind: 'task' as const, task: t }))]
})

watch(
  () => props.modelValue,
  () => {
    if (!open.value) syncQueryFromSelection()
  },
  { immediate: true },
)

watch(
  () => props.options,
  () => {
    if (!open.value) syncQueryFromSelection()
  },
)

function syncQueryFromSelection() {
  query.value = selected.value?.title || ''
}

function openMenu() {
  if (props.disabled) return
  open.value = true
  highlight.value = 0
  // Clear to search; keep selection until a new pick is made.
  query.value = ''
}

function closeMenu(restore = true) {
  open.value = false
  if (restore) syncQueryFromSelection()
}

function selectNone() {
  emit('update:modelValue', '')
  emit('change')
  query.value = ''
  open.value = false
}

function selectTask(task: Task) {
  emit('update:modelValue', task.id)
  emit('change')
  query.value = task.title
  open.value = false
}

function onInput() {
  if (props.disabled) return
  open.value = true
  highlight.value = 0
  // Typing clears a previous selection until a match is chosen again.
  if (props.modelValue !== '') {
    emit('update:modelValue', '')
    emit('change')
  }
}

function onBlur() {
  // Delay so option mousedown/click can fire first.
  window.setTimeout(() => {
    if (!open.value) return
    closeMenu(true)
  }, 120)
}

function onKeydown(e: KeyboardEvent) {
  if (props.disabled) return
  if (!open.value && (e.key === 'ArrowDown' || e.key === 'Enter')) {
    openMenu()
    e.preventDefault()
    return
  }
  if (!open.value) return

  if (e.key === 'Escape') {
    closeMenu(true)
    e.preventDefault()
    return
  }
  if (e.key === 'ArrowDown') {
    highlight.value = Math.min(highlight.value + 1, menuItems.value.length - 1)
    scrollHighlightIntoView()
    e.preventDefault()
    return
  }
  if (e.key === 'ArrowUp') {
    highlight.value = Math.max(highlight.value - 1, 0)
    scrollHighlightIntoView()
    e.preventDefault()
    return
  }
  if (e.key === 'Enter') {
    const item = menuItems.value[highlight.value]
    if (!item || item.kind === 'none') selectNone()
    else selectTask(item.task)
    e.preventDefault()
  }
}

async function scrollHighlightIntoView() {
  await nextTick()
  const active = listEl.value?.querySelector<HTMLElement>('[data-active="true"]')
  active?.scrollIntoView({ block: 'nearest' })
}

function clearSelection() {
  if (props.disabled) return
  selectNone()
  void nextTick(() => inputEl.value?.focus())
}
</script>

<template>
  <div class="parent-task-combobox position-relative">
    <div class="input-group">
      <input
        :id="inputId"
        ref="inputEl"
        v-model="query"
        type="text"
        class="form-control"
        role="combobox"
        autocomplete="off"
        :aria-expanded="open"
        aria-autocomplete="list"
        :aria-controls="`${inputId}-listbox`"
        :disabled="disabled"
        :placeholder="placeholder"
        @focus="openMenu"
        @input="onInput"
        @keydown="onKeydown"
        @blur="onBlur"
      />
      <button
        v-if="modelValue !== '' && !disabled"
        type="button"
        class="btn btn-outline-secondary"
        title="Clear parent"
        tabindex="-1"
        @mousedown.prevent
        @click="clearSelection"
      >
        <i class="bi bi-x" />
      </button>
    </div>

    <ul
      v-if="open && !disabled"
      :id="`${inputId}-listbox`"
      ref="listEl"
      class="parent-task-menu list-unstyled mb-0"
      role="listbox"
    >
      <li
        role="option"
        class="parent-task-option"
        :class="{ active: highlight === 0 }"
        :data-active="highlight === 0 ? 'true' : undefined"
        :aria-selected="modelValue === ''"
        @mousedown.prevent
        @click="selectNone"
      >
        {{ noneLabel }}
      </li>
      <li
        v-for="(task, idx) in filtered"
        :key="task.id"
        role="option"
        class="parent-task-option"
        :class="{ active: highlight === idx + 1 }"
        :data-active="highlight === idx + 1 ? 'true' : undefined"
        :aria-selected="modelValue === task.id"
        @mousedown.prevent
        @click="selectTask(task)"
      >
        <span class="parent-task-option-title text-truncate">{{ task.title }}</span>
        <span v-if="task.project" class="parent-task-option-project text-truncate">{{ task.project }}</span>
      </li>
      <li v-if="!filtered.length" class="parent-task-option text-muted" aria-disabled="true">
        No matching tasks
      </li>
    </ul>
    <small v-if="selected?.project" class="form-hint d-block mt-1">In project {{ selected.project }}</small>
  </div>
</template>

<style scoped>
.parent-task-menu {
  position: absolute;
  z-index: 20;
  left: 0;
  right: 0;
  top: calc(100% + 2px);
  max-height: 220px;
  overflow-y: auto;
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  border-radius: 0.375rem;
  background: var(--ordryn-card-bg, #fff);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.parent-task-option {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  padding: 0.45rem 0.75rem;
  cursor: pointer;
  color: var(--ordryn-text, inherit);
}

.parent-task-option.active,
.parent-task-option:hover {
  background: color-mix(in srgb, var(--ordryn-accent, #0d6efd) 14%, transparent);
}

.parent-task-option-title {
  font-weight: 600;
  font-size: 0.92rem;
}

.parent-task-option-project {
  font-size: 0.75rem;
  color: var(--ordryn-muted, #6c757d);
}
</style>
