document.addEventListener("DOMContentLoaded", () => {
  let sidebar = document.getElementById("sidebar");

  // Theme toggle functionality
  function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'light';
    const themeToggle = document.getElementById('theme-toggle');
    
    if (savedTheme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark');
      if (themeToggle) themeToggle.classList.add('active');
    } else {
      document.documentElement.setAttribute('data-theme', 'light');
      if (themeToggle) themeToggle.classList.remove('active');
    }
  }

  function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    const themeToggle = document.getElementById('theme-toggle');
    
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    
    if (newTheme === 'dark') {
      if (themeToggle) themeToggle.classList.add('active');
    } else {
      if (themeToggle) themeToggle.classList.remove('active');
    }
  }
  
  // Initialize theme on page load
  initTheme();
  
  // Set up theme toggle event listener
  const themeToggle = document.getElementById('theme-toggle');
  if (themeToggle) {
    themeToggle.addEventListener('click', toggleTheme);
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
    
    // Reattach theme toggle if needed
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
      themeToggle.removeEventListener('click', toggleTheme);
      themeToggle.addEventListener('click', toggleTheme);
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

  // Handle task deletion
  document.body.addEventListener("taskDeleted", function (evt) {
    // Get the current page from the page number display
    const pageDisplay = document.querySelector('.text-muted');
    const currentPage = pageDisplay ? parseInt(pageDisplay.textContent.match(/\d+/)[0]) : 1;

    // Reload the current page
    let url = `/api/fetch-tasks?page=${currentPage}`;
    const searchInput = document.getElementById("search");
    if (searchInput && searchInput.value) {
      url += `&search=${encodeURIComponent(searchInput.value)}`;
    }
    htmx.ajax('GET', url, { target: "#task-container", swap: "innerHTML" });
  });

  document.body.addEventListener("reload-previous-page", function (evt) {
    // Get the current page from the page number display
    const pageDisplay = document.querySelector('.text-muted');
    const currentPage = pageDisplay ? parseInt(pageDisplay.textContent.match(/\d+/)[0]) : 1;
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
    htmx.ajax('GET', url, { target: "#task-container", swap: "innerHTML" });
  });
});
