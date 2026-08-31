<script setup lang="ts">
import { computed } from 'vue'
import { splitCommentBody } from '@/utils/taskCommentBody'

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

const parts = computed(() => splitCommentBody(props.body || ''))

function label(id: number, title: string) {
  return `Task #${id} - ${title}`
}
</script>

<template>
  <span class="rich-body" :class="{ 'rich-body--compact': compact }">
    <template v-for="(part, i) in parts" :key="i">
      <span v-if="part.type === 'text'" class="rich-body-text">{{ part.value }}</span>
      <span v-else-if="part.type === 'mention'" class="rich-body-mention">{{ part.raw }}</span>
      <img
        v-else-if="part.type === 'image'"
        class="rich-body-image"
        :src="part.src"
        :alt="part.alt || 'image'"
        loading="lazy"
        decoding="async"
        @click.stop
      />
      <button
        v-else-if="part.type === 'task' && taskTitle?.(part.id)"
        type="button"
        class="rich-body-task-link"
        @click="emit('open-task', part.id)"
      >
        {{ label(part.id, taskTitle(part.id)!) }}
      </button>
      <span v-else-if="part.type === 'task'">{{ part.raw }}</span>
    </template>
  </span>
</template>

<style scoped>
.rich-body {
  display: inline;
  word-break: break-word;
}
.rich-body-text {
  white-space: pre-wrap;
}
.rich-body-mention {
  font-weight: 600;
  color: var(--ordryn-accent, #2563eb);
}
.rich-body-image {
  display: block;
  max-width: 100%;
  height: auto;
  max-height: 28rem;
  margin: 0.45rem 0;
  border-radius: 0.375rem;
  border: 1px solid var(--ordryn-card-border, #dee2e6);
  background: var(--ordryn-muted-bg, #f8f6ee);
}
.rich-body--compact .rich-body-image {
  max-height: 8rem;
  margin: 0.3rem 0;
}
.rich-body-task-link {
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
.rich-body-task-link:hover {
  filter: brightness(1.1);
}
</style>
