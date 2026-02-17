// Admin page functionality
// Handles character counter and validation for admin settings

/**
 * Initialize character counter for global announcement text
 */
export function initAnnouncementCharCounter() {
  const announcementText = document.getElementById("global_announcement_text");
  const announcementCharCount = document.getElementById(
    "announcement-char-count",
  );

  if (announcementText && announcementCharCount) {
    announcementText.addEventListener("input", function () {
      const length = this.value.length;
      announcementCharCount.textContent = length;

      // Add visual feedback when approaching limit
      if (length > 450) {
        announcementCharCount.classList.add("text-warning");
      } else {
        announcementCharCount.classList.remove("text-warning");
      }
      if (length > 480) {
        announcementCharCount.classList.add("text-danger");
      } else {
        announcementCharCount.classList.remove("text-danger");
      }
    });

    // Clear error when user starts typing
    announcementText.addEventListener("input", function () {
      const errorDiv = document.getElementById("announcement-text-error");
      if (errorDiv) {
        errorDiv.innerHTML = "";
      }
    });
  }
}
