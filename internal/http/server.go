package http

import (
	"log/slog"
	"net/http"
	"time"

	postgres_repo "github.com/lorem-ipsum-team/geode/internal/postgres"
)

type Server struct {
	Server *http.Server
	repo   postgres_repo.Repo
	log    *slog.Logger
}

func New(
	log *slog.Logger,
	addr string,
	repo postgres_repo.Repo,
) Server {
	log = log.WithGroup("http_server")
	serv := Server{
		Server: &http.Server{ //nolint:exhaustruct
			Addr:              addr,
			ReadHeaderTimeout: time.Second / 2,
		},
		repo: repo,
		log:  log,
	}

	log.Info("register handlers")
	serv.registerHandlers()

	return serv
}

func (s Server) registerHandlers() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /geo", s.handleGetSwipes)
	mux.HandleFunc("GET /healthy", s.handleHealthy)

	s.Server.Handler = loggingMiddleware(mux, s.log)
}

func loggingMiddleware(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		log.DebugContext(r.Context(), "request", slog.Group(
			"request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", lrw.statusCode),
			slog.Duration("dur", time.Since(start)),
			slog.String("remote_ip", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
