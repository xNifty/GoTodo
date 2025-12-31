// Helper to build correct API URLs that work on both localhost and subpaths
window.apiPath = function (endpoint) {
  // Remove leading slash if present
  const path = endpoint.startsWith("/") ? endpoint.slice(1) : endpoint;
  // Use relative path with dot prefix so HTMX resolves it relative to current location
  return "./" + path;
};

document.addEventListener("DOMContentLoaded", () => {
  let sidebar = document.getElementById("sidebar");
  let openSidebarBtn = document.getElementById("openSidebar");
  let closeSidebarBtn = document.getElementById("closeSidebar");
  let description = document.getElementById("description");
  let charCount = document.getElementById("char-count");

  // Character counter for description
  if (description && charCount) {
    description.addEventListener("input", function () {
      const length = this.value.length;
      charCount.textContent = length;

      // Add visual feedback when approaching limit
      if (length > 90) {
        charCount.classList.add("text-warning");
      } else {
        charCount.classList.remove("text-warning");
      }
      if (length > 95) {
        charCount.classList.add("text-danger");
      } else {
        charCount.classList.remove("text-danger");
      }
    });
  }

  // Theme toggle functionality
  function initTheme() {
    const themeToggle = document.getElementById("theme-toggle");
    // Prefer an existing data-theme (may be set server-side), fall back to localStorage, then default to light
    const existing = document.documentElement.getAttribute("data-theme");
    const saved = existing || localStorage.getItem("theme") || "light";
    document.documentElement.setAttribute("data-theme", saved);
    if (saved === "dark") {
      if (themeToggle) themeToggle.classList.add("active");
    } else {
      if (themeToggle) themeToggle.classList.remove("active");
    }
  }

  function toggleTheme() {
    const currentTheme =
      document.documentElement.getAttribute("data-theme") || "light";
    const newTheme = currentTheme === "light" ? "dark" : "light";
    const themeToggle = document.getElementById("theme-toggle");

    document.documentElement.setAttribute("data-theme", newTheme);
    localStorage.setItem("theme", newTheme);

    // Persist to cookie so server-side rendering can pick it up as a fallback
    try {
      document.cookie =
        "theme=" + newTheme + "; path=/; max-age=31536000; SameSite=Lax";
    } catch (e) {
      // ignore
    }

    if (newTheme === "dark") {
      if (themeToggle) themeToggle.classList.add("active");
    } else {
      if (themeToggle) themeToggle.classList.remove("active");
    }
  }

  // Initialize theme on page load
  initTheme();

  // Set up theme toggle event listener
  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle) {
    themeToggle.addEventListener("click", toggleTheme);
  }

  function openSidebar() {
    sidebar = document.getElementById("sidebar"); // Refetch in case HTMX replaced this element
    if (sidebar) {
      sidebar.classList.add("active");
    }
  }

  function closeSidebar() {
    sidebar = document.getElementById("sidebar"); // Refetch in case HTMX replaced this element
    if (sidebar) {
      sidebar.classList.remove("active");
    }
  }

  function initializeSidebarEventListeners() {
    // Re-query buttons in case HTMX replaced the DOM inside #task-container
    const openBtn = document.getElementById("openSidebar");
    const closeBtn = document.getElementById("closeSidebar");

    if (openBtn) {
      openBtn.removeEventListener("click", openSidebar); // Prevent duplicate bindings
      openBtn.addEventListener("click", openSidebar);
    }

    if (closeBtn) {
      closeBtn.removeEventListener("click", closeSidebar); // Prevent duplicate bindings
      closeBtn.addEventListener("click", closeSidebar);
    }

    // Reattach theme toggle if needed
    const themeToggle = document.getElementById("theme-toggle");
    if (themeToggle) {
      themeToggle.removeEventListener("click", toggleTheme);
      themeToggle.addEventListener("click", toggleTheme);
    }
  }

  // Attach initial event listeners
  initializeSidebarEventListeners();

  // Reattach event listeners after HTMX swaps
  document.body.addEventListener("htmx:afterSettle", (event) => {
    if (event.target.id === "task-container") {
      initializeSidebarEventListeners();

      // Reapply the active class if necessary
      sidebar = document.getElementById("sidebar"); // Refetch updated sidebar
      if (sidebar && sidebar.classList.contains("htmx-added")) {
        sidebar.classList.add("active");
      }
    }
  });

  // Initialize Sortable on favorite and regular task lists to allow drag-and-drop
  function initSortable() {
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
              '#task-container [name="currentPage"]'
            );
            if (pageEl && pageEl.value) form.append("page", pageEl.value);

            fetch("/api/reorder-tasks", { method: "POST", body: form })
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

  // Ensure table element inside #task-container has expected classes
  function ensureTableClasses() {
    const container = document.getElementById("task-container");
    if (!container) return;
    const table = container.querySelector("table");
    if (!table) return;
    const classes = [
      "table",
      "table-striped",
      "table-bordered",
      "w-100",
      "mb-3",
    ];
    classes.forEach((c) => {
      if (!table.classList.contains(c)) table.classList.add(c);
    });
  }

  // If HTMX returned rows or tbody elements without a surrounding table, wrap them
  function ensureTableStructure() {
    const container = document.getElementById("task-container");
    if (!container) return;
    // If a table already exists, nothing to do
    if (container.querySelector("table")) return;

    // Look for tbody elements or lists that should be wrapped
    const fav = container.querySelector("#favorite-task-list");
    const reg = container.querySelector("#task-list");

    const nodesToWrap = [];
    if (fav) nodesToWrap.push(fav);
    if (reg) nodesToWrap.push(reg);

    // If there are no known tbody containers, look for direct <tr> children
    if (nodesToWrap.length === 0) {
      const trs = Array.from(container.querySelectorAll(":scope > tr"));
      if (trs.length) {
        nodesToWrap.push(...trs);
      }
    }

    if (nodesToWrap.length === 0) return;

    // Build a table structure and move the nodes into it
    const table = document.createElement("table");
    table.className = "table table-striped table-bordered w-100 mb-3";

    // If there's a thead elsewhere in the container, move it into the table
    const thead = container.querySelector("thead");
    if (thead) {
      table.appendChild(thead.cloneNode(true));
      try {
        thead.remove();
      } catch (e) {}
    }

    // Append each node into a tbody. If node is already a tbody, append directly.
    const tbody = document.createElement("tbody");
    nodesToWrap.forEach((n) => {
      tbody.appendChild(n);
    });
    table.appendChild(tbody);

    // Insert the table at the top of the container
    container.insertBefore(table, container.firstChild);
  }

  // Optional: Close sidebar after form submission
  const taskForm = document.getElementById("newTaskForm");
  if (taskForm) {
    taskForm.addEventListener("htmx:afterRequest", (event) => {
      // Only close sidebar if the request was successful and not a validation error
      // event.detail.successful will be true for a 200 status code response, even with HX-Trigger/HX-Retarget
      // Check for a validation header set by the server (preferred) or fall back to triggerSpec for compatibility
      let isValidationError = false;
      try {
        const xhr = event.detail && event.detail.xhr;
        const header =
          xhr && xhr.getResponseHeader
            ? xhr.getResponseHeader("X-Validation-Error")
            : null;
        if (header && header.toLowerCase() === "true") {
          isValidationError = true;
        } else if (
          event.detail &&
          event.detail.triggerSpec &&
          event.detail.triggerSpec.trigger === "description-error"
        ) {
          // Backwards compat: if older handlers used triggerSpec, respect that too
          isValidationError = true;
        }
      } catch (e) {
        // ignore and treat as not a validation error
      }

      if (event.detail.successful && !isValidationError) {
        closeSidebar();
        // Clear the form fields on successful submission
        const tEl = document.getElementById("title");
        if (tEl) tEl.value = "";
        const dEl = document.getElementById("description");
        if (dEl) dEl.value = "";
        // Reset character counter
        let charCount = document.getElementById("char-count");
        if (charCount) charCount.textContent = "0";
        // Clear any old validation message
        let errorDiv = document.getElementById("description-error");
        if (errorDiv) errorDiv.innerHTML = "";
      } else if (isValidationError) {
        // Keep the sidebar open and show the error (HTMX will swap the error message into #description-error due to HX-Retarget)
        // The form fields and char counter are retained automatically by the browser
        return; // Stop further processing so sidebar remains open
      }
      // Note: For network errors (non-2xx status), event.detail.successful will be false,
      // and this handler will not close the sidebar, which is the desired behavior.
    });
  }

  const modalElement = document.getElementById("modal");

  if (modalElement) {
    modalElement.addEventListener("hide.bs.modal", () => {
      if (document.activeElement instanceof HTMLElement) {
        document.activeElement.blur();
      }
    });

    // Optional: Set aria-hidden to true when the modal is hidden (if necessary)
    modalElement.addEventListener("hidden.bs.modal", () => {
      modalElement.setAttribute("aria-hidden", "true");
    });
  }

  // Handle task deletion
  document.body.addEventListener("taskDeleted", function (evt) {
    // Get the current page from the page number display
    // Find the pagination span that contains "Page X of Y"
    const spans = document.querySelectorAll("#task-container span");
    let currentPage = 1;
    for (let span of spans) {
      const match = span.textContent.match(/Page\s+(\d+)\s+of\s+(\d+)/);
      if (match) {
        currentPage = parseInt(match[1]);
        break;
      }
    }

    // Reload the current page
    let url = `/api/fetch-tasks?page=${currentPage}`;
    const searchInput = document.getElementById("search");
    if (searchInput && searchInput.value) {
      url += `&search=${encodeURIComponent(searchInput.value)}`;
    }
    htmx.ajax("GET", url, { target: "#task-container", swap: "innerHTML" });
  });

  // Handle reload with specific page number
  document.body.addEventListener("reloadPage", function (evt) {
    const page = evt.detail.page || 1;
    let url = `/api/fetch-tasks?page=${page}`;
    const searchInput = document.getElementById("search");
    if (searchInput && searchInput.value) {
      url += `&search=${encodeURIComponent(searchInput.value)}`;
    }
    htmx.ajax("GET", url, { target: "#task-container", swap: "innerHTML" });
  });

  document.body.addEventListener("reload-previous-page", function (evt) {
    // Get the current page from the page number display
    // Find the pagination span that contains "Page X of Y"
    const spans = document.querySelectorAll("#task-container span");
    let currentPage = 1;
    for (let span of spans) {
      const match = span.textContent.match(/Page\s+(\d+)\s+of\s+(\d+)/);
      if (match) {
        currentPage = parseInt(match[1]);
        break;
      }
    }
    const prevPage = Math.max(currentPage - 1, 1);

    // Optionally, preserve search query if present
    const searchInput = document.getElementById("search");
    const searchQuery = searchInput ? searchInput.value : "";

    // Build the URL for the previous page
    let url = `/api/fetch-tasks?page=${prevPage}`;
    if (searchQuery) {
      url += `&search=${encodeURIComponent(searchQuery)}`;
    }

    // Use HTMX to load the previous page into the task container
    htmx.ajax("GET", url, { target: "#task-container", swap: "innerHTML" });
  });

  // Handle login success
  document.body.addEventListener("login-success", function (evt) {
    // Close the login modal
    const modal = bootstrap.Modal.getInstance(document.getElementById("modal"));
    if (modal) {
      modal.hide();
    }

    // Optionally reload the page to show logged-in state
    window.location.reload();
  });

  // Update completed/incomplete badges when server notifies via HX-Trigger
  document.body.addEventListener("taskCountsChanged", function (evt) {
    try {
      const d = evt.detail || {};
      if (typeof d.completed !== "undefined") {
        const el = document.getElementById("completed-tasks-badge");
        if (el) el.textContent = `Completed: ${d.completed}`;
      }
      if (typeof d.incomplete !== "undefined") {
        const el2 = document.getElementById("incomplete-tasks-badge");
        if (el2) el2.textContent = `Incomplete: ${d.incomplete}`;
      }
      // update total if both present
      if (
        typeof d.completed !== "undefined" &&
        typeof d.incomplete !== "undefined"
      ) {
        const totalEl = document.getElementById("total-tasks-badge");
        if (totalEl)
          totalEl.textContent = `Total Tasks: ${d.completed + d.incomplete}`;
      }
    } catch (e) {
      // ignore
    }
  });

  // Toast helper: create container and show transient toasts matching theme
  function ensureToastContainer() {
    let c = document.querySelector(".app-toast-container");
    if (!c) {
      c = document.createElement("div");
      c.className = "app-toast-container";
      document.body.appendChild(c);
    }
    return c;
  }

  function showToast(message, opts) {
    opts = opts || {};
    const container = ensureToastContainer();
    const t = document.createElement("div");
    t.className = "app-toast" + (opts.error ? " app-toast--error" : "");
    t.setAttribute("role", "status");
    t.setAttribute("aria-live", "polite");
    t.textContent = message;
    container.appendChild(t);

    // ensure next frame for animation
    requestAnimationFrame(() => {
      t.classList.add("show");
    });

    const timeout = typeof opts.duration === "number" ? opts.duration : 3500;
    const remove = () => {
      t.classList.remove("show");
      setTimeout(() => {
        try {
          t.remove();
        } catch (e) {}
      }, 220);
    };

    // Auto-remove
    const to = setTimeout(remove, timeout);

    // Allow manual dismissal on click
    t.addEventListener("click", function () {
      clearTimeout(to);
      remove();
    });
  }

  // Listen for favorite-limit-reached HTMX trigger (server sets HX-Trigger header)
  document.body.addEventListener("favorite-limit-reached", function (evt) {
    try {
      // If server provided a detail.message, prefer that, otherwise default text
      const msg =
        (evt && evt.detail && evt.detail.message) ||
        "You can only favorite up to 5 tasks";
      showToast(msg, { error: true });
    } catch (e) {
      showToast("You can only favorite up to 5 tasks", { error: true });
    }
  });

  // Note: Logout now uses HX-Redirect in the handler, so no event listener needed

  // Re-initialize character counter and theme toggle after HTMX swaps if sidebar is active
  document.body.addEventListener("htmx:afterSwap", (event) => {
    // Check if the sidebar element exists and is currently active
    const sidebarElement = document.getElementById("sidebar");
    if (sidebarElement && sidebarElement.classList.contains("active")) {
      // Re-initialize character counter if elements are present
      let description = document.getElementById("description");
      let charCount = document.getElementById("char-count");
      if (description && charCount) {
        // Ensure we don't add duplicate listeners
        // You might need a more robust way to handle this if HTMX swaps parts of the sidebar body frequently.
        // For now, re-setting innerHTML for #description-error means the textarea and char-count span themselves are usually not replaced unless the whole form is swapped.
        // Let's re-attach the listener defensively.

        // Remove any existing listener before adding a new one to prevent duplicates
        // (Requires storing the listener function reference if we want to remove specifically, or rely on HTMX replacing element)
        // Since HTMX often replaces the element, a simpler approach is to just re-add the listener.
        // If the textarea element itself is *not* replaced by the swap (only content around it),
        // you might get duplicate listeners. However, given the hx-swap="innerHTML" on the form,
        // the textarea and char-count span should be new elements, making this approach okay.

        charCount.textContent = description.value.length; // Initialize count

        // It's safer to check if a listener already exists or if the element was replaced.
        // A simpler way for now is to rely on the element being replaced on form swap.
        // For robustness, consider using htmx.onLoad or a mutation observer if needed.

        // For now, re-attach the input listener assuming the element might be new.
        // This listener might be added multiple times if the description element isn't fully replaced,
        // but HTMX swap="innerHTML" on the form usually replaces the whole form content.
        // Let's add a check to see if the element has a marker class indicating listener already added.
        if (!description.classList.contains("char-count-listener-added")) {
          description.addEventListener("input", function () {
            const length = this.value.length;
            charCount.textContent = length;
            // Add visual feedback when approaching limit
            if (length > 90) {
              charCount.classList.add("text-warning");
            } else {
              charCount.classList.remove("text-warning");
            }
            if (length > 95) {
              charCount.classList.add("text-danger");
            } else {
              charCount.classList.remove("text-danger");
            }
          });
          description.classList.add("char-count-listener-added"); // Mark as having listener
        }
      }
      // Re-initialize theme toggle if needed
      if (typeof initTheme === "function") {
        initTheme(); // This function should be idempotent or handle re-running safely
      }
    }
  });
});
