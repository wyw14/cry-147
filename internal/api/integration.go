package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-147/internal/service"
)

func (server *Server) equipment(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, server.runtime.Equipment())
}

func (server *Server) pollThermal(writer http.ResponseWriter, request *http.Request) {
	reading, err := server.runtime.PollThermal(request.Context(), chi.URLParam(request, "zone"))
	if err != nil {
		writeError(writer, http.StatusBadGateway, err)
		return
	}
	writeJSON(writer, http.StatusOK, reading)
}

func (server *Server) interlocks(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, server.runtime.Interlocks())
}

func (server *Server) isolate(writer http.ResponseWriter, request *http.Request) {
	var input service.IsolationRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if err := server.runtime.Isolate(input); err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]string{"status": "latched"})
}
