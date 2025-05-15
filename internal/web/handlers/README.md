# Go Sentinel Web Handlers (internal/web/handlers/)

This package contains HTTP and API route handlers for the Go Sentinel web server.

## Responsibilities
- Serve test results, metrics, history, coverage, and settings via HTTP and API endpoints
- Render templates for web pages and partials
- Handle WebSocket connections for real-time updates

## Key Files
- `test_results.go`: Test results API and page handlers
- `metrics.go`: Metrics API and dashboard handlers
- `websocket.go`: WebSocket handler for real-time communication
- `history.go`: Test run history endpoints
- `coverage.go`: Coverage data endpoints
- `settings.go`: Settings API and page handlers

---
See `../server/README.md` for server setup and template integration.
