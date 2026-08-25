<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { TaskComment } from '@/api/types'
import { useConfirm } from '@/composables/useConfirm'
import { useTaskSidebar } from '@/composables/useTaskSidebar'
import { useToast } from '@/composables/useToast'
import { USER_SEARCH_MIN_QUERY, useUserSearch } from '@/composables/useUserSearch'
import {
  extractTaskRefIDs,
  insertMention,
  insertTaskRef,
  isInsertedTaskRef,
  mentionTokenAtCursor,
  splitCommentBody,
  type MentionToken,
} from '@/utils/taskCommentBody'

const props = defineProps<{
  taskId: number
  projectId?: number | null
  currentUserId: number | null
  isOwner: boolean
  /** Grow with the parent column instead of capping the thread at 360px. */
  fillHeight?: boolean
}>()

const toast = useToast()
const { askConfirm } = useConfirm()
const { openEdit } = useTaskSidebar()

const comments = ref<TaskComment[]>([])
const loading = ref(false)
const posting = ref(false)
const draft = ref('')
const draftEl = ref<HTMLTextAreaElement | null>(null)
const mentionListEl = ref<HTMLElement | null>(null)
const bottomEl = ref<HTMLElement | null>(null)
const MAX_BODY = 2000

type DraftLink = {
  id: number
  title: string
  inserted: boolean
}

const draftLinks = ref<DraftLink[]>([])
const titleCache = new Map<number, string | null>()
let lookupTimer: ReturnType<typeof setTimeout> | null = null

const mentionOpen = ref(false)
const mentionToken = ref<MentionToken | null>(null)
const mentionMenuStyle = ref<Record<string, string>>({})

const { filtered: mentionHits, loading: mentionLoading, highlight: mentionHighlight, scheduleSearch, cancelPending, applyLocal, cachedNames } =
  useUserSearch({
    projectId: () => props.projectId,
  })

const showMentionMenu = computed(
  () =>
    mentionOpen.value &&
    !!props.projectId &&
    (mentionToken.value?.query.length ?? 0) >= USER_SEARCH_MIN_QUERY,
)

async function reload() {
  if (!props.taskId) return
  loading.value = true
  try {
    comments.value = await api.listTaskComments(props.taskId)
    await nextTick()
    bottomEl.value?.scrollIntoView({ block: 'nearest' })
  } catch (err) {
    toast.push(err instanceof Error ? err.message : 'Could not load discussion', 'error')
  } finally {
    loading.value = false
  }
}

function authorLabel(c: TaskComment) {
  if (props.currentUserId && c.user_id === props.currentUserId) return 'You'
  return c.user_name || `User #${c.user_id}`
}

function initials(c: TaskComment) {
  const name = authorLabel(c).trim()
  if (name === 'You') return 'YO'
  const parts = name.split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase()
  return name.slice(0, 2).toUpperCase() || '?'
}

function formatWhen(iso: string) {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

function tombstone(c: TaskComment) {
  if (c.deleted_by_kind === 'owner') return 'Message deleted by project owner'
  return 'Message deleted by user'
}

function canDelete(c: TaskComment) {
  if (c.deleted) return false
  if (props.currentUserId && c.user_id === props.currentUserId) return true
  return props.isOwner
}

function linkTitle(c: TaskComment, id: number) {
  return c.links?.find((l) => l.id === id)?.title
}

function taskLinkLabel(id: number, title?: string) {
  return title ? `Task #${id} - ${title}` : `Task #${id}`
}

function openLinkedTask(id: number) {
  if (id === props.taskId) return
  openEdit(id)
}

async function lookupDraftLinks() {
  const ids = extractTaskRefIDs(draft.value)
  const next: DraftLink[] = []
  for (const id of ids) {
    if (!titleCache.has(id)) {
      try {
        const task = await api.getTask(id)
        titleCache.set(id, task.title || `Task #${id}`)
      } catch {
        titleCache.set(id, null)
      }
    }
    const title = titleCache.get(id)
    if (!title) continue
    next.push({
      id,
      title,
      inserted: isInsertedTaskRef(draft.value, id),
    })
  }
  draftLinks.value = next
}

function scheduleDraftLookup() {
  if (lookupTimer) clearTimeout(lookupTimer)
  lookupTimer = setTimeout(() => {
    void lookupDraftLinks()
  }, 250)
}

function insertLink(id: number) {
  draft.value = insertTaskRef(draft.value, id)
  void lookupDraftLinks()
}

function updateMentionMenuPosition() {
  const el = draftEl.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const gap = 2
  const maxHeight = 220
  const spaceBelow = window.innerHeight - r.bottom - gap
  const openAbove = spaceBelow < 120 && r.top > spaceBelow
  if (openAbove) {
    mentionMenuStyle.value = {
      position: 'fixed',
      top: 'auto',
      bottom: `${window.innerHeight - r.top + gap}px`,
      left: `${r.left}px`,
      width: `${Math.max(r.width, 160)}px`,
      maxHeight: `${Math.min(maxHeight, r.top - gap)}px`,
      zIndex: '2000',
    }
  } else {
    mentionMenuStyle.value = {
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

function closeMentionMenu() {
  mentionOpen.value = false
  mentionToken.value = null
  cancelPending()
}

function syncMentionFromEl() {
  const el = draftEl.value
  if (!el || !props.projectId) {
    closeMentionMenu()
    return
  }
  const token = mentionTokenAtCursor(el.value, el.selectionStart ?? 0)
  mentionToken.value = token
  if (!token || token.query.length < USER_SEARCH_MIN_QUERY) {
    cancelPending()
    mentionOpen.value = false
    return
  }
  mentionOpen.value = true
  const local = cachedNames(token.query)
  if (local) {
    cancelPending()
    applyLocal(local)
  } else {
    scheduleSearch(token.query)
  }
  void nextTick(() => updateMentionMenuPosition())
}

function selectMention(name: string) {
  const token = mentionToken.value
  if (!token) return
  const inserted = insertMention(draft.value, token, name)
  draft.value = inserted.body
  closeMentionMenu()
  void nextTick(() => {
    const el = draftEl.value
    if (el) {
      el.focus()
      el.setSelectionRange(inserted.cursor, inserted.cursor)
    }
    scheduleDraftLookup()
  })
}

async function scrollMentionHighlight() {
  await nextTick()
  const active = mentionListEl.value?.querySelector<HTMLElement>('[data-active="true"]')
  active?.scrollIntoView({ block: 'nearest' })
}

function onDraftInput(e: Event) {
  const el = e.target as HTMLTextAreaElement
  draft.value = el.value
  syncMentionFromEl()
}

function onDraftKeydown(e: KeyboardEvent) {
  if (!showMentionMenu.value) return
  if (e.key === 'Escape') {
    closeMentionMenu()
    e.preventDefault()
    return
  }
  if (e.key === 'ArrowDown') {
    if (!mentionHits.value.length) return
    mentionHighlight.value = Math.min(mentionHighlight.value + 1, mentionHits.value.length - 1)
    void scrollMentionHighlight()
    e.preventDefault()
    return
  }
  if (e.key === 'ArrowUp') {
    mentionHighlight.value = Math.max(mentionHighlight.value - 1, 0)
    void scrollMentionHighlight()
    e.preventDefault()
    return
  }
  if (e.key === 'Enter' || e.key === 'Tab') {
    const name = mentionHits.value[mentionHighlight.value]
    if (name) {
      selectMention(name)
      e.preventDefault()
    }
  }
}

function onMentionReposition() {
  if (showMentionMenu.value) updateMentionMenuPosition()
}

const charCount = computed(() => draft.value.length)

async function post() {
  const body = draft.value.trim()
  if (!body || posting.value) return
  posting.value = true
  try {
    const created = await api.addTaskComment(props.taskId, body)
    comments.value = [...comments.value, created]
    draft.value = ''
    draftLinks.value = []
    closeMentionMenu()
    await nextTick()
    bottomEl.value?.scrollIntoView({ block: 'nearest' })
  } catch (err) {
    toast.push(err instanceof Error ? err.message : 'Could not post comment', 'error')
  } finally {
    posting.value = false
  }
}

async function remove(c: TaskComment) {
  const ok = await askConfirm({
    title: 'Delete comment',
    message: 'Delete this comment? Others will see that it was removed.',
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  try {
    await api.deleteTaskComment(props.taskId, c.id)
    await reload()
  } catch (err) {
    toast.push(err instanceof Error ? err.message : 'Could not delete comment', 'error')
  }
}

watch(
  () => props.taskId,
  () => {
    void reload()
  },
)

watch(draft, () => {
  scheduleDraftLookup()
})

watch(showMentionMenu, (visible) => {
  if (visible) void nextTick(() => updateMentionMenuPosition())
})

onMounted(() => {
  void reload()
  window.addEventListener('scroll', onMentionReposition, true)
  window.addEventListener('resize', onMentionReposition)
})

onUnmounted(() => {
  if (lookupTimer) clearTimeout(lookupTimer)
  cancelPending()
  window.removeEventListener('scroll', onMentionReposition, true)
  window.removeEventListener('resize', onMentionReposition)
})

defineExpose({ reload })
</script>

<template>
  <div class="task-discussion" :class="fillHeight ? 'task-discussion--fill' : 'mt-3'">
    <h3 class="h6 mb-2">Discussion</h3>
    <p v-if="loading && !comments.length" class="text-muted small mb-2">Loading discussion…</p>
    <div v-else class="task-discussion-thread mb-2">
      <ul v-if="comments.length" class="list-unstyled mb-0">
        <li v-for="c in comments" :key="c.id" class="task-discussion-post">
          <article class="task-post-card" :class="{ 'task-post-card--deleted': c.deleted }">
            <header class="task-post-header">
              <div class="task-post-avatar" aria-hidden="true">{{ initials(c) }}</div>
              <div class="task-post-meta">
                <div class="task-post-author text-truncate">{{ authorLabel(c) }}</div>
                <time class="task-post-time" :datetime="c.created_at">{{ formatWhen(c.created_at) }}</time>
              </div>
              <button
                v-if="canDelete(c)"
                class="btn btn-sm btn-link text-danger p-0 ms-auto"
                type="button"
                @click="remove(c)"
              >
                Delete
              </button>
            </header>
            <div v-if="c.deleted" class="task-post-body fst-italic text-muted">
              {{ tombstone(c) }}
            </div>
            <div v-else class="task-post-body">
              <template v-for="(part, i) in splitCommentBody(c.body)" :key="i">
                <span v-if="part.type === 'text'" style="white-space: pre-wrap;">{{ part.value }}</span>
                <span v-else-if="part.type === 'mention'" class="task-post-mention">{{ part.raw }}</span>
                <button
                  v-else-if="linkTitle(c, part.id)"
                  type="button"
                  class="task-post-task-link"
                  @click="openLinkedTask(part.id)"
                >
                  {{ taskLinkLabel(part.id, linkTitle(c, part.id)) }}
                </button>
                <span v-else>#{{ part.id }}</span>
              </template>
            </div>
          </article>
        </li>
      </ul>
      <p v-else class="small text-muted mb-0">No comments yet. Start the discussion.</p>
      <div ref="bottomEl" />
    </div>
    <label class="form-label small mb-1" for="task-comment-body">Add a comment</label>
    <textarea
      id="task-comment-body"
      ref="draftEl"
      :value="draft"
      class="form-control"
      rows="3"
      :maxlength="MAX_BODY"
      placeholder="Write a comment… Type @ to mention a member, or paste #123 to link a task."
      @input="onDraftInput"
      @keydown="onDraftKeydown"
      @click="syncMentionFromEl"
      @keyup="syncMentionFromEl"
    />
    <Teleport to="body">
      <ul
        v-if="showMentionMenu"
        id="task-comment-mention-listbox"
        ref="mentionListEl"
        class="user-search-menu list-unstyled mb-0"
        role="listbox"
        :style="mentionMenuStyle"
      >
        <li
          v-for="(name, idx) in mentionHits"
          :key="name"
          role="option"
          class="user-search-option"
          :class="{ active: mentionHighlight === idx }"
          :data-active="mentionHighlight === idx ? 'true' : undefined"
          :aria-selected="mentionToken?.query.toLowerCase() === name.toLowerCase()"
          @mousedown.prevent
          @click="selectMention(name)"
        >
          {{ name }}
        </li>
        <li
          v-if="mentionLoading && !mentionHits.length"
          class="user-search-option text-muted"
          aria-disabled="true"
        >
          Searching…
        </li>
        <li
          v-else-if="!mentionHits.length"
          class="user-search-option text-muted"
          aria-disabled="true"
        >
          No matching project members
        </li>
      </ul>
    </Teleport>
    <div v-if="draftLinks.length" class="task-discussion-suggestions mt-2">
      <div v-for="link in draftLinks" :key="link.id" class="task-discussion-suggestion">
        <span class="small">{{ taskLinkLabel(link.id, link.title) }}</span>
        <button
          v-if="!link.inserted"
          type="button"
          class="btn btn-sm btn-outline-primary"
          @click="insertLink(link.id)"
        >
          Insert link
        </button>
        <span v-else class="badge text-bg-secondary">Linked</span>
      </div>
    </div>
    <div class="d-flex justify-content-between align-items-center mt-1">
      <small class="text-muted">{{ charCount }}/{{ MAX_BODY }}</small>
      <button
        type="button"
        class="btn btn-sm btn-primary"
        :disabled="posting || !draft.trim()"
        @click="post"
      >
        {{ posting ? 'Posting…' : 'Post' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.task-discussion--fill {
  margin-top: 1.25rem;
}
.task-discussion-thread {
  max-height: 360px;
  overflow-y: auto;
  padding: 0.5rem;
  border-radius: 0.5rem;
  background: var(--ordryn-muted-bg, #f8f6ee);
  border: 1px solid var(--ordryn-card-border, #dee2e6);
}
.task-discussion--fill .task-discussion-thread {
  max-height: none;
  overflow: visible;
  min-height: 8rem;
}
.task-discussion-post + .task-discussion-post {
  margin-top: 0.75rem;
}
.task-post-card {
  background: var(--ordryn-card-bg, #fff);
  color: var(--ordryn-text);
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  border-radius: 0.5rem;
  box-shadow: var(--ordryn-card-shadow, 0 2px 8px rgba(0, 0, 0, 0.06));
  overflow: hidden;
}
.task-post-card--deleted {
  opacity: 0.85;
}
.task-post-header {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.65rem 0.75rem;
  background: color-mix(in srgb, var(--ordryn-muted-bg, #f1f5f9) 80%, var(--ordryn-card-bg, #fff));
  border-bottom: 1px solid var(--ordryn-card-border, #dee2e6);
}
.task-post-avatar {
  width: 2rem;
  height: 2rem;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--ordryn-filter-active-text, #fff);
  background: var(--ordryn-accent, #2563eb);
  flex-shrink: 0;
}
.task-post-meta {
  min-width: 0;
}
.task-post-author {
  font-weight: 600;
  font-size: 0.875rem;
  line-height: 1.2;
}
.task-post-time {
  display: block;
  font-size: 0.75rem;
  color: var(--ordryn-muted, #64748b);
}
.task-post-body {
  padding: 0.75rem;
  font-size: 0.9rem;
  line-height: 1.45;
  word-break: break-word;
}
.task-post-task-link {
  display: inline;
  padding: 0;
  margin: 0;
  border: 0;
  background: none;
  color: var(--ordryn-accent, #2563eb);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 2px;
  cursor: pointer;
}
.task-post-task-link:hover {
  filter: brightness(1.1);
}
.task-post-mention {
  font-weight: 600;
  color: var(--ordryn-accent, #2563eb);
}
.task-discussion-suggestions {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.task-discussion-suggestion {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.4rem 0.55rem;
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  border-radius: 0.375rem;
  background: var(--ordryn-muted-bg, #f8f6ee);
}
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
