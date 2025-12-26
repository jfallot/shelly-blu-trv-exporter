package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDebugBluHandler(t *testing.T) {
	jsonBody := `{"components":[{"key":"ble","status":{}},{"key":"blugw","status":{}},{"key":"blutrv:200","status":{"id":200,"target_C":17.8,"current_C":18.5,"pos":54,"connected":true,"rssi":-60,"battery":100,"packet_id":243,"last_updated_ts":1766497610,"paired":true,"rpc":true,"rsv":56,"fw_ver":"v1.2.10"},"config":{},"attrs":{"flags":17,"model_id":8}}],"cfg_rev":29}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(jsonBody))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	cfg := Config{Address: addr, Username: "u", Password: "p", Timeout: 2}

	h := debugBluHandler(cfg)
	req := httptest.NewRequest("GET", "/debug/blu", nil)
	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	var out map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if _, ok := out["200"]; !ok {
		t.Fatalf("expected 200 in response")
	}
	if v := out["200"]["battery"].(float64); int(v) != 100 {
		t.Fatalf("unexpected battery value: %v", v)
	}
}
