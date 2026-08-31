<script setup lang="ts">
import { ref } from 'vue'
import { useImageUpload } from '@/composables/useImageUpload'
import { IMAGE_ACCEPT, armTaskOverlayFileGuard } from '@/utils/imageUpload'

const props = withDefaults(
  defineProps<{
    disabled?: boolean
    compact?: boolean
  }>(),
  { disabled: false, compact: false },
)

const emit = defineEmits<{
  insert: [markdown: string]
}>()

const { enabled, busy, uploadImageFile } = useImageUpload()
const inputEl = ref<HTMLInputElement | null>(null)

function pick() {
  if (props.disabled || busy.value || !enabled.value) return
  armTaskOverlayFileGuard()
  inputEl.value?.click()
}

async function onFile(ev: Event) {
  armTaskOverlayFileGuard()
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  const markdown = await uploadImageFile(file)
  if (markdown) emit('insert', markdown)
}
</script>

<template>
  <span v-if="enabled" class="image-insert">
    <input
      ref="inputEl"
      type="file"
      class="d-none"
      :accept="IMAGE_ACCEPT"
      @change="onFile"
    />
    <button
      type="button"
      class="btn btn-outline-secondary"
      :class="compact ? 'btn-sm' : ''"
      :disabled="disabled || busy"
      title="Upload an image into this comment"
      @click.stop="pick"
    >
      {{ busy ? 'Uploading…' : 'Insert image' }}
    </button>
  </span>
</template>
