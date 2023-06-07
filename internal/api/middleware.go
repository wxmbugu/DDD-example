package api

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"net/http"
	"strings"

	"github.com/patienttracker/internal/services"
)

// json set header middleware
func jsonmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type PayloadKey string

const (
	authPayloadKey PayloadKey = "auth_payload"
	authHeaderKey             = "authorization"
)

func (server *Server) authmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqtoken := r.Header.Get(authHeaderKey)
		tokenvalue := strings.Split(reqtoken, "Bearer ")
		if len(tokenvalue) == 0 {
			server.Log.Debug("authorization header not provided")
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": "authorization header not provided"})
			return
		}
		if len(tokenvalue) < 2 {
			server.Log.Debug("invalid authorization header format")
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": "invalid authorization header format"})
			return
		}

		reqtoken = tokenvalue[1]
		payload, err := server.Auth.VerifyToken(reqtoken)
		if err != nil {
			server.Log.Debug(err.Error())
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": err.Error()})
			return
		}
		ctx := context.WithValue(r.Context(), authPayloadKey, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
	return
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func (server *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Redirect(w, r, "/500", 300)
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
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/login", 300)
		}
		user := getUser(session)
		if !user.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/login", 300)
		}
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) sessionadminmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "admin")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/admin/login", 300)
		}
		user := getAdmin(session)
		if !user.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/admin/login", 300)
		}
		ctx := context.WithValue(r.Context(), "session-admin", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) sessionstaffmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "staff")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/staff/login", 300)
		}
		user := getStaff(session)
		if !user.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/staff/login", 300)
		}
		ctx := context.WithValue(r.Context(), "staff", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (server *Server) sessionnursemiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "nurse")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/nurse/login", 300)
		}
		user := getNurse(session)
		if !user.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/nurse/login", 300)
		}
		ctx := context.WithValue(r.Context(), "nurse", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Check accepts a built-in or a custom checker type and instructs it to
// check if the required permissions were satisfied or not. Based on the
// result, it either returns a 403 response or continues with the request.
func (server *Server) CheckPermissions(next http.HandlerFunc, c services.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Store.Get(r, "admin")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/admin/login", 300)
		}
		user := getAdmin(session)
		if !user.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/admin/login", 300)
		}

		if ok := c.IsSatisfied(user.Permission); !ok {
			w.WriteHeader(http.StatusForbidden)
			server.Templates.Render(w, "403.html", nil)
			return
		}

		next.ServeHTTP(w, r)
	}
}
