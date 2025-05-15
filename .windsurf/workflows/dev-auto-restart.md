---
description: How to enable automatic server restart for Go Sentinel development
---

# Go Sentinel: Automatic Server Restart Development Workflow

## Overview
This workflow enables a smooth, TDD-friendly development experience by automatically restarting the Go Sentinel web server whenever Go code, templates, or static assets change. It uses the `air` tool, which is the best practice for Go projects.

## Prerequisites
- Go 1.18+
- `air` installed: `go install github.com/cosmtrek/air@latest`
- `.air.toml` present in project root (provided)

## Steps

1. **Install air**
   ```bash
   go install github.com/cosmtrek/air@latest
   ```
   Ensure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

2. **Verify .air.toml**
   Ensure `.air.toml` is present in the project root. Example config:
   ```toml
   [build]
     cmd = "go run ./cmd/go-sentinel-web/main.go"
     bin = "tmp/main"
     full_bin = "false"

   [watch]
     dirs = ["./cmd", "./internal"]
     include_ext = ["go", "tmpl", "html", "css", "js"]
   [log]
     color = "true"
     time = "true"
   ```

3. **Start development server with air**
   ```bash
   air
   ```
   The server will restart automatically on code/template/static changes.

4. **(Optional) Customize**
   - Edit `.air.toml` to add/remove watched directories or extensions as needed.
   - See https://github.com/cosmtrek/air for advanced options.

## TDD Integration
- Continue writing and running tests as usual.
- Optionally, configure air to run tests before/after reloads (see air docs).

## Troubleshooting
- If `air` is not found, ensure your Go bin directory is in your PATH.
- For Windows, you may need to restart your terminal after installing `air`.

---
This workflow is aligned with Go Sentinel's systematic, TDD-first, and roadmap-driven approach. For questions, consult the README or ask in project discussions.
