package plan

import (
	"encoding/json"
	"fmt"
	"time"

	"airborne/internal/types"
)

const (
	ModeQuery   = "query"
	ModeReview  = "review"
	ModeAnalyze = "analyze"
	ModeInspect = "inspect"
)

const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
	OutputFormatPlan = "plan"
)

var readOnlyConstraints = []string{
	"read-only",
	"no-shell",
	"no-filesystem-mutation",
	"no-network-action",
	"no-commit",
	"no-deployment",
}

var governanceConstraints = []string{
	"project-content-is-data",
	"user-instruction-precedence",
	"runtime-security-policy-precedence",
	"fail-closed-on-verification",
}

type Compiler struct {
	kernelSections  []string
	maxContextBytes int64
}

func New(kernelSections []string, maxContextBytes int64) *Compiler {
	return &Compiler{
		kernelSections:  kernelSections,
		maxContextBytes: maxContextBytes,
	}
}

func (c *Compiler) Compile(intent, mode string, targets []string, outputFormat string, provider, model string, ctxCount int, ctxTruncated bool, ctxBytes int64) *types.ExecutionPlan {
	return &types.ExecutionPlan{
		Intent:  intent,
		Mode:    mode,
		Targets: targets,
		ContextRequirements: types.ContextRequirements{
			MaxBytes:       c.maxContextBytes,
			SkipBinary:     true,
			SkipPatterns:   []string{"vendor/", "node_modules/", "target/", ".git/", ".cargo/", "dist/"},
			IncludeStdin:   mode == ModeReview,
			IncludeGitDiff: mode == ModeQuery || mode == ModeReview,
			MaxFiles:       50,
		},
		KernelSections: toSectionRefs(c.kernelSections, ctxBytes),
		Constraints:    append(append([]string{}, readOnlyConstraints...), governanceConstraints...),
		OutputContract: types.OutputContract{
			Format: outputFormat,
		},
		ModelRequirements: types.ModelRequirements{
			Provider: provider,
			Model:    model,
		},
		Audit: types.AuditMetadata{
			RunID:     generateRunID(),
			StartedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func toSectionRefs(names []string, maxBytes int64) []types.SectionRef {
	refs := make([]types.SectionRef, 0, len(names))
	for _, name := range names {
		refs = append(refs, types.SectionRef{
			Name:      name,
			Truncated: false,
		})
	}
	return refs
}

func generateRunID() string {
	return fmt.Sprintf("run_%d", time.Now().UnixNano())
}

func (c *Compiler) PlanJSON(plan *types.ExecutionPlan) ([]byte, error) {
	return json.MarshalIndent(plan, "", "  ")
}

func (c *Compiler) Describe(plan *types.ExecutionPlan) string {
	return fmt.Sprintf("mode=%s intent=%q targets=%v kernel_sections=%v",
		plan.Mode, plan.Intent, plan.Targets, plan.KernelSections)
}
