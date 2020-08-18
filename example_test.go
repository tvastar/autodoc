package autodoc_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/tvastar/autodoc"
)

func ExampleTransportMarkdownRecorder() {
	f, err := ioutil.TempFile("", "autodoc")
	if err != nil {
		fmt.Println("Got error", err)
		return
	}
	defer os.Remove(f.Name())
	fname := f.Name()
	f.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		if _, err := w.Write([]byte(`{"foo": "bar"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	md, err := autodoc.NewMarkdown(fname)
	if err != nil {
		fmt.Println("Got error", err)
		return
	}

	transport := md.Transport(nil).
		WithSkipHeaders("Date").
		WithRequestInfo("## Request\n", "\nthat was the request\n").
		WithResponseInfo("## Response\n", "\nthat was the response\n")
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", server.URL+"/some_endpoint", strings.NewReader("some body"))
	if err != nil {
		fmt.Println("Got error", err)
		return
	}
	req.Header.Add("Content-Type", "application/json;charset=utf8")
	if _, err2 := client.Do(req); err2 != nil {
		fmt.Println("Got error", err2)
		return
	}

	if err2 := md.Writer.Close(); err2 != nil {
		fmt.Println("Got error", err2)
		return
	}

	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println("Got error", err)
		return
	}
	fmt.Println(strings.ReplaceAll(string(data), "\r", "\n"))

	// Output:
	// ## Request
	// GET /some_endpoint
	// Content-Type: application/json;charset=utf8
	//
	// some body
	// that was the request
	// ## Response
	// HTTP/1.1 200 OK
	// Content-Length: 14
	// Content-Type: application/json
	//
	// {"foo": "bar"}
	// that was the response

}
