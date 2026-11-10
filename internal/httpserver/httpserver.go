package httpserver

import (
	"fmt"
	"net/http"

	"chat/internal/client"
	"chat/internal/hub"
)

type Server struct {
	Hub *hub.Hub
}

func New(h *hub.Hub) *http.ServeMux {
	s := &Server{Hub: h}
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/ws", s.websocketHandler)

	return mux
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	cl := client.New(newWebSocketConn(conn), s.Hub)
	s.Hub.Register(cl)
	go cl.WriteLoop()
	go cl.Run()
}
