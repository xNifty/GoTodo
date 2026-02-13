// Helper to build correct API URLs that work on both localhost and subpaths
export function apiPath(endpoint) {
  // Remove leading slash if present
  const path = endpoint.startsWith("/") ? endpoint.slice(1) : endpoint;
  // Use relative path with dot prefix so HTMX resolves it relative to current location
  return "./" + path;
}

// Global helper to restore footer if accidentally removed by HTMX
let __originalFooterHTML = null;

export function captureFooterHTML() {
  try {
    const f = document.querySelector("footer");
    if (f) __originalFooterHTML = f.outerHTML;
  } catch (e) {}
}

export function restoreFooterIfMissing() {
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

export function ensureToastContainer() {
  let c = document.querySelector(".app-toast-container");
  if (!c) {
    c = document.createElement("div");
    c.className = "app-toast-container";
    document.body.appendChild(c);
  }
  return c;
}

export function ensureTableClasses() {
  const container = document.getElementById("task-container");
  if (!container) return;
  const table = container.querySelector("table");
  if (!table) return;
  const classes = ["table", "table-striped", "table-bordered", "w-100", "mb-3"];
  classes.forEach((c) => {
    if (!table.classList.contains(c)) table.classList.add(c);
  });
}

// If HTMX returned rows or tbody elements without a surrounding table, wrap them
export function ensureTableStructure() {
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
