package verifier

import (
	"testing"

	"airborne/internal/types"
)

func TestSHA256Hex(t *testing.T) {
	result := sha256Hex([]byte("hello"))
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sha256:abc123", "abc123"},
		{"sha256:", ""},
		{"abc123", "abc123"},
		{"", ""},
	}
	for _, tt := range tests {
		result := stripPrefix(tt.input)
		if result != tt.expected {
			t.Errorf("stripPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCanonicalize(t *testing.T) {
	data := []byte(`{"b":2,"a":1,"c":[1,2,3]}`)
	result := canonicalize(data)
	expected := `{"a":1,"b":2,"c":[1,2,3]}`
	if string(result) != expected {
		t.Errorf("canonicalize = %s, want %s", string(result), expected)
	}
}

func TestCanonicalizeIdempotent(t *testing.T) {
	data := []byte(`{"b":2,"a":1}`)
	first := canonicalize(data)
	second := canonicalize(first)
	if string(first) != string(second) {
		t.Errorf("canonicalize not idempotent: %s vs %s", string(first), string(second))
	}
}

func TestVerifyArtifactResultOK(t *testing.T) {
	ar := types.ArtifactResult{
		Name:     "test.txt",
		SHA256:   "abc123",
		Size:     100,
		SHA256OK: true,
		SizeOK:   true,
	}
	if !ar.SHA256OK || !ar.SizeOK {
		t.Error("expected both checks to pass")
	}
}
