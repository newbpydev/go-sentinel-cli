# Go Sentinel Web Server (internal/web/)

This package contains the Go HTTP web server and related handlers for Go Sentinel.

## Key Features
- Uses `chi` for routing and middleware
- Robust template rendering system (layouts → partials → pages)
- Static file serving from `/static/`
- Page routes and API endpoints under `/api/`
- WebSocket handler for real-time test updates

## Main Entry Point
- `server/server.go`: Main server logic, route registration, template system, WebSocket setup

## Handlers
- `handlers/`: HTTP and API handlers for tests, metrics, history, coverage, and settings
- `middleware/`: Logging, error handling, and other middleware

## WebSocket
- WebSocket handler is started on server init
- Broadcasts test results and status updates to all connected clients

---
See `../README.md` for backend architecture and API details.
