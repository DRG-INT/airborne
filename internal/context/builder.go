package context

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"airborne/internal/types"
)

const DefaultMaxBytes = 128 * 1024
const DefaultMaxFiles = 50

var skipDirs = map[string]bool{
	"vendor":              true,
	"node_modules":        true,
	"target":              true,
	".git":                true,
	".cargo":              true,
	"__pycache__":         true,
	".ruff_cache":         true,
	"dist":                true,
	".next":               true,
	".nuxt":               true,
	"build":               true,
	".gradle":             true,
	".m2":                 true,
	".terraform":          true,
	".venv":               true,
	".tox":                true,
	"cmake-build-debug":   true,
	"cmake-build-release": true,
}

var textExtensions = map[string]bool{
	".go": true, ".py": true, ".rs": true, ".ts": true, ".js": true,
	".c": true, ".h": true, ".cpp": true, ".cc": true, ".hpp": true,
	".cs": true, ".java": true, ".rb": true, ".php": true, ".swift": true,
	".kt": true, ".scala": true, ".lua": true, ".sh": true, ".bash": true,
	".md": true, ".txt": true, ".json": true, ".yaml": true, ".yml": true,
	".toml": true, ".xml": true, ".csv": true, ".proto": true, ".sql": true,
	".css": true, ".html": true, ".svelte": true, ".vue": true,
	".mod": true, ".sum": true, ".lock": true, ".env": true,
	".cmake": true, ".mk": true, ".dockerfile": true,
	".tf": true, ".hcl": true,
}

type Builder struct {
	maxBytes       int64
	maxFiles       int
	skipBinary     bool
	skipPatterns   []string
	collectStdin   bool
	collectGitDiff bool
	items          []types.ContextItem
	totalBytes     int64
	truncated      bool
}

func New(maxBytes int64, maxFiles int, skipPatterns []string, includeStdin, includeGitDiff bool) *Builder {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if maxFiles <= 0 {
		maxFiles = DefaultMaxFiles
	}
	return &Builder{
		maxBytes:       maxBytes,
		maxFiles:       maxFiles,
		skipBinary:     true,
		skipPatterns:   skipPatterns,
		collectStdin:   includeStdin,
		collectGitDiff: includeGitDiff,
	}
}

func (b *Builder) Build(targets []string) ([]types.ContextItem, bool) {
	if b.collectStdin {
		b.addStdin()
	}
	if b.collectGitDiff {
		b.addGitDiff()
	}
	for _, target := range targets {
		b.addPath(target)
	}
	return b.items, b.truncated
}

func (b *Builder) addStdin() {
	if stdinFileInfo, err := os.Stdin.Stat(); err == nil {
		if (stdinFileInfo.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(io.LimitReader(os.Stdin, b.maxBytes))
			if err != nil {
				return
			}
			b.addItem("stdin", data, len(data) > 0)
		}
	}
}

func (b *Builder) addGitDiff() {
	workDir, err := os.Getwd()
	if err != nil {
		return
	}
	cmd := exec.Command("git", "diff")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return
	}
	b.addItem("git_diff", output, len(output) > 0)
}

func (b *Builder) addPath(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		b.walkDir(path)
	} else {
		b.addFile(path, info)
	}
}

func (b *Builder) walkDir(root string) {
	if len(b.items) >= b.maxFiles {
		b.truncated = true
		return
	}

	relTo, err := filepath.Abs(root)
	if err != nil {
		return
	}

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if path != root && skipDirs[strings.ToLower(base)] {
				return filepath.SkipDir
			}
			return nil
		}

		if len(b.items) >= b.maxFiles {
			b.truncated = true
			return filepath.SkipAll
		}

		rel, err := filepath.Rel(relTo, path)
		if err != nil {
			rel = path
		}

		for _, pattern := range b.skipPatterns {
			if strings.Contains(rel, pattern) {
				return nil
			}
		}

		b.addFile(path, info)
		return nil
	})
}

func (b *Builder) addFile(path string, info os.FileInfo) {
	if b.totalBytes >= b.maxBytes {
		b.truncated = true
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	if b.skipBinary && !textExtensions[ext] && !isTextFile(path) {
		return
	}

	remaining := b.maxBytes - b.totalBytes
	if remaining <= 0 {
		b.truncated = true
		return
	}

	data, err := readFileBounded(path, remaining)
	if err != nil {
		return
	}
	if len(data) == 0 {
		return
	}

	truncated := int64(len(data)) >= remaining
	b.addItem(path, data, truncated)
}

func (b *Builder) addItem(source string, data []byte, truncated bool) {
	item := types.ContextItem{
		Source:    source,
		Content:   string(data),
		Size:      len(data),
		Truncated: truncated,
	}
	b.items = append(b.items, item)
	b.totalBytes += int64(len(data))
	if truncated {
		b.truncated = true
	}
}

func readFileBounded(path string, limit int64) ([]byte, error) {
	// #nosec G304 - path comes from file walking over trusted project tree
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if limit > 0 {
		return io.ReadAll(io.LimitReader(bufio.NewReader(f), limit))
	}
	return io.ReadAll(bufio.NewReader(f))
}

func isTextFile(path string) bool {
	// #nosec G304 - path comes from file walking over trusted project tree
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return true
	}

	textChars := 0
	for _, b := range buf[:n] {
		if b == 0 || b > 127 {
			continue
		}
		if b >= 32 || b == '\n' || b == '\r' || b == '\t' {
			textChars++
		}
	}
	return float64(textChars)/float64(n) > 0.80
}

func FormatContext(items []types.ContextItem) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("--- %s ---\n", item.Source))
		sb.WriteString(item.Content)
		if !strings.HasSuffix(item.Content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
