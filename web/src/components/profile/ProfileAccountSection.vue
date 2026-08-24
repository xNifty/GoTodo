<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
import MfaModal from '@/components/MfaModal.vue'
import { useToast } from '@/composables/useToast'

defineProps<{
  email: string
  username: string
}>()

const { push } = useToast()
const showPasswordForm = ref(false)
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const busy = ref(false)
const mfaEnabled = ref(false)
const mfaRecoveryRemaining = ref(0)
const mfaModalOpen = ref(false)

function closePasswordForm() {
  showPasswordForm.value = false
  currentPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
}

async function changePassword() {
  busy.value = true
  try {
    await api.changePassword({
      current_password: currentPassword.value,
      new_password: newPassword.value,
      confirm_password: confirmPassword.value,
    })
    closePasswordForm()
    push('Password updated', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Password change failed', 'error')
  } finally {
    busy.value = false
  }
}

function onMfaUpdated(status: { enabled: boolean; recovery_codes_remaining: number }) {
  mfaEnabled.value = status.enabled
  mfaRecoveryRemaining.value = status.recovery_codes_remaining
}

onMounted(async () => {
  try {
    const status = await api.getMFA()
    onMfaUpdated(status)
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Failed to load MFA status', 'error')
  }
})
</script>

<template>
  <div class="card mb-4">
    <div class="card-header">
      <h3 class="card-title mb-0">Account</h3>
    </div>
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label fw-bold">Email</label>
        <input type="text" class="form-control-plaintext" :value="email" readonly tabindex="-1" />
      </div>
      <div class="mb-0">
        <label class="form-label fw-bold">Username</label>
        <input type="text" class="form-control-plaintext" :value="username" readonly tabindex="-1" />
        <div class="form-text">
          Usernames cannot be changed except by an administrator.
        </div>
      </div>
    </div>
  </div>

  <div class="card mb-4">
    <div class="card-header">
      <h3 class="card-title mb-0">Password</h3>
    </div>
    <div class="card-body">
      <template v-if="!showPasswordForm">
        <p class="text-muted small mb-3">Change the password used to sign in to this account.</p>
        <button type="button" class="btn btn-outline-primary" @click="showPasswordForm = true">
          Change password
        </button>
      </template>
      <form v-else @submit.prevent="changePassword">
        <div class="mb-3">
          <label class="form-label" for="profile-current-password">Current password</label>
          <input
            id="profile-current-password"
            v-model="currentPassword"
            type="password"
            class="form-control"
            required
            autocomplete="current-password"
          />
        </div>
        <div class="mb-3">
          <label class="form-label" for="profile-new-password">New password</label>
          <input
            id="profile-new-password"
            v-model="newPassword"
            type="password"
            class="form-control"
            required
            autocomplete="new-password"
          />
        </div>
        <div class="mb-3">
          <label class="form-label" for="profile-confirm-password">Confirm new password</label>
          <input
            id="profile-confirm-password"
            v-model="confirmPassword"
            type="password"
            class="form-control"
            required
            autocomplete="new-password"
          />
        </div>
        <div class="d-flex flex-wrap gap-2">
          <button type="submit" class="btn btn-primary" :disabled="busy">
            {{ busy ? 'Saving…' : 'Change password' }}
          </button>
          <button type="button" class="btn btn-secondary" :disabled="busy" @click="closePasswordForm">
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>

  <div id="mfa-section" class="card mb-4">
    <div class="card-header"><h3 class="card-title mb-0">Two-factor authentication</h3></div>
    <div class="card-body d-flex flex-wrap align-items-center justify-content-between gap-2">
      <div>
        <span v-if="mfaEnabled" class="badge text-bg-success">Enabled</span>
        <span v-else class="badge text-bg-secondary">Off</span>
        <span v-if="mfaEnabled" class="text-muted small ms-2">
          {{ mfaRecoveryRemaining }} unused recovery code{{ mfaRecoveryRemaining === 1 ? '' : 's' }} remaining
        </span>
        <span v-else class="text-muted small ms-2">Optional authenticator-app login codes</span>
      </div>
      <button type="button" class="btn btn-outline-primary" @click="mfaModalOpen = true">
        {{ mfaEnabled ? 'Manage' : 'Set up' }}
      </button>
    </div>
  </div>
  <MfaModal v-model="mfaModalOpen" @updated="onMfaUpdated" />
</template>
