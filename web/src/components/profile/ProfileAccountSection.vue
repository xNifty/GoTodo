<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
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
</template>
