package router

import (
	"net/http"
	"os"
	"strings"

	"example.com/auth"
	"example.com/entity"
)

var Mux *http.ServeMux

var corsMiddleware func(http.HandlerFunc) http.HandlerFunc = func(next http.HandlerFunc) http.HandlerFunc {
	return next
}

func corsHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(200)
}

func standarizedContextPath() string {
	t := os.Getenv("CONTEXT_PATH")
	restul := ""
	parts := strings.Split(t, "/")
	for _, p := range parts {
		if p == "" {
			continue
		}
		restul += "/" + p
	}
	return restul
}

func init() {
	var corsOrigin = os.Getenv("CORS_ORIGIN")
	if corsOrigin != "" {
		corsMiddleware = func(next http.HandlerFunc) http.HandlerFunc {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Access-Control-Allow-Origin", corsOrigin)
				rw.Header().Set("Access-Control-Allow-Headers", "authorization,content-type")
				rw.Header().Set("Access-Control-Allow-Methods", "GET,DELETE,POST,PUT")
				next.ServeHTTP(rw, req)
			})
		}
	}
	var contextPath string = standarizedContextPath()
	Mux = http.NewServeMux()
	Mux.HandleFunc("/", http.NotFound)
	Mux.HandleFunc("OPTIONS /", corsMiddleware(corsHandler))
	Mux.HandleFunc("GET "+contextPath+"/entities", corsMiddleware(auth.Middleware(entity.GetEntities, "entity-read")))
	Mux.HandleFunc("GET "+contextPath+"/entities/{id}", corsMiddleware(auth.Middleware(entity.GetEntity, "entity-read")))
	Mux.HandleFunc("POST "+contextPath+"/entities", corsMiddleware(auth.Middleware(entity.AddEntity, "entity-create")))
	Mux.HandleFunc("DELETE "+contextPath+"/entities/{id}", corsMiddleware(auth.Middleware(entity.RemoveEntity, "entity-delete")))
	Mux.HandleFunc("PUT "+contextPath+"/entities/{id}", corsMiddleware(auth.Middleware(entity.UpdateEntity, "entity-update")))
}
