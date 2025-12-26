package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestBluCollectorMetrics(t *testing.T) {
	// Set up a test server that returns the sample JSON
	jsonBody := `{"components":[{"key":"ble","status":{}},{"key":"blugw","status":{}},{"key":"blutrv:200","status":{"id":200,"target_C":17.8,"current_C":18.5,"pos":54,"connected":true,"rssi":-60,"battery":100,"packet_id":243,"last_updated_ts":1766497610,"paired":true,"rpc":true,"rsv":56,"fw_ver":"v1.2.10"},"config":{},"attrs":{"flags":17,"model_id":8}},{"key":"blutrv:201","status":{"id":201,"target_C":19.5,"current_C":20.1,"pos":1,"connected":true,"rssi":-60,"battery":100,"packet_id":80,"last_updated_ts":1766497602,"paired":true,"rpc":true,"rsv":16,"fw_ver":"v1.2.10"},"config":{},"attrs":{"flags":17,"model_id":8}},{"key":"blutrv:202","status":{"id":202,"target_C":17.5,"current_C":16.9,"pos":4,"connected":true,"rssi":-72,"battery":100,"packet_id":52,"last_updated_ts":1766497618,"paired":true,"rpc":true,"rsv":1,"fw_ver":"v1.2.10"},"config":{},"attrs":{"flags":17,"model_id":8}}],"cfg_rev":29}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(jsonBody))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	cfg := Config{Address: addr, Username: "u", Password: "p", Timeout: 2}

	reg := prometheus.NewRegistry()
	reg.MustRegister(NewBluCollector(cfg))

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	// helper to find metric family
	find := func(name string) *dto.MetricFamily {
		for _, mf := range mfs {
			if mf.GetName() == name {
				return mf
			}
		}
		return nil
	}

	// check components total
	mf := find("blutrv_components_total")
	if mf == nil || len(mf.Metric) == 0 {
		t.Fatalf("blu_components_total missing")
	}
	if *mf.Metric[0].Gauge.Value != 4 {
		t.Fatalf("expected 4 blu components, got %v", *mf.Metric[0].Gauge.Value)
	}

	// check battery for 200
	mf = find("blutrv_battery")
	if mf == nil {
		t.Fatalf("blutrv_battery missing")
	}
	found := false
	for _, m := range mf.Metric {
		for _, l := range m.Label {
			if l.GetName() == "component" && l.GetValue() == "200" {
				if *m.Gauge.Value != 100 {
					t.Fatalf("unexpected battery value: %v", *m.Gauge.Value)
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("could not find battery metric for 200")
	}

	// check fw info
	mf = find("blutrv_info")
	if mf == nil {
		t.Fatalf("blutrv_info missing")
	}
	infoFound := false
	for _, m := range mf.Metric {
		var comp, fw string
		for _, l := range m.Label {
			if l.GetName() == "component" {
				comp = l.GetValue()
			}
			if l.GetName() == "fw_ver" {
				fw = l.GetValue()
			}
		}
		if comp == "200" && fw == "v1.2.10" && *m.Gauge.Value == 1 {
			infoFound = true
		}
	}
	if !infoFound {
		t.Fatalf("blutrv_info for 200 not found or incorrect")
	}

	// check identity metric (component -> original key)
	mf = find("blutrv_identity")
	if mf == nil {
		t.Fatalf("blutrv_identity missing")
	}
	identFound := false
	for _, m := range mf.Metric {
		var comp, orig string
		for _, l := range m.Label {
			if l.GetName() == "component" {
				comp = l.GetValue()
			}
			if l.GetName() == "orig_key" {
				orig = l.GetValue()
			}
		}
		if comp == "200" && orig == "blutrv:200" && *m.Gauge.Value == 1 {
			identFound = true
		}
	}
	if !identFound {
		t.Fatalf("blutrv_identity for 200 not found or incorrect")
	}
}
