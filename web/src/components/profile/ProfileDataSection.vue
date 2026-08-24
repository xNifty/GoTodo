<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'

const { push } = useToast()

async function exportTasks(format: 'json' | 'csv') {
  try {
    await api.downloadExport(format)
    push(`Exported ${format.toUpperCase()}`, 'success')
  } catch (err) {
    push(err instanceof APIError ? err.message : 'Export failed', 'error')
  }
}
</script>

<template>
  <div class="card mb-4">
    <div class="card-header">
      <h3 class="card-title mb-0">Export &amp; import</h3>
    </div>
    <div class="card-body">
      <p class="text-muted small mb-3">
        Download your tasks, or import a CSV from the dedicated import page.
      </p>
      <div class="d-flex flex-wrap gap-2">
        <button type="button" class="btn btn-primary" @click="exportTasks('json')">Export JSON</button>
        <button type="button" class="btn btn-outline-secondary" @click="exportTasks('csv')">Export CSV</button>
        <RouterLink to="/import" class="btn btn-outline-primary">Import CSV</RouterLink>
      </div>
    </div>
  </div>
</template>
