<script setup lang="ts">
import { computed, ref } from 'vue'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
import { useSite } from '@/composables/useSite'
import { useToast } from '@/composables/useToast'

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

const { siteInfo } = useSite()
const toast = useToast()
const inputEl = ref<HTMLInputElement | null>(null)
const busy = ref(false)

const enabled = computed(() => !!siteInfo.value?.image_hosting_enabled)

function pick() {
  if (props.disabled || busy.value || !enabled.value) return
  inputEl.value?.click()
}

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  const max = siteInfo.value?.image_max_bytes || 5 * 1024 * 1024
  if (file.size > max) {
    toast.push(`Image is larger than the ${Math.round(max / (1024 * 1024))} MB limit`, 'error')
    return
  }
  busy.value = true
  try {
    const uploaded = await api.uploadImage(file)
    const alt = file.name.replace(/\.[^.]+$/, '') || 'image'
    emit('insert', `![${alt}](${uploaded.url})`)
  } catch (err) {
    toast.push(err instanceof APIError ? err.message : 'Image upload failed', 'error')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <span v-if="enabled" class="image-insert">
    <input
      ref="inputEl"
      type="file"
      class="d-none"
      accept="image/jpeg,image/png,image/gif,image/webp,.jpg,.jpeg,.png,.gif,.webp"
      @change="onFile"
    />
    <button
      type="button"
      class="btn btn-outline-secondary"
      :class="compact ? 'btn-sm' : ''"
      :disabled="disabled || busy"
      @click="pick"
    >
      {{ busy ? 'Uploading…' : 'Insert image' }}
    </button>
  </span>
</template>
