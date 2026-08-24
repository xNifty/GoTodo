<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/api/types'
import { APIError } from '@/api/types'
import { useToast } from '@/composables/useToast'
import ProjectSharePanel from '@/components/ProjectSharePanel.vue'
import ProjectWorkflowPanel from '@/components/ProjectWorkflowPanel.vue'
import ProjectGitHubPanel from '@/components/ProjectGitHubPanel.vue'
import ProjectTagsPanel from '@/components/ProjectTagsPanel.vue'

const props = defineProps<{
  open: boolean
  project: Project | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
  changed: []
  'columns-changed': []
}>()

const toast = useToast()
const name = ref('')
const description = ref('')
const saving = ref(false)
const isOwner = computed(() => (props.project?.role || 'owner') === 'owner')

watch(
  () => [props.open, props.project] as const,
  ([open, project]) => {
    if (open && project) {
      name.value = project.name
      description.value = project.description || ''
    }
  },
  { immediate: true },
)

function close() {
  emit('close')
}

async function saveBasics() {
  if (!props.project || !name.value.trim() || !isOwner.value) return
  saving.value = true
  try {
    await api.updateProject(props.project.id, {
      name: name.value.trim(),
      description: description.value.trim(),
    })
    toast.push('Project updated', 'success')
    emit('saved')
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Failed to update project', 'error')
  } finally {
    saving.value = false
  }
}

function onPanelChanged() {
  emit('changed')
}

function onColumnsChanged() {
  emit('columns-changed')
}
</script>

<template>
  <div
    v-if="open && project"
    class="modal fade show d-block"
    style="background: rgba(0,0,0,0.5);"
    tabindex="-1"
    @click.self="close"
  >
    <div class="modal-dialog modal-dialog-centered modal-lg modal-dialog-scrollable">
      <div
        class="modal-content border-0 shadow"
        style="background: var(--ordryn-card-bg); color: var(--ordryn-text);"
      >
        <div class="modal-header border-0 pb-0">
          <h5 class="modal-title fw-bold">{{ isOwner ? 'Edit Project' : 'Project settings' }}</h5>
          <button type="button" class="btn-close" aria-label="Close" @click="close" />
        </div>
        <div class="modal-body py-3">
          <div class="mb-3">
            <label for="edit-project-name" class="form-label small fw-bold">Project Name</label>
            <input
              id="edit-project-name"
              v-model="name"
              type="text"
              class="form-control"
              maxlength="50"
              placeholder="Project Name"
              :readonly="!isOwner"
            />
            <div class="d-flex justify-content-between">
              <small class="form-hint">Max 50 characters</small>
              <small class="text-muted">{{ name.length }}/50</small>
            </div>
          </div>

          <div class="mb-3">
            <label for="edit-project-description" class="form-label small fw-bold">Description</label>
            <textarea
              id="edit-project-description"
              v-model="description"
              class="form-control"
              rows="3"
              maxlength="1000"
              placeholder="Optional details about this project"
              :readonly="!isOwner"
            />
            <div class="d-flex justify-content-end">
              <small class="text-muted">{{ description.length }}/1000</small>
            </div>
          </div>

          <div v-if="isOwner" class="d-flex justify-content-end mb-4">
            <button
              type="button"
              class="btn btn-sm btn-primary px-3"
              :disabled="saving || !name.trim()"
              @click="saveBasics"
            >
              Save details
            </button>
          </div>

          <hr class="my-3 opacity-25" />

          <div class="mb-4">
            <ProjectWorkflowPanel
              :project="project"
              @changed="onPanelChanged"
              @columns-changed="onColumnsChanged"
            />
          </div>

          <hr class="my-3 opacity-25" />

          <div class="mb-4">
            <ProjectGitHubPanel :project="project" @changed="onPanelChanged" />
          </div>

          <hr class="my-3 opacity-25" />

          <div class="mb-4">
            <ProjectTagsPanel :project="project" @changed="onPanelChanged" />
          </div>

          <hr class="my-3 opacity-25" />

          <ProjectSharePanel :project="project" @changed="onPanelChanged" />
        </div>
        <div class="modal-footer border-0 pt-0 justify-content-end">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="close">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>
