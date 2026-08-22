package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/rr173/task159-licensecompat/internal/service"
	"github.com/rr173/task159-licensecompat/internal/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkflowHTTP(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	api := New(service.New(repository))
	policy := request(t, api, "POST", "/v1/policies", map[string]any{"name": "test", "version": 1, "rules": []map[string]any{{"id": "gpl", "name": "no gpl", "forbidden": []string{"GPL-3.0"}, "allow_waiver": true}}})
	policyID := policy["id"].(string)
	request(t, api, "POST", "/v1/policies/"+policyID+"/activate", nil)
	submission := request(t, api, "POST", "/v1/submissions", map[string]any{"name": "app", "release": "1", "components": []map[string]any{{"id": "app", "name": "app", "version": "1", "license": "PROPRIETARY"}, {"id": "dep", "name": "dep", "version": "1", "license": "GPL-3.0"}}, "edges": []map[string]any{{"from": "app", "to": "dep", "scope": "runtime"}}})
	sub := submission["submission"].(map[string]any)
	decision := request(t, api, "POST", "/v1/submissions/"+sub["id"].(string)+"/analyze", nil)
	analysis := decision["analysis"].(map[string]any)
	if analysis["status"] != "blocked" {
		t.Fatalf("status=%v", analysis["status"])
	}
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), "GET", "/v1/analyses/"+analysis["id"].(string)+"/explain", nil))
	if response.Code != http.StatusOK {
		t.Fatal(response.Body.String())
	}
}
func request(t *testing.T, api *API, method, path string, body any) map[string]any {
	t.Helper()
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(method, path, bytes.NewReader(data)))
	if response.Code < 200 || response.Code > 299 {
		t.Fatalf("%s %s: %d %s", method, path, response.Code, response.Body.String())
	}
	var value map[string]any
	if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
