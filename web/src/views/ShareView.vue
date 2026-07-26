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
          <ul class="list-group list-group-flush">
            <li
              v-for="t in view?.tasks || []"
              :key="t.id"
              class="list-group-item d-flex justify-content-between align-items-start"
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
            </li>
            <li v-if="!(view?.tasks || []).length" class="list-group-item text-muted">No tasks in this list.</li>
          </ul>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import type { ShareLinkView } from '@/api/types'
import NotFoundView from '@/views/NotFoundView.vue'

const route = useRoute()
const view = ref<ShareLinkView | null>(null)
const loading = ref(true)
const notFound = ref(false)

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
