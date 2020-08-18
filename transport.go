package autodoc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// HeadersSkipper allows users of TransportMarkdownRecorder to skip
// specific headers (such as `Authorization` or `Date`).
//
// For fixed headers, just use autodoc.SkipHeaders{"Date"}.
type HeadersSkipper interface {
	SkipHeaders(header string, value []string) bool
}

// SkipHeaders implements the HeadersSkipper interface for the
// specified set of headers.  Header comparision is case insensitive.
type SkipHeaders []string

// SkipHeaders implements HeadersSkipper.
func (s SkipHeaders) SkipHeaders(header string, _ []string) bool {
	for _, hdr := range s {
		if strings.EqualFold(hdr, header) {
			return true
		}
	}
	return false
}

// TransportMarkdownRecorder wraps a http Transport automatically
// recording the output as markdown in the provided input file.
//
// This implements the http.RoundTripper interface.
type TransportMarkdownRecorder struct {
	// Writer is where the data is written to.
	Writer io.Writer

	// Underlying is the underlying transport.  If one is not
	// provided, the default http transport is used.
	Underlying http.RoundTripper

	// SkipHeaders represents an optional interface that can be
	// provided to skip some headers from being logged.
	SkipHeaders HeadersSkipper

	// RequestPremable is written to the markdown before the
	// request.
	RequestPreamble string

	// RequestPostamble is written to the markdown after the
	// request.
	RequestPostamble string

	// ResponsePreamble is written to the markdown before the
	// response.
	ResponsePreamble string

	// ResponsePostamble is written to the markdown after the
	// response.
	ResponsePostamble string
}

// WithRequestInfo updates the request preamble and postamble.
func (t *TransportMarkdownRecorder) WithRequestInfo(preamble, postamble string) *TransportMarkdownRecorder {
	t.RequestPreamble = preamble
	t.RequestPostamble = postamble
	return t
}

// WithResponseInfo updates the response preamble and postamble.
func (t *TransportMarkdownRecorder) WithResponseInfo(preamble, postamble string) *TransportMarkdownRecorder {
	t.ResponsePreamble = preamble
	t.ResponsePostamble = postamble
	return t
}

// WithSkipHeaders setsup the static list of headers to skip.
func (t *TransportMarkdownRecorder) WithSkipHeaders(skip ...string) *TransportMarkdownRecorder {
	t.SkipHeaders = SkipHeaders(skip)
	return t
}

func (t *TransportMarkdownRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.recordRequest(t.Writer, req); err != nil {
		return nil, err
	}

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if err := t.recordResponse(t.Writer, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (t *TransportMarkdownRecorder) recordRequest(f io.Writer, req *http.Request) error {
	if _, err := f.Write([]byte(t.RequestPreamble)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f, "```\n%s %s\n", req.Method, req.URL.RequestURI()); err != nil {
		return err
	}

	var buf strings.Builder
	if err := req.Header.WriteSubset(&buf, t.skipHeaders(req.Header)); err != nil {
		return err
	}

	headers := strings.ReplaceAll(buf.String(), "\r", "")
	if _, err := fmt.Fprintf(f, "%s\n", headers); err != nil {
		return err
	}

	if err := t.writeBody(f, &req.Body); err != nil {
		return err
	}

	if _, err := f.Write([]byte("\n```\n")); err != nil {
		return err
	}

	if _, err := f.Write([]byte(t.RequestPostamble)); err != nil {
		return err
	}
	return nil
}

func (t *TransportMarkdownRecorder) recordResponse(f io.Writer, res *http.Response) error {
	if _, err := f.Write([]byte(t.ResponsePreamble)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f, "```\n%s %s\n", res.Proto, res.Status); err != nil {
		return err
	}

	var buf strings.Builder
	if err := res.Header.WriteSubset(&buf, t.skipHeaders(res.Header)); err != nil {
		return err
	}

	headers := strings.ReplaceAll(buf.String(), "\r", "")
	if _, err := fmt.Fprintf(f, "%s\n", headers); err != nil {
		return err
	}

	if err := t.writeBody(f, &res.Body); err != nil {
		return err
	}

	if _, err := f.Write([]byte("\n```\n")); err != nil {
		return err
	}

	if _, err := f.Write([]byte(t.ResponsePostamble)); err != nil {
		return err
	}

	return nil
}

func (t *TransportMarkdownRecorder) skipHeaders(h http.Header) map[string]bool {
	if t.SkipHeaders == nil {
		return nil
	}

	result := map[string]bool{}
	for k, v := range h {
		if t.SkipHeaders.SkipHeaders(k, v) {
			result[k] = true
		}
	}
	return result
}

func (t *TransportMarkdownRecorder) writeBody(w io.Writer, r *io.ReadCloser) error {
	if *r == nil {
		return nil
	}

	data, err := ioutil.ReadAll(*r)
	if err != nil {
		return err
	}
	if _, err2 := w.Write(data); err2 != nil {
		return err
	}

	(*r).Close()
	*r = ioutil.NopCloser(strings.NewReader(string(data)))
	return nil
}

func (t *TransportMarkdownRecorder) transport() http.RoundTripper {
	if t.Underlying != nil {
		return t.Underlying
	}
	return http.DefaultTransport
}
