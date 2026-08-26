package tray_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wyw14/cry-147/internal/api"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
	"github.com/wyw14/cry-147/internal/service"
)

func TestAlarmNotificationDoesNotReenterTrayLock(t *testing.T) {
	system, err := cellruntime.New(t.TempDir(), "http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	operation := system.Operations()[0]
	body, _ := json.Marshal(service.IsolationRequest{RunID: operation.RunID, TrayID: operation.TrayID, CellID: operation.TrayID + "-C001", Channel: 1, Reason: "temperature high"})
	handler := api.NewServer(service.NewRuntime(system), t.TempDir()).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/interlocks", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(recorder, request)
		close(done)
	}()
	select {
	case <-done:
		if recorder.Code != http.StatusAccepted {
			t.Fatalf("interlock status %d: %s", recorder.Code, recorder.Body.String())
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("interlock HTTP request deadlocked while notifying tray listeners")
	}
}
