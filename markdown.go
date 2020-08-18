package autodoc

import (
	"io"
	"net/http"
	"os"
)

// NewMarkdown returns a new markdown instance
func NewMarkdown(fname string) (*Markdown, error) {
	w, err := os.OpenFile(fname, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Markdown{Writer: w}, nil
}

// Markdown implements markdown documentation.
type Markdown struct {
	Writer io.WriteCloser
}

// Transport returns a http.RoundTripper that wraps the provided
// roundtripper.
//
// If no roundtripper is provided, the default transport is used.
func (m *Markdown) Transport(tr http.RoundTripper) *TransportMarkdownRecorder {
	return &TransportMarkdownRecorder{Writer: m.Writer, Underlying: tr}
}
