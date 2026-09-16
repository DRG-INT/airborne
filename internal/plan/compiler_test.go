package plan

import (
	"encoding/json"
	"testing"

	"airborne/internal/types"
)

func TestCompile(t *testing.T) {
	c := New([]string{"governance", "architecture"}, 128*1024)
	p := c.Compile("test intent", ModeQuery, []string{"src/"}, OutputFormatText, "openai", "gpt-4", 3, false, 1024)

	if p.Intent != "test intent" {
		t.Errorf("expected intent 'test intent', got %s", p.Intent)
	}
	if p.Mode != ModeQuery {
		t.Errorf("expected mode %s, got %s", ModeQuery, p.Mode)
	}
	if len(p.Targets) != 1 || p.Targets[0] != "src/" {
		t.Errorf("expected targets [src/], got %v", p.Targets)
	}
	if len(p.KernelSections) != 2 {
		t.Errorf("expected 2 kernel sections, got %d", len(p.KernelSections))
	}
	if p.ModelRequirements.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", p.ModelRequirements.Model)
	}
	if len(p.Constraints) == 0 {
		t.Error("expected constraints to be set")
	}
	if p.Audit.RunID == "" {
		t.Error("expected run_id to be set")
	}
}

func TestCompileHasReadOnlyConstraints(t *testing.T) {
	c := New(nil, 128*1024)
	p := c.Compile("test", ModeQuery, nil, OutputFormatText, "openai", "gpt-4", 0, false, 0)

	foundReadOnly := false
	for _, c := range p.Constraints {
		if c == "read-only" {
			foundReadOnly = true
		}
	}
	if !foundReadOnly {
		t.Error("expected 'read-only' constraint")
	}
}

func TestPlanJSON(t *testing.T) {
	c := New([]string{"governance"}, 128*1024)
	p := c.Compile("test", ModeQuery, nil, OutputFormatText, "openai", "gpt-4", 0, false, 0)

	data, err := c.PlanJSON(p)
	if err != nil {
		t.Fatalf("PlanJSON failed: %v", err)
	}

	var parsed types.ExecutionPlan
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if parsed.Intent != "test" {
		t.Errorf("expected intent 'test', got %s", parsed.Intent)
	}
}
