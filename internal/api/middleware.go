package api

import (
	"context"
	//	"fmt"
	// "fmt"
	"net/http"
	"strings"
	// "github.com/patienttracker/internal/auth"
)

//	func (server *Server) contentTypeMiddleware(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.Header().Add("Content-Type", "application/json")
//			next.ServeHTTP(w, r)
//		})
//	}
//
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

func (server Server) authmiddleware(next http.Handler) http.Handler {
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
