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

  // Save original footer HTML so we can restore it if an HTMX swap accidentally removes it
  let __originalFooterHTML = null;
  (function captureFooter() {
    try {
      const f = document.querySelector("footer");
      if (f) __originalFooterHTML = f.outerHTML;
    } catch (e) {}
  })();

  function restoreFooterIfMissing() {
    try {
      if (!document.querySelector("footer") && __originalFooterHTML) {
        // Insert footer after the main container if present, otherwise append to body
        const container = document.getElementById("task-container");
        if (container && container.parentNode) {
          // Create a temporary node and insert
          const temp = document.createElement("div");
          temp.innerHTML = __originalFooterHTML;
          // append after container
          if (container.nextSibling) {
            container.parentNode.insertBefore(
              temp.firstElementChild,
              container.nextSibling,
            );
          } else {
            container.parentNode.appendChild(temp.firstElementChild);
          }
        } else {
          const temp = document.createElement("div");
          temp.innerHTML = __originalFooterHTML;
          document.body.appendChild(temp.firstElementChild);
        }
      }
    } catch (e) {
      // ignore
    }
  }
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

  // Character counter for project name
  const projectNameInput = document.getElementById("project-name");
  const projectNameCharCount = document.getElementById(
    "project-name-char-count",
  );
  if (projectNameInput && projectNameCharCount) {
    projectNameInput.addEventListener("input", function () {
      const length = this.value.length;
      projectNameCharCount.textContent = length;

      // Add visual feedback when approaching limit
      if (length > 40) {
        projectNameCharCount.classList.add("text-warning");
      } else {
        projectNameCharCount.classList.remove("text-warning");
      }
      if (length > 45) {
        projectNameCharCount.classList.add("text-danger");
      } else {
        projectNameCharCount.classList.remove("text-danger");
      }
    });
  }

  // Handle project form validation errors and clear error on input
  function initializeProjectFormHandlers() {
    const projectForm = document.getElementById("createProjectForm");
    const projectNameInput = document.getElementById("project-name");
    const projectNameError = document.getElementById("project-name-error");

    if (projectForm && projectNameInput && projectNameError) {
      // Clear error when user starts typing
      projectNameInput.addEventListener("input", function () {
        projectNameError.innerHTML = "";
      });

      // Handle validation errors from server
      projectForm.addEventListener("htmx:afterRequest", function (event) {
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
            event.detail.triggerSpec.trigger === "project-name-error"
          ) {
            isValidationError = true;
          }
        } catch (e) {}

        // Clear form on successful submission (not validation error)
        if (event.detail.successful && !isValidationError) {
          projectNameInput.value = "";
          const charCount = document.getElementById("project-name-char-count");
          if (charCount) charCount.textContent = "0";
          projectNameError.innerHTML = "";
        }
      });
    }
  }

  // Initialize project form handlers
  initializeProjectFormHandlers();

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
      openBtn.addEventListener("click", function (ev) {
        try {
          const tf = document.getElementById("newTaskForm");
          if (tf) {
            const titleEl = tf.querySelector("#title");
            if (titleEl) titleEl.value = "";
            const descEl = tf.querySelector("#description");
            if (descEl) descEl.value = "";
            const projEl = tf.querySelector("#project_id");
            if (projEl) projEl.value = "";
            const dueEl = tf.querySelector("#due_date");
            if (dueEl) dueEl.value = "";
            const idInput = tf.querySelector('input[name="id"]');
            if (idInput) idInput.remove();
            const submit = tf.querySelector('button[type="submit"]');
            if (submit) submit.textContent = "Add Task";
            // Ensure the form posts to the add endpoint
            try {
              tf.setAttribute("hx-post", apiPath("/api/add-task"));
            } catch (e) {}
            const cp = tf.querySelector('input[name="currentPage"]');
            if (cp) cp.value = "1";
            // Ensure the form carries the current toolbar project filter so server can decide refresh
            try {
              let projField = tf.querySelector('input[name="project"]');
              const toolbar = document.querySelector("select#project-filter");
              const toolbarVal = toolbar ? toolbar.value : "";
              if (!projField) {
                projField = document.createElement("input");
                projField.type = "hidden";
                projField.name = "project";
                tf.appendChild(projField);
              }
              projField.value = toolbarVal;
            } catch (e) {}
            const sbTitle = document.querySelector(
              "#sidebar .sidebar-header h5",
            );
            if (sbTitle) sbTitle.textContent = "Add Task";
            const charCount = document.getElementById("char-count");
            if (charCount) charCount.textContent = "0";
          }
        } catch (e) {}
        openSidebar();
      });
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

    // Reattach task form submit listener so dynamically swapped forms behave the same
    try {
      const tf = document.getElementById("newTaskForm");
      if (tf && !tf.classList.contains("task-form-initialized")) {
        // Ensure hidden project field exists and is kept up-to-date before submit
        try {
          let projField = tf.querySelector('input[name="project"]');
          const toolbar = document.querySelector("select#project-filter");
          const toolbarVal = toolbar ? toolbar.value : "";
          if (!projField) {
            projField = document.createElement("input");
            projField.type = "hidden";
            projField.name = "project";
            tf.appendChild(projField);
          }
          projField.value = toolbarVal;
          // Update it on submit in case toolbar changed while form open
          tf.addEventListener("submit", function () {
            try {
              const tb = document.querySelector("select#project-filter");
              if (tb) projField.value = tb.value;
            } catch (e) {}
          });
        } catch (e) {}
        tf.addEventListener("htmx:afterRequest", (event) => {
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
              isValidationError = true;
            }
          } catch (e) {}

          if (event.detail.successful && !isValidationError) {
            closeSidebar();
            const tEl = document.getElementById("title");
            if (tEl) tEl.value = "";
            const dEl = document.getElementById("description");
            if (dEl) dEl.value = "";
            const charCount = document.getElementById("char-count");
            if (charCount) charCount.textContent = "0";
            const errorDiv = document.getElementById("description-error");
            if (errorDiv) errorDiv.innerHTML = "";
          }
        });
        tf.classList.add("task-form-initialized");
      }
    } catch (e) {}
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

  // Modal initialization helper - re-attach listeners after HTMX swaps
  function initializeModalEventListeners() {
    const modalEl = document.getElementById("modal");
    if (!modalEl) return;
    if (!modalEl.classList.contains("modal-listeners-initialized")) {
      modalEl.addEventListener("hide.bs.modal", () => {
        if (document.activeElement instanceof HTMLElement) {
          document.activeElement.blur();
        }
      });
      modalEl.addEventListener("hidden.bs.modal", () => {
        modalEl.setAttribute("aria-hidden", "true");
      });
      modalEl.classList.add("modal-listeners-initialized");
    }
    // ensure bootstrap modal instance exists so data-bs-dismiss works
    try {
      if (
        typeof bootstrap !== "undefined" &&
        bootstrap.Modal &&
        typeof bootstrap.Modal.getOrCreateInstance === "function"
      ) {
        bootstrap.Modal.getOrCreateInstance(modalEl);
      }
    } catch (e) {}
  }

  // Call once on initial load
  initializeModalEventListeners();

  // Changelog modal: fetch and render entries when opened
  function renderChangelog(entries) {
    const container = document.getElementById("changelog-body");
    if (!container) return;
    if (!entries || !entries.length) {
      container.innerHTML =
        '<div class="text-center text-muted">No changelog entries available.</div>';
      return;
    }
    // Build HTML with collapsible entries (collapsed by default).
    const out = document.createElement("div");
    out.className = "changelog-list";
    const MAX_MODAL = 5;
    const recent = entries.slice(0, MAX_MODAL);
    recent.forEach((e, idx) => {
      const card = document.createElement("div");
      card.className = "card mb-3";
      const cardBody = document.createElement("div");
      cardBody.className = "card-body";

      // Header button with caret and badge (shows version, title, tag, date)
      const headerBtn = document.createElement("button");
      headerBtn.type = "button";
      headerBtn.className =
        "btn btn-link text-start w-100 p-0 d-flex align-items-center";
      headerBtn.style.textDecoration = "none";

      const arrow = document.createElement("span");
      arrow.className = "chev me-2";
      arrow.textContent = "►";
      headerBtn.appendChild(arrow);

      const titleWrap = document.createElement("div");
      titleWrap.className = "flex-grow-1 text-start";
      const strong = document.createElement("strong");
      strong.textContent = e.title || "";
      titleWrap.appendChild(strong);

      const span = document.createElement("span");
      span.className =
        "badge releasetag ms-3 " +
        (e.prerelease ? "bg-warning text-dark" : "bg-success");
      span.textContent =
        (e.prerelease ? "Prerelease" : "Release") + " • " + (e.date || "");
      titleWrap.appendChild(span);

      headerBtn.appendChild(titleWrap);
      cardBody.appendChild(headerBtn);

      const collapseDiv = document.createElement("div");
      collapseDiv.id = `changelog-modal-${idx}-${Date.now()}`;
      collapseDiv.className = "collapse mt-2";

      // Body content
      if (e.html) {
        const bodyDiv = document.createElement("div");
        bodyDiv.className = "changelog-entry-body";
        bodyDiv.innerHTML = e.html;

        // If the rendered HTML includes a heading with an anchor/permalink,
        // shorten that anchor's visible text to only the release title so the
        // clickable area doesn't show the entire breadcrumb.
        try {
          const heading = bodyDiv.querySelector("h1,h2,h3,h4,h5,h6");
          if (heading) {
            const anchor = heading.querySelector("a");
            if (anchor) {
              anchor.textContent =
                e.title ||
                anchor.getAttribute("title") ||
                anchor.textContent ||
                "";
            }
          }
        } catch (err) {}

        // Remove any leading paragraph/div that duplicates the version/date
        const first = bodyDiv.firstElementChild;
        try {
          if (
            first &&
            (first.tagName === "P" ||
              first.tagName === "DIV" ||
              first.tagName === "PRE")
          ) {
            const txt = (first.textContent || "").trim().toLowerCase();
            const v = (e.version || "").toLowerCase();
            const d = (e.date || "").toLowerCase();
            if (
              (v && txt.includes(v)) ||
              (d && txt.includes(d)) ||
              txt.includes(" - ")
            ) {
              first.remove();
            }
          }
        } catch (err) {}

        collapseDiv.appendChild(bodyDiv);
      } else {
        const ul = document.createElement("ul");
        if (Array.isArray(e.notes)) {
          e.notes.forEach((n) => {
            const li = document.createElement("li");
            li.textContent = n;
            ul.appendChild(li);
          });
        }
        collapseDiv.appendChild(ul);
      }

      // Toggle behavior: open/close collapse and swap arrow
      headerBtn.addEventListener("click", function (ev) {
        ev.preventDefault();
        const opened = collapseDiv.classList.toggle("show");
        arrow.textContent = opened ? "▼" : "►";
      });

      cardBody.appendChild(collapseDiv);
      card.appendChild(cardBody);
      out.appendChild(card);
    });

    // If there are more entries, add a link to view the full changelog page
    if (entries.length > recent.length) {
      const more = document.createElement("div");
      more.className = "text-center mt-3";
      const a = document.createElement("a");
      a.href = apiPath("/changelog/page");
      //a.target = "_blank";
      a.textContent = "View full changelog";
      more.appendChild(a);
      out.appendChild(more);
    }
    container.innerHTML = "";
    container.appendChild(out);
  }

  function loadChangelog() {
    const container = document.getElementById("changelog-body");
    if (container)
      container.innerHTML =
        '<div class="text-center text-muted">Loading...</div>';
    fetch(apiPath("/changelog"))
      .then((res) => {
        if (!res.ok) throw new Error("failed to load changelog");
        return res.json();
      })
      .then((data) => {
        renderChangelog(data);
      })
      .catch((err) => {
        const container = document.getElementById("changelog-body");
        if (container)
          container.innerHTML =
            '<div class="text-danger">Unable to load changelog.</div>';
        console.error(err);
      });
  }

  const changelogModalEl = document.getElementById("changelogModal");
  if (changelogModalEl) {
    try {
      changelogModalEl.addEventListener("show.bs.modal", loadChangelog);
    } catch (e) {
      // If bootstrap isn't present or event fails, attempt to load on click of link
      const link = document.querySelector('[data-bs-target="#changelogModal"]');
      if (link) link.addEventListener("click", loadChangelog);
    }
  }

  // Debug helper: when ?cssdebug=1 is present in the URL, log which media queries match.
  (function cssDebugHelper() {
    try {
      const params = new URLSearchParams(window.location.search);
      if (!params.get("cssdebug")) return;

      const queries = {
        "max-420": "(max-width: 420px)",
        "max-600": "(max-width: 600px)",
        "max-768": "(max-width: 768px)",
        "max-1024": "(max-width: 1024px)",
        "pointer-coarse": "(pointer: coarse)",
        "hover-none": "(hover: none)",
      };

      console.groupCollapsed("CSS Debug — media query matches");
      Object.entries(queries).forEach(([k, q]) => {
        try {
          const m = window.matchMedia(q);
          console.log(q + ":", m.matches);
        } catch (e) {
          console.log(q + ": error");
        }
      });
      // Also log touch-capability and maxTouchPoints
      console.log("navigator.maxTouchPoints:", navigator.maxTouchPoints);
      console.log("ontouchstart in window:", "ontouchstart" in window);
      console.groupEnd();
    } catch (e) {}
  })();

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

  // Mobile accordion behavior removed: task titles no longer toggle rows.
  // Previous click handler intentionally removed to make all task details visible on mobile.

  // When an edit button is clicked, open the sidebar immediately so the user sees the form loading
  document.body.addEventListener("click", function (e) {
    try {
      const btn = e.target && e.target.closest && e.target.closest(".edit-btn");
      if (!btn) return;
      const sb = document.getElementById("sidebar");
      if (sb) sb.classList.add("active");
    } catch (e) {}
  });

  // Delegated close button handler: works even if the sidebar markup was swapped
  document.body.addEventListener("click", function (e) {
    try {
      const close =
        e.target && e.target.closest && e.target.closest("#closeSidebar");
      if (!close) return;
      closeSidebar();
    } catch (e) {}
  });

  // Keyboard activation for task-toggle no longer toggles expansion;
  // task rows are always expanded on mobile so no JS toggle is needed.

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

  // If HTMX swapped the sidebar, ensure it's opened and listeners attached
  document.body.addEventListener("htmx:afterSwap", function (evt) {
    try {
      const target =
        evt.detail && evt.detail.target ? evt.detail.target : evt.target;
      if (target && target.id === "sidebar") {
        try {
          initializeSidebarEventListeners();
        } catch (e) {}
        const sb = document.getElementById("sidebar");
        if (sb) sb.classList.add("active");
      }
    } catch (e) {}
  });

  // Also handle cases where we replace only the sidebar body via innerHTML
  document.body.addEventListener("htmx:afterSwap", function (evt) {
    try {
      const detail = evt && evt.detail;
      // If the swapped target is the sidebar body (selector used by edit button), open the sidebar
      const swapped = detail && detail.target ? detail.target : evt.target;
      if (swapped && swapped.id === "sidebar") return; // handled above
      // when swapping innerHTML into '#sidebar .sidebar-body', the event target will be that element
      if (
        swapped &&
        swapped.classList &&
        swapped.classList.contains("sidebar-body")
      ) {
        try {
          initializeSidebarEventListeners();
        } catch (e) {}
        const sb = document.getElementById("sidebar");
        if (sb) sb.classList.add("active");
      }
    } catch (e) {}
  });

  // Additionally, respond to edit requests specifically: if an HTMX request to /api/edit succeeded,
  // open the sidebar (covers cases where server returns a fragment without a clear swapped target)
  document.body.addEventListener("htmx:afterRequest", function (evt) {
    try {
      const xhr = evt && evt.detail && evt.detail.xhr;
      if (!xhr || !xhr.responseURL) return;
      if (xhr.responseURL.includes("/api/edit")) {
        // Only open on success (2xx)
        const status = xhr.status || 0;
        if (status >= 200 && status < 300) {
          try {
            initializeSidebarEventListeners();
          } catch (e) {}
          const sb = document.getElementById("sidebar");
          if (sb) sb.classList.add("active");
        }
      }
      // Clear create-project form after successful HTMX create
      if (xhr.responseURL.includes("/api/projects/create")) {
        const status = xhr.status || 0;
        // Check if this is a validation error
        const isValidationError =
          xhr.getResponseHeader &&
          xhr.getResponseHeader("X-Validation-Error") === "true";

        if (status >= 200 && status < 300 && !isValidationError) {
          try {
            const form = document.getElementById("createProjectForm");
            if (form) {
              const nameInput = form.querySelector('input[name="name"]');
              if (nameInput) nameInput.value = "";
              const charCount = document.getElementById(
                "project-name-char-count",
              );
              if (charCount) charCount.textContent = "0";
              const errorDiv = document.getElementById("project-name-error");
              if (errorDiv) errorDiv.innerHTML = "";
            }
            // Reinitialize project form handlers
            initializeProjectFormHandlers();
            if (typeof showToast === "function") showToast("Project created.");
          } catch (e) {}
        }
      }

      // When server notifies projects changed, refresh project selects
      try {
        if (
          xhr &&
          xhr.getResponseHeader &&
          xhr.getResponseHeader("HX-Trigger")
        ) {
          const trig = xhr.getResponseHeader("HX-Trigger");
          if (trig && trig.indexOf("projects-changed") !== -1) {
            fetch(apiPath("/api/projects/json"))
              .then((res) => res.json())
              .then((data) => {
                try {
                  // Update all selects with id project_id
                  const selects =
                    document.querySelectorAll("select#project_id");
                  selects.forEach((sel) => {
                    // preserve current value
                    const cur = sel.value;
                    // clear existing options
                    while (sel.options.length > 1) sel.remove(1);
                    data.forEach((p) => {
                      const opt = document.createElement("option");
                      opt.value = p.id;
                      opt.textContent = p.name;
                      sel.appendChild(opt);
                    });
                    // restore value if still present
                    try {
                      sel.value = cur;
                    } catch (e) {}
                  });
                } catch (e) {}
              })
              .catch(() => {});
          }
          // If server asked to reset the toolbar project filter, do that too
          if (trig && trig.indexOf("reset-project-filter") !== -1) {
            try {
              const pf = document.querySelector("select#project-filter");
              if (pf) {
                pf.value = "";
                // dispatch change so HTMX will fetch the full task list
                pf.dispatchEvent(new Event("change", { bubbles: true }));
              }
            } catch (e) {}
          }
          // If server requested setting the toolbar project filter, apply it
          if (trig && trig.indexOf("set-project-filter") !== -1) {
            try {
              const m = trig.match(/set-project-filter:([^\s]+)/);
              if (m && m[1] !== undefined) {
                const val = m[1];
                const pf = document.querySelector("select#project-filter");
                if (pf) {
                  pf.value = val;
                  // Do not dispatch change here — server already returned the correct view
                }
              }
            } catch (e) {}
          }
        }
      } catch (e) {}
    } catch (e) {}
  });

  // Reattach modal listeners after HTMX swaps that replace task container
  document.body.addEventListener("htmx:afterSwap", function (evt) {
    if (evt.target && evt.target.id === "task-container") {
      try {
        initializeModalEventListeners();
      } catch (e) {}
    }
  });
});
