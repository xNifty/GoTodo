<template>
  <div class="container mt-4">
    <div class="row">
      <div class="col-md-8">
        <div v-if="pendingInvites.length" class="card mb-4 border-primary">
          <div class="card-header">
            <h3 class="mb-0 h5">Pending project invites</h3>
          </div>
          <div class="card-body">
            <div
              v-for="inv in pendingInvites"
              :key="inv.id"
              class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2"
            >
              <div>
                <strong>{{ inv.project_name || 'Project' }}</strong>
                <span class="text-muted"> as {{ inv.role }}</span>
                <div class="small text-muted" v-if="inv.inviter_user_name || inv.inviter_email">
                  From {{ inv.inviter_user_name || inv.inviter_email }}
                </div>
              </div>
              <div class="d-flex gap-2">
                <button class="btn btn-sm btn-primary" type="button" @click="acceptInvite(inv)">Accept</button>
                <button class="btn btn-sm btn-outline-secondary" type="button" @click="declineInvite(inv)">
                  Decline
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <h3 class="mb-0">Your Projects</h3>
          </div>
          <div class="card-body">
            <p class="text-muted">
              Create projects to organize your tasks. Share a project with household members or create a
              read-only link. Deleting a project will unassign tasks from it.
            </p>
            <div class="table-responsive">
              <table class="table table-striped projects-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th style="width: 100px">Role</th>
                    <th style="width: 160px">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <template v-for="p in ownedProjects" :key="p.id">
                    <tr>
                      <td data-label="Name">
                        <template v-if="renameProjectId === p.id">
                          <form
                            class="d-inline-flex align-items-center gap-1 flex-wrap"
                            @submit.prevent="saveRenameProject"
                          >
                            <input
                              v-model="renameProjectValue"
                              type="text"
                              class="form-control form-control-sm"
                              maxlength="50"
                              required
                              aria-label="Project name"
                            />
                            <button class="btn btn-sm btn-primary" type="submit" aria-label="Save project name">
                              <i class="bi bi-check" />
                            </button>
                            <button
                              class="btn btn-sm btn-secondary"
                              type="button"
                              aria-label="Cancel rename"
                              @click="renameProjectId = null"
                            >
                              <i class="bi bi-x" />
                            </button>
                          </form>
                        </template>
                        <template v-else>
                          <span class="project-name-display">{{ p.name }}</span>
                          <button
                            class="btn btn-sm btn-link edit-project-btn p-0 ms-1"
                            type="button"
                            aria-label="Rename project"
                            @click="beginRenameProject(p)"
                          >
                            <i class="bi bi-pencil" />
                          </button>
                        </template>
                      </td>
                      <td data-label="Role"><span class="badge text-bg-secondary">owner</span></td>
                      <td data-label="Actions">
                        <button
                          class="btn btn-sm btn-outline-primary me-1"
                          type="button"
                          @click="toggleSharePanel(p.id)"
                        >
                          Share
                        </button>
                        <button
                          class="btn btn-sm btn-danger"
                          type="button"
                          aria-label="Delete project"
                          @click="removeProject(p)"
                        >
                          <i class="bi bi-trash" />
                        </button>
                      </td>
                    </tr>
                    <tr v-if="sharePanelId === p.id">
                      <td colspan="3" class="bg-body-tertiary">
                        <ProjectSharePanel :project="p" @changed="load" />
                      </td>
                    </tr>
                  </template>
                  <tr v-if="!ownedProjects.length">
                    <td colspan="3" class="text-muted">No owned projects yet.</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div class="card mt-4">
          <div class="card-header">
            <h3 class="mb-0 h5">Shared with me</h3>
          </div>
          <div class="card-body">
            <div class="table-responsive" v-if="sharedProjects.length">
              <table class="table table-striped">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Owner</th>
                    <th>Role</th>
                    <th style="width: 100px">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in sharedProjects" :key="p.id">
                    <td>{{ p.name }}</td>
                    <td class="text-muted small">{{ p.owner_user_name || p.owner_email }}</td>
                    <td><span class="badge text-bg-info">{{ p.role }}</span></td>
                    <td>
                      <button
                        v-if="p.role !== 'owner'"
                        class="btn btn-sm btn-outline-secondary"
                        type="button"
                        @click="leaveProject(p)"
                      >
                        Leave
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else class="text-muted mb-0">No shared projects yet.</p>
          </div>
        </div>

        <div class="card mt-4">
          <div class="card-header">
            <h3 class="mb-0 h5">Your Tags</h3>
          </div>
          <div class="card-body">
            <p class="text-muted small mb-3">
              Tags stay private by default. Optionally create a read-only share link for tasks with that tag,
              then revoke it anytime to make them private again.
            </p>
            <div class="table-responsive">
              <table class="table table-striped tags-table">
                <thead>
                  <tr>
                    <th>Tag</th>
                    <th style="width: 280px">Sharing</th>
                    <th style="width: 80px">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="tag in tags" :key="tag.id">
                    <td data-label="Tag">
                      <template v-if="renameTagId === tag.id">
                        <form class="d-inline-flex align-items-center gap-1 flex-wrap" @submit.prevent="saveRenameTag">
                          <input
                            v-model="renameTagValue"
                            type="text"
                            class="form-control form-control-sm"
                            maxlength="50"
                            required
                          />
                          <button class="btn btn-sm btn-primary" type="submit"><i class="bi bi-check" /></button>
                          <button class="btn btn-sm btn-secondary" type="button" @click="renameTagId = null">
                            <i class="bi bi-x" />
                          </button>
                        </form>
                      </template>
                      <template v-else>
                        {{ tag.name }}
                        <button class="btn btn-sm btn-link p-0 ms-1" type="button" @click="beginRenameTag(tag)">
                          <i class="bi bi-pencil" />
                        </button>
                      </template>
                    </td>
                    <td data-label="Sharing">
                      <template v-if="tagLinks[tag.id]?.length">
                        <div
                          v-for="link in tagLinks[tag.id]"
                          :key="link.id"
                          class="d-flex flex-wrap align-items-center gap-1 mb-1"
                        >
                          <span class="badge text-bg-success">Shared</span>
                          <button class="btn btn-sm btn-outline-secondary" type="button" @click="copyTagLink(link.url)">
                            Copy
                          </button>
                          <button
                            class="btn btn-sm btn-outline-danger"
                            type="button"
                            @click="revokeTagLink(tag, link.id)"
                          >
                            Make private
                          </button>
                        </div>
                      </template>
                      <template v-else>
                        <span class="text-muted small me-2">Private</span>
                        <button class="btn btn-sm btn-outline-primary" type="button" @click="shareTag(tag)">
                          Create share link
                        </button>
                      </template>
                    </td>
                    <td data-label="Actions">
                      <button class="btn btn-sm btn-danger" type="button" @click="removeTag(tag)">
                        <i class="bi bi-trash" />
                      </button>
                    </td>
                  </tr>
                  <tr v-if="!tags.length">
                    <td colspan="3" class="text-muted">No tags yet.</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card">
          <div class="card-header">
            <h3 class="mb-0 h5">Create Project</h3>
          </div>
          <div class="card-body">
            <form @submit.prevent="createProject">
              <div class="mb-3">
                <label class="form-label" for="project-name">Name</label>
                <input
                  id="project-name"
                  v-model="name"
                  type="text"
                  class="form-control"
                  maxlength="50"
                  required
                />
                <div class="d-flex justify-content-between">
                  <small class="form-hint">Max 50 Characters</small>
                  <small class="text-muted">{{ name.length }}/50</small>
                </div>
              </div>
              <div class="d-flex gap-2">
                <button class="btn btn-primary" type="submit">Create</button>
                <RouterLink class="btn btn-secondary" to="/">Cancel</RouterLink>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '@/api/client'
import type { Project, ProjectInvite, ShareLink, Tag } from '@/api/types'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'
import { useAuth } from '@/composables/useAuth'
import { useConfirm } from '@/composables/useConfirm'
import ProjectSharePanel from '@/components/ProjectSharePanel.vue'

const projects = ref<Project[]>([])
const pendingInvites = ref<ProjectInvite[]>([])
const tags = ref<Tag[]>([])
const tagLinks = ref<Record<number, ShareLink[]>>({})
const name = ref('')
const renameProjectId = ref<number | null>(null)
const renameProjectValue = ref('')
const renameTagId = ref<number | null>(null)
const renameTagValue = ref('')
const sharePanelId = ref<number | null>(null)
const toast = useToast()
const auth = useAuth()
const { askConfirm } = useConfirm()

const ownedProjects = computed(() => projects.value.filter((p) => (p.role || 'owner') === 'owner'))
const sharedProjects = computed(() => projects.value.filter((p) => p.role && p.role !== 'owner'))

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
  try {
    const [p, t, invites] = await Promise.all([
      api.listProjects(),
      api.listTags(),
      api.listMyProjectInvites(),
    ])
    projects.value = p
    tags.value = t
    pendingInvites.value = invites
    await loadTagLinks(t)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load projects', 'error')
  }
}

async function createProject() {
  if (!name.value.trim()) return
  try {
    await api.createProject(name.value.trim())
    name.value = ''
    toast.push('Project created', 'success')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Create failed', 'error')
  }
}

function beginRenameProject(p: Project) {
  renameProjectId.value = p.id
  renameProjectValue.value = p.name
  renameTagId.value = null
}

async function saveRenameProject() {
  if (renameProjectId.value == null || !renameProjectValue.value.trim()) return
  try {
    await api.renameProject(renameProjectId.value, renameProjectValue.value.trim())
    renameProjectId.value = null
    toast.push('Renamed', 'success')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Rename failed', 'error')
  }
}

async function removeProject(p: Project) {
  const ok = await askConfirm({
    title: 'Delete project?',
    message: 'Delete this project? Tasks will be unassigned but not deleted.',
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  try {
    await api.deleteProject(p.id)
    if (sharePanelId.value === p.id) sharePanelId.value = null
    toast.push('Project deleted', 'info')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Delete failed', 'error')
  }
}

function toggleSharePanel(id: number) {
  sharePanelId.value = sharePanelId.value === id ? null : id
}

async function leaveProject(p: Project) {
  const me = auth.user.value?.id
  if (!me) return
  const ok = await askConfirm({
    title: 'Leave project?',
    message: `Leave project “${p.name}”?`,
    confirmLabel: 'Leave',
    danger: true,
  })
  if (!ok) return
  try {
    await api.removeProjectMember(p.id, me)
    toast.push('Left project', 'info')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Leave failed', 'error')
  }
}

async function acceptInvite(inv: ProjectInvite) {
  try {
    await api.acceptProjectInvite(inv.id)
    toast.push('Joined project', 'success')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Accept failed', 'error')
  }
}

async function declineInvite(inv: ProjectInvite) {
  try {
    await api.declineProjectInvite(inv.id)
    toast.push('Invite declined', 'info')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Decline failed', 'error')
  }
}

function beginRenameTag(tag: Tag) {
  renameTagId.value = tag.id
  renameTagValue.value = tag.name
  renameProjectId.value = null
}

async function saveRenameTag() {
  if (renameTagId.value == null || !renameTagValue.value.trim()) return
  try {
    await api.renameTag(renameTagId.value, renameTagValue.value.trim())
    renameTagId.value = null
    toast.push('Tag renamed', 'success')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Rename failed', 'error')
  }
}

async function removeTag(tag: Tag) {
  const ok = await askConfirm({
    title: 'Delete tag?',
    message: 'Delete this tag? It will be removed from all tasks.',
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  try {
    await api.deleteTag(tag.id)
    toast.push('Tag deleted', 'info')
    await load()
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
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to revoke link', 'error')
  }
}

onMounted(load)
</script>
