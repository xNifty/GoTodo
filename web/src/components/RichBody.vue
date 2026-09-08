<script setup lang="ts">
import { computed } from 'vue'
import MarkdownNodes from '@/components/MarkdownNodes.vue'
import { parseCommentMarkdown } from '@/utils/commentMarkdown'

const props = withDefaults(
  defineProps<{
    body: string
    compact?: boolean
    taskTitle?: (id: number) => string | undefined
  }>(),
  { compact: false },
)

const emit = defineEmits<{
  'open-task': [id: number]
}>()

const nodes = computed(() => parseCommentMarkdown(props.body || ''))
</script>

<template>
  <div class="rich-body" :class="{ 'rich-body--compact': compact }">
    <MarkdownNodes :nodes="nodes" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
  </div>
</template>

<style scoped>
.rich-body {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow-wrap: anywhere;
  word-break: break-word;
}
.rich-body :deep(.rich-body-paragraph) {
  margin: 0 0 0.5em;
}
.rich-body :deep(.rich-body-paragraph:last-child) {
  margin-bottom: 0;
}
.rich-body :deep(.rich-body-list) {
  margin: 0.35em 0 0.35em 1.25em;
  padding: 0;
}
.rich-body :deep(.rich-body-list:last-child) {
  margin-bottom: 0;
}
.rich-body :deep(.rich-body-list-item > .rich-body-paragraph) {
  margin: 0;
}
.rich-body :deep(.rich-body-text) {
  white-space: pre-wrap;
}
.rich-body :deep(.rich-body-mention) {
  font-weight: 600;
  color: var(--ordryn-accent, #2563eb);
}
.rich-body :deep(.rich-body-underline) {
  text-decoration: underline;
}
.rich-body :deep(.rich-body-link) {
  color: var(--ordryn-accent, #2563eb);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.rich-body :deep(.rich-body-image-link) {
  display: block;
  max-width: 100%;
  margin: 0.5rem 0;
}
.rich-body :deep(.rich-body-image) {
  display: block;
  max-width: 100%;
  width: auto;
  height: auto;
  max-height: min(240px, 40vh);
  object-fit: contain;
  border-radius: 0.375rem;
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  background: var(--ordryn-muted-bg, #f8f6ee);
}
.rich-body--compact :deep(.rich-body-image-link) {
  margin: 0.25rem 0;
}
.rich-body--compact :deep(.rich-body-image) {
  max-height: 6rem;
}
.rich-body :deep(.rich-body-task-link) {
  display: inline;
  padding: 0;
  margin: 0;
  border: 0;
  background: none;
  color: var(--ordryn-accent, #2563eb);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 2px;
  cursor: pointer;
}
.rich-body :deep(.rich-body-task-link:hover),
.rich-body :deep(.rich-body-link:hover) {
  filter: brightness(1.1);
}
</style>
