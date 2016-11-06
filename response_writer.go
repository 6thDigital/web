package web

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
)

// ResponseWriter includes net/http's ResponseWriter and adds a StatusCode() method to obtain the written status code.
// A ResponseWriter is sent to handlers on each request.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	http.CloseNotifier

	// StatusCode returns the written status code, or 0 if none has been written yet.
	StatusCode() int
	// Written returns whether the header has been written yet.
	Written() bool
	// Size returns the size in bytes of the body written so far.
	Size() int
	// Body returns a copy of the response body written so far.
	Body() ([]byte, error)
}

type appResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int

	body io.ReadWriter
}

// Don't need this yet because we get it for free:
func (w *appResponseWriter) Write(data []byte) (n int, err error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	// Store a copy of the data in the body
	// to run through validation later
	if w.body == nil {
		w.body = bytes.NewBuffer(data)
	} else {
		_, err = w.body.Write(data)
	}
	var size int
	size, err = w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

func (w *appResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *appResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *appResponseWriter) Written() bool {
	return w.statusCode != 0
}

func (w *appResponseWriter) Size() int {
	return w.size
}

func (w *appResponseWriter) Body() ([]byte, error) {
	if w.body == nil {
		return nil, io.EOF
	}
	var b []byte = make([]byte, w.Size())
	_, err := w.body.Read(b)
	w.body = bytes.NewBuffer(b)
	return b, err
}

func (w *appResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (w *appResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *appResponseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
