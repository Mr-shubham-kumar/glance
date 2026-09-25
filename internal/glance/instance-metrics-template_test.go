package glance

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"html/template"
)

func TestInstanceMetricsTemplateRendersLivePayload(t *testing.T) {
	raw, err := os.ReadFile("../../config/widgets/instance-metrics.yml")
	if err != nil {
		t.Fatal(err)
	}
	var config []struct {
		Template string `yaml:"template"`
	}
	if err := yaml.Unmarshal(raw, &config); err != nil || len(config) != 1 {
		t.Fatalf("invalid widget yaml: %v", err)
	}
	tpl, err := template.New("metrics").Funcs(customAPITemplateFuncs).Parse(config[0].Template)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"available":true,"cpuReady":true,"cpuPercent":23.4,"cpuHeadroomPercent":76.6,"cpuLimitCores":0.15,"memoryMiB":42,"memoryRemainingMiB":470,"memoryLimitMiB":512,"memoryPercent":8.2,"revision":"0123456789abcdef","observedAt":"2026-09-25T12:00:00Z","startedAt":"2026-09-25T11:00:00Z","uptimeMinutes":60,"pageContentLoads":7,"pageContentFailures":0,"samples":[{"x":120,"cpu":23.4,"cpuReady":true,"memory":8.2}]}`
	data := &customAPITemplateData{customAPIResponseData: &customAPIResponseData{JSON: decoratedGJSONResult{gjson.Parse(payload)}}}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"23.4%", "42 MiB", "470 MiB", "01234567", "Page loads"} {
		if !strings.Contains(buf.String(), text) {
			t.Fatalf("expected %q in metrics widget: %s", text, buf.String())
		}
	}
	if strings.Contains(buf.String(), "ZgotmplZ") {
		t.Fatal("HTML templating rejected output")
	}
}
