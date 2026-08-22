<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api/client'
import type { APIKey } from '@/api/types'
import { APIError } from '@/api/types'
import { useConfirm } from '@/composables/useConfirm'
import { useToast } from '@/composables/useToast'

const { push } = useToast()
const { askConfirm } = useConfirm()
const keys = ref<APIKey[]>([])
const keyName = ref('')
const mintedKey = ref('')
const renameKeyId = ref<number | null>(null)
const renameKeyValue = ref('')

function formatKeyTime(iso: string) {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

async function loadKeys() {
  try {
    keys.value = await api.listAPIKeys()
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Failed to load API keys', 'error')
  }
}

async function createKey() {
  if (!keyName.value.trim()) return
  try {
    const created = await api.createAPIKey(keyName.value.trim())
    mintedKey.value = created.key
    keyName.value = ''
    keys.value = await api.listAPIKeys()
    push('API key created — copy it now', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Create failed', 'error')
  }
}

function beginRenameKey(key: APIKey) {
  renameKeyId.value = key.id
  renameKeyValue.value = key.name
}

async function saveRenameKey() {
  if (renameKeyId.value == null || !renameKeyValue.value.trim()) return
  try {
    await api.renameAPIKey(renameKeyId.value, renameKeyValue.value.trim())
    renameKeyId.value = null
    keys.value = await api.listAPIKeys()
    push('API key renamed', 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Rename failed', 'error')
  }
}

async function revokeKey(key: APIKey) {
  const ok = await askConfirm({
    title: 'Revoke API key?',
    message: `Revoke API key “${key.name}”? Apps using it will stop working.`,
    confirmLabel: 'Revoke',
    danger: true,
  })
  if (!ok) return
  try {
    await api.revokeAPIKey(key.id)
    keys.value = await api.listAPIKeys()
    push('API key revoked', 'info')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Revoke failed', 'error')
  }
}

onMounted(() => {
  void loadKeys()
})
</script>

<template>
  <div id="api-keys-section" class="card mb-4">
    <div class="card-header">
      <h3 class="card-title mb-0">API keys</h3>
    </div>
    <div class="card-body">
      <form class="row g-2 mb-3" @submit.prevent="createKey">
        <div class="col-sm-8">
          <input v-model="keyName" type="text" class="form-control" placeholder="Key name" required maxlength="80" />
        </div>
        <div class="col-sm-4">
          <button type="submit" class="btn btn-primary w-100">Create key</button>
        </div>
      </form>
      <div v-if="mintedKey" class="api-key-reveal-panel">
        <p class="api-key-reveal-title">New key (shown once)</p>
        <p class="text-break mb-0"><code>{{ mintedKey }}</code></p>
      </div>
      <div class="api-key-list">
        <div v-for="key in keys" :key="key.id" class="api-key-card">
          <div class="api-key-card-header">
            <template v-if="renameKeyId === key.id">
              <input
                v-model="renameKeyValue"
                type="text"
                class="form-control form-control-sm api-key-rename-input"
                maxlength="80"
                @keyup.enter="saveRenameKey"
              />
              <div class="d-flex gap-2">
                <button type="button" class="btn btn-sm btn-primary" :disabled="!renameKeyValue.trim()" @click="saveRenameKey">Save</button>
                <button type="button" class="btn btn-sm btn-secondary" @click="renameKeyId = null">Cancel</button>
              </div>
            </template>
            <template v-else>
              <span class="api-key-name">{{ key.name }}</span>
              <div class="d-flex gap-2">
                <button type="button" class="btn btn-sm btn-outline-secondary" @click="beginRenameKey(key)">Rename</button>
                <button type="button" class="btn btn-sm btn-outline-danger" @click="revokeKey(key)">Revoke</button>
              </div>
            </template>
          </div>
          <div class="api-key-card-fields">
            <div class="api-key-field">
              <span class="api-key-field-label">Prefix</span>
              <span class="api-key-field-value api-key-prefix">{{ key.key_prefix }}</span>
            </div>
            <div class="api-key-field">
              <span class="api-key-field-label">Created</span>
              <span class="api-key-field-value">{{ formatKeyTime(key.created_at) }}</span>
            </div>
            <div class="api-key-field">
              <span class="api-key-field-label">Last used</span>
              <span class="api-key-field-value">{{ key.last_used_at ? formatKeyTime(key.last_used_at) : 'Never' }}</span>
            </div>
          </div>
        </div>
        <p v-if="!keys.length" class="text-muted mb-0">No API keys.</p>
      </div>
    </div>
  </div>
</template>
