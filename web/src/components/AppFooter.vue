<script setup lang="ts">
import { RouterLink } from "vue-router";
</script>

<template>
  <footer
    class="text-center py-3 app-footer mt-auto border-top border-opacity-10 w-100"
  >
    <p class="text-muted small mb-0">
      Made with <i class="bi bi-heart-fill text-danger" /> by xNifty |
      <a
        href="https://github.com/SentientTD-Studios/Ordryn"
        target="_blank"
        rel="noopener noreferrer"
        class="text-decoration-none text-muted"
        >Source</a
      >
      |
      <RouterLink to="/docs/guide" class="text-decoration-none text-muted"
        >Documentation</RouterLink
      >
      |
      <RouterLink to="/docs/api/v1" class="text-decoration-none text-muted"
        >API</RouterLink
      >
      |
      <a
        href="https://github.com/SentientTD-Studios/Ordryn/issues"
        target="_blank"
        rel="noopener noreferrer"
        class="text-decoration-none text-muted"
        >Report a Bug</a
      >
      |
      <a
        href="https://ordryn.com"
        target="_blank"
        rel="noopener noreferrer"
        class="text-decoration-none text-muted"
        >Powered by Ordryn © 2026</a
      >
    </p>
  </footer>
</template>
