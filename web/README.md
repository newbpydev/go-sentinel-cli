# Go Sentinel Frontend (web/)

This directory contains the web frontend for Go Sentinel.

## Key Technologies
- **Go server-side templates** (html/template)
- **HTMX** for dynamic, real-time UI (included via CDN)
- **Standard CSS** for styling (utility classes)
- **Minimal JavaScript** (custom extensions, e.g., htmx-ws.js)

## Structure
- `static/`: CSS, JS, and image assets
- `templates/`: HTML templates (layouts, partials, pages)
- `templates/layouts/`: Base layouts (e.g., base.tmpl)
- `templates/partials/`: Reusable UI components
- `templates/pages/`: Route-specific templates

## Template Hierarchy
- Layouts → Partials → Pages (strict order)
- Pages inherit from base layout using `{{define "pagename"}}{{template "base" .}}{{end}}`
- Common UI patterns extracted into partials

## Real-Time Features
- HTMX WebSocket extension for live test updates
- Connection status indicator and keyboard shortcuts
- Animated UI updates for real-time test results

---
See `ROADMAP-FRONTEND.md` for feature plans and best practices.
