<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { Project, ShareLink, Tag } from '@/api/types'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'

const props = defineProps<{
  project: Project
}>()

const emit = defineEmits<{
  changed: []
}>()

const toast = useToast()
const { askConfirm } = useConfirm()
const tags = ref<Tag[]>([])
const tagLinks = ref<Record<number, ShareLink[]>>({})
const tagName = ref('')
const renameTagId = ref<number | null>(null)
const renameTagValue = ref('')
const loading = ref(false)

const isOwner = computed(() => (props.project.role || 'owner') === 'owner')
const canManage = computed(() => {
  const role = props.project.role || 'owner'
  return role === 'owner' || role === 'editor'
})

async function loadTagLinks(tagList: Tag[]) {
  const entries = await Promise.all(
    tagList.map(async (tag) => {
      try {
        const links = await api.listShareLinks('tag', tag.id)
        return [tag.id, links] as const
      } catch {
        return [tag.id, [] as ShareLink[]] as const
      }
    }),
  )
  const next: Record<number, ShareLink[]> = {}
  for (const [id, links] of entries) {
    next[id] = links
  }
  tagLinks.value = next
}

async function load() {
  loading.value = true
  try {
    const list = await api.listTags({ project_id: props.project.id })
    tags.value = list
    await loadTagLinks(list)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load tags', 'error')
  } finally {
    loading.value = false
  }
}

watch(
  () => props.project.id,
  () => {
    void load()
  },
  { immediate: true },
)

async function createTag() {
  if (!tagName.value.trim() || !canManage.value) return
  try {
    await api.createTag(tagName.value.trim(), props.project.id)
    tagName.value = ''
    toast.push('Tag created', 'success')
    await load()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Create failed', 'error')
  }
}

function beginRenameTag(tag: Tag) {
  renameTagId.value = tag.id
  renameTagValue.value = tag.name
}

async function saveRenameTag() {
  if (renameTagId.value == null || !renameTagValue.value.trim()) return
  try {
    await api.renameTag(renameTagId.value, renameTagValue.value.trim())
    renameTagId.value = null
    toast.push('Tag renamed', 'success')
    await load()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Rename failed', 'error')
  }
}

async function removeTag(tag: Tag) {
  const ok = await askConfirm({
    title: 'Delete tag?',
    message: `Delete tag “${tag.name}”? It will be removed from all tasks in this project.`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  try {
    await api.deleteTag(tag.id)
    toast.push('Tag deleted', 'info')
    await load()
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Delete failed', 'error')
  }
}

async function shareTag(tag: Tag) {
  try {
    const link = await api.createShareLink('tag', tag.id)
    await navigator.clipboard.writeText(link.url)
    toast.push('Share link created and copied', 'success')
    await loadTagLinks(tags.value)
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to create share link', 'error')
  }
}

async function copyTagLink(url: string) {
  try {
    await navigator.clipboard.writeText(url)
    toast.push('Copied', 'success')
  } catch {
    toast.push(url, 'info')
  }
}

async function revokeTagLink(tag: Tag, linkId: number) {
  const ok = await askConfirm({
    title: 'Make private?',
    message: `Revoke the share link for “${tag.name}”? Anyone with that URL will lose access.`,
    confirmLabel: 'Make private',
    danger: true,
  })
  if (!ok) return
  try {
    await api.revokeShareLink(linkId)
    toast.push('Link revoked — tag is private again', 'info')
    await loadTagLinks(tags.value)
    emit('changed')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to revoke link', 'error')
  }
}
</script>

<template>
  <div>
    <h4 class="h6 mb-2">Tags</h4>
    <p class="small text-muted mb-3">
      Tags on this project are shared with members. Owners and editors can add or delete them.
    </p>
    <form v-if="canManage" class="row g-2 mb-3" @submit.prevent="createTag">
      <div class="col-sm-8">
        <input v-model="tagName" type="text" class="form-control form-control-sm" placeholder="New tag" maxlength="50" required />
      </div>
      <div class="col-sm-4">
        <button type="submit" class="btn btn-sm btn-primary w-100">Add tag</button>
      </div>
    </form>
    <p v-if="loading" class="small text-muted mb-0">Loading tags…</p>
    <ul v-else class="list-group">
      <li v-for="tag in tags" :key="tag.id" class="list-group-item">
        <div class="d-flex flex-wrap gap-2 align-items-center">
          <template v-if="renameTagId === tag.id">
            <input v-model="renameTagValue" type="text" class="form-control form-control-sm" maxlength="50" />
            <button type="button" class="btn btn-sm btn-primary" @click="saveRenameTag">Save</button>
            <button type="button" class="btn btn-sm btn-secondary" @click="renameTagId = null">Cancel</button>
          </template>
          <template v-else>
            <span class="tag-chip" :style="{ backgroundColor: tag.color || '#6c757d' }">{{ tag.name }}</span>
            <span class="flex-grow-1" />
            <button v-if="canManage" type="button" class="btn btn-sm btn-outline-secondary" @click="beginRenameTag(tag)">
              Rename
            </button>
            <button v-if="canManage" type="button" class="btn btn-sm btn-outline-danger" @click="removeTag(tag)">
              Delete
            </button>
          </template>
        </div>
        <div class="mt-2 small">
          <template v-if="tagLinks[tag.id]?.length">
            <div v-for="link in tagLinks[tag.id]" :key="link.id" class="d-flex flex-wrap align-items-center gap-1 mb-1">
              <span class="badge text-bg-success">Shared</span>
              <button class="btn btn-sm btn-outline-secondary" type="button" @click="copyTagLink(link.url)">Copy</button>
              <button
                v-if="isOwner"
                class="btn btn-sm btn-outline-danger"
                type="button"
                @click="revokeTagLink(tag, link.id)"
              >
                Make private
              </button>
            </div>
          </template>
          <template v-else-if="isOwner">
            <button class="btn btn-sm btn-outline-primary" type="button" @click="shareTag(tag)">Create share link</button>
          </template>
        </div>
      </li>
      <li v-if="!tags.length" class="list-group-item text-muted">No tags on this project yet.</li>
    </ul>
  </div>
</template>
