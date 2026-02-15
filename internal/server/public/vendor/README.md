# Vendor Assets

This directory contains third-party libraries downloaded for self-hosting to improve performance and eliminate external CDN dependencies.

## Current Versions

- **Bootstrap**: 5.3.0-alpha1
  - Source: https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/
  - Files: `bootstrap/css/bootstrap.min.css`, `bootstrap/js/bootstrap.min.js`, `bootstrap/js/bootstrap.bundle.min.js`

- **Bootstrap Icons**: 1.7.2
  - Source: https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/
  - Files: `bootstrap-icons/bootstrap-icons.css`, `bootstrap-icons/fonts/`

- **HTMX**: 2.0.3
  - Source: https://unpkg.com/htmx.org@2.0.3/
  - Files: `htmx/htmx.min.js`

- **SortableJS**: 1.15.0
  - Source: https://cdn.jsdelivr.net/npm/sortablejs@1.15.0/
  - Files: `sortable/Sortable.min.js`

- **Popper.js**: 2.11.6
  - Source: https://cdn.jsdelivr.net/npm/@popperjs/core@2.11.6/
  - Files: `popper/popper.min.js`

## Updating Vendor Libraries

To update a library:

1. Download the new version files
2. Replace the existing files in the appropriate directory
3. Update this README with the new version number
4. Test thoroughly to ensure compatibility
5. Update the version number in the CSP if needed
