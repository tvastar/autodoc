# autodoc

Auto generate documentation for golang protocol flows etc.

## Generating http protocol flows into markdown

Autodoc supports recording a http request/response into a markdown file.

```golang
        // import "github.com/tvastar/autodoc"

	rec := &autodoc.TransportMarkdownRecorder{
		MarkdownFileName:  "protocol.md",
		SkipHeaders:       autodoc.SkipHeaders{"Date"},
		RequestPreamble:   "## Request\n",
		RequestPostamble:  "\nthat was the request\n",
		ResponsePreamble:  "## Response\n",
		ResponsePostamble: "\nthat was the response\n",
	}
	client := &http.Client{Transport: rec}

	... now use client as a regular http client ...
```
