package quality_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wyw14/cry-147/internal/api"
	"github.com/wyw14/cry-147/internal/model"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
	"github.com/wyw14/cry-147/internal/service"
)

func TestGradePreviewCannotMutateLiveTraySnapshot(t *testing.T) {
	system, err := cellruntime.New(t.TempDir(), "http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	operation := system.Operations()[0]
	handler := api.NewServer(service.NewRuntime(system), t.TempDir()).Handler()
	cellID := operation.TrayID + "-C001"
	sampleBody, _ := json.Marshal(service.SampleRequest{RunID: operation.RunID, TrayID: operation.TrayID, CellID: cellID, Channel: 1, Sequence: 1, Voltage: 3.7, Temperature: 25})
	sampleRecorder := httptest.NewRecorder()
	handler.ServeHTTP(sampleRecorder, httptest.NewRequest(http.MethodPost, "/api/operations/sample", bytes.NewReader(sampleBody)))
	if sampleRecorder.Code != http.StatusAccepted {
		t.Fatalf("sample status %d: %s", sampleRecorder.Code, sampleRecorder.Body.String())
	}
	readCurve := func() []model.Cell {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/trays/"+operation.TrayID+"/curve", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("curve status %d", recorder.Code)
		}
		var values []model.Cell
		if err := json.Unmarshal(recorder.Body.Bytes(), &values); err != nil {
			t.Fatal(err)
		}
		return values
	}
	before := readCurve()[0]
	previewBody, _ := json.Marshal(map[string]string{"run_id": operation.RunID, "tray_id": operation.TrayID})
	previewRecorder := httptest.NewRecorder()
	handler.ServeHTTP(previewRecorder, httptest.NewRequest(http.MethodPost, "/api/operations/grade/preview", bytes.NewReader(previewBody)))
	if previewRecorder.Code != http.StatusOK {
		t.Fatalf("preview status %d: %s", previewRecorder.Code, previewRecorder.Body.String())
	}
	after := readCurve()[0]
	if after.Capacity != before.Capacity || after.Grade != before.Grade {
		t.Fatalf("preview mutated live cell from %+v to %+v", before, after)
	}
}
