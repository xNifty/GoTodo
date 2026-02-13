export function initTheme() {
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

export function toggleTheme() {
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

export function attachThemeToggle() {
  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle) {
    themeToggle.removeEventListener("click", toggleTheme);
    themeToggle.addEventListener("click", toggleTheme);
  }
}
