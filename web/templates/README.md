# Go Sentinel Templates (web/templates/)

This directory contains all HTML templates for the Go Sentinel frontend, organized in a strict three-tier hierarchy:

- `layouts/`: Base templates (e.g., `base.tmpl`)
- `partials/`: Reusable components (e.g., nav, status bar, test row)
- `pages/`: Route-specific templates (e.g., dashboard, tests, history)

## Template Loading Order
Templates are loaded in this order:
1. Layouts
2. Partials
3. Pages

## Inheritance Model
- Pages use `{{define "pagename"}}{{template "base" .}}{{end}}`
- All blocks are explicitly defined in base layouts
- No dynamic template name construction or complex logic

## Best Practices
- Extract common UI into partials
- Use HTMX attributes for dynamic UI
- Test templates before implementation (TDD)

See the main `web/README.md` for more details.
