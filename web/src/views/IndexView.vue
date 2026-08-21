<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { useAuth } from '@/composables/useAuth'

defineProps<{
  mobileSidebarOpen?: boolean
}>()

const emit = defineEmits<{
  'close-mobile-sidebar': []
}>()

const { isAuthenticated } = useAuth()

const TasksView = defineAsyncComponent(() => import('@/views/TasksView.vue'))
const GuestHomeView = defineAsyncComponent(() => import('@/views/GuestHomeView.vue'))
</script>

<template>
  <TasksView
    v-if="isAuthenticated"
    :mobile-sidebar-open="mobileSidebarOpen"
    @close-mobile-sidebar="emit('close-mobile-sidebar')"
  />
  <GuestHomeView v-else />
</template>
