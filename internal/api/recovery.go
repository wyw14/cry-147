package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (server *Server) incidents(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, server.runtime.Incidents())
}

func (server *Server) ackIncident(writer http.ResponseWriter, request *http.Request) {
	incident, err := server.runtime.AcknowledgeIncident(chi.URLParam(request, "id"))
	if err != nil {
		writeError(writer, http.StatusNotFound, err)
		return
	}
	writeJSON(writer, http.StatusOK, incident)
}
