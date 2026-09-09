import { computed, ref } from 'vue'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
import { useSite } from '@/composables/useSite'
import { useToast } from '@/composables/useToast'
import { altFromFilename, isAllowedImageFile, toImageMarkdown } from '@/utils/imageUpload'

export function useImageUpload() {
  const { siteInfo } = useSite()
  const toast = useToast()
  const busy = ref(false)
  const enabled = computed(() => {
    if (siteInfo.value === null) return true
    return !!siteInfo.value?.image_hosting_enabled
  })

  async function uploadImageFile(file: File): Promise<string | null> {
    if (!isAllowedImageFile(file)) {
      toast.push('Use a JPEG, PNG, GIF, or WebP image', 'error')
      return null
    }
    const max = siteInfo.value?.image_max_bytes || 5 * 1024 * 1024
    if (file.size > max) {
      toast.push(`Image is larger than the ${Math.round(max / (1024 * 1024))} MB limit`, 'error')
      return null
    }
    busy.value = true
    try {
      const uploaded = await api.uploadImage(file)
      return toImageMarkdown(uploaded.url, altFromFilename(file.name || 'image'))
    } catch (err) {
      const msg = err instanceof APIError ? err.message : err instanceof Error ? err.message : 'Image upload failed'
      toast.push(msg, 'error')
      return null
    } finally {
      busy.value = false
    }
  }

  return { enabled, busy, uploadImageFile }
}
