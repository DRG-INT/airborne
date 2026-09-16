package types

import (
	"encoding/json"
	"testing"
)

func TestManifestUnmarshal(t *testing.T) {
	data := `{
		"schema_version": "airborne-manifest-v1",
		"build_id": "UNICAGD_289X",
		"key_id": "key-v1",
		"crypto_manifest_digest": "sha256:abc123",
		"generated_at": "2026-01-01T00:00:00Z",
		"artifacts": [
			{"name": "test.json", "sha256": "def456", "size": 42}
		]
	}`

	var m Manifest
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if m.BuildID != "UNICAGD_289X" {
		t.Errorf("expected build_id UNICAGD_289X, got %s", m.BuildID)
	}
	if m.KeyID != "key-v1" {
		t.Errorf("expected key_id key-v1, got %s", m.KeyID)
	}
	if len(m.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(m.Artifacts))
	}
	if m.Artifacts[0].Name != "test.json" {
		t.Errorf("expected artifact name test.json, got %s", m.Artifacts[0].Name)
	}
	if m.Artifacts[0].Size != 42 {
		t.Errorf("expected size 42, got %d", m.Artifacts[0].Size)
	}
}

func TestExecutionPlanJSON(t *testing.T) {
	plan := ExecutionPlan{
		Intent:         "test query",
		Mode:           "query",
		Targets:        []string{"src/"},
		OutputContract: OutputContract{Format: "text"},
	}
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var parsed ExecutionPlan
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if parsed.Intent != "test query" {
		t.Errorf("expected intent 'test query', got %s", parsed.Intent)
	}
}
