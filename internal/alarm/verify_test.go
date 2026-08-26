package alarm_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wyw14/cry-147/internal/api"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
	"github.com/wyw14/cry-147/internal/service"
)

func TestNoAlarmDoesNotBecomeTypedNilFailure(t *testing.T) {
	system, err := cellruntime.New(t.TempDir(), "http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	operation := system.Operations()[0]
	input := service.SampleRequest{RunID: operation.RunID, TrayID: operation.TrayID, CellID: operation.TrayID + "-C001", Channel: 1, Sequence: 1, Voltage: 3.7, Temperature: 25}
	body, _ := json.Marshal(input)
	handler := api.NewServer(service.NewRuntime(system), t.TempDir()).Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/operations/sample", bytes.NewReader(body)))
	if recorder.Code != http.StatusAccepted {
		var failure map[string]string
		_ = json.Unmarshal(recorder.Body.Bytes(), &failure)
		t.Fatalf("normal sample became HTTP %d: %s", recorder.Code, failure["error"])
	}
}
