package api

import (
	"net/http"

	"github.com/wyw14/cry-147/internal/service"
)

func (server *Server) sample(writer http.ResponseWriter, request *http.Request) {
	var input service.SampleRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	value, err := server.runtime.Sample(input)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, value)
}

type gradeRequest struct {
	RunID  string `json:"run_id"`
	TrayID string `json:"tray_id"`
}

func (server *Server) previewGrade(writer http.ResponseWriter, request *http.Request) {
	var input gradeRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	result, tray, err := server.runtime.PreviewGrade(input.RunID, input.TrayID)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"result": result, "tray": tray})
}

func (server *Server) publishGrade(writer http.ResponseWriter, request *http.Request) {
	var input gradeRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	result, err := server.runtime.PublishGrade(input.RunID, input.TrayID)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, result)
}
