<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { api } from '@/api/client'

const SEARCH_LIMIT = 10
const MIN_QUERY = 2
const DEBOUNCE_MS = 300

type CacheEntry = { names: string[]; complete: boolean }

const searchCache = new Map<string, CacheEntry>()

function cacheKey(q: string) {
  return q.trim().toLowerCase()
}

function filterPrefix(names: string[], key: string) {
  return names.filter((n) => n.toLowerCase().startsWith(key))
}

/** Cached names for q, or null if a network fetch is still needed. */
function namesFromCache(q: string): string[] | null {
  const key = cacheKey(q)
  if (key.length < MIN_QUERY) return []

  const exact = searchCache.get(key)
  if (exact) return filterPrefix(exact.names, key)

  for (let len = key.length - 1; len >= MIN_QUERY; len--) {
    const entry = searchCache.get(key.slice(0, len))
    if (entry?.complete) return filterPrefix(entry.names, key)
  }
  return null
}

function remember(q: string, names: string[]) {
  searchCache.set(cacheKey(q), { names, complete: names.length < SEARCH_LIMIT })
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    inputId?: string
    excludeUsernames?: string[]
    disabled?: boolean
  }>(),
  {
    inputId: 'invite-username',
    excludeUsernames: () => [],
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const highlight = ref(0)
const inputEl = ref<HTMLInputElement | null>(null)
const listEl = ref<HTMLElement | null>(null)
const hits = ref<string[]>([])
const loading = ref(false)
const menuStyle = ref<Record<string, string>>({})

const excludeSet = computed(
  () => new Set(props.excludeUsernames.map((n) => n.toLowerCase()).filter(Boolean)),
)

const filtered = computed(() =>
  hits.value.filter((name) => !excludeSet.value.has(name.toLowerCase())),
)

const showMenu = computed(() => open.value && !props.disabled && !!props.modelValue.trim())

let debounceTimer: number | undefined
let searchGen = 0
let abort: AbortController | undefined

function abortInFlight() {
  abort?.abort()
  abort = undefined
}

function cancelPending() {
  window.clearTimeout(debounceTimer)
  abortInFlight()
  searchGen += 1
  loading.value = false
}

watch(
  () => props.modelValue,
  (q) => {
    if (!q.trim()) {
      cancelPending()
      hits.value = []
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

function applyLocal(names: string[]) {
  hits.value = names
  open.value = true
  highlight.value = 0
  loading.value = false
  void nextTick(() => updateMenuPosition())
}

function scheduleSearch(raw: string) {
  window.clearTimeout(debounceTimer)
  const q = raw.trim()
  if (q.length < MIN_QUERY) {
    cancelPending()
    hits.value = []
    open.value = false
    return
  }

  const local = namesFromCache(q)
  if (local) {
    cancelPending()
    applyLocal(local)
    return
  }

  open.value = true
  loading.value = true
  void nextTick(() => updateMenuPosition())
  debounceTimer = window.setTimeout(() => {
    void runSearch(q)
  }, DEBOUNCE_MS)
}

async function runSearch(q: string) {
  const local = namesFromCache(q)
  if (local) {
    applyLocal(local)
    return
  }

  const gen = ++searchGen
  abortInFlight()
  abort = new AbortController()
  const { signal } = abort
  loading.value = true
  try {
    const results = await api.searchUsers(q, { signal })
    if (gen !== searchGen) return
    const names = results.map((h) => h.user_name)
    remember(q, names)
    hits.value = names
    open.value = true
    highlight.value = 0
    void nextTick(() => updateMenuPosition())
  } catch {
    if (gen !== searchGen || signal.aborted) return
    hits.value = []
    open.value = true
  } finally {
    if (gen === searchGen) loading.value = false
  }
}

function onInput(e: Event) {
  if (props.disabled) return
  const value = (e.target as HTMLInputElement).value
  emit('update:modelValue', value)
  scheduleSearch(value)
}

function onFocus() {
  if (props.disabled) return
  const q = props.modelValue.trim()
  if (q.length < MIN_QUERY) return
  const local = namesFromCache(q)
  if (local) {
    applyLocal(local)
    return
  }
  scheduleSearch(q)
}

function selectName(name: string) {
  emit('update:modelValue', name)
  hits.value = []
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
