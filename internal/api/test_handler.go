package api

import "net/http"

func (s *Server) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
