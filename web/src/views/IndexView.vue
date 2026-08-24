<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { useAuth } from '@/composables/useAuth'

const { isAuthenticated } = useAuth()

const TasksView = defineAsyncComponent(() => import('@/views/TasksView.vue'))
const GuestHomeView = defineAsyncComponent(() => import('@/views/GuestHomeView.vue'))
</script>

<template>
  <TasksView v-if="isAuthenticated" />
  <GuestHomeView v-else />
</template>
