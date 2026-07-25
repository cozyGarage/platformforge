package api

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/platformforge/platformforge/internal/content"
	"github.com/platformforge/platformforge/internal/lab"
	"github.com/platformforge/platformforge/internal/progress"
	"github.com/platformforge/platformforge/internal/ui"
)

type server struct {
	catalog *content.Catalog
	paths   *content.PathCatalog
	engine  *lab.Engine
	store   *progress.Store
}

type terminalControl struct {
	Type string `json:"type"`
	Rows int    `json:"rows"`
	Cols int    `json:"cols"`
}

func Serve(ctx context.Context, addr, _ string, catalog *content.Catalog, paths *content.PathCatalog, engine *lab.Engine, store *progress.Store) error {
	s := &server{catalog: catalog, paths: paths, engine: engine, store: store}
	r, err := s.routes()
	if err != nil {
		return err
	}

	srv := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	err = srv.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func NewHandler(catalog *content.Catalog, paths *content.PathCatalog, engine *lab.Engine, store *progress.Store) (http.Handler, error) {
	s := &server{catalog: catalog, paths: paths, engine: engine, store: store}
	return s.routes()
}

func (s *server) routes() (*mux.Router, error) {
	r := mux.NewRouter()
	r.Use(securityHeaders)
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", s.health).Methods(http.MethodGet)
	api.HandleFunc("/labs", s.labs).Methods(http.MethodGet)
	api.HandleFunc("/labs/{id}", s.labDetail).Methods(http.MethodGet)
	api.HandleFunc("/labs/{id}/status", s.status).Methods(http.MethodGet)
	api.HandleFunc("/labs/{id}/start", s.start).Methods(http.MethodPost)
	api.HandleFunc("/labs/{id}/validate", s.validate).Methods(http.MethodPost)
	api.HandleFunc("/labs/{id}/reset", s.reset).Methods(http.MethodPost)
	api.HandleFunc("/labs/{id}/stop", s.stop).Methods(http.MethodPost)
	api.HandleFunc("/labs/{id}/terminal", s.terminal)
	api.HandleFunc("/progress", s.progress).Methods(http.MethodGet)
	api.HandleFunc("/progress/{id}", s.progressDetail).Methods(http.MethodGet)
	api.HandleFunc("/paths", s.pathList).Methods(http.MethodGet)
	api.HandleFunc("/paths/{id}", s.pathDetail).Methods(http.MethodGet)
	dist, err := fs.Sub(ui.Assets, "dist")
	if err != nil {
		return nil, err
	}
	files := http.FileServer(http.FS(dist))
	r.PathPrefix("/").Handler(spaHandler(files))
	return r, nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' ws:; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func spaHandler(files http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(strings.TrimPrefix(r.URL.Path, "/"), ".") {
			r.URL.Path = "/"
		}
		files.ServeHTTP(w, r)
	})
}

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, err error) {
	respond(w, status, map[string]string{"error": err.Error()})
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) labs(w http.ResponseWriter, _ *http.Request) {
	v, err := s.catalog.List()
	if err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, v)
}

func (s *server) labDetail(w http.ResponseWriter, r *http.Request) {
	v, err := s.catalog.Get(mux.Vars(r)["id"])
	if err != nil {
		fail(w, 404, err)
		return
	}
	respond(w, 200, v)
}

func (s *server) status(w http.ResponseWriter, r *http.Request) {
	respond(w, 200, s.engine.Status(mux.Vars(r)["id"]))
}

func (s *server) start(w http.ResponseWriter, r *http.Request) {
	v, err := s.engine.Start(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		fail(w, 400, err)
		return
	}
	respond(w, 201, v)
}

func (s *server) validate(w http.ResponseWriter, r *http.Request) {
	v, err := s.engine.Validate(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		fail(w, 400, err)
		return
	}
	respond(w, 200, v)
}

func (s *server) reset(w http.ResponseWriter, r *http.Request) {
	v, err := s.engine.Reset(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		fail(w, 400, err)
		return
	}
	respond(w, 200, v)
}

func (s *server) stop(w http.ResponseWriter, r *http.Request) {
	if err := s.engine.Stop(r.Context(), mux.Vars(r)["id"]); err != nil {
		fail(w, 400, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) progress(w http.ResponseWriter, _ *http.Request) {
	v, err := s.store.List()
	if err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, v)
}

func (s *server) progressDetail(w http.ResponseWriter, r *http.Request) {
	labID := mux.Vars(r)["id"]
	p, err := s.store.Get(labID)
	if err != nil {
		fail(w, 500, err)
		return
	}
	hintCounts := map[string]int{}
	if m, err := s.catalog.Get(labID); err == nil {
		for _, task := range m.Tasks {
			hintCounts[task.ID] = len(task.Hints)
		}
	}
	tasks, err := s.store.TaskProgress(labID, hintCounts)
	if err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, map[string]any{
		"progress":       p,
		"taskProgress":   tasks,
		"ghostHintEvery": progress.GhostHintEvery,
	})
}

func (s *server) pathList(w http.ResponseWriter, _ *http.Request) {
	v, err := s.paths.List()
	if err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, v)
}

func (s *server) pathDetail(w http.ResponseWriter, r *http.Request) {
	v, err := s.paths.Get(mux.Vars(r)["id"])
	if err != nil {
		fail(w, 404, err)
		return
	}
	respond(w, 200, v)
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	host := strings.Split(r.Host, ":")[0]
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}}

type wsStream struct {
	conn *websocket.Conn
	buf  []byte
}

func (w *wsStream) Read(p []byte) (int, error) {
	for len(w.buf) == 0 {
		mt, data, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
			continue
		}
		w.buf = append([]byte(nil), data...)
	}
	n := copy(p, w.buf)
	w.buf = w.buf[n:]
	return n, nil
}

func (w *wsStream) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if err := w.conn.WriteMessage(websocket.BinaryMessage, append([]byte(nil), p...)); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *server) terminal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	labID := mux.Vars(r)["id"]
	pr, pw := io.Pipe()
	stream := &wsStream{conn: conn}

	go func() {
		defer pw.Close()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.TextMessage && strings.HasPrefix(string(data), "{\"type\":") {
				var ctrl terminalControl
				if json.Unmarshal(data, &ctrl) == nil && ctrl.Type == "resize" {
					_ = s.engine.Resize(r.Context(), labID, ctrl.Rows, ctrl.Cols)
					continue
				}
			}
			if _, err := pw.Write(data); err != nil {
				return
			}
		}
	}()

	if err := s.engine.Terminal(r.Context(), labID, pr, stream); err != nil && err != io.EOF {
		log.Printf("terminal: %v", err)
	}
}
