// Global Announcement module
// Handles dismissal and cookie-based persistence

/**
 * Get a cookie value by name
 * @param {string} name - Cookie name
 * @returns {string|null} - Cookie value or null if not found
 */
function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(";").shift();
  return null;
}

/**
 * Set a cookie with expiration
 * @param {string} name - Cookie name
 * @param {string} value - Cookie value
 * @param {number} days - Days until expiration
 */
function setCookie(name, value, days) {
  const date = new Date();
  date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
  const expires = `expires=${date.toUTCString()}`;
  document.cookie = `${name}=${value};${expires};path=/;SameSite=Lax`;
}

/**
 * Dismiss the global announcement and set a cookie
 */
export function dismissGlobalAnnouncement() {
  const announcement = document.getElementById("global-announcement");
  if (announcement) {
    // Fade out animation
    announcement.classList.remove("show");
    announcement.classList.add("fade-out");

    // Remove from DOM after animation
    setTimeout(() => {
      announcement.remove();
    }, 300);

    // Set cookie to remember dismissal for 30 days
    setCookie("announcement_dismissed", "true", 30);
  }
}

/**
 * Initialize announcement banner on page load
 * Checks if user has previously dismissed the announcement
 */
export function initGlobalAnnouncement() {
  const announcement = document.getElementById("global-announcement");
  if (announcement) {
    // Check if user has dismissed the announcement
    const isDismissed = getCookie("announcement_dismissed");

    if (isDismissed === "true") {
      // Hide the announcement if previously dismissed
      announcement.style.display = "none";
    }

    // Attach event listener to close button
    const closeButton = announcement.querySelector(".btn-close");
    if (closeButton) {
      closeButton.addEventListener("click", dismissGlobalAnnouncement);
    }
  }
}
