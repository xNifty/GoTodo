<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { TaskComment } from '@/api/types'
import { useConfirm } from '@/composables/useConfirm'
import { useTaskSidebar } from '@/composables/useTaskSidebar'
import { useToast } from '@/composables/useToast'
import {
  extractTaskRefIDs,
  insertTaskRef,
  isInsertedTaskRef,
  splitCommentBody,
} from '@/utils/taskCommentBody'

const props = defineProps<{
  taskId: number
  currentUserId: number | null
  isOwner: boolean
}>()

const toast = useToast()
const { askConfirm } = useConfirm()
const { openEdit } = useTaskSidebar()

const comments = ref<TaskComment[]>([])
const loading = ref(false)
const posting = ref(false)
const draft = ref('')
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

onMounted(() => {
  void reload()
})

onUnmounted(() => {
  if (lookupTimer) clearTimeout(lookupTimer)
})

defineExpose({ reload })
</script>

<template>
  <div class="task-discussion mt-3">
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
      v-model="draft"
      class="form-control"
      rows="3"
      :maxlength="MAX_BODY"
      placeholder="Write a comment… Paste #123 to link a task."
    />
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
.task-discussion-thread {
  max-height: 360px;
  overflow-y: auto;
  padding: 0.5rem;
  border-radius: 0.5rem;
  background: var(--ordryn-muted-bg, #f8f6ee);
  border: 1px solid var(--ordryn-card-border, #dee2e6);
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
</style>
