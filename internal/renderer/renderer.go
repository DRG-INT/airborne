package renderer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"airborne/internal/provider"
)

type Renderer struct {
	output io.Writer
}

func New(w io.Writer) *Renderer {
	r := &Renderer{output: w}
	if r.output == nil {
		r.output = os.Stdout
	}
	return r
}

func (r *Renderer) Stream(ch <-chan provider.StreamEvent) {
	for event := range ch {
		if event.Error != nil {
			fmt.Fprintf(os.Stderr, "\n[error] %s\n", event.Error.Error())
			continue
		}
		if event.Delta != "" {
			fmt.Fprint(r.output, event.Delta)
		}
		if event.Done {
			if event.Usage != nil && event.Usage.TotalTokens > 0 {
				fmt.Fprintf(os.Stderr, "\n[usage] prompt=%d completion=%d total=%d\n",
					event.Usage.PromptTokens, event.Usage.CompletionTokens, event.Usage.TotalTokens)
			}
		}
	}
	fmt.Fprintln(r.output)
}

func (r *Renderer) JSON(data interface{}) error {
	enc := json.NewEncoder(r.output)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (r *Renderer) Text(msg string) {
	fmt.Fprint(r.output, msg)
}

func (r *Renderer) Textf(format string, args ...interface{}) {
	fmt.Fprintf(r.output, format, args...)
}

func (r *Renderer) Err(msg string) {
	fmt.Fprint(os.Stderr, msg)
	if !hasNewline(msg) {
		fmt.Fprintln(os.Stderr)
	}
}

func hasNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}
