document.addEventListener("DOMContentLoaded", () => {
  const sidebar = document.getElementById("sidebar");
  const openSidebarBtn = document.getElementById("openSidebar");
  const closeSidebarBtn = document.getElementById("closeSidebar");

  // Open sidebar
  openSidebarBtn.addEventListener("click", () => {
    sidebar.classList.add("active");
  });

  // Close sidebar
  closeSidebarBtn.addEventListener("click", () => {
    sidebar.classList.remove("active");
  });

  // Optional: Close sidebar when form is submitted
  document.getElementById("newTaskForm").addEventListener("submit", () => {
    sidebar.classList.remove("active");
  });
});
