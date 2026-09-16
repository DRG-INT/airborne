package cx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsTextFile(t *testing.T) {
	tmp := t.TempDir()

	textFile := filepath.Join(tmp, "test.txt")
	os.WriteFile(textFile, []byte("hello world\nthis is text"), 0644)
	if !isTextFile(textFile) {
		t.Error("expected test.txt to be text file")
	}

	binFile := filepath.Join(tmp, "test.bin")
	os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE}, 0644)
	if isTextFile(binFile) {
		t.Error("expected test.bin to NOT be text file")
	}
}

func TestBuilderNewFile(t *testing.T) {
	tmp := t.TempDir()
	srcFile := filepath.Join(tmp, "src.go")
	os.WriteFile(srcFile, []byte("package main\nfunc main() {}\n"), 0644)

	b := New(DefaultMaxBytes, DefaultMaxFiles, nil, false, false)
	items, truncated := b.Build([]string{srcFile})

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Source != srcFile {
		t.Errorf("expected source %s, got %s", srcFile, items[0].Source)
	}
	if truncated {
		t.Error("did not expect truncation")
	}
}

func TestBuilderSkipBinary(t *testing.T) {
	tmp := t.TempDir()
	binFile := filepath.Join(tmp, "binary.dat")
	os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}, 0644)

	b := New(DefaultMaxBytes, DefaultMaxFiles, nil, false, false)
	items, _ := b.Build([]string{binFile})

	if len(items) != 0 {
		t.Errorf("expected 0 items for binary file, got %d", len(items))
	}
}

func TestBuilderSkipDirs(t *testing.T) {
	tmp := t.TempDir()
	subDir := filepath.Join(tmp, "node_modules")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "hidden.js"), []byte("console.log('hidden')"), 0644)
	os.WriteFile(filepath.Join(tmp, "visible.go"), []byte("package main"), 0644)

	b := New(128*1024, 50, nil, false, false)
	items, _ := b.Build([]string{tmp})

	foundVisible := false
	for _, item := range items {
		if filepath.Base(item.Source) == "visible.go" {
			foundVisible = true
		}
	}
	if !foundVisible {
		t.Error("expected to find visible.go")
	}
	for _, item := range items {
		if strings.Contains(item.Source, "node_modules") {
			t.Error("should have skipped node_modules")
		}
	}
}
