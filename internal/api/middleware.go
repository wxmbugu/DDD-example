package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"net/http"

	"github.com/patienttracker/internal/services"
	"golang.org/x/time/rate"
)

type key string

const (
	session        key = "session"
	admin_session  key = "admin"
	nurse_sesssion key = "nurse"
	staff_session  key = "staff"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func (server *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Redirect(w, r, "/500", http.StatusMovedPermanently)
				server.Log.Error(errors.New(err.(string)), debug.Stack())
			}
		}()
		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		server.Log.Info(
			fmt.Sprintf("status=%d", wrapped.status),
			fmt.Sprintf("method=%s", r.Method),
			fmt.Sprintf("path=%s", r.URL),
			fmt.Sprintf("duration=%s", time.Since(start)),
		)
	})
}

func (server *Server) sessionmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "user-session")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		user := getUser(session)
		if !user.Authenticated {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		ctx := context.WithValue(r.Context(), session, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) sessionadminmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "admin")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		user := getAdmin(session)
		if !user.Authenticated {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		ctx := context.WithValue(r.Context(), admin_session, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) sessionstaffmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "staff")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		user := getStaff(session)
		if !user.Authenticated {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		ctx := context.WithValue(r.Context(), staff_session, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (server *Server) sessionnursemiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "nurse")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		user := getNurse(session)
		if !user.Authenticated {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		ctx := context.WithValue(r.Context(), nurse_sesssion, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CheckPermissions accepts a built-in or a custom checker type and instructs it to
// check if the required permissions were satisfied or not. Based on the
// result, it either returns a 403 response or continues with the request.
func (server *Server) CheckPermissions(next http.HandlerFunc, c services.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "admin")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		user := getAdmin(session)
		if !user.Authenticated {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		if ok := c.IsSatisfied(user.Permission); !ok {
			w.WriteHeader(http.StatusForbidden)
			server.Templates.Render(w, "403.html", nil)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (server *Server) Ipratelimiter(next http.Handler) http.Handler {
	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}

			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusMovedPermanently)
			return
		}
		mu.Lock()
		if _, found := clients[ip]; !found {
			// Create and add a new client struct to the map if it doesn't already exist.
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}
		// Update the last seen time for the client.
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			http.Redirect(w, r, "/429", http.StatusMovedPermanently)
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
