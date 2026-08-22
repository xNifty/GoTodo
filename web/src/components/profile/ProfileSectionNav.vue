<script setup lang="ts">
import { PROFILE_SECTION_ITEMS, type ProfileSection } from '@/utils/profileSections'

defineProps<{
  active: ProfileSection
}>()

const emit = defineEmits<{
  select: [section: ProfileSection]
}>()
</script>

<template>
  <nav class="profile-section-nav" aria-label="Profile sections">
    <ul class="nav nav-pills profile-section-pills gap-1 flex-nowrap overflow-auto flex-md-column">
      <li v-for="item in PROFILE_SECTION_ITEMS" :key="item.id" class="nav-item">
        <button
          type="button"
          class="nav-link profile-section-link"
          :class="{ active: active === item.id }"
          :aria-current="active === item.id ? 'page' : undefined"
          @click="emit('select', item.id)"
        >
          {{ item.label }}
        </button>
      </li>
    </ul>
  </nav>
</template>
