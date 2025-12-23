package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Component represents a single component entry in the RPC JSON response.
type Component struct {
	Key    string                 `json:"key"`
	Status map[string]interface{} `json:"status"`
	Config map[string]interface{} `json:"config"`
	Attrs  map[string]interface{} `json:"attrs"`
}

// ComponentsResponse models the subset of fields we care about.
type ComponentsResponse struct {
	Components []Component `json:"components"`
}

// ParseBluAttrs parses the JSON response and returns a map from component key
// to a map containing both `attrs` and `status` fields (status fields overwrite
// attrs on key collisions) for all components whose key contains the substring
// "blu" (case-insensitive). If a matching component has no `attrs` or
// `status` field, an empty map is returned for that key.
func ParseBluAttrs(jsonBytes []byte) (map[string]map[string]interface{}, error) {
	var resp ComponentsResponse
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		return nil, err
	}

	out := make(map[string]map[string]interface{})
	for _, c := range resp.Components {
		if strings.Contains(strings.ToLower(c.Key), "blutrv:") {

			m := make(map[string]interface{})
			id := strings.TrimPrefix(c.Key, "blutrv:")
			m["id"] = id

			if c.Status != nil {
				for k, v := range c.Status {
					m[k] = v
				}
			}
			if c.Config != nil {
				for k, v := range c.Config {
					m[k] = v
				}
			}
			if c.Attrs != nil {
				for k, v := range c.Attrs {
					m[k] = v
				}
			}

			// prefer explicit name in config if present and non-empty (normalize)
			outKey := id
			if c.Config != nil {
				if rawName, ok := c.Config["name"]; ok && rawName != nil {
					if s, ok := rawName.(string); ok {
						s = strings.TrimSpace(s)
						if s != "" {
							outKey = s
						}
					}
				}
			}

			// ensure unique keys: if outKey already used, append suffix (" (1)", " (2)", ...)
			origOutKey := outKey
			if _, exists := out[outKey]; exists {
				i := 1
				for {
					candidate := fmt.Sprintf("%s (%d)", origOutKey, i)
					if _, ok := out[candidate]; !ok {
						outKey = candidate
						break
					}
					i++
				}
			}

			// record original key for traceability
			m["__orig_key"] = c.Key

			out[outKey] = m
		}
	}

	return out, nil
}
