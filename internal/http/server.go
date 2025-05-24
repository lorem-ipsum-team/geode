package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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

	corsMiddleware := CORS(CORSOptions{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		AllowCredentials: true,
	})

	s.Server.Handler = loggingMiddleware(corsMiddleware(mux), s.log)
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

type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func CORS(opts CORSOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if len(opts.AllowedOrigins) > 0 {
				allowed := false

				for _, o := range opts.AllowedOrigins {
					if o == "*" || o == origin {
						allowed = true

						break
					}
				}

				if !allowed {
					next.ServeHTTP(w, r)

					return
				}
			}

			// Set headers
			if len(opts.AllowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if len(opts.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(opts.AllowedMethods, ", "))
			}

			if len(opts.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(opts.AllowedHeaders, ", "))
			}

			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if opts.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
