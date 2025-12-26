package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// BluCollector exposes metrics for blu components parsed from the RPC response.
type BluCollector struct {
	config         Config
	componentsDesc *prometheus.Desc
	targetDesc     *prometheus.Desc
	currentDesc    *prometheus.Desc
	batteryDesc    *prometheus.Desc
	rssiDesc       *prometheus.Desc
	flagsDesc      *prometheus.Desc
	infoDesc       *prometheus.Desc // info metric with fw_ver as label
	identityDesc   *prometheus.Desc // mapping between chosen component key and original key
	mutex          sync.Mutex
}

func NewBluCollector(config Config) *BluCollector {
	return &BluCollector{
		config:         config,
		componentsDesc: prometheus.NewDesc("blutrv_components_total", "Number of blu components found in last scrape", nil, nil),
		targetDesc:     prometheus.NewDesc("blutrv_target_c", "Target temperature C for blu components", []string{"component"}, nil),
		currentDesc:    prometheus.NewDesc("blutrv_current_c", "Current temperature C for blu components", []string{"component"}, nil),
		batteryDesc:    prometheus.NewDesc("blutrv_battery", "Battery percentage for blu components", []string{"component"}, nil),
		rssiDesc:       prometheus.NewDesc("blutrv_rssi", "RSSI value for blu components", []string{"component"}, nil),
		flagsDesc:      prometheus.NewDesc("blutrv_flags", "Flags attribute for blu components", []string{"component"}, nil),
		infoDesc:       prometheus.NewDesc("blutrv_info", "Info labels for blu components", []string{"component", "fw_ver"}, nil),
		identityDesc:   prometheus.NewDesc("blutrv_identity", "Identity mapping for blu components (component -> original key)", []string{"component", "orig_key"}, nil),
	}
}

func (c *BluCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.componentsDesc
	ch <- c.targetDesc
	ch <- c.currentDesc
	ch <- c.batteryDesc
	ch <- c.rssiDesc
	ch <- c.flagsDesc
	ch <- c.infoDesc
}

func (c *BluCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	bluMap, err := fetchComponentsMap(c.config)
	if err != nil {
		// still attempt to export collected values if any
		log.Fatalf("Error collecting stats: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.componentsDesc, prometheus.GaugeValue, float64(len(bluMap)))

	for key, attrs := range bluMap {
		// flags
		if v, ok := attrs["flags"]; ok {
			if f, ok := v.(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.flagsDesc, prometheus.GaugeValue, f, key)
			}
		}
		// battery
		if v, ok := attrs["battery"]; ok {
			if f, ok := v.(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.batteryDesc, prometheus.GaugeValue, f, key)
			}
		}
		// rssi
		if v, ok := attrs["rssi"]; ok {
			if f, ok := v.(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.rssiDesc, prometheus.GaugeValue, f, key)
			}
		}
		// target/current temperatures
		if v, ok := attrs["target_C"]; ok {
			if f, ok := v.(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.targetDesc, prometheus.GaugeValue, f, key)
			}
		}
		if v, ok := attrs["current_C"]; ok {
			if f, ok := v.(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.currentDesc, prometheus.GaugeValue, f, key)
			}
		}
		// fw_ver -> info metric with label
		if v, ok := attrs["fw_ver"]; ok {
			if s, ok := v.(string); ok {
				ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1, key, s)
			}
		}
		// emit identity metric with original key if present
		if orig, ok := attrs["__orig_key"]; ok {
			if s, ok := orig.(string); ok {
				ch <- prometheus.MustNewConstMetric(c.identityDesc, prometheus.GaugeValue, 1, key, s)
			}
		}
	}

	duration := time.Since(start).Seconds()
	// reuse lastScrapeDuration metric name from port collector? not accessible here; skip timing metric
	_ = duration
}

// fetchComponentsMap calls the Shelly RPC endpoint and returns the parsed blu map.
func fetchComponentsMap(config Config) (map[string]map[string]interface{}, error) {
	url := "http://" + config.Address + "/rpc/Shelly.GetComponents"
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if config.Username != "" || config.Password != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	m, err := ParseBluAttrs(body)
	if err != nil {
		return nil, fmt.Errorf("parsing blu json: %w", err)
	}

	return m, nil
}
