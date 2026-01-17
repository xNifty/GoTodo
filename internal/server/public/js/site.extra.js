// Extra helpers loaded after the main bundle to ensure dev-time hooks are active
(function () {
  document.addEventListener("DOMContentLoaded", function () {
    // Open sidebar immediately when edit button is clicked so user sees loading
    document.body.addEventListener("click", function (e) {
      try {
        var btn = e.target && e.target.closest && e.target.closest(".edit-btn");
        if (!btn) return;
        var sb = document.getElementById("sidebar");
        if (sb) {
          sb.classList.add("active");
        }
      } catch (err) {}
    });

    // Delegated close button handler for the extra helper as well
    document.body.addEventListener("click", function (e) {
      try {
        var close =
          e.target && e.target.closest && e.target.closest("#closeSidebar");
        if (!close) return;
        var sb = document.getElementById("sidebar");
        if (sb) sb.classList.remove("active");
      } catch (err) {}
    });

    // Log /api/edit responses and open sidebar on success
    document.body.addEventListener("htmx:afterRequest", function (evt) {
      try {
        var xhr = evt && evt.detail && evt.detail.xhr;
        if (!xhr || !xhr.responseURL) return;
        if (xhr.responseURL.indexOf("/api/edit") !== -1) {
          try {
            // console.log(
            //   "response start:",
            //   (xhr.responseText || "").slice(0, 300)
            // );
          } catch (e) {}
          if (xhr.status >= 200 && xhr.status < 300) {
            var sb2 = document.getElementById("sidebar");
            if (sb2) {
              sb2.classList.add("active");
            }
          }
        }
      } catch (e) {}
    });
  });
})();
