FROM golang:1.27.3-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /chat-server ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 app
COPY --from=build /chat-server /chat-server

USER app
EXPOSE 8080 8081
ENTRYPOINT ["/chat-server"]
