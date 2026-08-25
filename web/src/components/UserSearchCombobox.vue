<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { USER_SEARCH_MIN_QUERY, useUserSearch } from '@/composables/useUserSearch'

const props = withDefaults(
  defineProps<{
    modelValue: string
    inputId?: string
    excludeUsernames?: string[]
    projectId?: number | null
    disabled?: boolean
  }>(),
  {
    inputId: 'invite-username',
    excludeUsernames: () => [],
    projectId: null,
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const inputEl = ref<HTMLInputElement | null>(null)
const listEl = ref<HTMLElement | null>(null)
const menuStyle = ref<Record<string, string>>({})

const { filtered, loading, highlight, scheduleSearch, cancelPending, applyLocal, cachedNames } =
  useUserSearch({
    projectId: () => props.projectId,
    excludeUsernames: () => props.excludeUsernames,
  })

const showMenu = computed(() => open.value && !props.disabled && !!props.modelValue.trim())

watch(
  () => props.modelValue,
  (q) => {
    if (!q.trim()) {
      cancelPending()
      open.value = false
    }
  },
)

watch(showMenu, (visible) => {
  if (visible) {
    void nextTick(() => updateMenuPosition())
  }
})

function updateMenuPosition() {
  const el = inputEl.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const gap = 2
  const maxHeight = 220
  const spaceBelow = window.innerHeight - r.bottom - gap
  const openAbove = spaceBelow < 120 && r.top > spaceBelow
  if (openAbove) {
    menuStyle.value = {
      position: 'fixed',
      top: 'auto',
      bottom: `${window.innerHeight - r.top + gap}px`,
      left: `${r.left}px`,
      width: `${Math.max(r.width, 160)}px`,
      maxHeight: `${Math.min(maxHeight, r.top - gap)}px`,
      zIndex: '2000',
    }
  } else {
    menuStyle.value = {
      position: 'fixed',
      top: `${r.bottom + gap}px`,
      bottom: 'auto',
      left: `${r.left}px`,
      width: `${Math.max(r.width, 160)}px`,
      maxHeight: `${Math.min(maxHeight, Math.max(spaceBelow, 80))}px`,
      zIndex: '2000',
    }
  }
}

function onInput(e: Event) {
  if (props.disabled) return
  const value = (e.target as HTMLInputElement).value
  emit('update:modelValue', value)
  const q = value.trim()
  if (q.length < USER_SEARCH_MIN_QUERY) {
    cancelPending()
    open.value = false
    return
  }
  open.value = true
  scheduleSearch(value)
  void nextTick(() => updateMenuPosition())
}

function onFocus() {
  if (props.disabled) return
  const q = props.modelValue.trim()
  if (q.length < USER_SEARCH_MIN_QUERY) return
  const local = cachedNames(q)
  if (local) {
    applyLocal(local)
    open.value = true
    void nextTick(() => updateMenuPosition())
    return
  }
  open.value = true
  scheduleSearch(q)
}

function selectName(name: string) {
  emit('update:modelValue', name)
  open.value = false
}

function closeMenu() {
  open.value = false
}

function onBlur() {
  window.setTimeout(() => {
    if (!open.value) return
    const active = document.activeElement
    if (listEl.value && active && listEl.value.contains(active)) return
    closeMenu()
  }, 120)
}

function onKeydown(e: KeyboardEvent) {
  if (props.disabled) return
  if (!open.value) return

  if (e.key === 'Escape') {
    closeMenu()
    e.preventDefault()
    return
  }
  if (e.key === 'ArrowDown') {
    if (!filtered.value.length) return
    highlight.value = Math.min(highlight.value + 1, filtered.value.length - 1)
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
    const name = filtered.value[highlight.value]
    if (name) selectName(name)
    e.preventDefault()
  }
}

async function scrollHighlightIntoView() {
  await nextTick()
  const active = listEl.value?.querySelector<HTMLElement>('[data-active="true"]')
  active?.scrollIntoView({ block: 'nearest' })
}

function onReposition() {
  if (showMenu.value) updateMenuPosition()
}

onMounted(() => {
  window.addEventListener('scroll', onReposition, true)
  window.addEventListener('resize', onReposition)
})

onBeforeUnmount(() => {
  cancelPending()
  window.removeEventListener('scroll', onReposition, true)
  window.removeEventListener('resize', onReposition)
})
</script>

<template>
  <div class="user-search-combobox">
    <input
      :id="inputId"
      ref="inputEl"
      :value="modelValue"
      type="text"
      class="form-control form-control-sm"
      role="combobox"
      autocomplete="off"
      minlength="3"
      maxlength="32"
      pattern="[A-Za-z0-9_]+"
      required
      :aria-expanded="open"
      aria-autocomplete="list"
      :aria-controls="`${inputId}-listbox`"
      :disabled="disabled"
      @focus="onFocus"
      @input="onInput"
      @keydown="onKeydown"
      @blur="onBlur"
    />

    <Teleport to="body">
      <ul
        v-if="showMenu"
        :id="`${inputId}-listbox`"
        ref="listEl"
        class="user-search-menu list-unstyled mb-0"
        role="listbox"
        :style="menuStyle"
      >
        <li
          v-for="(name, idx) in filtered"
          :key="name"
          role="option"
          class="user-search-option"
          :class="{ active: highlight === idx }"
          :data-active="highlight === idx ? 'true' : undefined"
          :aria-selected="modelValue.toLowerCase() === name.toLowerCase()"
          @mousedown.prevent
          @click="selectName(name)"
        >
          {{ name }}
        </li>
        <li v-if="loading && !filtered.length" class="user-search-option text-muted" aria-disabled="true">
          Searching…
        </li>
        <li
          v-else-if="!filtered.length"
          class="user-search-option text-muted"
          aria-disabled="true"
        >
          No matching users
        </li>
      </ul>
    </Teleport>
  </div>
</template>

<style scoped>
.user-search-menu {
  overflow-y: auto;
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  border-radius: 0.375rem;
  background: var(--ordryn-card-bg, #fff);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.user-search-option {
  padding: 0.45rem 0.75rem;
  cursor: pointer;
  color: var(--ordryn-text, inherit);
  font-weight: 600;
  font-size: 0.92rem;
}

.user-search-option[aria-disabled='true'] {
  cursor: default;
  font-weight: 400;
}

.user-search-option.active,
.user-search-option:hover:not([aria-disabled='true']) {
  background: color-mix(in srgb, var(--ordryn-accent, #0d6efd) 14%, transparent);
}
</style>
