<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '@/api/client'
import { APIError } from '@/api/types'
import { useSite } from '@/composables/useSite'

const { siteName, enableJoinRequests, openJoin, loaded } = useSite()

const email = ref('')
const message = ref('')
const busy = ref(false)
const error = ref('')
const submitted = ref(false)

const remaining = computed(() => 500 - message.value.length)

async function onSubmit() {
  busy.value = true
  error.value = ''
  try {
    await api.createJoinRequest(email.value.trim(), message.value.trim())
    submitted.value = true
  } catch (err) {
    error.value = err instanceof APIError ? err.message : 'Could not submit request'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card">
          <div class="card-body">
            <h1 class="card-title">Request to Join</h1>

            <template v-if="loaded && !enableJoinRequests">
              <p class="text-muted">{{ siteName }} is not accepting join requests right now.</p>
              <p class="mb-0">
                <RouterLink v-if="openJoin" to="/register">Join today</RouterLink>
                <template v-else>
                  <RouterLink to="/login">Sign in</RouterLink>
                  if you already have an account.
                </template>
              </p>
            </template>

            <template v-else-if="submitted">
              <p class="mb-0">
                If this site accepts join requests, we will be in touch. You can
                <RouterLink to="/login">sign in</RouterLink>
                if you already have an account.
              </p>
            </template>

            <form v-else @submit.prevent="onSubmit">
              <p class="text-muted">Ask to join {{ siteName }}. An administrator will review your request.</p>
              <div class="mb-3">
                <label for="join-email" class="form-label">Email</label>
                <input
                  id="join-email"
                  v-model="email"
                  type="email"
                  class="form-control"
                  required
                  autocomplete="email"
                />
              </div>
              <div class="mb-3">
                <label for="join-message" class="form-label">Message <span class="text-muted">(optional)</span></label>
                <textarea
                  id="join-message"
                  v-model="message"
                  class="form-control"
                  rows="4"
                  maxlength="500"
                />
                <div class="form-text">{{ remaining }} characters left</div>
              </div>
              <div v-if="error" class="text-danger mb-3">{{ error }}</div>
              <button type="submit" class="btn btn-primary" :disabled="busy || !loaded">
                {{ busy ? 'Sending…' : 'Submit request' }}
              </button>
            </form>

            <p class="mt-3 mb-0 text-muted small">
              Already registered?
              <RouterLink to="/login">Sign in</RouterLink>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
