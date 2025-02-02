package index

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	servicegoindex "github.com/go-mod-proxy/go-mod-proxy/go/internal/pkg/service/goindex"
)

type ServerOptions struct {
	IndexService *servicegoindex.Service
	Router       *mux.Router
}

type Server struct {
	service *servicegoindex.Service
	router  *mux.Router
}

func NewServer(opts ServerOptions) *Server {
	s := &Server{
		service: opts.IndexService,
		router:  opts.Router,
	}
	s.router.PathPrefix("/index").Methods(http.MethodGet).Handler(s)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var since time.Time
	var limit int
	query := req.URL.Query()
	if sinceParam, err := time.Parse(time.RFC3339, query.Get("since")); err == nil {
		since = sinceParam
	}
	if limitParam, err := strconv.Atoi(query.Get("limit")); err == nil {
		limit = limitParam
	}

	modules, err := s.service.GetIndex(req.Context(), since, limit)
	if err != nil {
		status := http.StatusInternalServerError
		log.Errorf("failed to query storage due to error: %v", err)
		w.WriteHeader(status)
		return
	}

	enc := json.NewEncoder(w)
	for _, module := range modules {
		if err = enc.Encode(module); err != nil {
			log.Trace("error encoding module, assuming client closed connection")
			return
		}
	}
}
