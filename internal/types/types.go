package types

import (
	"encoding/json"
	"time"
)

type Manifest struct {
	SchemaVersion        string     `json:"schema_version"`
	BuildID              string     `json:"build_id"`
	KeyID                string     `json:"key_id"`
	CryptoManifestDigest string     `json:"crypto_manifest_digest"`
	GeneratedAt          string     `json:"generated_at"`
	Artifacts            []Artifact `json:"artifacts"`
}

type Artifact struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type CryptoManifest struct {
	BuildID     string     `json:"build_id"`
	GeneratedAt string     `json:"generated_at"`
	Artifacts   []Artifact `json:"artifacts"`
}

type Signature struct {
	KeyID                string `json:"key_id"`
	ManifestSHA256       string `json:"manifest_sha256"`
	CryptoManifestDigest string `json:"crypto_manifest_digest"`
	Signature            string `json:"signature"`
	SignedAt             string `json:"signed_at"`
}

type VerificationResult struct {
	BuildID              string
	KeyID                string
	ManifestSHA256       string
	CryptoManifestDigest string
	ManifestVerified     bool
	Artifacts            []ArtifactResult
	SignatureValid       bool
	KeyVerified          bool
	AllPassed            bool
	VerifiedAt           time.Time
}

type ArtifactResult struct {
	Name     string
	SHA256   string
	Size     int64
	SHA256OK bool
	SizeOK   bool
}

type KernelJSON struct {
	BuildID   string                     `json:"build_id"`
	Sections  map[string]json.RawMessage `json:"sections"`
	BuildInfo BuildInfo                  `json:"build_info"`
	Schema    string                     `json:"schema"`
}

type BuildInfo struct {
	Version  string `json:"version"`
	BuildID  string `json:"build_id"`
	Compiler string `json:"compiler"`
	Target   string `json:"target"`
	Package  string `json:"package"`
}

type SectionRef struct {
	Name      string `json:"name"`
	Truncated bool   `json:"truncated"`
}

type ContextItem struct {
	Source    string `json:"source"`
	Content   string `json:"content"`
	Size      int    `json:"size"`
	Truncated bool   `json:"truncated"`
}

type ExecutionPlan struct {
	Intent              string              `json:"intent"`
	Mode                string              `json:"mode"`
	Targets             []string            `json:"targets"`
	ContextRequirements ContextRequirements `json:"context_requirements"`
	KernelSections      []SectionRef        `json:"kernel_sections"`
	Constraints         []string            `json:"constraints"`
	OutputContract      OutputContract      `json:"output_contract"`
	ModelRequirements   ModelRequirements   `json:"model_requirements"`
	Audit               AuditMetadata       `json:"audit"`
}

type ContextRequirements struct {
	MaxBytes       int64    `json:"max_bytes"`
	SkipBinary     bool     `json:"skip_binary"`
	SkipPatterns   []string `json:"skip_patterns"`
	IncludeStdin   bool     `json:"include_stdin"`
	IncludeGitDiff bool     `json:"include_git_diff"`
	MaxFiles       int      `json:"max_files"`
}

type OutputContract struct {
	Format string `json:"format"`
}

type ModelRequirements struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type AuditMetadata struct {
	RunID     string `json:"run_id"`
	StartedAt string `json:"started_at"`
}

type AuditRecord struct {
	RunID            string              `json:"run_id"`
	Mode             string              `json:"mode"`
	Intent           string              `json:"intent"`
	StartedAt        string              `json:"started_at"`
	EndedAt          string              `json:"ended_at"`
	DurationMs       int64               `json:"duration_ms"`
	Status           string              `json:"status"`
	ExitCode         int                 `json:"exit_code"`
	Verification     *VerificationResult `json:"verification"`
	ContextCount     int                 `json:"context_count"`
	ContextTruncated bool                `json:"context_truncated"`
	ContextBytes     int64               `json:"context_bytes"`
	KernelSections   []string            `json:"kernel_sections"`
	Plan             json.RawMessage     `json:"plan"`
	Error            string              `json:"error,omitempty"`
}
