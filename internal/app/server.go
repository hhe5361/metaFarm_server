package app

import "metafarm/internal/storage"

type Server struct {
	listenAddr string
	router     *serverRouter
	openaiKey  string
}

func NewServer(listenAddr string, storage storage.Storage, openaiKey string) *Server {
	router := newServerRouter(storage, openaiKey)
	return &Server{listenAddr: listenAddr, router: router, openaiKey: openaiKey}
}

func (s *Server) Run() error {
	return s.router.Run(s.listenAddr)
}
