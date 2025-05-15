# Go Sentinel Web Middleware (internal/web/middleware/)

This package contains custom middleware for the Go Sentinel web server.

## Responsibilities
- Logging HTTP requests and responses
- Error handling and toast notifications
- Real IP extraction and recovery

## Key Files
- `logger.go`: Custom logging middleware
- `toast_error.go`: Middleware for handling errors and displaying toast notifications

---
See `../server/README.md` for integration details and usage examples.
