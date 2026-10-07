package glance

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

var intelligenceWidgetFiles = []string{
	"intelligence-today.yml",
	"intelligence-radar.yml",
	"intelligence-build.yml",
	"intelligence-systems.yml",
}

func renderIntelligenceWidget(t *testing.T, file string, payload map[string]any) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "config", "widgets", file))
	if err != nil {
		t.Fatal(err)
	}
	var config []struct {
		Template string `yaml:"template"`
	}
	if err := yaml.Unmarshal(raw, &config); err != nil || len(config) != 1 {
		t.Fatalf("invalid widget yaml %s: %v", file, err)
	}
	tpl, err := template.New(file).Funcs(customAPITemplateFuncs).Parse(config[0].Template)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	data := &customAPITemplateData{customAPIResponseData: &customAPIResponseData{
		JSON: decoratedGJSONResult{gjson.Parse(string(body))},
	}}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("execute %s: %v", file, err)
	}
	if strings.Contains(buf.String(), "ZgotmplZ") {
		t.Fatalf("HTML templating rejected output in %s: %s", file, buf.String())
	}
	if strings.Contains(buf.String(), "collected unavailable") {
		t.Fatalf("legacy 'collected unavailable' prose rendered in %s: %s", file, buf.String())
	}
	return buf.String()
}

func validIntelligencePayload(now, generated time.Time, fresh bool) map[string]any {
	return map[string]any{
		"schema_version": 1,
		"generated_at":   generated.UTC().Format(time.RFC3339),
		"data_fresh":     fresh,
		"brief": []any{
			map[string]any{
				"label": "Investigate", "title": "Test signal",
				"reason": "New primary-source announcement in a tracked area",
				"published_at": "2026-10-06T10:00:00Z",
				"url":          "/radar", "source_url": "https://example.com/signal",
			},
			map[string]any{
				"label": "Check", "title": "Feed failed",
				"reason":       "ParseError",
				"published_at": "2026-10-07T08:00:00Z",
				"url":          "/systems", "source_url": "https://example.com/feed",
			},
		},
		"sources": []any{
			map[string]any{
				"name": "OpenAI", "status": "ok", "url": "https://example.com/rss.xml",
				"input_count": 25, "last_success_at": "2026-10-07T08:52:26Z",
			},
		},
		"items": map[string]any{
			"radar": []any{
				map[string]any{
					"title": "Test development", "url": "https://example.com/dev",
					"source": "OpenAI", "published_at": "2026-10-06T10:00:00Z",
					"reason": "New primary-source announcement in a tracked area",
				},
			},
			"experiments": []any{
				map[string]any{
					"url": "https://example.com/release", "source": "OpenCode", "title": "v2",
					"experiment_url": "/build", "experiment": "Try",
					"published_at": "2026-10-06", "reason": "release notes",
					"matched_phrase": "agent", "evidence": "agent support landed",
				},
			},
		},
		"alerts": []any{
			map[string]any{"title": "Source failed", "url": "https://example.com/feed",
				"reason": "ParseError", "kind": "source"},
		},
		"receipt": map[string]any{
			"input_count": 124, "unique_count": 124, "radar_count": 1,
			"experiment_count": 0, "excluded_count": 10,
			"failed_sources": 1, "stale_sources": 2,
		},
	}
}

func TestIntelligenceWidgetsRenderUnavailableState(t *testing.T) {
	payload := unavailableIntelligence()
	for _, file := range intelligenceWidgetFiles {
		renderIntelligenceWidget(t, file, payload)
	}
	today := renderIntelligenceWidget(t, "intelligence-today.yml", payload)
	if !strings.Contains(today, "Public snapshot unavailable") {
		t.Fatalf("today must show unavailable header: %s", today)
	}
	if !strings.Contains(today, "No supported decision signal") {
		t.Fatalf("today must show empty brief message: %s", today)
	}
	radar := renderIntelligenceWidget(t, "intelligence-radar.yml", payload)
	if !strings.Contains(radar, "Public snapshot unavailable") {
		t.Fatalf("radar must show unavailable header: %s", radar)
	}
	build := renderIntelligenceWidget(t, "intelligence-build.yml", payload)
	if !strings.Contains(build, "Public snapshot unavailable") {
		t.Fatalf("build must show unavailable header: %s", build)
	}
	systems := renderIntelligenceWidget(t, "intelligence-systems.yml", payload)
	if !strings.Contains(systems, "Collection unavailable · no snapshot verified yet") {
		t.Fatalf("systems must show unavailable header: %s", systems)
	}
	if !strings.Contains(systems, "Intelligence snapshot unavailable") {
		t.Fatalf("systems must show the delivery alert: %s", systems)
	}
}

func TestIntelligenceWidgetsRenderStaleState(t *testing.T) {
	now := time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
	payload := validIntelligencePayload(now, now.Add(-24*time.Hour), true)
	applyIntelligenceDeliveryState(payload, false, now)
	if payload["delivery_status"] != "stale" || payload["data_fresh"] != false {
		t.Fatalf("payload must be stale: %v", payload["delivery_status"])
	}
	for _, file := range intelligenceWidgetFiles {
		renderIntelligenceWidget(t, file, payload)
	}
	today := renderIntelligenceWidget(t, "intelligence-today.yml", payload)
	generated := payload["generated_at"].(string)
	if !strings.Contains(today, "Stale snapshot · collected "+generated) {
		t.Fatalf("today must show stale header: %s", today)
	}
	if strings.Contains(today, "fresh collection") {
		t.Fatalf("stale today must not claim a fresh collection: %s", today)
	}
	if !strings.Contains(today, "No supported decision signal") {
		t.Fatalf("stale today must clear the brief: %s", today)
	}
	if !strings.Contains(today, "Intelligence snapshot is stale") {
		t.Fatalf("stale today must show the delivery alert: %s", today)
	}
	radar := renderIntelligenceWidget(t, "intelligence-radar.yml", payload)
	if !strings.Contains(radar, "last known evidence") {
		t.Fatalf("stale radar must label evidence: %s", radar)
	}
	systems := renderIntelligenceWidget(t, "intelligence-systems.yml", payload)
	if !strings.Contains(systems, "Intelligence snapshot is stale") {
		t.Fatalf("stale systems must show the delivery alert first: %s", systems)
	}
	if !strings.Contains(systems, "124 inputs · 124 unique · 1 failed · 2 stale sources") {
		t.Fatalf("stale systems must keep receipt counts: %s", systems)
	}
}

func TestIntelligenceWidgetsTrimBriefWhenNotFresh(t *testing.T) {
	now := time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
	payload := validIntelligencePayload(now, now.Add(-5*time.Minute), false)
	applyIntelligenceDeliveryState(payload, true, now)
	if payload["delivery_status"] != "ok" {
		t.Fatalf("delivery must stay ok: %v", payload["delivery_status"])
	}
	today := renderIntelligenceWidget(t, "intelligence-today.yml", payload)
	if !strings.Contains(today, "no fresh source data") {
		t.Fatalf("today must label the not-fresh header: %s", today)
	}
	if !strings.Contains(today, "Feed failed") {
		t.Fatalf("health Check line must be kept: %s", today)
	}
	if strings.Contains(today, "Test signal") {
		t.Fatalf("Investigate line must be dropped without fresh data: %s", today)
	}
	if !strings.Contains(today, "https://example.com/feed") {
		t.Fatalf("kept line must link its dated source: %s", today)
	}
}

func TestIntelligenceWidgetsRenderFreshState(t *testing.T) {
	now := time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
	payload := validIntelligencePayload(now, now.Add(-5*time.Minute), true)
	applyIntelligenceDeliveryState(payload, true, now)
	if payload["delivery_status"] != "ok" {
		t.Fatalf("delivery must be ok: %v", payload["delivery_status"])
	}
	today := renderIntelligenceWidget(t, "intelligence-today.yml", payload)
	generated := payload["generated_at"].(string)
	if !strings.Contains(today, "Updated "+generated+" · fresh collection") {
		t.Fatalf("today must show fresh header: %s", today)
	}
	if !strings.Contains(today, `href="https://example.com/signal"`) {
		t.Fatalf("brief must link the dated source: %s", today)
	}
	if !strings.Contains(today, "2026-10-06T10:00:00Z") {
		t.Fatalf("brief must show published_at: %s", today)
	}
	if !strings.Contains(today, "Attention ·") {
		t.Fatalf("today must show the first alert line: %s", today)
	}
	radar := renderIntelligenceWidget(t, "intelligence-radar.yml", payload)
	if !strings.Contains(radar, `href="https://example.com/dev"`) {
		t.Fatalf("radar must render items: %s", radar)
	}
	if strings.Contains(radar, "last known evidence") {
		t.Fatalf("fresh radar must not label evidence as last known: %s", radar)
	}
	build := renderIntelligenceWidget(t, "intelligence-build.yml", payload)
	if !strings.Contains(build, "OpenCode: v2") || !strings.Contains(build, `href="/build"`) {
		t.Fatalf("build must render the experiment match: %s", build)
	}
	systems := renderIntelligenceWidget(t, "intelligence-systems.yml", payload)
	if !strings.Contains(systems, "124 inputs · 124 unique · 1 failed · 2 stale sources") {
		t.Fatalf("systems must show receipt counts: %s", systems)
	}
	if !strings.Contains(systems, "OpenAI</a> · ok · 25 entries") {
		t.Fatalf("systems must show the source receipt: %s", systems)
	}
}

func TestIntelligenceWidgetsRenderLegacySnapshotFields(t *testing.T) {
	now := time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
	payload := validIntelligencePayload(now, now.Add(-5*time.Minute), true)
	receipt := payload["receipt"].(map[string]any)
	delete(receipt, "stale_sources")
	brief := payload["brief"].([]any)
	for _, entry := range brief {
		item := entry.(map[string]any)
		delete(item, "source_url")
		delete(item, "published_at")
	}
	applyIntelligenceDeliveryState(payload, true, now)
	today := renderIntelligenceWidget(t, "intelligence-today.yml", payload)
	if !strings.Contains(today, `href="/radar"`) {
		t.Fatalf("brief must fall back to the internal url: %s", today)
	}
	if strings.Contains(today, "2026-10-06T10:00:00Z") {
		t.Fatalf("legacy snapshot has no published_at to show: %s", today)
	}
	systems := renderIntelligenceWidget(t, "intelligence-systems.yml", payload)
	if !strings.Contains(systems, "0 stale sources") {
		t.Fatalf("missing stale_sources must render as 0: %s", systems)
	}
}
