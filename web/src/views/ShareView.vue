<template>
  <NotFoundView v-if="notFound" />
  <div v-else class="container mt-4 mb-5">
    <div class="card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h1 class="h4 mb-0">Shared list</h1>
        <span class="badge text-bg-secondary">Read only</span>
      </div>
      <div class="card-body">
        <p v-if="loading" class="text-muted">Loading…</p>
        <template v-else>
          <p class="text-muted small">
            Viewing a shared {{ view?.scope_type }} list. Sign in to collaborate on shared projects.
          </p>
          <div class="list-group list-group-flush">
            <button
              v-for="t in view?.tasks || []"
              :key="t.id"
              type="button"
              class="list-group-item list-group-item-action d-flex justify-content-between align-items-start text-start"
              @click="selectedTask = t"
            >
              <div>
                <span :class="{ 'text-decoration-line-through text-muted': t.completed }">{{ t.title }}</span>
                <div class="small text-muted" v-if="t.project || t.due_date">
                  <span v-if="t.project">{{ t.project }}</span>
                  <span v-if="t.project && t.due_date"> · </span>
                  <span v-if="t.due_date">Due {{ t.due_date }}</span>
                </div>
              </div>
              <span v-if="t.completed" class="badge text-bg-success">Done</span>
            </button>
            <div v-if="!(view?.tasks || []).length" class="list-group-item text-muted">No tasks in this list.</div>
          </div>
        </template>
      </div>
    </div>
    <ShareTaskModal v-if="selectedTask" :task="selectedTask" @close="selectedTask = null" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import type { ShareLinkTask, ShareLinkView } from '@/api/types'
import ShareTaskModal from '@/components/ShareTaskModal.vue'
import NotFoundView from '@/views/NotFoundView.vue'

const route = useRoute()
const view = ref<ShareLinkView | null>(null)
const loading = ref(true)
const notFound = ref(false)
const selectedTask = ref<ShareLinkTask | null>(null)

onMounted(async () => {
  const token = String(route.params.token || '')
  if (!token) {
    notFound.value = true
    loading.value = false
    return
  }
  try {
    view.value = await api.viewShareLink(token)
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
})
</script>
