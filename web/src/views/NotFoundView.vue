<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <div class="container mt-5 mb-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card text-center">
          <div class="card-body py-5">
            <h1 class="display-4 mb-2">404</h1>
            <p class="text-muted mb-4">This page isn’t available or doesn’t exist.</p>
            <RouterLink to="/" class="btn btn-primary">Home</RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
