<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api } from '@/api/client'
import type { AdminSettings, AdminSettingsPatch, AdminUser } from '@/api/types'
import { APIError } from '@/api/types'
import { useSite } from '@/composables/useSite'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'

const toast = useToast()
const { askConfirm } = useConfirm()
const { refresh: refreshSite } = useSite()
const users = ref<AdminUser[]>([])
const busy = ref(false)
const emailBusy = ref(false)
const editingUsernameId = ref<number | null>(null)
const editUsernameValue = ref('')
const mailgunApiKeyInput = ref('')
const smtpPasswordInput = ref('')
const settings = reactive<AdminSettings>({
  site_name: '',
  default_timezone: 'UTC',
  show_changelog: true,
  site_version: '',
  enable_registration: true,
  invite_only: false,
  meta_description: '',
  enable_global_announcement: false,
  global_announcement_text: '',
  enable_api: false,
  email_provider: '',
  email_from_address: '',
  email_from_name: '',
  email_mailgun_domain: '',
  email_mailgun_api_key_set: false,
  email_smtp_host: '',
  email_smtp_port: 587,
  email_smtp_username: '',
  email_smtp_password_set: false,
  email_smtp_tls: true,
})

async function load() {
  try {
    const [s, u] = await Promise.all([api.getAdminSettings(), api.listAdminUsers()])
    Object.assign(settings, s)
    mailgunApiKeyInput.value = ''
    smtpPasswordInput.value = ''
    users.value = u
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load admin data', 'error')
  }
}

async function saveSettings() {
  busy.value = true
  try {
    const saved = await api.patchAdminSettings({
      site_name: settings.site_name,
      default_timezone: settings.default_timezone,
      show_changelog: settings.show_changelog,
      enable_registration: settings.enable_registration,
      invite_only: settings.invite_only,
      meta_description: settings.meta_description,
      enable_global_announcement: settings.enable_global_announcement,
      global_announcement_text: settings.global_announcement_text,
      enable_api: settings.enable_api,
    })
    Object.assign(settings, saved)
    await refreshSite()
    toast.push('Settings saved', 'success')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Save failed', 'error')
  } finally {
    busy.value = false
  }
}

async function saveEmailSettings() {
  emailBusy.value = true
  try {
    const payload: AdminSettingsPatch = {
      email_provider: settings.email_provider || '',
      email_from_address: settings.email_from_address,
      email_from_name: settings.email_from_name,
      email_mailgun_domain: settings.email_mailgun_domain,
      email_smtp_host: settings.email_smtp_host,
      email_smtp_port: settings.email_smtp_port,
      email_smtp_username: settings.email_smtp_username,
      email_smtp_tls: settings.email_smtp_tls,
    }
    if (mailgunApiKeyInput.value !== '') {
      payload.email_mailgun_api_key = mailgunApiKeyInput.value
    }
    if (smtpPasswordInput.value !== '') {
      payload.email_smtp_password = smtpPasswordInput.value
    }
    const saved = await api.patchAdminSettings(payload)
    Object.assign(settings, saved)
    mailgunApiKeyInput.value = ''
    smtpPasswordInput.value = ''
    toast.push('Email settings saved', 'success')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Save failed', 'error')
  } finally {
    emailBusy.value = false
  }
}

async function toggleBan(user: AdminUser) {
  const ok = await askConfirm({
    title: user.is_banned ? 'Unban user?' : 'Ban user?',
    message: user.is_banned
      ? `Unban ${user.email}?`
      : `Ban ${user.email}? They will no longer be able to sign in.`,
    confirmLabel: user.is_banned ? 'Unban' : 'Ban',
    danger: !user.is_banned,
  })
  if (!ok) return
  try {
    if (user.is_banned) {
      await api.unbanUser(user.id)
      toast.push('User unbanned', 'success')
    } else {
      await api.banUser(user.id)
      toast.push('User banned', 'info')
    }
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Update failed', 'error')
  }
}

function startEditUsername(user: AdminUser) {
  editingUsernameId.value = user.id
  editUsernameValue.value = user.user_name || ''
}

function cancelEditUsername() {
  editingUsernameId.value = null
  editUsernameValue.value = ''
}

async function saveUsername(user: AdminUser) {
  const next = editUsernameValue.value.trim()
  if (!next) {
    toast.push('Username is required', 'error')
    return
  }
  try {
    await api.setAdminUsername(user.id, next)
    toast.push('Username updated', 'success')
    cancelEditUsername()
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to update username', 'error')
  }
}

onMounted(load)
</script>

<template>
  <div class="container mt-3">
    <h1>Admin</h1>
    <div class="card mb-4">
      <div class="card-header"><h2 class="h5 mb-0">Site settings</h2></div>
      <div class="card-body">
        <form @submit.prevent="saveSettings">
          <div class="mb-3">
            <label class="form-label">Site name</label>
            <input v-model="settings.site_name" type="text" class="form-control" required />
          </div>
          <div class="mb-3">
            <label class="form-label">Default timezone</label>
            <input v-model="settings.default_timezone" type="text" class="form-control" required />
          </div>
          <div class="mb-3">
            <label class="form-label">Meta description</label>
            <textarea v-model="settings.meta_description" class="form-control" rows="2" />
          </div>
          <div class="form-check mb-2">
            <input id="admin-registration" v-model="settings.enable_registration" class="form-check-input" type="checkbox" />
            <label class="form-check-label" for="admin-registration">Enable registration</label>
          </div>
          <div class="form-check mb-2">
            <input id="admin-invite-only" v-model="settings.invite_only" class="form-check-input" type="checkbox" />
            <label class="form-check-label" for="admin-invite-only">Invite only</label>
          </div>
          <div class="form-check mb-2">
            <input id="admin-changelog" v-model="settings.show_changelog" class="form-check-input" type="checkbox" />
            <label class="form-check-label" for="admin-changelog">Show changelog</label>
          </div>
          <div class="form-check mb-2">
            <input id="admin-api" v-model="settings.enable_api" class="form-check-input" type="checkbox" />
            <label class="form-check-label" for="admin-api">Enable external REST API (API keys &amp; Android)</label>
          </div>
          <p class="text-muted small">The web app always uses the JSON API with your session cookie. This toggle controls Bearer access for scripts and mobile clients.</p>
          <div class="form-check mb-2">
            <input id="admin-announcement" v-model="settings.enable_global_announcement" class="form-check-input" type="checkbox" />
            <label class="form-check-label" for="admin-announcement">Global announcement</label>
          </div>
          <div class="mb-3">
            <label class="form-label">Announcement text</label>
            <textarea v-model="settings.global_announcement_text" class="form-control" rows="2" maxlength="500" />
          </div>
          <p class="text-muted">Site Version: {{ settings.site_version || '—' }}</p>
          <button type="submit" class="btn btn-primary" :disabled="busy">
            {{ busy ? 'Saving…' : 'Save settings' }}
          </button>
        </form>
      </div>
    </div>

    <div class="card mb-4">
      <div class="card-header"><h2 class="h5 mb-0">Email</h2></div>
      <div class="card-body">
        <form @submit.prevent="saveEmailSettings">
          <div class="mb-3">
            <label class="form-label" for="email-provider">Provider</label>
            <select id="email-provider" v-model="settings.email_provider" class="form-select">
              <option value="">Disabled</option>
              <option value="mailgun">Mailgun</option>
              <option value="smtp">SMTP</option>
            </select>
          </div>

          <template v-if="settings.email_provider === 'mailgun' || settings.email_provider === 'smtp'">
            <div class="mb-3">
              <label class="form-label" for="email-from-name">From name</label>
              <input id="email-from-name" v-model="settings.email_from_name" type="text" class="form-control" placeholder="GoTodo" />
            </div>
            <div class="mb-3">
              <label class="form-label" for="email-from-address">From address</label>
              <input
                id="email-from-address"
                v-model="settings.email_from_address"
                type="email"
                class="form-control"
                placeholder="noreply@ordryn.com"
                required
              />
              <div class="form-text">
                Shown to recipients. May differ from the Mailgun or SMTP sending domain.
              </div>
            </div>
          </template>

          <template v-if="settings.email_provider === 'mailgun'">
            <div class="mb-3">
              <label class="form-label" for="email-mailgun-domain">Mailgun domain</label>
              <input
                id="email-mailgun-domain"
                v-model="settings.email_mailgun_domain"
                type="text"
                class="form-control"
                placeholder="mydomain.com"
                required
              />
              <div class="form-text">Sending domain used by Mailgun (signing/routing), not necessarily the From address.</div>
            </div>
            <div class="mb-3">
              <label class="form-label" for="email-mailgun-key">Mailgun API key</label>
              <input
                id="email-mailgun-key"
                v-model="mailgunApiKeyInput"
                type="password"
                class="form-control"
                autocomplete="new-password"
                :placeholder="settings.email_mailgun_api_key_set ? '•••• configured (leave blank to keep)' : 'Enter API key'"
              />
            </div>
          </template>

          <template v-if="settings.email_provider === 'smtp'">
            <div class="mb-3">
              <label class="form-label" for="email-smtp-host">SMTP host</label>
              <input id="email-smtp-host" v-model="settings.email_smtp_host" type="text" class="form-control" placeholder="smtp.example.com" required />
            </div>
            <div class="mb-3">
              <label class="form-label" for="email-smtp-port">SMTP port</label>
              <input id="email-smtp-port" v-model.number="settings.email_smtp_port" type="number" class="form-control" min="1" max="65535" required />
              <div class="form-text">Use 587 with STARTTLS, or 465 for implicit TLS.</div>
            </div>
            <div class="mb-3">
              <label class="form-label" for="email-smtp-username">SMTP username</label>
              <input id="email-smtp-username" v-model="settings.email_smtp_username" type="text" class="form-control" required />
            </div>
            <div class="mb-3">
              <label class="form-label" for="email-smtp-password">SMTP password</label>
              <input
                id="email-smtp-password"
                v-model="smtpPasswordInput"
                type="password"
                class="form-control"
                autocomplete="new-password"
                :placeholder="settings.email_smtp_password_set ? '•••• configured (leave blank to keep)' : 'Enter password'"
              />
            </div>
            <div class="form-check mb-3">
              <input id="email-smtp-tls" v-model="settings.email_smtp_tls" class="form-check-input" type="checkbox" />
              <label class="form-check-label" for="email-smtp-tls">Use STARTTLS (ignored for port 465, which always uses TLS)</label>
            </div>
          </template>

          <p v-if="!settings.email_provider" class="text-muted small">
            Outbound email is disabled. Password resets and digests will not send until a provider is configured.
          </p>

          <button type="submit" class="btn btn-primary" :disabled="emailBusy">
            {{ emailBusy ? 'Saving…' : 'Save email settings' }}
          </button>
        </form>
      </div>
    </div>

    <div class="card">
      <div class="card-header"><h2 class="h5 mb-0">Users</h2></div>
      <ul class="list-group list-group-flush">
        <li v-for="user in users" :key="user.id" class="list-group-item">
          <div class="d-flex justify-content-between align-items-start gap-3 flex-wrap">
            <div class="flex-grow-1">
              <template v-if="editingUsernameId === user.id">
                <div class="input-group input-group-sm" style="max-width: 280px">
                  <input
                    v-model="editUsernameValue"
                    type="text"
                    class="form-control"
                    minlength="3"
                    maxlength="32"
                    pattern="[A-Za-z0-9_]+"
                    @keyup.enter="saveUsername(user)"
                  />
                  <button type="button" class="btn btn-primary" @click="saveUsername(user)">Save</button>
                  <button type="button" class="btn btn-outline-secondary" @click="cancelEditUsername">Cancel</button>
                </div>
              </template>
              <template v-else>
                <strong>{{ user.user_name || user.email }}</strong>
                <button type="button" class="btn btn-link btn-sm py-0" @click="startEditUsername(user)">
                  Edit username
                </button>
              </template>
              <div class="text-muted small">
                {{ user.email }}
                <span v-if="user.is_banned">· banned</span>
              </div>
            </div>
            <button type="button" class="btn btn-sm" :class="user.is_banned ? 'btn-outline-secondary' : 'btn-outline-danger'" @click="toggleBan(user)">
              {{ user.is_banned ? 'Unban' : 'Ban' }}
            </button>
          </div>
        </li>
        <li v-if="!users.length" class="list-group-item text-muted">No users found.</li>
      </ul>
    </div>
  </div>
</template>
