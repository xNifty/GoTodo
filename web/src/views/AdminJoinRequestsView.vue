<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/client'
import type { JoinRequest } from '@/api/types'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import AdminSubnav from '@/components/AdminSubnav.vue'

const toast = useToast()
const { askConfirm } = useConfirm()
const requests = ref<JoinRequest[]>([])
const busyId = ref<number | null>(null)

async function load() {
  try {
    requests.value = await api.listAdminJoinRequests()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to load join requests', 'error')
  }
}

async function approve(jr: JoinRequest) {
  const ok = await askConfirm({
    title: 'Approve join request?',
    message: `Approve ${jr.email}? They will receive an invite email if outbound email is configured.`,
    confirmLabel: 'Approve',
  })
  if (!ok) return
  busyId.value = jr.id
  try {
    await api.approveJoinRequest(jr.id)
    toast.push(`Invite sent to ${jr.email}`, 'success')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Approve failed', 'error')
  } finally {
    busyId.value = null
  }
}

async function deny(jr: JoinRequest) {
  const ok = await askConfirm({
    title: 'Deny join request?',
    message: `Deny ${jr.email}?`,
    confirmLabel: 'Deny',
    danger: true,
  })
  if (!ok) return
  busyId.value = jr.id
  try {
    await api.denyJoinRequest(jr.id)
    toast.push('Request denied', 'info')
    await load()
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Deny failed', 'error')
  } finally {
    busyId.value = null
  }
}

function statusBadge(status: JoinRequest['status']) {
  if (status === 'pending') return 'bg-warning text-dark'
  if (status === 'approved') return 'bg-success'
  return 'bg-secondary'
}

const pending = computed(() => requests.value.filter((r) => r.status === 'pending'))
const reviewed = computed(() => requests.value.filter((r) => r.status !== 'pending'))

onMounted(load)
</script>

<template>
  <div class="container mt-3">
    <AdminSubnav :pending-count="pending.length" />
    <h1>Join requests</h1>
    <p class="text-muted">Review visitors who asked to join this site.</p>

    <div class="card mb-4">
      <div class="card-header"><h2 class="h5 mb-0">Pending</h2></div>
      <ul class="list-group list-group-flush">
        <li v-for="jr in pending" :key="jr.id" class="list-group-item">
          <div class="d-flex justify-content-between align-items-start gap-3 flex-wrap">
            <div>
              <strong>{{ jr.email }}</strong>
              <span class="badge ms-2" :class="statusBadge(jr.status)">{{ jr.status }}</span>
              <div class="text-muted small">{{ jr.created_at }}</div>
              <p v-if="jr.message" class="mb-0 mt-2">{{ jr.message }}</p>
            </div>
            <div class="d-flex gap-2">
              <button
                type="button"
                class="btn btn-sm btn-primary"
                :disabled="busyId === jr.id"
                @click="approve(jr)"
              >
                Approve
              </button>
              <button
                type="button"
                class="btn btn-sm btn-outline-danger"
                :disabled="busyId === jr.id"
                @click="deny(jr)"
              >
                Deny
              </button>
            </div>
          </div>
        </li>
        <li v-if="!pending.length" class="list-group-item text-muted">No pending requests.</li>
      </ul>
    </div>

    <div class="card">
      <div class="card-header"><h2 class="h5 mb-0">Reviewed</h2></div>
      <ul class="list-group list-group-flush">
        <li v-for="jr in reviewed" :key="jr.id" class="list-group-item">
          <div>
            <strong>{{ jr.email }}</strong>
            <span class="badge ms-2" :class="statusBadge(jr.status)">{{ jr.status }}</span>
            <div class="text-muted small">
              {{ jr.created_at }}
              <span v-if="jr.reviewed_at"> · reviewed {{ jr.reviewed_at }}</span>
            </div>
            <p v-if="jr.message" class="mb-0 mt-2 text-muted">{{ jr.message }}</p>
          </div>
        </li>
        <li v-if="!reviewed.length" class="list-group-item text-muted">No reviewed requests yet.</li>
      </ul>
    </div>
  </div>
</template>
