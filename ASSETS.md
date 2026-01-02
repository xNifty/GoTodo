# Asset build (JS/CSS minification)

This project uses `esbuild` to minify and bundle the frontend assets.

Install dev dependencies:

```powershell
npm install
```

Build assets (minifies in-place):

```powershell
npm run build:assets
```

Notes:

- Scripts overwrite `internal/server/public/js/site.js` and `internal/server/public/css/site.css` with minified output so template includes don't need changing.
- Bump `ASSET_VERSION` (or set via env/CI) to force client cache refresh.

CI integration:

- A GitHub Actions workflow is included at `.github/workflows/build-assets.yml`. On pushes to `main` it will build assets, compute an `ASSET_VERSION` timestamp, write `internal/server/public/.asset_version`, and commit the built assets + `.asset_version` back to the repository.
- Ensure branch protection or permissions allow the workflow to push; the default `GITHUB_TOKEN` should be sufficient for simple repos.

File location:

- The asset version file is written to `internal/server/public/.asset_version` and is read by the server at runtime. This keeps asset versioning next to the built assets and avoids committing secrets in `.env`.

If you prefer using a git SHA instead of a timestamp, edit the workflow step that computes `ASSET_VERSION`.
