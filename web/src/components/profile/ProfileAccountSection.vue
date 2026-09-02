<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
import AvatarCropModal from '@/components/AvatarCropModal.vue'
import MfaModal from '@/components/MfaModal.vue'
import { useAuth } from '@/composables/useAuth'
import { useConfirm } from '@/composables/useConfirm'
import { useSite } from '@/composables/useSite'
import { useToast } from '@/composables/useToast'
import { AVATAR_ACCEPT, isAllowedAvatarFile } from '@/utils/imageUpload'

defineProps<{
  email: string
  username: string
}>()

const { user, uploadAvatar, deleteAvatar } = useAuth()
const { siteInfo } = useSite()
const { push } = useToast()
const { askConfirm } = useConfirm()

const imageHostingEnabled = computed(() => !!siteInfo.value?.image_hosting_enabled)
const maxImageBytes = computed(() => siteInfo.value?.image_max_bytes || 5 * 1024 * 1024)

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const cropModalOpen = ref(false)
const avatarBusy = ref(false)

const showPasswordForm = ref(false)
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const busy = ref(false)
const mfaEnabled = ref(false)
const mfaRecoveryRemaining = ref(0)
const mfaModalOpen = ref(false)

function onSelectFileClick() {
  if (!imageHostingEnabled.value) return
  fileInput.value?.click()
}

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (!isAllowedAvatarFile(file)) {
    push('Profile pictures must be a PNG or JPEG image.', 'error')
    target.value = ''
    return
  }

  if (file.size > maxImageBytes.value) {
    const mb = Math.round(maxImageBytes.value / (1024 * 1024))
    push(`Image is larger than the ${mb} MB limit.`, 'error')
    target.value = ''
    return
  }

  selectedFile.value = file
  cropModalOpen.value = true
}

async function onAvatarCropped(blob: Blob) {
  avatarBusy.value = true
  try {
    await uploadAvatar(blob)
    push('Profile picture updated', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Failed to upload profile picture', 'error')
  } finally {
    avatarBusy.value = false
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  }
}

async function onRemoveAvatar() {
  const ok = await askConfirm({
    title: 'Remove profile picture',
    message: 'Are you sure you want to remove your profile picture?',
    confirmLabel: 'Remove',
    danger: true,
  })
  if (!ok) return

  avatarBusy.value = true
  try {
    await deleteAvatar()
    push('Profile picture removed', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Failed to remove profile picture', 'error')
  } finally {
    avatarBusy.value = false
  }
}

const userInitials = computed(() => {
  const name = user.value?.user_name || user.value?.email || ''
  return name.slice(0, 2).toUpperCase()
})

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
      <h3 class="card-title mb-0">Profile Picture</h3>
    </div>
    <div class="card-body">
      <div class="d-flex flex-column flex-sm-row align-items-sm-center gap-3">
        <div class="avatar-container position-relative flex-shrink-0">
          <img
            v-if="user?.avatar_url"
            :src="user.avatar_url"
            alt="Profile picture"
            class="rounded-circle object-fit-cover avatar-img"
          />
          <div v-else class="rounded-circle avatar-placeholder d-flex align-items-center justify-content-center fw-bold">
            {{ userInitials }}
          </div>
        </div>

        <div class="flex-grow-1">
          <template v-if="imageHostingEnabled">
            <div class="d-flex flex-wrap gap-2 mb-2">
              <input
                ref="fileInput"
                type="file"
                class="d-none"
                :accept="AVATAR_ACCEPT"
                @change="onFileChange"
              />
              <button
                type="button"
                class="btn btn-outline-primary btn-sm d-inline-flex align-items-center gap-1"
                :disabled="avatarBusy"
                @click="onSelectFileClick"
              >
                <i class="bi bi-upload" />
                <span>{{ user?.avatar_url ? 'Change picture' : 'Upload picture' }}</span>
              </button>
              <button
                v-if="user?.avatar_url"
                type="button"
                class="btn btn-outline-danger btn-sm d-inline-flex align-items-center gap-1"
                :disabled="avatarBusy"
                @click="onRemoveAvatar"
              >
                <i class="bi bi-trash" />
                <span>Remove</span>
              </button>
            </div>
            <div class="text-muted small">
              PNG or JPG only. Supports cropping before upload. Max size: {{ Math.round(maxImageBytes / (1024 * 1024)) }} MB.
            </div>
          </template>
          <template v-else>
            <div class="text-muted small">
              <i class="bi bi-info-circle me-1" />
              Profile picture uploads are disabled because image hosting is not configured on this server.
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>

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
  <AvatarCropModal
    v-model="cropModalOpen"
    :image-file="selectedFile"
    @cropped="onAvatarCropped"
  />
</template>

<style scoped>
.avatar-img {
  width: 72px;
  height: 72px;
  border: 2px solid var(--ordryn-header-border, #dee2e6);
}
.avatar-placeholder {
  width: 72px;
  height: 72px;
  background-color: var(--ordryn-muted-bg, #e9ecef);
  color: var(--ordryn-text, #495057);
  font-size: 1.4rem;
  border: 2px solid var(--ordryn-header-border, #dee2e6);
}
</style>
