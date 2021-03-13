package server

import (
	"log"
	"net/http"
	"runtime/debug"
)

func (srv *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("[PANIC] recovered", err, string(debug.Stack()))
				http.Error(w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)

			}
		}()
		next.ServeHTTP(w, r)
	})
}
func (srv *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s | User-Agent: %s\n", r.Method, r.URL, r.Header.Get("User-Agent"))
		next.ServeHTTP(w, r)
	})
}
