package api

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wyw14/cry-147/internal/service"
)

type Server struct {
	runtime *service.Runtime
	router  chi.Router
	webRoot string
}

func NewServer(runtime *service.Runtime, webRoot string) *Server {
	server := &Server{runtime: runtime, router: chi.NewRouter(), webRoot: webRoot}
	server.routes()
	return server
}

func (server *Server) routes() {
	server.router.Use(middleware.RequestID)
	server.router.Use(middleware.RealIP)
	server.router.Use(middleware.Recoverer)
	server.router.Get("/healthz", server.health)
	server.router.Route("/api", func(router chi.Router) {
		router.Get("/operations", server.operations)
		router.Post("/operations", server.loadCampaign)
		router.Post("/operations/current", server.setCurrent)
		router.Post("/operations/sample", server.sample)
		router.Post("/operations/grade/preview", server.previewGrade)
		router.Post("/operations/grade/publish", server.publishGrade)
		router.Get("/operations/{runID}/samples", server.operationSamples)
		router.Post("/operations/{runID}/checkpoint", server.saveRecovery)
		router.Post("/operations/{runID}/rest", server.startRest)
		router.Get("/operations/{runID}/rest", server.restStatus)
		router.Post("/operations/{runID}/rest/pause", server.pauseRest)
		router.Post("/operations/{runID}/rest/resume", server.resumeRest)
		router.Post("/operations/{runID}/rest/complete", server.completeRest)
		router.Get("/operations/{runID}/grade", server.gradeResult)
		router.Get("/trays/{trayID}/curve", server.trayCurve)
		router.Get("/trays/{trayID}/isolated-channels", server.isolatedChannels)
		router.Post("/recovery/validate", server.validateRecovery)
		router.Post("/recovery/restore", server.restoreRecovery)
		router.Get("/equipment", server.equipment)
		router.Post("/equipment/result", server.equipmentResult)
		router.Post("/equipment/thermal/{zone}", server.pollThermal)
		router.Get("/equipment/thermal/{zone}", server.thermalReading)
		router.Post("/equipment/thermal/{zone}/wait", server.waitThermalStable)
		router.Get("/interlocks", server.interlocks)
		router.Post("/interlocks", server.isolate)
		router.Post("/interlocks/clear", server.clearIsolation)
		router.Get("/incidents", server.incidents)
		router.Post("/incidents/{id}/ack", server.ackIncident)
	})
	server.router.Get("/", server.page("operations.html"))
	server.router.Get("/operations", server.page("operations.html"))
	server.router.Get("/equipment", server.page("equipment.html"))
	server.router.Get("/interlocks", server.page("interlocks.html"))
	server.router.Get("/incidents", server.page("incidents.html"))
	server.router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(server.webRoot))))
}

func (server *Server) Handler() http.Handler {
	return server.router
}

func (server *Server) page(name string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, filepath.Join(server.webRoot, name))
	}
}
