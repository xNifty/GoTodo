import {
  apiPath,
  restoreFooterIfMissing,
  ensureTableClasses,
  ensureTableStructure,
} from "./utils.js";
import { initializeSidebarEventListeners } from "./sidebar.js";
import { initializeModalEventListeners } from "./modal.js";

export function initSortable() {
  try {
    if (typeof Sortable === "undefined") return;

    const favList = document.getElementById("favorite-task-list");
    const regList = document.getElementById("task-list");

    const createSortable = (el, isFav) => {
      if (!el) return;
      // Destroy existing Sortable instance if present
      if (el._sortable) {
        try {
          el._sortable.destroy();
        } catch (e) {}
      }
      el._sortable = Sortable.create(el, {
        handle: ".drag-handle",
        animation: 150,
        onEnd: function (evt) {
          // Build order of ids
          const ids = Array.from(evt.to.children)
            .map((row) => {
              const id = row.id || "";
              return id.replace("task-", "");
            })
            .filter(Boolean)
            .join(",");

          // Post new order to server
          const form = new URLSearchParams();
          form.append("order", ids);
          form.append("is_favorite", isFav ? "true" : "false");
          // include current page if present
          const pageEl = document.querySelector(
            '#task-container [name="currentPage"]',
          );
          if (pageEl && pageEl.value) form.append("page", pageEl.value);
          // include current toolbar project filter so server can respect scoped reorder
          try {
            const toolbar = document.querySelector("select#project-filter");
            const toolbarVal = toolbar ? toolbar.value : "";
            if (typeof toolbarVal !== "undefined" && toolbarVal !== null) {
              form.append("project", toolbarVal);
            }
          } catch (e) {}

          fetch(apiPath("/api/reorder-tasks"), { method: "POST", body: form })
            .then((resp) => {
              if (resp.ok) return resp.text();
              throw new Error("Failed to save order");
            })
            .then((html) => {
              // Replace task container with returned HTML
              const container = document.getElementById("task-container");
              if (container) {
                container.innerHTML = html;
                // Let HTMX process any hx-* attributes in the newly inserted content
                try {
                  if (typeof htmx !== "undefined") htmx.process(container);
                } catch (e) {}
                // Reinitialize sortable after DOM update
                try {
                  initSortable();
                } catch (e) {}
                // Reattach sidebar and modal listeners which may have been lost
                try {
                  if (typeof initializeSidebarEventListeners === "function") {
                    initializeSidebarEventListeners();
                  }
                } catch (e) {}
                try {
                  if (typeof initializeModalEventListeners === "function") {
                    initializeModalEventListeners();
                  }
                } catch (e) {}
                // Ensure footer still exists after manual replacement
                try {
                  restoreFooterIfMissing();
                } catch (e) {}
              }
            })
            .catch((err) => {
              console.error("Reorder failed", err);
            });
        },
      });
    };

    createSortable(favList, true);
    createSortable(regList, false);
  } catch (e) {
    // ignore
  }
}

export function attachSortableInitializers() {
  // Initialize sortable on initial load and after HTMX swaps
  initSortable();
  document.body.addEventListener("htmx:afterSwap", function (evt) {
    if (evt.target.id === "task-container") {
      // Ensure table retains expected Bootstrap classes after HTMX replaces content
      try {
        ensureTableStructure();
        ensureTableClasses();
      } catch (e) {}

      initSortable();
    }
  });
}
