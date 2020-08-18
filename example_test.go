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

func ExampleMarkdown_writeFieldType() {
	fname, cleanup, err := tempFile()
	if err != nil {
		return
	}
	defer cleanup()

	md, err := autodoc.NewMarkdown(fname)
	if err != nil {
		fmt.Println("Got error", err)
		return
	}

	err = md.WriteStructTable(&struct {
		Hello  string
		World  int
		Nested *struct {
			Hello uint
			World string
		}
	}{})
	if err != nil {
		fmt.Println("Got error", err)
		return
	}

	if err2 := md.Writer.Close(); err2 != nil {
		fmt.Println("Got error", err2)
		return
	}

	dumpFile(fname)

	// Output:
	// | Field | Type | Description |
	// | ----- | ---- | ----------- |
	// | Hello | string  |  |
	// | World | number  |  |
	// | Nested.Hello | number  |  |
	// | Nested.World | string  |  |
}

func ExampleMarkdown_transport() {
	fname, cleanup, err := tempFile()
	if err != nil {
		return
	}
	defer cleanup()

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

	dumpFile(fname)

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

func tempFile() (string, func(), error) {
	f, err := ioutil.TempFile("", "autodoc")
	if err != nil {
		fmt.Println("Got error", err)
		return "", func() {}, err
	}
	name := f.Name()
	f.Close()
	return name, func() { os.Remove(name) }, nil
}

func dumpFile(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println("Got error", err)
		return
	}

	fmt.Println(string(data))
}
