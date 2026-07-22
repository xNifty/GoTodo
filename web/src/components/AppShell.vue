<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useSite } from '@/composables/useSite'
import { api } from '@/api/client'
import ModernHeader from '@/components/modern/ModernHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import ToastHost from '@/components/ToastHost.vue'
import TaskSidebar from '@/components/TaskSidebar.vue'
import ChangelogModal from '@/components/ChangelogModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'

const route = useRoute()
const { isAuthenticated } = useAuth()
const { siteInfo, refresh: refreshSite } = useSite()
const overdueCount = ref(0)
const mobileSidebarOpen = ref(false)

const showAnnouncement = computed(
  () =>
    !!siteInfo.value?.enable_global_announcement &&
    !!siteInfo.value?.global_announcement_text &&
    !siteInfo.value?.announcement_dismissed,
)

const showChangelog = computed(() => siteInfo.value?.show_changelog !== false)

async function loadOverdue() {
  if (!isAuthenticated.value) return
  try {
    const stats = await api.dashboard()
    overdueCount.value = stats.overdue_count
  } catch {
    overdueCount.value = 0
  }
}

async function dismissAnnouncement() {
  try {
    await api.dismissAnnouncement()
    if (siteInfo.value) {
      siteInfo.value = { ...siteInfo.value, announcement_dismissed: true }
    }
  } catch {
    if (siteInfo.value) {
      siteInfo.value = { ...siteInfo.value, announcement_dismissed: true }
    }
  }
}

function toggleMobileSidebar() {
  mobileSidebarOpen.value = !mobileSidebarOpen.value
}

onMounted(() => {
  void loadOverdue()
  void refreshSite()
})
</script>

<template>
  <div class="ordryn-app-shell">
    <ModernHeader @toggle-mobile-sidebar="toggleMobileSidebar" />

    <div
      v-if="showAnnouncement && siteInfo?.global_announcement_text"
      id="global-announcement"
      class="global-announcement-wrapper"
    >
      <div class="container py-2">
        <div class="alert alert-primary alert-dismissible fade show mb-0" role="alert">
          <i class="bi bi-megaphone-fill me-2" />
          <strong>{{ siteInfo.global_announcement_text }}</strong>
          <button
            type="button"
            class="btn-close no-invert"
            aria-label="Close"
            @click="dismissAnnouncement"
          />
        </div>
      </div>
    </div>

    <main class="site-main flex-grow-1 d-flex flex-column" data-page="spa">
      <RouterView :mobile-sidebar-open="mobileSidebarOpen" @close-mobile-sidebar="mobileSidebarOpen = false" />
    </main>

    <!-- Global Footer on all pages EXCEPT sidebar layout pages (/, /dashboard, /calendar) -->
    <AppFooter v-if="!['/', '/dashboard', '/calendar'].includes(route.path)" class="container" />

    <ToastHost />
    <ConfirmModal />
    <TaskSidebar v-if="isAuthenticated" />
    <ChangelogModal v-if="showChangelog" />
  </div>
</template>
