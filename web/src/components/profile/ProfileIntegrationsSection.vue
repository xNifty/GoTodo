<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/client'
import type { CalendarInfo, GitHubConnection } from '@/api/types'
import { APIError } from '@/api/types'
import { useConfirm } from '@/composables/useConfirm'
import { useSite } from '@/composables/useSite'
import { useToast } from '@/composables/useToast'

const { push } = useToast()
const { askConfirm } = useConfirm()
const { siteInfo, refresh: refreshSite } = useSite()

const github = ref<GitHubConnection | null>(null)
const githubPAT = ref('')
const githubBusy = ref(false)
const githubOAuthEnabled = computed(() => !!siteInfo.value?.github_oauth_configured)

const calendar = ref<CalendarInfo | null>(null)
const icsFile = ref<File | null>(null)

async function load() {
  try {
    const [c, g] = await Promise.all([api.getCalendar(), api.getGitHubConnection()])
    calendar.value = c
    github.value = g
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Failed to load integrations', 'error')
  }
}

async function connectGitHubPAT() {
  if (!githubPAT.value.trim()) return
  githubBusy.value = true
  try {
    github.value = await api.connectGitHubPAT(githubPAT.value.trim())
    githubPAT.value = ''
    push('GitHub connected', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'GitHub connect failed', 'error')
  } finally {
    githubBusy.value = false
  }
}

async function connectGitHubOAuth() {
  githubBusy.value = true
  try {
    const { authorize_url } = await api.startGitHubOAuth()
    window.location.href = authorize_url
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Could not start GitHub OAuth', 'error')
    githubBusy.value = false
  }
}

async function disconnectGitHub() {
  const ok = await askConfirm({
    title: 'Disconnect GitHub?',
    message: 'You will need to reconnect before linking repositories or creating issues.',
    confirmLabel: 'Disconnect',
    danger: true,
  })
  if (!ok) return
  githubBusy.value = true
  try {
    await api.disconnectGitHub()
    github.value = { connected: false }
    push('GitHub disconnected', 'info')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Disconnect failed', 'error')
  } finally {
    githubBusy.value = false
  }
}

async function regenerateCalendar() {
  const ok = await askConfirm({
    title: 'Regenerate calendar link?',
    message: 'This invalidates the current calendar feed URL. Anyone using the old link will lose access.',
    confirmLabel: 'Regenerate',
    danger: true,
  })
  if (!ok) return
  try {
    calendar.value = await api.regenerateCalendar()
    push('Calendar token regenerated', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Regenerate failed', 'error')
  }
}

function onIcsFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  icsFile.value = input.files?.[0] ?? null
}

async function syncCalendar() {
  if (!icsFile.value) return
  try {
    const result = await api.syncCalendar(icsFile.value)
    push(`Updated ${result.updated} task due dates`, 'success')
    icsFile.value = null
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Calendar sync failed', 'error')
  }
}

onMounted(() => {
  void refreshSite()
  void load()
})
</script>

<template>
  <div id="github-section" class="card mb-4">
    <div class="card-header">
      <h3 class="card-title mb-0">GitHub</h3>
    </div>
    <div class="card-body">
      <p class="text-muted small">
        Connect GitHub to link repositories on projects you own and create or attach issues from board tasks.
      </p>
      <div v-if="github?.connected" class="d-flex flex-wrap align-items-center gap-2 mb-3">
        <span class="badge text-bg-success">Connected as @{{ github.github_login }}</span>
        <span v-if="github.auth_method" class="text-muted small">via {{ github.auth_method }}</span>
        <button
          type="button"
          class="btn btn-sm btn-outline-danger"
          :disabled="githubBusy"
          @click="disconnectGitHub"
        >
          Disconnect
        </button>
      </div>
      <template v-else>
        <div v-if="githubOAuthEnabled" class="mb-3">
          <button
            type="button"
            class="btn btn-primary"
            :disabled="githubBusy"
            @click="connectGitHubOAuth"
          >
            {{ githubBusy ? 'Redirecting…' : 'Connect with GitHub' }}
          </button>
        </div>
        <form class="row g-2" @submit.prevent="connectGitHubPAT">
          <div class="col-12">
            <label class="form-label" for="github-pat">Personal access token</label>
            <input
              id="github-pat"
              v-model="githubPAT"
              type="password"
              class="form-control"
              autocomplete="off"
              placeholder="ghp_… or github_pat_…"
              :disabled="githubBusy"
            />
            <div class="form-text">
              Needs access to the repositories you will link (fine-grained: Contents metadata + Issues read/write; classic: <code>repo</code>).
            </div>
          </div>
          <div class="col-12">
            <button type="submit" class="btn btn-outline-primary" :disabled="githubBusy || !githubPAT.trim()">
              {{ githubBusy ? 'Connecting…' : 'Connect with token' }}
            </button>
          </div>
        </form>
      </template>
    </div>
  </div>

  <div id="calendar-feed" class="card mb-4">
    <div class="card-header">
      <h3 class="card-title mb-0">Calendar feed</h3>
    </div>
    <div class="card-body">
      <p v-if="calendar" class="text-break"><code>{{ calendar.feed_url }}</code></p>
      <button type="button" class="btn btn-outline-secondary mb-3" @click="regenerateCalendar">Regenerate token</button>
      <form @submit.prevent="syncCalendar">
        <div class="mb-3">
          <label class="form-label">Sync due dates from ICS export</label>
          <input type="file" class="form-control" accept=".ics,text/calendar" @change="onIcsFileChange" />
        </div>
        <button type="submit" class="btn btn-outline-primary" :disabled="!icsFile">Sync calendar</button>
      </form>
    </div>
  </div>
</template>
