// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flamego/flamego"
)

func TestGzip(t *testing.T) {
	before := false

	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(Gzip(Options{-10}))
	f.Use(func(r http.ResponseWriter) {
		r.(flamego.ResponseWriter).Before(func(rw flamego.ResponseWriter) {
			before = true
		})
	})
	f.Get("/", func() string { return "hello wolrd!" })

	// Not yet gzip.
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	f.ServeHTTP(resp, req)

	_, ok := resp.Result().Header[headerContentEncoding]
	assert.False(t, ok)

	ce := resp.Header().Get(headerContentEncoding)
	assert.False(t, strings.EqualFold(ce, "gzip"))

	// Gzip now.
	resp = httptest.NewRecorder()
	req.Header.Set(headerAcceptEncoding, "gzip")
	f.ServeHTTP(resp, req)

	_, ok = resp.Result().Header[headerContentEncoding]
	assert.True(t, ok)

	ce = resp.Header().Get(headerContentEncoding)
	assert.True(t, strings.EqualFold(ce, "gzip"))

	assert.True(t, before)
}

type hijackableResponse struct {
	Hijacked bool
	header   http.Header
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{header: make(http.Header)}
}

func (h *hijackableResponse) Header() http.Header           { return h.header }
func (h *hijackableResponse) Write(buf []byte) (int, error) { return 0, nil }
func (h *hijackableResponse) WriteHeader(code int)          {}
func (h *hijackableResponse) Flush()                        {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func TestResponseWriterHijack(t *testing.T) {
	hijackable := newHijackableResponse()

	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(Gzip())
	f.Use(func(rw http.ResponseWriter) {
		hj, ok := rw.(http.Hijacker)
		assert.True(t, ok)

		_, _, err := hj.Hijack()
		assert.Nil(t, err)
	})

	r, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	r.Header.Set(headerAcceptEncoding, "gzip")
	f.ServeHTTP(hijackable, r)

	assert.True(t, hijackable.Hijacked)
}
