# Go Sentinel API (internal/api/)

This package contains the API layer for Go Sentinel, exposing endpoints for test results, metrics, history, coverage, settings, and WebSocket communication.

## Key Technologies
- **Go** (Golang) for all API logic
- **chi** router for HTTP routing and middleware
- **gorilla/websocket** for WebSocket endpoints

## Structure
- `server/`: HTTP server implementation and Swagger UI
- `handlers/`: API route handlers (test results, metrics, etc.)
- `middleware/`: HTTP middleware for API
- `websocket/`: WebSocket connection management and message handling
- `models/`: Data models for API requests and responses

## API Documentation
- OpenAPI/Swagger docs available at `/docs` and `/docs/ui`
- See `ROADMAP-API.md` for API design and endpoint details

---
For integration details, see `../web/README.md` and `ROADMAP-INTEGRATION.md`.
