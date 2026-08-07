<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useSite } from '@/composables/useSite'
import { useTheme } from '@/composables/useTheme'
import { useToast } from '@/composables/useToast'

const emit = defineEmits<{
  'toggle-mobile-sidebar': []
}>()

const { isAuthenticated, user, logout, hasPermission } = useAuth()
const { siteName, siteInfo } = useSite()
const { currentTheme, availableThemes, setTheme } = useTheme()
const { push } = useToast()
const router = useRouter()

const showChangelog = computed(() => siteInfo.value?.show_changelog !== false)

async function onLogout() {
  try {
    await logout()
    push('Signed out', 'info')
    await router.push({ name: 'login' })
  } catch (err) {
    push(err instanceof Error ? err.message : 'Logout failed', 'error')
  }
}
</script>

<template>
  <header class="ordryn-header border-bottom py-2 px-3" style="background-color: var(--ordryn-header-bg); border-color: var(--ordryn-header-border) !important;">
    <div class="d-flex align-items-center justify-content-between">
      <!-- Left: Mobile Toggle & Brand Logo -->
      <div class="d-flex align-items-center gap-3">
        <button
          type="button"
          class="btn btn-link text-decoration-none p-0 d-md-none me-1"
          style="color: var(--ordryn-text); font-size: 1.25rem;"
          aria-label="Toggle mobile menu"
          @click="emit('toggle-mobile-sidebar')"
        >
          <i class="bi bi-list" />
        </button>

        <RouterLink to="/" class="d-flex align-items-center gap-2 text-decoration-none" style="color: var(--ordryn-text);">
          <div class="brand-logo-icon d-flex align-items-center justify-content-center rounded-3 px-2 py-1" style="background: var(--ordryn-accent-light); color: var(--ordryn-accent); font-weight: 800;">
            <i class="bi bi-layers-half" style="font-size: 1.2rem;" />
          </div>
          <span class="fw-bold fs-5 tracking-tight">{{ siteName }}</span>
        </RouterLink>

        <!-- Top Navigation Links (Desktop) -->
        <nav class="d-none d-md-flex align-items-center gap-3 ms-4">
          <template v-if="showChangelog">
            <a
              href="#changelogModal"
              data-bs-toggle="modal"
              data-bs-target="#changelogModal"
              class="nav-link-item text-decoration-none small fw-medium"
              style="color: var(--ordryn-muted);"
            >Changelog</a>
          </template>
          <a
            href="#shortcutsModal"
            data-bs-toggle="modal"
            data-bs-target="#shortcutsModal"
            class="nav-link-item text-decoration-none small fw-medium"
            style="color: var(--ordryn-muted);"
          >Shortcuts</a>
          <RouterLink to="/docs/guide" class="nav-link-item text-decoration-none small fw-medium" style="color: var(--ordryn-muted);">How to use</RouterLink>
          <RouterLink to="/docs/api/v1" class="nav-link-item text-decoration-none small fw-medium" style="color: var(--ordryn-muted);">API</RouterLink>
          <template v-if="isAuthenticated">
            <RouterLink v-if="hasPermission('admin')" to="/admin" class="nav-link-item text-decoration-none small fw-medium" style="color: var(--ordryn-muted);">Admin</RouterLink>
            <RouterLink v-if="hasPermission('createinvites')" to="/invites" class="nav-link-item text-decoration-none small fw-medium" style="color: var(--ordryn-muted);">Create Invite</RouterLink>
          </template>
        </nav>
      </div>

      <!-- Right Actions: Multi-Theme Dropdown & Auth Controls -->
      <div class="d-flex align-items-center gap-2">
        <!-- Theme Picker Dropdown -->
        <div class="dropdown me-2">
          <button
            class="btn btn-sm dropdown-toggle d-flex align-items-center gap-1 border-0 shadow-none"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
            style="background: var(--ordryn-muted-bg); color: var(--ordryn-text);"
          >
            <i :class="availableThemes.find(t => t.id === currentTheme)?.icon || 'bi-palette'" />
            <span class="d-none d-sm-inline ms-1 capitalize">{{ currentTheme }}</span>
          </button>
          <ul class="dropdown-menu dropdown-menu-end shadow-sm border-0 mt-1">
            <li v-for="t in availableThemes" :key="t.id">
              <button
                class="dropdown-item d-flex align-items-center gap-2 small py-2"
                :class="{ active: currentTheme === t.id }"
                @click="setTheme(t.id)"
              >
                <i :class="t.icon" />
                <span>{{ t.label }}</span>
              </button>
            </li>
          </ul>
        </div>

        <template v-if="isAuthenticated">
          <span class="text-muted small d-none d-lg-inline me-1">{{ user?.user_name || user?.email }}</span>
          <RouterLink to="/settings" class="btn btn-outline-secondary btn-sm d-flex align-items-center gap-1" title="Profile Settings">
            <i class="bi bi-person-circle" />
            <span class="d-none d-sm-inline">Profile</span>
          </RouterLink>
          <button type="button" class="btn btn-outline-danger btn-sm d-flex align-items-center gap-1" @click="onLogout">
            <i class="bi bi-box-arrow-right" />
            <span class="d-none d-sm-inline">Logout</span>
          </button>
        </template>
        <template v-else>
          <RouterLink to="/register" class="btn btn-outline-secondary btn-sm">Sign Up</RouterLink>
          <RouterLink to="/login" class="btn btn-primary btn-sm ms-1">Login</RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>

<style scoped>
.nav-link-item:hover {
  color: var(--ordryn-text) !important;
}
.brand-logo-icon {
  width: 32px;
  height: 32px;
}
</style>
