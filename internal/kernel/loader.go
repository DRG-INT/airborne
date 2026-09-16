package kernel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"airborne/internal/types"
)

const KernelFile = "UNICAGD_289X.json"

type Loader struct {
	kernelDir string
	kernel    *types.KernelJSON
}

func New(kernelDir string) *Loader {
	return &Loader{kernelDir: kernelDir}
}

func (l *Loader) Load() error {
	path := filepath.Join(l.kernelDir, KernelFile)
	// #nosec G304 - path is constructed from trusted kernel directory
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read kernel: %w", err)
	}

	var k types.KernelJSON
	if err := json.Unmarshal(data, &k); err != nil {
		return fmt.Errorf("parse kernel: %w", err)
	}

	l.kernel = &k
	return nil
}

func (l *Loader) Kernel() *types.KernelJSON {
	return l.kernel
}

func (l *Loader) SelectSections(names []string) (map[string]json.RawMessage, []string) {
	if l.kernel == nil {
		return nil, nil
	}

	selected := make(map[string]json.RawMessage)
	var found []string

	for _, name := range names {
		if data, ok := l.kernel.Sections[name]; ok {
			selected[name] = data
			found = append(found, name)
		}
	}

	return selected, found
}

func (l *Loader) SelectSectionsWithBound(names []string, maxBytes int) (map[string]json.RawMessage, []string, bool) {
	selected, found := l.SelectSections(names)
	if maxBytes <= 0 {
		return selected, found, false
	}

	total := 0
	var truncated []string
	for name, data := range selected {
		total += len(data)
		if total > maxBytes {
			truncated = append(truncated, name)
		}
	}

	return selected, found, len(truncated) > 0
}

var defaultTaskSections = []string{
	"governance",
	"architecture",
	"ontology",
	"operational_context",
	"constraints",
}

func (l *Loader) SelectTaskRelevant() (map[string]json.RawMessage, []string) {
	return l.SelectSections(defaultTaskSections)
}

func (l *Loader) Describe() (string, error) {
	if l.kernel == nil {
		return "", fmt.Errorf("kernel not loaded")
	}
	sectionNames := make([]string, 0, len(l.kernel.Sections))
	for name := range l.kernel.Sections {
		sectionNames = append(sectionNames, name)
	}
	return fmt.Sprintf("build_id=%s schema=%s sections=%v", l.kernel.BuildID, l.kernel.Schema, sectionNames), nil
}
