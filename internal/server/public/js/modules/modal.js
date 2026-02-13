import { apiPath } from "./utils.js";

export function initializeModalEventListeners() {
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

export function renderChangelog(entries) {
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
    a.textContent = "View full changelog";
    more.appendChild(a);
    out.appendChild(more);
  }
  container.innerHTML = "";
  container.appendChild(out);
}

export function loadChangelog() {
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

export function attachChangelogListener() {
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
}
