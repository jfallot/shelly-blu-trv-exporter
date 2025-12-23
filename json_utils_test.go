package main

import (
	"testing"
)

func TestParseBluAttrs(t *testing.T) {
	jsonStr := `{"components":[{"key":"ble","status":{},"config":{"enable":true,"rpc":{"enable":true}}},{"key":"blugw","status":{},"config":{"sys_led_enable":true}},{"key":"blutrv:200","status":{"id":200,"target_C":17.8,"current_C":18.5,"pos":54,"connected":true,"rssi":-60,"battery":100,"packet_id":243,"last_updated_ts":1766497610,"paired":true,"rpc":true,"rsv":56,"fw_ver":"v1.2.10"},"config":{"id":200,"addr":"f8:44:77:1d:05:54","name":"TRV-200","key":null,"trv":"bthomedevice:200","temp_sensors":[],"dw_sensors":[],"override_delay":30,"meta":null},"attrs":{"flags":17,"model_id":8}},{"key":"blutrv:201","status":{"id":201,"target_C":19.5,"current_C":20.1,"pos":1,"connected":true,"rssi":-60,"battery":100,"packet_id":80,"last_updated_ts":1766497602,"paired":true,"rpc":true,"rsv":16,"fw_ver":"v1.2.10"},"config":{"id":201,"addr":"28:68:47:ef:8f:c9","name":null,"key":null,"trv":"bthomedevice:201","temp_sensors":[],"dw_sensors":[],"override_delay":30,"meta":null},"attrs":{"flags":17,"model_id":8}},{"key":"blutrv:202","status":{"id":202,"target_C":17.5,"current_C":16.9,"pos":4,"connected":true,"rssi":-72,"battery":100,"packet_id":52,"last_updated_ts":1766497618,"paired":true,"rpc":true,"rsv":1,"fw_ver":"v1.2.10"},"config":{"id":202,"addr":"28:68:47:fd:b9:d8","name":null,"key":null,"trv":"bthomedevice:202","temp_sensors":[],"dw_sensors":[],"override_delay":30,"meta":null},"attrs":{"flags":17,"model_id":8}},{"key":"bthome","status":{},"config":{}},{"key":"bthomedevice:200","status":{"id":200,"rssi":-60,"battery":100,"packet_id":243,"last_updated_ts":1766497610,"paired":true,"rpc":true,"rsv":56,"fw_ver":"v1.2.10"},"config":{"id":200,"addr":"f8:44:77:1d:05:54","name":null,"key":null,"meta":null},"attrs":{"flags":17,"model_id":8}},{"key":"bthomedevice:201","status":{"id":201,"rssi":-60,"battery":100,"packet_id":80,"last_updated_ts":1766497602,"paired":true,"rpc":true,"rsv":16,"fw_ver":"v1.2.10"},"config":{"id":201,"addr":"28:68:47:ef:8f:c9","name":null,"key":null,"meta":null},"attrs":{"flags":17,"model_id":8}}],"cfg_rev":29,"offset":0,"total":23}`

	m, err := ParseBluAttrs([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseBluAttrs failed: %v", err)
	}

	expectedKeys := []string{"blugw", "TRV-200", "201", "202"}
	if len(m) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(m))
	}

	for _, k := range expectedKeys {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing key %s in result", k)
		}
	}

	// Verify numeric attr exists and equals 17 for blutrv:200
	attrs := m["TRV-200"]
	if attrs == nil {
		t.Fatalf("attrs for TRV-200 is nil")
	}
	if f, ok := attrs["flags"].(float64); !ok || int(f) != 17 {
		t.Fatalf("unexpected flags value for TRV-200: %#v", attrs["flags"])
	}

	// Verify battery and rssi are captured from status
	if b, ok := attrs["battery"].(float64); !ok || int(b) != 100 {
		t.Fatalf("unexpected battery value for TRV-200: %#v", attrs["battery"])
	}
	if r, ok := attrs["rssi"].(float64); !ok || int(r) != -60 {
		t.Fatalf("unexpected rssi value for TRV-200: %#v", attrs["rssi"])
	}
	// original key should be recorded
	if orig, ok := attrs["__orig_key"].(string); !ok || orig != "blutrv:200" {
		t.Fatalf("unexpected orig key for TRV-200: %#v", attrs["__orig_key"])
	}
}
