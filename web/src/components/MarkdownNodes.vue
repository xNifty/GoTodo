<script setup lang="ts">
import type { MdNode } from '@/utils/commentMarkdown'

defineOptions({ name: 'MarkdownNodes' })

defineProps<{
  nodes: MdNode[]
  taskTitle?: (id: number) => string | undefined
}>()

const emit = defineEmits<{
  'open-task': [id: number]
}>()

function label(id: number, title: string) {
  return `Task #${id} - ${title}`
}
</script>

<template>
  <template v-for="(node, i) in nodes" :key="i">
    <p v-if="node.type === 'paragraph'" class="rich-body-paragraph">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </p>
    <ul v-else-if="node.type === 'bullet_list'" class="rich-body-list">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </ul>
    <ol v-else-if="node.type === 'ordered_list'" class="rich-body-list rich-body-list--ordered">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </ol>
    <li v-else-if="node.type === 'list_item'" class="rich-body-list-item">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </li>
    <strong v-else-if="node.type === 'strong'">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </strong>
    <em v-else-if="node.type === 'em'">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </em>
    <u v-else-if="node.type === 'underline'" class="rich-body-underline">
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </u>
    <a
      v-else-if="node.type === 'link'"
      class="rich-body-link"
      :href="node.href"
      target="_blank"
      rel="noopener noreferrer"
      @click.stop
    >
      <MarkdownNodes :nodes="node.children" :task-title="taskTitle" @open-task="emit('open-task', $event)" />
    </a>
    <a
      v-else-if="node.type === 'image'"
      class="rich-body-image-link"
      :href="node.src"
      target="_blank"
      rel="noopener noreferrer"
      :title="node.alt ? `Open ${node.alt}` : 'Open image'"
      @click.stop
    >
      <img
        class="rich-body-image"
        :src="node.src"
        :alt="node.alt || 'image'"
        loading="lazy"
        decoding="async"
      />
    </a>
    <span v-else-if="node.type === 'mention'" class="rich-body-mention">{{ node.raw }}</span>
    <button
      v-else-if="node.type === 'task' && taskTitle?.(node.id)"
      type="button"
      class="rich-body-task-link"
      @click="emit('open-task', node.id)"
    >
      {{ label(node.id, taskTitle(node.id)!) }}
    </button>
    <span v-else-if="node.type === 'task'">{{ node.raw }}</span>
    <span v-else-if="node.type === 'text'" class="rich-body-text">{{ node.value }}</span>
    <br v-else-if="node.type === 'break'" />
  </template>
</template>
