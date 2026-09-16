package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"airborne/internal/types"
)

const AuditFileName = "last-run.json"

type Store struct {
	stateDir string
}

func New(stateDir string) *Store {
	return &Store{stateDir: stateDir}
}

func (s *Store) Save(record *types.AuditRecord) error {
	dir := s.stateDir
	if dir == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal audit record: %w", err)
	}

	path := filepath.Join(dir, AuditFileName)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write audit: %w", err)
	}
	return os.Rename(tmpPath, path)
}

func (s *Store) Load() (*types.AuditRecord, error) {
	path := filepath.Join(s.stateDir, AuditFileName)
	// #nosec G304 - path is constructed from trusted state directory
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read audit: %w", err)
	}

	var record types.AuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("parse audit: %w", err)
	}
	return &record, nil
}
