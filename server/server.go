package server

import (
	"encoding/json"
	"net/http"

	"github.com/cumbreras/shortener/service"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

// Server struct defines the server
type Server struct {
	logger            hclog.Logger
	router            *mux.Router
	repositoryService service.RepositoryService
}

// New generates a new server
func New(router *mux.Router, logger hclog.Logger, repositoryService service.RepositoryService) *Server {
	srv := &Server{router: router, logger: logger, repositoryService: repositoryService}
	srv.routes()
	return srv
}

// GenericError is a generic error message returned by a server
type GenericError struct {
	Message string `json:"message"`
}

func (s *Server) handlerShortenerGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Getting ShortenURL")
		vars := mux.Vars(r)
		code := vars["code"]
		s.logger.Info("Code found as parameter", code)

		shortenURLEnt, err := s.repositoryService.Find(r.Context(), code)

		if err != nil {
			s.respond(w, r, &GenericError{Message: err.Error()}, http.StatusNotFound)
			return
		}

		http.Redirect(w, r, shortenURLEnt.URL, http.StatusMovedPermanently)
	}
}

func (s *Server) handlerShortenerCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Creating ShortenURL")
		us := s.repositoryService.NewUS()
		err := us.FromJSON(r.Body)

		if err != nil {
			s.respond(w, r, &GenericError{Message: err.Error()}, http.StatusConflict)
		}

		shortenURLEnt, err := s.repositoryService.Create(r.Context(), us)

		if err != nil {
			s.respond(w, r, &GenericError{Message: err.Error()}, http.StatusConflict)
		}

		s.respond(w, r, shortenURLEnt, http.StatusCreated)
	}
}

func (s *Server) handlerShortenerDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Deleting ShortenURL")
		vars := mux.Vars(r)
		code := vars["code"]
		s.logger.Info("Code found as parameter", code)

		err := s.repositoryService.Destroy(r.Context(), code)

		if err != nil {
			s.respond(w, r, &GenericError{Message: err.Error()}, http.StatusNotFound)
		}

		s.respond(w, r, nil, http.StatusNoContent)
	}
}

func (s *Server) commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.router.Use(s.commonMiddleware)
	s.router.HandleFunc("/", s.handlerShortenerCreate()).Methods(http.MethodPost)
	s.router.HandleFunc("/{code}", s.handlerShortenerGet()).Methods(http.MethodGet)
	s.router.HandleFunc("/{code}", s.handlerShortenerDelete()).Methods(http.MethodDelete)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			s.logger.Error(err.Error())
		}
	}
}
