package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-147/internal/service"
)

func (server *Server) operationSamples(writer http.ResponseWriter, request *http.Request) {
	values, err := server.runtime.Samples(chi.URLParam(request, "runID"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	writeJSON(writer, http.StatusOK, values)
}

func (server *Server) trayCurve(writer http.ResponseWriter, request *http.Request) {
	values, err := server.runtime.Curve(chi.URLParam(request, "trayID"))
	if err != nil {
		writeError(writer, http.StatusNotFound, err)
		return
	}
	writeJSON(writer, http.StatusOK, values)
}

func (server *Server) saveRecovery(writer http.ResponseWriter, request *http.Request) {
	value, err := server.runtime.SaveRecovery(chi.URLParam(request, "runID"))
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusCreated, value)
}

type recoveryRequest struct {
	Path string `json:"path"`
}

func (server *Server) validateRecovery(writer http.ResponseWriter, request *http.Request) {
	var input recoveryRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if err := server.runtime.ValidateRecovery(input.Path); err != nil {
		writeError(writer, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "valid"})
}

func (server *Server) restoreRecovery(writer http.ResponseWriter, request *http.Request) {
	var input recoveryRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	value, err := server.runtime.RestoreRecovery(input.Path)
	if err != nil {
		writeError(writer, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

type restRequest struct {
	Seconds     int     `json:"seconds"`
	Temperature float64 `json:"temperature"`
}

func (server *Server) startRest(writer http.ResponseWriter, request *http.Request) {
	var input restRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if err := server.runtime.StartRest(chi.URLParam(request, "runID"), input.Seconds); err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	value, _ := server.runtime.RestStatus(chi.URLParam(request, "runID"))
	writeJSON(writer, http.StatusAccepted, value)
}

func (server *Server) restStatus(writer http.ResponseWriter, request *http.Request) {
	value, err := server.runtime.RestStatus(chi.URLParam(request, "runID"))
	if err != nil {
		writeError(writer, http.StatusNotFound, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) pauseRest(writer http.ResponseWriter, request *http.Request) {
	value, err := server.runtime.PauseRest(chi.URLParam(request, "runID"))
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) resumeRest(writer http.ResponseWriter, request *http.Request) {
	value, err := server.runtime.ResumeRest(chi.URLParam(request, "runID"))
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) completeRest(writer http.ResponseWriter, request *http.Request) {
	var input restRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	value, err := server.runtime.CompleteRest(request.Context(), chi.URLParam(request, "runID"), input.Temperature)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) equipmentResult(writer http.ResponseWriter, request *http.Request) {
	var input service.EquipmentResultRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	value, err := server.runtime.EquipmentResult(input)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) clearIsolation(writer http.ResponseWriter, request *http.Request) {
	var input service.ClearIsolationRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if err := server.runtime.ClearIsolation(input); err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "cleared"})
}

func (server *Server) gradeResult(writer http.ResponseWriter, request *http.Request) {
	value, ok := server.runtime.GradeResult(chi.URLParam(request, "runID"))
	if !ok {
		writeError(writer, http.StatusNotFound, fmt.Errorf("grade result not found"))
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) thermalReading(writer http.ResponseWriter, request *http.Request) {
	value, ok := server.runtime.ThermalReading(chi.URLParam(request, "zone"))
	if !ok {
		writeError(writer, http.StatusNotFound, fmt.Errorf("thermal reading not found"))
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (server *Server) waitThermalStable(writer http.ResponseWriter, request *http.Request) {
	if err := server.runtime.WaitThermalStable(request.Context(), chi.URLParam(request, "zone"), 250*time.Millisecond); err != nil {
		writeError(writer, http.StatusBadGateway, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "stable"})
}

func (server *Server) isolatedChannels(writer http.ResponseWriter, request *http.Request) {
	value, err := server.runtime.IsolatedChannels(chi.URLParam(request, "trayID"))
	if err != nil {
		writeError(writer, http.StatusNotFound, err)
		return
	}
	writeJSON(writer, http.StatusOK, value)
}
