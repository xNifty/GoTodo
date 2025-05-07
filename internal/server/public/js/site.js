document.addEventListener("DOMContentLoaded", () => {
  let sidebar = document.getElementById("sidebar");

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
    const openSidebarBtn = document.getElementById("openSidebar");
    const closeSidebarBtn = document.getElementById("closeSidebar");

    if (openSidebarBtn) {
      openSidebarBtn.removeEventListener("click", openSidebar); // Prevent duplicate bindings
      openSidebarBtn.addEventListener("click", openSidebar);
    }

    if (closeSidebarBtn) {
      closeSidebarBtn.removeEventListener("click", closeSidebar); // Prevent duplicate bindings
      closeSidebarBtn.addEventListener("click", closeSidebar);
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

  // Optional: Close sidebar after form submission
  const taskForm = document.getElementById("newTaskForm");
  if (taskForm) {
    taskForm.addEventListener("submit", () => {
      closeSidebar();
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
});
