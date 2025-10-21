document.addEventListener("DOMContentLoaded", () => {
  let sidebar = document.getElementById("sidebar");
  let openSidebarBtn = document.getElementById("openSidebar");
  let closeSidebarBtn = document.getElementById("closeSidebar");
  let description = document.getElementById("description");
  let charCount = document.getElementById("char-count");

  // Character counter for description
  if (description && charCount) {
    description.addEventListener("input", function() {
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
    taskForm.addEventListener("htmx:afterRequest", (event) => {
      // Only close sidebar if the request was successful and not a validation error
      // event.detail.successful will be true for a 200 status code response, even with HX-Trigger/HX-Retarget
      // We check the triggerSpec to see if it was the validation error response
      const isValidationError = event.detail.triggerSpec && event.detail.triggerSpec.trigger === 'description-error';

      if (event.detail.successful && !isValidationError) {
        closeSidebar();
        // Clear the form fields on successful submission
        document.getElementById("title").value = "";
        document.getElementById("description").value = "";
        // Reset character counter
        let charCount = document.getElementById("char-count");
        if (charCount) charCount.textContent = "0";
        // Clear any old validation message
        let errorDiv = document.getElementById("description-error");
        if (errorDiv) errorDiv.innerHTML = "";
      } else if (isValidationError) {
        // Keep the sidebar open and show the error (HTMX will swap the error message into #description-error due to HX-Retarget)
        // The form fields and char counter are retained automatically by the browser
        event.preventDefault(); // Prevent default HTMX swap action on the main target if it somehow gets here
        return false; // Stop further event propagation
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

  // Handle login success
  document.body.addEventListener("login-success", function (evt) {
    // Close the login modal
    const modal = bootstrap.Modal.getInstance(document.getElementById('modal'));
    if (modal) {
      modal.hide();
    }
    
    // Optionally reload the page to show logged-in state
    window.location.reload();
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
        if (!description.classList.contains('char-count-listener-added')) {
            description.addEventListener("input", function() {
                const length = this.value.length;
                charCount.textContent = length;
                // Add visual feedback when approaching limit
                if (length > 90) { charCount.classList.add("text-warning"); } else { charCount.classList.remove("text-warning"); }
                if (length > 95) { charCount.classList.add("text-danger"); } else { charCount.classList.remove("text-danger"); }
            });
             description.classList.add('char-count-listener-added'); // Mark as having listener
        }
      }
      // Re-initialize theme toggle if needed
      if (typeof initTheme === 'function') {
        initTheme(); // This function should be idempotent or handle re-running safely
      }
    }
  });
});
