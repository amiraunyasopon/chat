# Chat Server

Chat Server is a Go application that provides a TLS-protected TCP chat server
and an HTTP/WebSocket interface.

## What this project does

- Accept chat clients over a TLS TCP connection
- Broadcast messages to everyone in the current room
- Support nicknames, direct messages, rooms, and room member lists
- Provide a WebSocket endpoint for browser and WebSocket clients
- Provide a health check endpoint

## Tech stack

- Go
- TLS
- WebSocket
- Docker
- Docker Compose

## Run locally

The server requires a TLS certificate and private key. With `server.crt` and
`server.key` in the project root, run:

```bash
go run ./cmd/server
```

By default, the TLS chat server listens on port `8080`, and the HTTP/WebSocket
server listens on port `8081`.

## Run tests

```bash
go test ./...
```

## API quick examples

### Health check

```text
GET http://localhost:8081/healthz
```

### WebSocket

Connect to:

```text
ws://localhost:8081/ws
```

### Chat commands

Send newline-delimited commands over the TLS TCP connection or WebSocket:

- `PING`
- `ECHO <text>`
- `MSG <text>`
- `NICK <name>`
- `DM <user> <text>`
- `JOIN <room>`
- `LEAVE`
- `WHO`
- `ROOMS`
- `QUIT`

## Run with Docker

The image does not contain TLS credentials. The certificate and private key
must be mounted at runtime:

```bash
docker build -t chat-server .
docker run --rm \
  -p 8080:8080 \
  -p 8081:8081 \
  -v "$PWD/server.crt:/tls/server.crt:ro" \
  -v "$PWD/server.key:/tls/server.key:ro" \
  -e TLS_CERT_FILE=/tls/server.crt \
  -e TLS_KEY_FILE=/tls/server.key \
  chat-server
```

The server also accepts positional port arguments:

```bash
docker run --rm \
  -p 8080:8080 \
  -p 8081:8081 \
  -v "$PWD/server.crt:/tls/server.crt:ro" \
  -v "$PWD/server.key:/tls/server.key:ro" \
  chat-server 8080 8081
```

`CHAT_PORT` changes the TLS chat port, and `WS_PORT` changes the HTTP/WebSocket
port. The first and second positional arguments override those environment
variables. `TLS_CERT_FILE` and `TLS_KEY_FILE` default to `server.crt` and
`server.key`.

## Run with Docker Compose

```bash
docker compose up --build
```

After the first build, use `docker compose up` to start the existing image.
Run it in the background with `docker compose up -d`, and stop it with:

```bash
docker compose down
```

Docker Compose mounts `server.crt` and `server.key` read-only into the
container and publishes ports `8080` and `8081`.
