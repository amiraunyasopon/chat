Learning about Golang, TCP, and TLS by making a chat application

## Running with Docker

The image does not contain TLS credentials. Mount the certificate and private key
at runtime:

```sh
docker build -t chat-server .
docker run --rm \
  -p 8080:8080 \
  -p 8081:8081 \
  -v "$PWD/server.crt:/tls/server.crt:ro" \
  -v "$PWD/server.key:/tls/server.key:ro" \
  -e TLS_CERT_FILE=/tls/server.crt \
  -e TLS_KEY_FILE=/tls/server.key \
  chat-server 8080 8081
```

`CHAT_PORT` changes the port the server listens on inside the container. A
positional first argument overrides `CHAT_PORT`. `WS_PORT` changes the HTTP and
WebSocket port, and a positional second argument overrides `WS_PORT`.
`TLS_CERT_FILE` and `TLS_KEY_FILE` default to `server.crt` and `server.key`.

Alternatively, use Docker Compose:

```sh
docker compose up --build
```

After the first build, use `docker compose up` when only starting the existing
image. Stop it with `Ctrl-C`, or run it in the background with
`docker compose up -d`. The TLS certificate and private key remain on the host
and are mounted read-only into the container.