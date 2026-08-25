# GoTodo

GoTodo (Ordryn) is a self-hosted task manager built with Go, PostgreSQL, Redis, and a Vue 3 SPA. It focuses on simplicity and a pleasant experience: user accounts, per-user tasks, invite flow, role-based permissions, and a JSON `/api/v1` for web and mobile clients.

**Current version:** v2.1.4

## Features

- User signup / login / logout with forgot-password flow and optional TOTP MFA
- Editable profile (display name, timezone, tasks-per-page preference)
- Per-user tasks: add, edit, duplicate, complete, delete, drag-and-drop reorder
- Projects with rename, archive, and delete; tags with create-on-type, rename, and delete
- Kanban boards with custom statuses, estimates, claims, and named sprints (descriptions, date ranges, lock dates, and a board sprint switcher)
- Priority levels (None / Low / Medium / High) with optional sort-by-priority view
- Due dates with smart filters (today, overdue, this week, no date) and relative labels
- Starred tasks pinned above pagination
- Search with project, status, tag, and due-date filters
- Markdown task descriptions with truncated list view and expand-in-place
- Task discussion comments with @-mentions of project members (notifies them) and #task links
- Bulk actions: complete, delete, move project, add/remove tag, set/clear due date, set priority
- Undo delete (toast with up to 120 seconds to restore, preserves task IDs when possible)
- ICS calendar feed for due tasks; in-app calendar view; ICS import to sync due dates
- Daily email digest (opt-in on Profile)
- CSV import with preview/confirm; CSV/JSON export (auto-creates projects and tags on import)
- Task activity timeline in the edit sidebar
- Dashboard with overdue/today counts, completion charts, and streak tracking
- Keyboard shortcuts for power users (`?` for help)
- Invite-only registration and role-based permissions (admin, create invites)
- Admin panel: site settings, user management, global announcements
- Dark and light themes
- Vue 3 SPA at the site root (or `BASE_PATH`, e.g. `/gotodo/`) over `/api/v1` (session cookie auth)
- Live updates over Server-Sent Events so shared projects and other tabs stay in sync without a refresh

## Requirements

- Go 1.24+
- PostgreSQL
- Redis (required for `/api/v1` auth, rate limits, device SSO, and live updates)
- Node.js + npm (to build or develop the Vue SPA)

## Quick start

One binary serves `/api/v1` and the Vue UI at `/` (or under `BASE_PATH`).

```bash
cp .env.example .env   # required; process will not start without it
npm run build:web      # writes web/dist; UI path is 503 without it
go run .
```

Set `DB_*`, `SESSION_KEY`, and `REDIS_URL` in `.env`. Open http://localhost:8080/

Install, reverse proxy, API-only mode, and upgrades: **[wiki](https://github.com/SentientTD-Studios/Ordryn/wiki)**.

## Docs

- [Wiki](https://github.com/SentientTD-Studios/Ordryn/wiki) — self-hosting, configuration, API guide, example clients
- [`openapi.yaml`](openapi.yaml) — `/api/v1` contract (also `GET /openapi.yaml` on a running instance)
- [`web/README.md`](web/README.md) — Vue SPA development
- [License](LICENSE)

SPA source lives in `web/`. Vite hot reload and tests: [Local development](https://github.com/SentientTD-Studios/Ordryn/wiki/Local-development).
