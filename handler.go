package webkitgtk

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

type uriSchemeRequestBody struct {
	stream ptr
	closed bool
}

func newUriSchemeRequestBody(stream ptr) *uriSchemeRequestBody {
	return &uriSchemeRequestBody{stream: stream}
}

// Read implements io.Reader
func (r *uriSchemeRequestBody) Read(p []byte) (int, error) {
	if r.closed {
		return 0, io.ErrClosedPipe
	}

	var content unsafe.Pointer
	var contentLen int
	if p != nil {
		content = unsafe.Pointer(&p[0])
		contentLen = len(p)
	}

	var n int
	var gErr *gError
	if !lib.g.InputStreamReadAll(r.stream, ptr(content), contentLen, &n, 0, gErr) {
		return 0, gErr.toError("stream read failed")
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

// Close implements io.Closer
func (r *uriSchemeRequestBody) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true

	var err error
	var gErr *gError
	if !lib.g.InputStreamClose(r.stream, 0, gErr) {
		err = gErr.toError("stream close failed")
	}

	lib.g.ObjectUnref(r.stream)
	r.stream = 0
	return err
}

type uriSchemeRequest struct {
	pointer ptr
	reqBody *uriSchemeRequestBody
}

type uriSchemeResponseWriter struct {
	request   *uriSchemeRequest
	header    http.Header
	code      int
	writer    io.WriteCloser
	writerErr error
	finished  bool
}

func (r *uriSchemeRequest) toResponseWriter() *uriSchemeResponseWriter {
	return &uriSchemeResponseWriter{
		request: r,
		header:  http.Header{},
		code:    math.MinInt,
	}
}

func (rw *uriSchemeResponseWriter) Close() error {
	if rw.code == math.MinInt {
		rw.WriteHeader(http.StatusNotImplemented)
	}
	if rw.finished {
		return nil
	}
	rw.finished = true
	if rw.writer != nil {
		return rw.writer.Close()
	}
	return nil
}

func (rw *uriSchemeResponseWriter) Header() http.Header {
	return rw.header
}

func (rw *uriSchemeResponseWriter) WriteHeader(code int) {
	if rw.finished || rw.code != math.MinInt {
		return
	}
	rw.code = code
	//////////////

	contentLength := int64(-1)
	if sLen := rw.Header().Get("Content-Length"); sLen != "" {
		if pLen, _ := strconv.ParseInt(sLen, 10, 64); pLen > 0 {
			contentLength = pLen
		}
	}

	rFD, w, err := rw.newPipe()
	if err != nil {
		rw.finishWithError(http.StatusInternalServerError, fmt.Errorf("unable to open pipe: %s", err))
		return
	}
	rw.writer = w

	stream := lib.g.UnixInputStreamNew(rFD, true)
	if err := rw.finishWithResponse(code, rw.Header(), stream, contentLength); err != nil {
		rw.finishWithError(http.StatusInternalServerError, fmt.Errorf("unable to finish request: %s", err))
		return
	}
}
func (rw *uriSchemeResponseWriter) Write(buf []byte) (n int, err error) {
	if rw.finished {
		return 0, fmt.Errorf("write after finish")
	}

	rw.WriteHeader(http.StatusOK)
	if rw.writerErr != nil {
		return 0, rw.writerErr
	}
	return rw.writer.Write(buf)
}

func (rw *uriSchemeResponseWriter) newPipe() (r int, f *os.File, err error) {
	var p [2]int
	e := syscall.Pipe2(p[0:], 0)
	if e != nil {
		return 0, nil, fmt.Errorf("pipe2: %s", e)
	}
	return p[0], os.NewFile(uintptr(p[1]), "|1"), nil
}
func (rw *uriSchemeResponseWriter) finishWithResponse(code int, header http.Header, stream ptr, streamLength int64) error {

	resp := lib.webkit.UriSchemeResponseNew(stream, streamLength)
	defer lib.g.ObjectUnref(resp)
	lib.webkit.UriSchemeResponseSetStatus(resp, code, "")
	lib.webkit.UriSchemeResponseSetContentType(resp, header.Get("Content-Type"))
	headers := lib.soup.MessageHeadersNew(1)
	for name, values := range header {
		for _, value := range values {
			lib.soup.MessageHeadersAppend(headers, name, value)
		}
	}
	lib.webkit.UriSchemeResponseSetHttpHeaders(resp, headers)
	lib.webkit.UriSchemeRequestFinishWithResponse(rw.request.pointer, resp)
	return nil
}

type uriSchemeResponseNopCloser struct {
	io.Writer
}

func (uriSchemeResponseNopCloser) Close() error { return nil }

func (rw *uriSchemeResponseWriter) finishWithError(code int, err error) {

	if rw.writer != nil {
		rw.writer.Close()
		rw.writer = &uriSchemeResponseNopCloser{io.Discard}
	}
	rw.writerErr = err

	msg := err.Error()
	gErr := lib.g.ErrorNewLiteral(1, msg, code, msg)
	defer lib.g.ErrorFree(gErr)
	lib.webkit.UriSchemeRequestFinishError(rw.request.pointer, gErr)
}

func (r *uriSchemeRequest) toHttpRequest() (*http.Request, error) {
	var req http.Request

	req.RequestURI = lib.webkit.UriSchemeRequestGetUri(r.pointer)
	reqUrl, err := url.ParseRequestURI(req.RequestURI)
	if err != nil {
		return nil, err
	}
	req.URL = reqUrl
	req.Method = lib.webkit.UriSchemeRequestGetHttpMethod(r.pointer)
	req.Body = http.NoBody
	reqBody := lib.webkit.UriSchemeRequestGetHttpBody(r.pointer)
	if reqBody != 0 {
		r.reqBody = newUriSchemeRequestBody(reqBody)
		req.Body = r.reqBody
	}
	return &req, nil
}

func newUriSchemeRequest(pointer ptr) *uriSchemeRequest {
	req := &uriSchemeRequest{pointer: pointer}
	lib.g.ObjectRef(req.pointer)
	return req
}

func (r *uriSchemeRequest) Close() error {
	if r.reqBody != nil {
		return r.reqBody.Close()
	}
	lib.g.ObjectUnref(r.pointer)
	r.pointer = 0
	return nil
}
