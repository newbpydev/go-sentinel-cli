# Go Sentinel Backend (internal/)

This directory contains the backend/server-side code for Go Sentinel.

## Key Technologies
- **Go** (Golang) for all backend logic
- **chi** router for HTTP routing and middleware
- **html/template** for server-side rendering
- **WebSocket** support for real-time updates

## Structure
- `web/server/`: Main HTTP server, template rendering, static file serving
- `api/`: API endpoints, handlers, middleware, models
- `web/handlers/`: HTTP handlers for tests, metrics, history, coverage, settings
- `web/middleware/`: Custom middleware (e.g., logging, error handling)

## Template System
- Strict three-tier hierarchy: layouts → partials → pages
- Explicit block definitions in base templates
- Templates loaded in strict order for reliable inheritance

## WebSocket Integration
- WebSocket handler for real-time test updates
- Broadcaster pattern for pushing updates to all clients

## API Overview
- RESTful endpoints for tests, metrics, history, coverage, and settings
- All endpoints documented in `ROADMAP-API.md`

---
For more details, see `web/server/server.go` and `ROADMAP-API.md`.
