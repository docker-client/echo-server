# echo-server

A cross-platform test server with support for **HTTP/1.1 connection upgrades** and **deterministic scripted responses**.

Multi-platform Docker images are available at
`gesellix/echo-server:<version>`
with `<version>` matching Git tags/releases on GitHub.

Supported platforms:
- `linux/amd64`
- `linux/arm64/v8`
- `windows-ltsc2022`

---

## 🚀 Features

- **HTTP/1.1 Upgrade Handling** to raw TCP
- **Scripted deterministic responses** via JSON body
- **Interactive commands** in default mode (e.g. `reset`, `set 10`)
- **Cross-platform & container-friendly**
- **Useful for client protocol testing** and upgrade simulations

---

## 🔌 HTTP/1.1 Upgrade Endpoint

```
POST /api/stream HTTP/1.1
Connection: Upgrade
Upgrade: testproto
```

The server upgrades the connection and speaks raw TCP after switching protocols.

---

## 🧪 Default Behavior

When no script is provided, the server enters a simple interactive loop:

- Sends a counter every second:
  ```
  [0] scripted: counter: 0
  [1] scripted: counter: 1
  ...
  ```

- Accepts basic commands from the client:
  - `reset` → resets counter
  - `set <n>` → sets counter to `<n>`
  - any other input → "unknown command"

You can connect using the built-in client or tools like `nc`:

```bash
go run ./src/hijack_client.go
# or
{ echo -en "GET /api/stream HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\nUpgrade: testproto\r\n\r\n"; cat; } | nc localhost 8080
```

---

## 📜 Scripted Responses (JSON)

Send a JSON payload with a `POST` request to `/api/stream`, along with Upgrade headers.

### Example Request:

```http
POST /api/stream HTTP/1.1
Host: localhost:8080
Connection: Upgrade
Upgrade: testproto
Content-Type: application/json

{
  "actions": [
    { "type": "response", "text": "first" },
    { "type": "delay",    "ms": 1000 },
    { "type": "response", "text": "second" },
    { "type": "close" }
  ]
}
```

### Supported Actions:

| Type       | Description                             | Example           |
|------------|-----------------------------------------|-------------------|
| `response` | Sends a line of text                    | `"text": "Hello"` |
| `delay`    | Waits before next action (milliseconds) | `"ms": 500`       |
| `close`    | Closes the TCP connection               | *(no parameters)* |

---

## 📁 Example Client

A simple Go client for interactive testing is available at:

```
./src/hijack_client.go
```

It connects to the server, performs the upgrade handshake, and allows typing commands live.

---

## 🐳 Docker

To run via Docker:

```bash
docker run --rm -it -p 8080:8080 gesellix/echo-server:<version>
```

---

## 📫 Feedback

Issues and contributions welcome via GitHub.
