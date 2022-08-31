package api

import "net/http"

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
