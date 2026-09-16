package kernel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderLoadAndSelect(t *testing.T) {
	tmp := t.TempDir()
	kernelJSON := `{
		"build_id": "TEST_001",
		"schema": "test-v1",
		"build_info": {"version": "1.0", "build_id": "TEST_001"},
		"sections": {
			"governance": {"rule": "fail_closed"},
			"architecture": {"pipeline": ["CLI", "verifier"]},
			"ontology": {"types": ["repo", "file"]},
			"other": {"data": "value"}
		}
	}`
	os.WriteFile(filepath.Join(tmp, "UNICAGD_289X.json"), []byte(kernelJSON), 0644)

	l := New(tmp)
	if err := l.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	k := l.Kernel()
	if k.BuildID != "TEST_001" {
		t.Errorf("expected build_id TEST_001, got %s", k.BuildID)
	}

	selected, found := l.SelectSections([]string{"governance", "architecture", "nonexistent"})
	if len(found) != 2 {
		t.Errorf("expected 2 found sections, got %d", len(found))
	}
	if _, ok := selected["governance"]; !ok {
		t.Error("expected governance in selected")
	}
	if _, ok := selected["nonexistent"]; ok {
		t.Error("should not find nonexistent section")
	}
}

func TestSelectTaskRelevant(t *testing.T) {
	tmp := t.TempDir()
	kernelJSON := `{
		"build_id": "TEST_001",
		"sections": {
			"governance": {},
			"architecture": {},
			"ontology": {}
		}
	}`
	os.WriteFile(filepath.Join(tmp, "UNICAGD_289X.json"), []byte(kernelJSON), 0644)

	l := New(tmp)
	if err := l.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	_, found := l.SelectTaskRelevant()
	expected := []string{"governance", "architecture", "ontology"}
	for _, e := range expected {
		ok := false
		for _, f := range found {
			if f == e {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("expected %s in task-relevant sections", e)
		}
	}
}

func TestLoaderMissingFile(t *testing.T) {
	tmp := t.TempDir()
	l := New(tmp)
	err := l.Load()
	if err == nil {
		t.Error("expected error when kernel file is missing")
	}
}
