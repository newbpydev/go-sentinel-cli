Below is a concise roadmap for building a browser-based test runner in Go with real-time updates. You can choose between Server-Sent Events (SSE), WebSockets, or even polling. HTMX supports both SSE and WebSocket workflows via official extensions, and on the Go side you can wire up either approach using `net/http`, `fsnotify` (or similar), and a small wrapper library.

## Summary

You have three main approaches for “real-time” test reporting in the browser:

1. **Server-Sent Events (SSE)**
   – Unidirectional (server→client) streams.
   – HTMX provides an `sse` extension so you can declaratively bind incoming events to HTML swaps ([htmx.org][1], [v1.htmx.org][2]).
   – On the Go side, you implement an `http.Handler` that writes `text/event-stream` responses, pushing test results as they arrive ([FreeCodeCamp][3]).

2. **WebSockets**
   – Full duplex (client⇄server), ideal for bidirectional controls like “rerun all” buttons.
   – HTMX’s `ws` extension lets you open a socket and map incoming messages to swaps with simple attributes ([htmx.org][4]).
   – In Go, use `gorilla/websocket` or the `net/http` plus `golang.org/x/net/websocket` packages to manage connections and broadcast test events ([DEV Community][5], [Medium][6]).

3. **Fallback Polling**
   – Simple `hx-get` with a small `interval` throttle.
   – Easiest to implement but least efficient; fine for small projects or prototypes.

Below we’ll deep-dive into each, plus “watch mode” file-change detection and how you might wire this all in Go.

---

## 1. Server-Sent Events (SSE) with HTMX

### How it works

* The browser creates an [`EventSource`](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) connection.
* The server keeps the HTTP response open and pushes messages prefixed with `data:` lines.
* HTMX’s SSE extension automatically listens and swaps content into your page.

### HTMX side

```html
<!-- Include the SSE extension -->
<script src="https://unpkg.com/htmx.org/dist/ext/sse.js"></script>
<div 
  hx-ext="sse"
  sse-connect="/tests/stream"
  sse-swap="innerHTML"
  id="test-output">
  <!-- incoming test results will replace this -->
</div>
```

This tells HTMX to open `/tests/stream` as an SSE connection and inject each event’s payload into `#test-output` ([htmx.org][1]).

### Go side

```go
func streamTests(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    flusher, _ := w.(http.Flusher)
    for result := range testResultsChannel {
        fmt.Fprintf(w, "data: %s\n\n", result)
        flusher.Flush()
    }
}
```

* Launch your `go test` runner in a goroutine.
* As each test completes, send its JSON or HTML snippet to `testResultsChannel`.
* The handler writes out SSE messages in real time ([FreeCodeCamp][3]).

---

## 2. WebSockets with HTMX

### How it works

* HTMX’s `ws` extension (`htmx.org/extensions/ws.js`) lets you use `hx-ws` attributes to open and send on a WebSocket ([htmx.org][4]).
* You can handle messages with `ws-swap` or `hx-swap-oob` for out-of-band updates.

### HTMX side

```html
<script src="https://unpkg.com/htmx.org/dist/ext/ws.js"></script>
<div 
  id="test-output"
  hx-ext="ws"
  hx-ws="connect:/ws/tests"
  hx-swap="innerHTML">
</div>
<button hx-ext="ws" hx-ws="send:/run-all">Run All</button>
```

* `connect:/ws/tests` opens the socket.
* Incoming messages auto-swap into `#test-output`.
* Buttons can `send:` messages to trigger actions ([TutorialsPoint][7]).

### Go side

Using Gorilla WebSocket:

```go
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{}

func wsTests(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    go runTestsAndBroadcast(conn)
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil { break }
        if string(msg) == "run-all" {
            // trigger full test suite
        }
    }
}

func runTestsAndBroadcast(conn *websocket.Conn) {
    for result := range testResultsChannel {
        conn.WriteMessage(websocket.TextMessage, []byte(result))
    }
}
```

This gives you true bidirectional control—for example, you can trigger “watch mode” toggles or reruns ([DEV Community][5]).

---

## 3. Watching File Changes in Go (Watch Mode)

To automatically rerun tests when files change, use a file-watcher:

```go
import "github.com/fsnotify/fsnotify"

func watchAndRun() {
    watcher, _ := fsnotify.NewWatcher()
    defer watcher.Close()
    watcher.Add("./") // watch your project directory
    for {
        select {
        case ev := <-watcher.Events:
            if ev.Op&fsnotify.Write == fsnotify.Write {
                triggerTestRun()
            }
        case err := <-watcher.Errors:
            log.Println("watch error:", err)
        }
    }
}
```

* This uses [`fsnotify`](https://github.com/fsnotify/fsnotify) for cross-platform events ([GitHub][8]).
* On each write event, push a “rerun” message into your SSE or WebSocket pipeline.

---

## 4. Other Frontend Options

* **React + React Query + WebSockets**
  You can combine TanStack Query with a WebSocket client to keep cache in sync ([LogRocket Blog][9], [TkDodo][10]).

* **Polling with HTMX**

  ```html
  <div hx-get="/tests/latest" hx-trigger="every 2s" hx-swap="innerHTML"></div>
  ```

  Simpler but generates more HTTP requests.

* **GraphQL Subscriptions** or **gRPC-Web** can also push streams if you prefer those protocols.

---

## Putting It All Together

1. **Run tests in a goroutine**, streaming results into a Go channel.
2. **Choose your transport**

   * SSE: simple, unidirectional → best for pure reporting.
   * WebSockets: bidirectional → best when you need the client to trigger actions.
3. **Wire HTMX** via `sse` or `ws` extensions to bind updates to DOM swaps.
4. **Use fsnotify** (or a custom file-watcher) to detect changes in `*_test.go` and push a “rerun” event back through your socket.

With this setup, the browser will reflect test results instantly, let users click “rerun all,” and even automatically watch for file saves—all with minimal JavaScript.

[1]: https://htmx.org/extensions/sse/?utm_source=chatgpt.com "htmx Server Sent Event (SSE) Extension"
[2]: https://v1.htmx.org/extensions/server-sent-events/?utm_source=chatgpt.com "The server-sent-events Extension - </> htmx"
[3]: https://www.freecodecamp.org/news/how-to-implement-server-sent-events-in-go/?utm_source=chatgpt.com "How to Implement Server-Sent Events in Go - freeCodeCamp"
[4]: https://htmx.org/extensions/ws/?utm_source=chatgpt.com "htmx Web Socket extension"
[5]: https://dev.to/neelp03/using-websockets-in-go-for-real-time-communication-4b3l?utm_source=chatgpt.com "Using WebSockets in Go for Real-Time Communication"
[6]: https://medium.com/wisemonks/implementing-websockets-in-golang-d3e8e219733b?utm_source=chatgpt.com "Implementing WebSockets in Golang: Real-Time Communication for ..."
[7]: https://www.tutorialspoint.com/htmx/htmx_websockets.htm?utm_source=chatgpt.com "HTMX WebSockets - Tutorialspoint"
[8]: https://github.com/fsnotify/fsnotify?utm_source=chatgpt.com "fsnotify/fsnotify: Cross-platform filesystem notifications for Go. - GitHub"
[9]: https://blog.logrocket.com/tanstack-query-websockets-real-time-react-data-fetching/?utm_source=chatgpt.com "TanStack Query and WebSockets: Real-time React data fetching"
[10]: https://tkdodo.eu/blog/using-web-sockets-with-react-query?utm_source=chatgpt.com "Using WebSockets with React Query | TkDodo's blog"
