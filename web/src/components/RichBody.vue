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
  <div class="rich-body" :class="{ 'rich-body--compact': compact }">
    <template v-for="(part, i) in parts" :key="i">
      <span v-if="part.type === 'text'" class="rich-body-text">{{ part.value }}</span>
      <span v-else-if="part.type === 'mention'" class="rich-body-mention">{{ part.raw }}</span>
      <a
        v-else-if="part.type === 'image'"
        class="rich-body-image-link"
        :href="part.src"
        target="_blank"
        rel="noopener noreferrer"
        :title="part.alt ? `Open ${part.alt}` : 'Open image'"
        @click.stop
      >
        <img
          class="rich-body-image"
          :src="part.src"
          :alt="part.alt || 'image'"
          loading="lazy"
          decoding="async"
        />
      </a>
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
.rich-body-text {
  white-space: pre-wrap;
}
.rich-body-mention {
  font-weight: 600;
  color: var(--ordryn-accent, #2563eb);
}
.rich-body-image-link {
  display: block;
  max-width: 100%;
  margin: 0.5rem 0;
}
.rich-body-image {
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
.rich-body--compact .rich-body-image-link {
  margin: 0.25rem 0;
}
.rich-body--compact .rich-body-image {
  max-height: 6rem;
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
