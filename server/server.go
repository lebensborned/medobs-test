package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lebensborned/medobs-test/store"
)

// Server wrapper
type Server struct {
	config *Config
	router *mux.Router
	store  *store.Store
}

// New creates config for server
func New(config *Config) *Server {
	return &Server{
		config: config,
		router: mux.NewRouter(),
	}
}

// Start init a store, routes, middlewares, http server
func (srv *Server) Start() error {
	store, err := store.New(srv.config.DBURL, srv.config.DBName)
	if err != nil {
		return err
	}
	srv.store = store
	if err := srv.store.Connect(); err != nil {
		return err
	}
	srv.configureRouter()
	appHandler := srv.logRequests(srv.router)
	appHandler = srv.recoverPanic(appHandler)
	log.Println("Starting server")
	return http.ListenAndServe(srv.config.BindAdrr, appHandler)
}

func (srv *Server) configureRouter() {
	http.Handle("/", srv.router)
	srv.router.HandleFunc("/login/{guid}", srv.GetTokens).Methods(http.MethodGet)
	srv.router.HandleFunc("/refresh", srv.RefreshTokens).Methods(http.MethodPost)
}
